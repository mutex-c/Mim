<template>
  <div class="w-screen h-screen flex overflow-hidden bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Leftmost Navigation Dock -->
    <NavigationDock @open-settings="showSettings = true" />

    <!-- Main Dynamic Area depending on active tab -->
    <template v-if="chatStore.activeTab === 'chats'">
      <!-- Middle Conversations List -->
      <ConversationList />

      <!-- Right Chat Stage -->
      <ChatStage />
    </template>

    <template v-else-if="chatStore.activeTab === 'contacts'">
      <ContactStage />
    </template>

    <template v-else-if="chatStore.activeTab === 'calls'">
      <CallStage />
    </template>

    <!-- Global Settings Modal -->
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import NavigationDock from './components/NavigationDock.vue';
import ConversationList from './components/ConversationList.vue';
import ChatStage from './components/ChatStage.vue';
import ContactStage from './components/ContactStage.vue';
import CallStage from './components/CallStage.vue';
import SettingsModal from './components/SettingsModal.vue';
import { useThemeStore } from './stores/theme';
import { useChatStore } from './stores/chat';

// Initialize Theme & Chat Store
useThemeStore();
const chatStore = useChatStore();
const showSettings = ref(false);
</script>
