<template>
  <aside class="w-16 h-full flex flex-col items-center py-4 bg-mim-light-dock dark:bg-mim-dark-dock select-none z-20">
    <!-- User Avatar -->
    <div class="relative mb-6">
      <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-500 flex items-center justify-center text-white font-semibold text-sm shadow-sm cursor-pointer hover:opacity-90 transition-opacity">
        AC
      </div>
      <span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-emerald-500 border-2 border-mim-light-dock dark:border-mim-dark-dock rounded-full"></span>
    </div>

    <!-- Main Navigation Icons -->
    <nav class="flex-1 flex flex-col gap-2 w-full px-2">
      <button
        v-for="item in navItems"
        :key="item.id"
        @click="activeTab = item.id"
        class="w-full h-11 rounded-xl flex items-center justify-center transition-all duration-150 relative group"
        :class="activeTab === item.id 
          ? 'bg-blue-600/10 dark:bg-blue-500/15 text-blue-600 dark:text-blue-400 font-medium' 
          : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
      >
        <component :is="item.icon" class="w-5 h-5 inline-icon" />
        <span 
          v-if="activeTab === item.id" 
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
        class="w-full h-11 rounded-xl flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 transition-colors relative group"
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
import { MessageSquare, Users, Phone, Settings, Sun, Moon, Monitor } from 'lucide-vue-next';
import { useThemeStore } from '../stores/theme';

const themeStore = useThemeStore();
const activeTab = ref('chats');

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
</script>
