<template>
  <div>
    <v-alert
      type="info"
      text
      color="hsl(348deg, 86%, 61%)"
      style="border-radius: 0; margin-top: -32px;"
      v-if="!premiumFeatures.terraform_backend"
    >
      Terraform/OpenTofu HTTP backend available only in <b>PRO</b> version.
      <v-btn
        class="ml-2"
        color="hsl(348deg, 86%, 61%)"
        href="https://semaphoreui.com/pro"
      >Upgrade
      </v-btn>
    </v-alert>

    <div class="px-4 py-3">
      <div v-for="alias of (aliases || [])" :key="alias.id">
        <code class="mr-2">{{ alias.url }}</code>
        <v-btn
          icon
          @click="copyToClipboard(alias.url, $t('aliasUrlCopied'))"
        >
          <v-icon>mdi-content-copy</v-icon>
        </v-btn>
        <v-btn icon @click="editAlias(alias.id)">
          <v-icon>mdi-pencil</v-icon>
        </v-btn>
        <v-btn icon @click="deleteAlias(alias.id)">
          <v-icon>mdi-delete</v-icon>
        </v-btn>
      </div>

      <v-btn
        color="primary"
        @click="addAlias()"
        :disabled="!premiumFeatures.terraform_backend"
      >
        {{ aliases == null ? $t('LoadAlias') : $t('AddAlias') }}
      </v-btn>
    </div>

    <TerraformStateView
      v-if="premiumFeatures.terraform_backend"
    />

    <v-container v-else>
      <div style="text-align: center; color: grey;">No state available.</div>
    </v-container>

  </div>
</template>
<script>

import axios from 'axios';
import EventBus from '@/event-bus';
import TerraformStateView from '@/components/TerraformStateView.vue';

export default {
  components: { TerraformStateView },
  props: {
    template: Object,
    repositories: Array,
    inventory: Array,
    environment: Array,
    premiumFeatures: Object,
  },

  data() {
    return {
      aliases: [],
      state: null,
    };
  },

  async created() {
    await this.loadAliases();
  },

  methods: {
    deleteState(id) {
      console.log(id);
    },

    async loadAliases() {
      this.aliases = (await axios({
        method: 'get',
        url: `/api/project/${this.template.project_id}/inventory/${this.template.id}/terraform/aliases`,
        responseType: 'json',
      })).data;
    },

    editAlias(alias) {
      console.log(alias);
    },

    async deleteAlias(alias) {
      await axios({
        method: 'delete',
        url: `/api/project/${this.template.project_id}/inventory/${this.template.inventory_id}/terraform/aliases/${alias}`,
      });
      await this.loadAliases();
    },

    async copyToClipboard(text) {
      try {
        await window.navigator.clipboard.writeText(text);
        EventBus.$emit('i-snackbar', {
          color: 'success',
          text: 'The command has been copied to the clipboard.',
        });
      } catch (e) {
        EventBus.$emit('i-snackbar', {
          color: 'error',
          text: `Can't copy the command: ${e.message}`,
        });
      }
    },

    async addAlias() {
      await axios({
        method: 'post',
        url: `/api/project/${this.template.project_id}/inventory/${this.template.inventory_id}/terraform/aliases`,
        responseType: 'json',
      });
      await this.loadAliases();
    },
  },
};
</script>
