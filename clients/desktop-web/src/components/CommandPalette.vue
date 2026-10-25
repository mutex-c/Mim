<template>
  <div
    class="fixed inset-0 z-[80] bg-black/50 backdrop-blur-sm flex items-start justify-center pt-24 px-4 select-none animate-in fade-in duration-150"
    @click.self="$emit('close')"
  >
    <div
      class="w-full max-w-xl rounded-2xl bg-mim-light-card dark:bg-mim-dark-card border border-black/10 dark:border-white/10 shadow-2xl overflow-hidden flex flex-col transition-all"
    >
      <!-- Search Input Header -->
      <div class="p-3.5 border-b border-black/[0.06] dark:border-white/[0.08] flex items-center gap-3">
        <Search class="w-4 h-4 text-mim-accent shrink-0 inline-icon" />
        <input
          ref="inputRef"
          v-model="searchQuery"
          type="text"
          :placeholder="langStore.currentLang === 'zh' ? '搜索会话、联系人或执行快捷指令...' : 'Search chats, contacts, or run commands...'"
          class="flex-1 bg-transparent text-sm text-mim-light-text dark:text-mim-dark-text placeholder-mim-light-muted dark:placeholder-mim-dark-muted outline-none border-none"
          @keydown.down.prevent="navigateDown"
          @keydown.up.prevent="navigateUp"
          @keydown.enter.prevent="selectCurrent"
          @keydown.esc.prevent="$emit('close')"
        />
        <kbd class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10 text-[10px] font-mono text-mim-light-muted dark:text-mim-dark-muted">
          ESC
        </kbd>
      </div>

      <!-- Category Filter Pills -->
      <div class="px-3.5 py-2 border-b border-black/[0.04] dark:border-white/[0.04] flex items-center gap-1.5 text-xs">
        <button
          v-for="cat in categories"
          :key="cat.id"
          @click="activeCategory = cat.id"
          class="px-2.5 py-1 rounded-lg transition-colors cursor-pointer text-[11px]"
          :class="activeCategory === cat.id
            ? 'bg-mim-accent text-white font-medium shadow-2xs'
            : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5'"
        >
          {{ cat.label }}
        </button>
      </div>

      <!-- Results List -->
      <div class="max-h-80 overflow-y-auto p-2 space-y-1">
        <div v-if="filteredResults.length === 0" class="py-8 text-center text-xs text-mim-light-muted dark:text-mim-dark-muted">
          {{ langStore.currentLang === 'zh' ? '未找到匹配结果' : 'No matching results found' }}
        </div>

        <div
          v-for="(item, index) in filteredResults"
          :key="item.id"
          @click="executeItem(item)"
          @mouseenter="selectedIndex = index"
          class="w-full px-3 py-2 rounded-xl flex items-center justify-between cursor-pointer transition-colors text-xs"
          :class="selectedIndex === index
            ? 'bg-mim-accent-subtle/60 text-mim-accent font-medium ring-1 ring-mim-accent/30'
            : 'text-mim-light-text dark:text-mim-dark-text hover:bg-black/[0.03] dark:hover:bg-white/[0.04]'"
        >
          <div class="flex items-center gap-3 min-w-0">
            <!-- Icon / Avatar -->
            <div
              v-if="item.avatarText"
              class="w-8 h-8 rounded-lg flex items-center justify-center font-bold text-xs shrink-0"
              :style="{ backgroundColor: getAvatarColor(item.id, item.type === 'group').bg, color: getAvatarColor(item.id, item.type === 'group').text }"
            >
              {{ item.avatarText }}
            </div>
            <div
              v-else
              class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/10 flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted shrink-0"
            >
              <component :is="item.icon" class="w-4 h-4 inline-icon" />
            </div>

            <!-- Details -->
            <div class="min-w-0">
              <div class="truncate font-medium flex items-center gap-1.5">
                <span>{{ item.title }}</span>
                <span v-if="item.badge" class="px-1.5 py-0.2 rounded-full text-[10px] bg-black/5 dark:bg-white/10 text-mim-light-muted dark:text-mim-dark-muted font-normal">
                  {{ item.badge }}
                </span>
              </div>
              <div class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted truncate">
                {{ item.subtitle }}
              </div>
            </div>
          </div>

          <div class="flex items-center gap-2 shrink-0 text-mim-light-muted dark:text-mim-dark-muted text-[11px]">
            <span v-if="selectedIndex === index" class="hidden sm:inline font-mono text-[10px]">
              回车选择
            </span>
            <CornerDownLeft v-if="selectedIndex === index" class="w-3.5 h-3.5 inline-icon" />
          </div>
        </div>
      </div>

      <!-- Keyboard Navigation Footer -->
      <div class="p-2.5 bg-black/[0.02] dark:bg-white/[0.02] border-t border-black/[0.04] dark:border-white/[0.04] flex items-center justify-between text-[11px] text-mim-light-muted dark:text-mim-dark-muted font-mono">
        <div class="flex items-center gap-3">
          <span><kbd class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/10">↑</kbd> <kbd class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/10">↓</kbd> 移动</span>
          <span><kbd class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10">Enter</kbd> 打开</span>
          <span><kbd class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/10">Esc</kbd> 退出</span>
        </div>
        <span class="text-[10px]">Mim Command Palette</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import {
  Search,
  CornerDownLeft,
  Settings,
  SunMoon,
  Languages,
  UserPlus,
  Users,
  MessageSquare,
  ShieldCheck,
} from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useThemeStore } from '../stores/theme';
import { getAvatarColor } from '../utils/avatar';

const emit = defineEmits(['close', 'open-settings']);
const router = useRouter();
const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();

const inputRef = ref<HTMLInputElement | null>(null);
const searchQuery = ref('');
const selectedIndex = ref(0);
const activeCategory = ref<'all' | 'chats' | 'contacts' | 'actions'>('all');

const categories = computed(() => [
  { id: 'all' as const, label: langStore.currentLang === 'zh' ? '全部' : 'All' },
  { id: 'chats' as const, label: langStore.currentLang === 'zh' ? '会话' : 'Chats' },
  { id: 'contacts' as const, label: langStore.currentLang === 'zh' ? '联系人' : 'Contacts' },
  { id: 'actions' as const, label: langStore.currentLang === 'zh' ? '指令' : 'Actions' },
]);

interface PaletteItem {
  id: string;
  type: 'chat' | 'contact' | 'action' | 'group';
  title: string;
  subtitle: string;
  avatarText?: string;
  badge?: string;
  icon?: any;
  action: () => void;
}

const allItems = computed<PaletteItem[]>(() => {
  const items: PaletteItem[] = [];

  // 1. Conversations
  chatStore.conversations.forEach((c) => {
    items.push({
      id: `conv_${c.id}`,
      type: c.type === 'group' ? 'group' : 'chat',
      title: c.title,
      subtitle: c.lastMessage || (c.type === 'group' ? '团队协作群' : '端到端加密私聊'),
      avatarText: c.avatarText,
      badge: c.type === 'group' ? '群聊' : (c.isOnline ? '在线' : '离线'),
      action: () => {
        chatStore.selectConversation(c.id);
        router.push(`/chats/${c.id}`);
        emit('close');
      },
    });
  });

  // 2. Contacts from demo users
  chatStore.demoUsers
    .filter((u) => u.user.id !== chatStore.currentUserId)
    .forEach((u) => {
      items.push({
        id: `user_${u.user.id}`,
        type: 'contact',
        title: u.user.nickname,
        subtitle: `@${u.user.account} · ID: ${u.user.id}`,
        avatarText: u.user.avatar || u.user.nickname.substring(0, 2).toUpperCase(),
        badge: '联系人',
        action: () => {
          const convId = `${Math.min(chatStore.currentUserId, u.user.id)}_${Math.max(chatStore.currentUserId, u.user.id)}`;
          chatStore.selectConversation(convId);
          router.push(`/chats/${convId}`);
          emit('close');
        },
      });
    });

  // 3. Quick System Actions
  items.push({
    id: 'action_settings',
    type: 'action',
    title: langStore.currentLang === 'zh' ? '打开系统设置' : 'Open System Settings',
    subtitle: '配置账号、安全、通知与加密偏好',
    icon: Settings,
    badge: '设置',
    action: () => {
      emit('open-settings');
      emit('close');
    },
  });

  items.push({
    id: 'action_theme',
    type: 'action',
    title: langStore.currentLang === 'zh' ? '切换色彩主题 (明 / 暗)' : 'Toggle Color Theme (Light / Dark)',
    subtitle: `当前主题: ${themeStore.currentMode}`,
    icon: SunMoon,
    badge: '外观',
    action: () => {
      themeStore.setTheme(themeStore.currentMode === 'dark' ? 'light' : 'dark');
      emit('close');
    },
  });

  items.push({
    id: 'action_lang',
    type: 'action',
    title: langStore.currentLang === 'zh' ? 'Switch Language (切换英文)' : '切换为简体中文',
    subtitle: `Current: ${langStore.currentLang === 'zh' ? '简体中文' : 'English'}`,
    icon: Languages,
    badge: '语言',
    action: () => {
      langStore.setLang(langStore.currentLang === 'zh' ? 'en' : 'zh');
      emit('close');
    },
  });

  return items;
});

const filteredResults = computed(() => {
  let list = allItems.value;

  if (activeCategory.value === 'chats') {
    list = list.filter((i) => i.type === 'chat' || i.type === 'group');
  } else if (activeCategory.value === 'contacts') {
    list = list.filter((i) => i.type === 'contact');
  } else if (activeCategory.value === 'actions') {
    list = list.filter((i) => i.type === 'action');
  }

  if (!searchQuery.value.trim()) {
    return list;
  }

  const q = searchQuery.value.toLowerCase().trim();
  return list.filter(
    (i) => i.title.toLowerCase().includes(q) || i.subtitle.toLowerCase().includes(q) || (i.badge && i.badge.toLowerCase().includes(q))
  );
});

const navigateDown = () => {
  if (filteredResults.value.length === 0) return;
  selectedIndex.value = (selectedIndex.value + 1) % filteredResults.value.length;
};

const navigateUp = () => {
  if (filteredResults.value.length === 0) return;
  selectedIndex.value = (selectedIndex.value - 1 + filteredResults.value.length) % filteredResults.value.length;
};

const selectCurrent = () => {
  const item = filteredResults.value[selectedIndex.value];
  if (item) {
    executeItem(item);
  }
};

const executeItem = (item: PaletteItem) => {
  item.action();
};

onMounted(() => {
  nextTick(() => {
    inputRef.value?.focus();
  });
});
</script>
