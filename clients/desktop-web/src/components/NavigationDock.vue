<template>
  <aside class="w-16 h-full flex flex-col items-center py-4 bg-mim-light-dock dark:bg-mim-dark-dock select-none z-20 relative">
    <!-- User Avatar & Quick Account Switcher -->
    <div class="relative mb-6">
      <button
        @click="showAccountMenu = !showAccountMenu"
        class="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center text-white font-semibold text-sm shadow-sm hover:opacity-90 transition-opacity focus:outline-none"
        title="Switch Active Account"
      >
        {{ chatStore.currentUserAvatar }}
      </button>

      <!-- Status Indicator -->
      <span
        class="absolute -bottom-0.5 -right-0.5 w-3 h-3 border-2 border-mim-light-dock dark:border-mim-dark-dock rounded-full transition-colors"
        :class="chatStore.isWsConnected ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'"
        :title="chatStore.isWsConnected ? 'Cluster WebSocket: Connected' : 'Cluster WebSocket: Connecting...'"
      ></span>

      <!-- Account Switcher Popover -->
      <div
        v-if="showAccountMenu"
        class="absolute left-16 top-0 w-56 rounded-xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 p-2 z-50 flex flex-col gap-1"
      >
        <div class="px-2 py-1.5 text-[11px] font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
          Switch Demo Account
        </div>
        <button
          v-for="item in chatStore.demoUsers"
          :key="item.user.id"
          @click="selectUser(item)"
          class="flex items-center gap-2.5 p-2 rounded-lg text-left text-xs transition-colors"
          :class="chatStore.currentUserId === item.user.id
            ? 'bg-blue-600/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 font-semibold'
            : 'text-mim-light-text dark:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5'"
        >
          <div class="w-7 h-7 rounded-lg bg-blue-600/20 text-blue-600 dark:text-blue-400 font-semibold flex items-center justify-center text-xs shrink-0">
            {{ item.user.avatar || item.user.nickname.substring(0, 2).toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="truncate">{{ item.user.nickname }}</div>
            <div class="text-[10px] text-mim-light-muted dark:text-mim-dark-muted font-mono truncate">ID: {{ item.user.id }} ({{ item.user.account }})</div>
          </div>
          <Check v-if="chatStore.currentUserId === item.user.id" class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400 shrink-0 inline-icon" />
        </button>

        <div class="mt-1 pt-1.5 border-t border-black/5 dark:border-white/10 px-2 py-1 text-[10px] text-mim-light-muted dark:text-mim-dark-muted flex items-center justify-between">
          <span>Backend WS:</span>
          <span :class="chatStore.isWsConnected ? 'text-emerald-500' : 'text-amber-500 font-medium'">
            {{ chatStore.isWsConnected ? 'Connected' : 'Offline' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Main Navigation Icons -->
    <nav class="flex-1 flex flex-col gap-2 w-full px-2">
      <button
        v-for="item in navItems"
        :key="item.id"
        @click="chatStore.activeTab = item.id"
        class="w-full h-11 rounded-xl flex items-center justify-center transition-all duration-150 relative group"
        :class="chatStore.activeTab === item.id 
          ? 'bg-blue-600/10 dark:bg-blue-500/15 text-blue-600 dark:text-blue-400 font-medium' 
          : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
      >
        <component :is="item.icon" class="w-5 h-5 inline-icon" />
        <span 
          v-if="chatStore.activeTab === item.id" 
          class="absolute left-0 top-3 bottom-3 w-1 bg-blue-600 dark:bg-blue-500 rounded-r-full"
        ></span>
        <!-- Tooltip -->
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          {{ item.name }}
        </span>
      </button>
    </nav>

    <!-- Bottom Controls: Theme Toggle & Settings -->
    <div class="flex flex-col gap-2 w-full px-2">
      <!-- Theme Switcher Button -->
      <button
        @click="cycleTheme"
        class="w-full h-11 rounded-xl flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 transition-colors relative group"
      >
        <Sun v-if="themeStore.currentMode === 'light'" class="w-5 h-5 inline-icon text-amber-500" />
        <Moon v-else-if="themeStore.currentMode === 'dark'" class="w-5 h-5 inline-icon text-blue-400" />
        <Monitor v-else class="w-5 h-5 inline-icon text-emerald-500" />
        
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          Theme: {{ themeStore.currentMode.toUpperCase() }}
        </span>
      </button>

      <button
        @click="$emit('open-settings')"
        class="w-full h-11 rounded-xl flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 transition-colors relative group"
        title="Settings"
      >
        <Settings class="w-5 h-5 inline-icon" />
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          Settings
        </span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { MessageSquare, Users, Phone, Settings, Sun, Moon, Monitor, Check } from 'lucide-vue-next';
import { useThemeStore } from '../stores/theme';
import { useChatStore, DemoUser } from '../stores/chat';

const themeStore = useThemeStore();
const chatStore = useChatStore();
const showAccountMenu = ref(false);

defineEmits(['open-settings']);

const navItems = [
  { id: 'chats', name: 'Conversations', icon: MessageSquare },
  { id: 'contacts', name: 'Contacts & Teams', icon: Users },
  { id: 'calls', name: 'Audio/Video Calls', icon: Phone },
];

const cycleTheme = () => {
  if (themeStore.currentMode === 'system') {
    themeStore.setTheme('dark');
  } else if (themeStore.currentMode === 'dark') {
    themeStore.setTheme('light');
  } else {
    themeStore.setTheme('system');
  }
};

const selectUser = (u: DemoUser) => {
  chatStore.switchAccount(u);
  showAccountMenu.value = false;
};
</script>
