<template>
  <div class="flex-1 h-full flex overflow-hidden">
    <!-- Middle Conversations List -->
    <ConversationList />

    <!-- Right Chat Stage -->
    <ChatStage />
  </div>
</template>

<script setup lang="ts">
import { watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import ConversationList from '../components/ConversationList.vue';
import ChatStage from '../components/ChatStage.vue';
import { useChatStore } from '../stores/chat';

const route = useRoute();
const router = useRouter();
const chatStore = useChatStore();

const syncConvFromRoute = () => {
  const convId = route.params.id as string;
  if (convId) {
    if (chatStore.activeConvId !== convId) {
      chatStore.selectConversation(convId);
    }
  } else if (chatStore.activeConvId) {
    // If on /chats with no id, redirect to current active id
    router.replace(`/chats/${chatStore.activeConvId}`);
  }
};

onMounted(() => {
  syncConvFromRoute();
});

watch(
  () => route.params.id,
  (newId) => {
    if (newId && typeof newId === 'string' && chatStore.activeConvId !== newId) {
      chatStore.selectConversation(newId);
    }
  }
);

watch(
  () => chatStore.activeConvId,
  (newId) => {
    if (newId && route.params.id !== newId && route.path.startsWith('/chats')) {
      router.push(`/chats/${newId}`);
    }
  }
);
</script>
