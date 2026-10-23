<template>
  <div class="w-screen h-screen flex overflow-hidden bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Leftmost Navigation Dock (Always persistent) -->
    <NavigationDock @open-settings="showSettings = true" />

    <!-- Dynamic Router View for Pages -->
    <router-view class="flex-1 h-full overflow-hidden" />

    <!-- Global Settings Modal -->
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import NavigationDock from './components/NavigationDock.vue';
import SettingsModal from './components/SettingsModal.vue';
import { useThemeStore } from './stores/theme';

// Initialize Theme
useThemeStore();
const showSettings = ref(false);
</script>
