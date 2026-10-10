<template>
  <v-app-bar app elevation="4" class="px-2">
    <v-icon color="primary" large class="mr-3">mdi-storefront</v-icon>
    <v-toolbar-title class="font-weight-bold text-h6 tracking-wide">
      Codzee Store &amp; Inventory Manager
    </v-toolbar-title>

    <v-spacer></v-spacer>

    <!-- API Status Badge -->
    <v-chip
      :color="isApiOnline ? 'success' : 'error'"
      outlined
      small
      class="mr-4 font-weight-medium"
    >
      <v-icon left small>
        {{ isApiOnline ? 'mdi-check-circle' : 'mdi-alert-circle' }}
      </v-icon>
      API: {{ isApiOnline ? 'ONLINE' : 'OFFLINE' }}
    </v-chip>

    <!-- Theme Switch Toggle -->
    <v-tooltip bottom>
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          icon
          v-bind="attrs"
          v-on="on"
          @click="$emit('toggle-theme')"
          color="amber lighten-1"
        >
          <v-icon>{{ $vuetify.theme.dark ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
        </v-btn>
      </template>
      <span>Toggle {{ $vuetify.theme.dark ? 'Light' : 'Dark' }} Theme</span>
    </v-tooltip>
  </v-app-bar>
</template>

<script>
export default {
  name: 'AppNavbar',
  props: {
    isApiOnline: {
      type: Boolean,
      default: false
    }
  }
};
</script>
