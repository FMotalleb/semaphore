package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type TerraformApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	Inventory  db.Inventory
	reader     terraformReader
	Name       string
	noChanges  bool
}

type terraformReaderResult int

const (
	terraformReaderConfirmed terraformReaderResult = iota
	terraformReaderFailed
)

type terraformReader struct {
	result *terraformReaderResult
}

func (t *TerraformApp) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = getEnvironmentVars()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	return cmd
}

func (t *TerraformApp) runCmd(command string, args []string) error {
	cmd := t.makeCmd(command, args, nil)
	t.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (t *TerraformApp) GetFullPath() string {
	return path.Join(t.Repository.GetFullPath(t.Template.ID), strings.TrimPrefix(t.Template.Playbook, "/"))
}

func (t *TerraformApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger

	t.Logger.AddLogListener(func(new time.Time, msg string) {
		if strings.Contains(msg, "No changes.") {
			t.noChanges = true
		}
	})

	t.Logger.AddStatusListener(func(status task_logger.TaskStatus) {
		var result terraformReaderResult

		switch status {
		case task_logger.TaskConfirmed:
			result = terraformReaderConfirmed
			t.reader.result = &result
		case task_logger.TaskFailStatus, task_logger.TaskStoppedStatus:
			result = terraformReaderFailed
			t.reader.result = &result
		}
	})

	return logger
}

func (t *TerraformApp) init(environmentVars *[]string, params *db.TerraformTaskParams) error {

	args := []string{"init"}

	if params.Upgrade {
		args = append(args, "-upgrade")
	}

	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (t *TerraformApp) isWorkspacesSupported(environmentVars *[]string) bool {
	cmd := t.makeCmd(string(t.Name), []string{"workspace", "list"}, environmentVars)
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

func (t *TerraformApp) selectWorkspace(workspace string, environmentVars *[]string) error {
	cmd := t.makeCmd(string(t.Name), []string{"workspace", "select", "-or-create=true", workspace}, environmentVars)
	t.Logger.LogCmd(cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (t *TerraformApp) InstallRequirements(environmentVars *[]string, params interface{}) (err error) {

	p := params.(*db.TerraformTaskParams)

	err = t.init(environmentVars, p)
	if err != nil {
		return
	}

	workspace := "default"

	if t.Inventory.Inventory != "" {
		workspace = t.Inventory.Inventory
	}

	if workspace == "default" && !t.isWorkspacesSupported(environmentVars) {
		return
	}

	err = t.selectWorkspace(workspace, environmentVars)
	return
}

func (t *TerraformApp) Plan(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"plan"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}

func (t *TerraformApp) Apply(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"apply", "-auto-approve"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}

func (t *TerraformApp) Run(args LocalAppRunningArgs) error {
	err := t.Plan(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	if err != nil {
		return err
	}

	params := args.TaskParams.(*db.TerraformTaskParams)

	if t.noChanges || params.Plan {
		t.Logger.SetStatus(task_logger.TaskSuccessStatus)
		return nil
	}

	if params.AutoApprove {
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	}

	t.Logger.SetStatus(task_logger.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if t.reader.result != nil {
			break
		}
	}

	switch *t.reader.result {
	case terraformReaderFailed:
		return nil
	case terraformReaderConfirmed:
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	default:
		return fmt.Errorf("unknown plan result")
	}
}
