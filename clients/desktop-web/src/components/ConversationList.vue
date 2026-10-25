<template>
  <section
    :style="{ width: `${chatStore.sidebarWidth}px` }"
    class="h-full flex flex-col bg-mim-light-list dark:bg-mim-dark-list select-none shrink-0 relative border-r border-black/[0.04] dark:border-white/[0.04]"
    :class="{ 'transition-none': isDragging, 'transition-[width] duration-75 ease-out': !isDragging }"
  >
    <!-- Full-screen drag overlay preventing iframe/selection interception (规范 5) -->
    <div
      v-if="isDragging"
      class="fixed inset-0 z-50 cursor-col-resize select-none pointer-events-auto"
    ></div>

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
        <button
          @click="showNewChatModal = true"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
          :title="langStore.t('conv.newChat')"
        >
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
        @click="selectConv(conv.id)"
        @contextmenu.prevent.stop="onConvContextMenu($event, conv)"
        class="w-full p-2.5 rounded-xl flex items-center gap-3 cursor-pointer transition-all duration-150 relative"
        :class="chatStore.activeConvId === conv.id
          ? (themeStore.currentPalette === 'vodyanitsa'
              ? 'vodyanitsa-conv-active'
              : (themeStore.currentPalette === 'genshin' ? 'bg-amber-500/10 dark:bg-amber-500/15 ring-1 ring-amber-500/35 shadow-xs' : 'bg-black/[0.06] dark:bg-white/[0.08] shadow-sm ring-1 ring-black/10 dark:ring-white/10'))
          : 'hover:bg-black/[0.03] dark:hover:bg-white/[0.04]'"
      >
        <!-- Active indicator bar -->
        <span
          v-if="chatStore.activeConvId === conv.id"
          class="absolute left-1 top-3.5 bottom-3.5 w-1 rounded-full"
          :class="themeStore.currentPalette === 'vodyanitsa' ? 'bg-cyan-400 shadow-[0_0_8px_#38bdf8]' : 'bg-mim-accent'"
        ></span>

        <!-- Avatar with Status Ring -->
        <div class="relative shrink-0">
          <div
            class="w-11 h-11 rounded-xl flex items-center justify-center text-sm font-bold shadow-xs transition-all overflow-hidden"
            :class="[
              getStatusRingClass(conv.type === 'group' ? 'group' : (conv.isOnline ? 'online' : 'offline')),
              themeStore.currentPalette === 'vodyanitsa' ? 'vodyanitsa-avatar-frame' : ''
            ]"
            :style="{
              backgroundColor: getAvatarColor(conv.targetId || conv.id, conv.type === 'group').bg,
              color: getAvatarColor(conv.targetId || conv.id, conv.type === 'group').text
            }"
          >
            <img v-if="themeStore.currentPalette === 'vodyanitsa' && (conv.title === 'Sarah Jenkins' || conv.title?.includes('沃雅'))" :src="vodyanitsaAvatar" alt="沃雅妮莎" class="w-full h-full object-cover" />
            <img v-else-if="conv.avatarText && (conv.avatarText.startsWith('http') || conv.avatarText.startsWith('data:image'))" :src="conv.avatarText" class="w-full h-full object-cover" />
            <span v-else>{{ conv.avatarText }}</span>
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
              {{ formatLastMessage(conv.lastMessage) }}
            </p>
            <div class="flex items-center gap-1.5 shrink-0">
              <!-- Calling indicator badge -->
              <span
                v-if="chatStore.activeCall && (chatStore.activeCall.convId === conv.id || chatStore.activeCall.peerId === conv.targetId)"
                class="px-1.5 py-0.5 text-[10px] font-medium bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 rounded-full flex items-center gap-1 animate-pulse"
                title="通话进行中"
              >
                <Phone class="w-2.5 h-2.5 inline-icon" />
                <span>通话中</span>
              </span>
              <Pin v-if="conv.isPinned" class="w-3 h-3 text-mim-light-muted dark:text-mim-dark-muted inline-icon" />
              <!-- Unread Badge (Rule 6: Tabular Nums) -->
              <span
                v-if="conv.unreadCount > 0"
                class="px-1.5 py-0.5 text-[10px] font-semibold bg-mim-accent text-white rounded-full tabular-nums"
              >
                {{ conv.unreadCount }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- New Chat Modal -->
    <NewChatModal v-if="showNewChatModal" @close="showNewChatModal = false" />

    <!-- Draggable Resize Handle on the Right Border (规范 5) -->
    <div
      class="absolute -right-1 top-0 bottom-0 w-2.5 cursor-col-resize z-20 group flex items-center justify-center hover:bg-mim-accent/10 transition-colors select-none"
      :class="{ 'bg-mim-accent/20': isDragging }"
      @mousedown.prevent="startResize"
      @dblclick="resetToHalf"
      title="按住左右拖拽调节会话栏宽度，双击恢复半屏或默认宽度"
    >
      <div
        class="w-[2px] h-full transition-colors"
        :class="isDragging ? 'bg-mim-accent' : 'bg-transparent group-hover:bg-mim-accent/70'"
      ></div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  Search,
  SquarePen,
  Pin,
  PinOff,
  Phone,
  Mail,
  CheckCheck,
  Bell,
  BellOff,
  Eraser,
  Trash2,
  LogOut,
} from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useThemeStore } from '../stores/theme';
import { useContextMenuStore, ContextMenuItem } from '../stores/contextMenu';
import { confirmDialog, alertDialog } from '../stores/dialog';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';
import NewChatModal from './NewChatModal.vue';
import vodyanitsaAvatar from '../assets/vodyanitsa/images/vodyanitsa_avatar.png';

const router = useRouter();
const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();
const contextMenuStore = useContextMenuStore();
const searchKeyword = ref('');
const showNewChatModal = ref(false);
const isDragging = ref(false);

const onConvContextMenu = (e: MouseEvent, conv: any) => {
  const items: ContextMenuItem[] = [
    {
      id: 'pin',
      label: conv.isPinned ? '取消置顶' : '置顶会话',
      icon: conv.isPinned ? PinOff : Pin,
      action: () => chatStore.togglePinConversation(conv.id),
    },
    {
      id: 'unread',
      label: conv.unreadCount > 0 ? '标为已读' : '标为未读',
      icon: conv.unreadCount > 0 ? CheckCheck : Mail,
      action: () => {
        if (conv.unreadCount > 0) {
          chatStore.markConversationAsRead(conv.id);
        } else {
          chatStore.markConversationAsUnread(conv.id);
        }
      },
    },
    {
      id: 'mute',
      label: conv.isMuted ? '关闭免打扰' : '消息免打扰',
      icon: conv.isMuted ? Bell : BellOff,
      action: () => chatStore.toggleMuteConversation(conv.id),
    },
    {
      id: 'clear',
      label: '清空聊天记录',
      icon: Eraser,
      divided: true,
      action: () => {
        chatStore.clearConversationMessages(conv.id);
      },
    },
    {
      id: 'delete',
      label: '从列表移除',
      icon: Trash2,
      danger: true,
      action: () => {
        chatStore.deleteConversation(conv.id);
      },
    },
  ];

  if (conv.type === 'group') {
    const match = conv.id.match(/^group_(\d+)$/);
    const groupId = match ? parseInt(match[1], 10) : conv.targetId;
    items.push({
      id: 'quit-group',
      label: '退出该群',
      icon: LogOut,
      danger: true,
      divided: true,
      action: async () => {
        const ok = await confirmDialog({
          title: '退出群聊',
          message: `确定要退出群聊 "${conv.title}" 吗？退出后将无法接收此群消息。`,
          confirmText: '确认退出',
          danger: true,
        });
        if (!ok) return;
        try {
          const res = await fetch(`/api/group/quit?user_id=${chatStore.currentUserId}`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Authorization: `Bearer ${chatStore.userToken}`,
            },
            body: JSON.stringify({ group_id: groupId }),
          });
          const json = await res.json();
          if (json.code === 0) {
            chatStore.deleteConversation(conv.id);
            chatStore.fetchConversations();
          } else {
            alertDialog({
              title: '退出失败',
              message: json.msg || '无法退出该群聊',
              type: 'danger',
            });
          }
        } catch (_) {}
      },
    });
  }

  contextMenuStore.show(e, items);
};

let initialWidth = 320;
let startX = 0;
let startWidth = 320;

const onMouseMove = (moveEvent: MouseEvent) => {
  const delta = moveEvent.clientX - startX;
  let newWidth = startWidth + delta;
  const maxRatio = 0.82;
  const maxWidth = window.innerWidth * maxRatio;
  if (newWidth > maxWidth) {
    newWidth = maxWidth;
  }
  chatStore.setSidebarWidth(newWidth);
};

const onKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    chatStore.setSidebarWidth(initialWidth);
    stopResize();
  }
};

const stopResize = () => {
  if (!isDragging.value) return;
  isDragging.value = false;
  window.removeEventListener('mousemove', onMouseMove);
  window.removeEventListener('mouseup', onMouseUp);
  window.removeEventListener('keydown', onKeyDown);
  document.body.style.removeProperty('cursor');
  document.body.style.removeProperty('user-select');
};

const onMouseUp = () => {
  stopResize();
};

const startResize = (e: MouseEvent) => {
  isDragging.value = true;
  startX = e.clientX;
  startWidth = chatStore.sidebarWidth;
  initialWidth = chatStore.sidebarWidth;

  document.body.style.cursor = 'col-resize';
  document.body.style.userSelect = 'none';

  window.addEventListener('mousemove', onMouseMove);
  window.addEventListener('mouseup', onMouseUp);
  window.addEventListener('keydown', onKeyDown);
};

const resetToHalf = () => {
  const half = Math.round(window.innerWidth / 2);
  chatStore.setSidebarWidth(half > 220 && half <= 520 ? half : 320);
};

const resetWidth = resetToHalf;

onUnmounted(() => {
  stopResize();
});

const formatLastMessage = (msg: string | undefined) => {
  if (!msg) return langStore.t('conv.noMessages');
  if (msg.startsWith('/api/file/download') || msg.includes('file_id')) {
    return '[文件]';
  }
  return msg;
};

const selectConv = (id: string) => {
  chatStore.selectConversation(id);
  router.push(`/chats/${id}`);
};

const filteredConversations = computed(() => {
  let list = chatStore.conversations;
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.toLowerCase();
    list = list.filter(
      (c) => c.title.toLowerCase().includes(kw) || c.lastMessage.toLowerCase().includes(kw)
    );
  }
  return [...list].sort((a, b) => (b.isPinned ? 1 : 0) - (a.isPinned ? 1 : 0));
});
</script>
