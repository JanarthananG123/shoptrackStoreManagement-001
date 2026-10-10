<template>
  <v-app>
    <AppNavbar :is-api-online="isApiOnline" @toggle-theme="toggleTheme" />

    <v-main>
      <v-container fluid class="pa-6">
        <router-view @api-status-change="handleApiStatusChange" />
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import AppNavbar from './components/AppNavbar.vue';
import apiClient from './api/axiosConfig';

export default {
  name: 'App',
  components: {
    AppNavbar
  },
  data() {
    return {
      isApiOnline: false,
      healthCheckInterval: null
    };
  },
  mounted() {
    this.checkHealth();
    this.healthCheckInterval = setInterval(this.checkHealth, 10000);
  },
  beforeDestroy() {
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
  },
  methods: {
    toggleTheme() {
      this.$vuetify.theme.dark = !this.$vuetify.theme.dark;
    },
    handleApiStatusChange(status) {
      this.isApiOnline = status;
    },
    async checkHealth() {
      try {
        const response = await apiClient.get('/health');
        this.isApiOnline = !!(response.data && response.data.success);
      } catch (error) {
        this.isApiOnline = false;
      }
    }
  }
};
</script>
