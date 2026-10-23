<template>
  <div class="w-screen h-screen flex overflow-hidden bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Leftmost Navigation Dock -->
    <NavigationDock />

    <!-- Middle Conversations List -->
    <ConversationList />

    <!-- Right Chat Stage -->
    <ChatStage />
  </div>
</template>

<script setup lang="ts">
import NavigationDock from './components/NavigationDock.vue';
import ConversationList from './components/ConversationList.vue';
import ChatStage from './components/ChatStage.vue';
import { useThemeStore } from './stores/theme';

// Initialize Theme
useThemeStore();
</script>
