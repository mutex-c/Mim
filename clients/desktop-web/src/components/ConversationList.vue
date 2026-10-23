<template>
  <section class="w-80 h-full flex flex-col bg-mim-light-list dark:bg-mim-dark-list select-none border-r border-black/[0.04] dark:border-white/[0.04]">
    <!-- Header with Search -->
    <div class="p-4 pb-3 flex flex-col gap-3">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-base font-semibold tracking-tight text-mim-light-text dark:text-mim-dark-text">
            {{ langStore.t('conv.messages') }}
          </h1>
          <div class="flex items-center gap-1.5 mt-0.5">
            <span
              class="w-1.5 h-1.5 rounded-full"
              :class="chatStore.isWsConnected ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'"
            ></span>
            <span class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted font-mono">
              {{ chatStore.currentUserName }} (ID: {{ chatStore.currentUserId }})
            </span>
          </div>
        </div>
        <button class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 transition-colors" :title="langStore.t('conv.newChat')">
          <SquarePen class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Search Bar (Rule 2: Bite Feel Input) -->
      <div class="relative w-full">
        <Search class="w-4 h-4 absolute left-3 top-2.5 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none inline-icon" />
        <input
          v-model="searchKeyword"
          type="text"
          :placeholder="langStore.t('conv.searchPlaceholder')"
          class="mim-input w-full h-9 pl-9 pr-3 rounded-lg bg-black/[0.04] dark:bg-white/[0.06] text-sm text-mim-light-text dark:text-mim-dark-text placeholder-mim-light-muted dark:placeholder-mim-dark-muted border-none"
        />
      </div>
    </div>

    <!-- Conversations Scroll Area -->
    <div class="flex-1 overflow-y-auto px-2 space-y-1">
      <div
        v-for="conv in filteredConversations"
        :key="conv.id"
        @click="chatStore.selectConversation(conv.id)"
        class="w-full p-2.5 rounded-xl flex items-center gap-3 cursor-pointer transition-all duration-150 relative"
        :class="chatStore.activeConvId === conv.id
          ? 'bg-black/[0.06] dark:bg-white/[0.08]'
          : 'hover:bg-black/[0.03] dark:hover:bg-white/[0.04]'"
      >
        <!-- Avatar -->
        <div class="relative shrink-0">
          <div
            class="w-11 h-11 rounded-xl flex items-center justify-center text-sm font-semibold"
            :class="conv.type === 'group' 
              ? 'bg-indigo-600/15 text-indigo-600 dark:text-indigo-400' 
              : 'bg-blue-600/15 text-blue-600 dark:text-blue-400'"
          >
            {{ conv.avatarText }}
          </div>
          <!-- Online Indicator -->
          <span
            v-if="conv.isOnline"
            class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-emerald-500 border-2 border-mim-light-list dark:border-mim-dark-list rounded-full"
          ></span>
        </div>

        <!-- Conversation Details -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between mb-0.5">
            <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text truncate">
              {{ conv.title }}
            </span>
            <!-- Rule 6: Tabular Nums on Timestamp -->
            <span class="text-xs text-mim-light-muted dark:text-mim-dark-muted tabular-nums shrink-0 ml-2">
              {{ conv.lastTime }}
            </span>
          </div>

          <div class="flex items-center justify-between">
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted truncate pr-2">
              {{ conv.lastMessage || langStore.t('conv.noMessages') }}
            </p>
            <div class="flex items-center gap-1 shrink-0">
              <Pin v-if="conv.isPinned" class="w-3 h-3 text-mim-light-muted dark:text-mim-dark-muted inline-icon" />
              <!-- Unread Badge (Rule 6: Tabular Nums) -->
              <span
                v-if="conv.unreadCount > 0"
                class="px-1.5 py-0.5 text-[10px] font-semibold bg-blue-600 text-white rounded-full tabular-nums"
              >
                {{ conv.unreadCount }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { Search, SquarePen, Pin } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';

const chatStore = useChatStore();
const langStore = useLangStore();
const searchKeyword = ref('');

const filteredConversations = computed(() => {
  if (!searchKeyword.value.trim()) {
    return chatStore.conversations;
  }
  const kw = searchKeyword.value.toLowerCase();
  return chatStore.conversations.filter(
    (c) => c.title.toLowerCase().includes(kw) || c.lastMessage.toLowerCase().includes(kw)
  );
});
</script>
