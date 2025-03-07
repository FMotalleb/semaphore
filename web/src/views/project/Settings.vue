<template xmlns:v-slot="http://www.w3.org/1999/XSL/Transform">
  <div>
    <YesNoDialog
      v-model="deleteProjectDialog"
      :title="$t('deleteProject')"
      :text="$t('askDeleteProj')"
      @yes="deleteProject()"
    />

    <v-toolbar flat >
      <v-app-bar-nav-icon @click="showDrawer()"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ $t('dashboard') }}</v-toolbar-title>
    </v-toolbar>

    <DashboardMenu
      :project-id="projectId"
      :project-type="projectType"
      :can-update-project="true"
    />

    <div class="project-settings-form">
      <div style="height: 300px;">
        <ProjectForm :item-id="projectId" ref="form" @error="onError" @save="onSave"/>
      </div>

      <div class="text-right">
        <v-btn color="primary" @click="saveProject()">{{ $t('save') }}</v-btn>
      </div>
    </div>

    <v-divider class="mb-8" />

    <div class="project-backup project-settings-button">

      <v-row align="center">
        <v-col class="shrink">

          <v-btn
            color="primary"
            @click="backupProject"
            :disabled="backupProgress"
            min-width="170"
          >{{ $t('backup') }}
          </v-btn>

          <v-progress-linear
            v-if="backupProgress"
            color="primary accent-4"
            indeterminate
            rounded
            height="36"
            style="margin-top: -36px"
          ></v-progress-linear>

        </v-col>
        <v-col class="grow">
          <div style="font-size: 14px;">
            {{ $t('downloadTheProjectBackupFile') }}
          </div>
        </v-col>
      </v-row>
    </div>
    <div class="project-delete-form project-settings-button">
      <v-row align="center">
        <v-col class="shrink">
          <v-btn
            color="error"
            min-width="170"
            @click="deleteProjectDialog = true"
          >{{ $t('deleteProject2') }}
          </v-btn>
        </v-col>
        <v-col class="grow">
          <div style="font-size: 14px; color: #ff5252">
            {{ $t('onceYouDeleteAProjectThereIsNoGoingBackPleaseBeCer') }}
          </div>
        </v-col>
      </v-row>
    </div>
  </div>
</template>
<style lang="scss">
  @import '~vuetify/src/styles/styles.sass';

  .project-settings-form {
    max-width: 400px;
    margin: 40px auto;
  }

  .project-settings-button {
    max-width: 400px;
    margin: 20px auto auto;

    @media #{map-get($display-breakpoints, 'sm-and-down')} {
      padding: 0 6px;
    }
  }
</style>
<script>
import EventBus from '@/event-bus';
import ProjectForm from '@/components/ProjectForm.vue';
import { getErrorMessage } from '@/lib/error';
import axios from 'axios';
import YesNoDialog from '@/components/YesNoDialog.vue';
import delay from '@/lib/delay';
import DashboardMenu from '@/components/DashboardMenu.vue';

export default {
  components: { DashboardMenu, YesNoDialog, ProjectForm },
  props: {
    projectId: Number,
    projectType: String,
  },

  data() {
    return {
      deleteProjectDialog: null,
      backupProgress: false,
    };
  },

  methods: {
    showDrawer() {
      EventBus.$emit('i-show-drawer');
    },

    onError(e) {
      EventBus.$emit('i-snackbar', {
        color: 'error',
        text: e.message,
      });
    },

    onSave(e) {
      EventBus.$emit('i-project', {
        action: 'edit',
        item: e.item,
      });
    },

    async saveProject() {
      await this.$refs.form.save();
    },

    async backupProject() {
      this.backupProgress = true;
      await delay(1000);

      try {
        const backup = await axios({
          method: 'get',
          url: `/api/project/${this.projectId}/backup`,
          transformResponse: (res) => res, // Necessary to not parse json
          responseType: 'json',
        });

        const a = document.createElement('a');
        const blob = new Blob([backup.data], { type: 'application/json' });
        a.download = `backup_${this.projectId}_${Date.now()}.json`;
        a.href = URL.createObjectURL(blob);
        a.click();

        await delay(1000);

        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'Project exported.',
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      } finally {
        this.backupProgress = false;
      }
    },

    async deleteProject() {
      try {
        await axios({
          method: 'delete',
          url: `/api/project/${this.projectId}`,
          responseType: 'json',
        });
        EventBus.$emit('i-project', {
          action: 'delete',
          item: {
            id: this.projectId,
          },
        });
      } catch (err) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: getErrorMessage(err),
        });
      }
    },
  },
};
</script>
