<template>
  <header
    data-tauri-drag-region
    class="h-9 w-full flex items-center justify-between px-3 bg-mim-light-dock dark:bg-mim-dark-dock border-b border-black/[0.04] dark:border-white/[0.04] select-none text-mim-light-muted dark:text-mim-dark-muted text-xs z-30 shrink-0 transition-colors"
    :class="themeStore.currentPalette === 'vodyanitsa' ? '!bg-[#070e20]/95 !border-cyan-400/20' : ''"
  >
    <!-- Left: App Branding & Status -->
    <div class="flex items-center gap-2.5 pointer-events-none" data-tauri-drag-region>
      <div class="flex items-center gap-1.5 font-semibold text-mim-light-text dark:text-mim-dark-text tracking-tight text-xs">
        <template v-if="themeStore.currentPalette === 'vodyanitsa'">
          <Snowflake class="w-4 h-4 text-cyan-300 drop-shadow-[0_0_6px_rgba(56,189,248,0.8)] inline-icon shrink-0" />
          <span class="text-white font-semibold tracking-wide text-xs drop-shadow-[0_0_6px_rgba(56,189,248,0.4)]">
            Vodyanitsa-Mim
          </span>
        </template>
        <template v-else>
          <div class="w-4 h-4 rounded-md flex items-center justify-center text-white shadow-2xs bg-mim-accent">
            <ShieldCheck class="w-2.5 h-2.5 inline-icon" />
          </div>
          <span>Mim</span>
        </template>
      </div>

      <!-- Network / Cluster Status Pill -->
      <div
        v-if="chatStore.isLoggedIn"
        class="flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-mono"
        :class="themeStore.currentPalette === 'vodyanitsa'
          ? 'bg-[#0a232c]/85 border border-[#14b8a6]/40 text-[#2dd4bf]'
          : 'bg-black/[0.03] dark:bg-white/[0.04]'"
      >
        <template v-if="themeStore.currentPalette === 'vodyanitsa'">
          <span class="text-xs text-[#2dd4bf] animate-pulse drop-shadow-[0_0_6px_#2dd4bf]">✦</span>
          <span>{{ chatStore.isWsConnected ? '24ms 正常' : '重连中' }}</span>
        </template>
        <template v-else>
          <span
            class="w-1.5 h-1.5 rounded-full"
            :class="chatStore.isWsConnected ? 'bg-emerald-400 animate-pulse' : 'bg-amber-500 animate-pulse'"
          ></span>
          <span>{{ chatStore.isWsConnected ? '24ms 正常' : '重连中' }}</span>
        </template>
      </div>
    </div>

    <!-- Center: Quick Command Palette Trigger (Ctrl+K) -->
    <div
      v-if="chatStore.isLoggedIn"
      data-tauri-drag-region
      class="flex-1 max-w-sm mx-4 flex items-center justify-center cursor-pointer"
      @click="$emit('open-command-palette')"
    >
      <button
        type="button"
        class="w-full h-7 px-3 rounded-xl flex items-center justify-between text-[11px] transition-all cursor-pointer group relative overflow-hidden"
        :class="themeStore.currentPalette === 'vodyanitsa'
          ? 'bg-[#0a1835]/75 border border-cyan-400/40 text-cyan-200 shadow-[0_0_12px_rgba(56,189,248,0.18)] hover:border-cyan-300'
          : 'bg-black/[0.03] dark:bg-white/[0.04] hover:bg-black/[0.06] dark:hover:bg-white/[0.08] border border-black/[0.04] dark:border-white/[0.04]'"
      >
        <!-- Corner Sparkles & Melody Staves in Vodyanitsa Mode -->
        <template v-if="themeStore.currentPalette === 'vodyanitsa'">
          <span class="absolute -bottom-0.5 left-1 text-[9px] text-cyan-300 leading-none pointer-events-none drop-shadow-[0_0_4px_#38bdf8]">✦</span>
          <span class="absolute -bottom-0.5 right-1 text-[9px] text-cyan-300 leading-none pointer-events-none drop-shadow-[0_0_4px_#38bdf8]">✦</span>
          <!-- Flowing Melody Staves on the Right -->
          <img
            :src="searchMelody"
            alt=""
            class="absolute right-14 top-0 h-full w-[140px] object-contain pointer-events-none opacity-50 group-hover:opacity-85 transition-opacity"
          />
        </template>

        <span
          class="flex items-center gap-1.5 transition-colors relative z-10"
          :class="themeStore.currentPalette === 'vodyanitsa'
            ? 'text-cyan-200 group-hover:text-white'
            : 'text-mim-light-muted dark:text-mim-dark-muted group-hover:text-mim-light-text dark:group-hover:text-mim-dark-text'"
        >
          <Search class="w-3.5 h-3.5 inline-icon text-cyan-400" />
          <span>{{ langStore.currentLang === 'zh' ? '搜索会话、联系人或群聊...' : 'Search chats, contacts, groups...' }}</span>
        </span>
        <kbd
          class="px-1.5 py-0.5 rounded text-[9px] font-mono relative z-10"
          :class="themeStore.currentPalette === 'vodyanitsa'
            ? 'bg-cyan-950/80 text-cyan-200 border border-cyan-400/40 shadow-xs'
            : 'bg-black/5 dark:bg-white/10 text-mim-light-muted dark:text-mim-dark-muted'"
        >
          Ctrl K
        </kbd>
      </button>
    </div>
    <div v-else class="flex-1" data-tauri-drag-region></div>

    <!-- Right: Shortcuts Help & Windows System Controls -->
    <div class="flex items-center">
      <!-- Shortcuts Help Button -->
      <button
        @click="$emit('open-shortcuts')"
        class="h-7 px-2 rounded-md hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors flex items-center gap-1 cursor-pointer mr-1"
        :title="langStore.currentLang === 'zh' ? '快捷键帮助 (Ctrl + /)' : 'Shortcuts Help (Ctrl + /)'"
      >
        <Keyboard class="w-3.5 h-3.5 inline-icon" />
      </button>

      <!-- Windows Native Window Buttons -->
      <div class="flex items-center -mr-3">
        <!-- Minimize -->
        <button
          @click="minimizeWindow"
          class="w-11 h-9 flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors cursor-pointer"
          title="最小化"
        >
          <Minus class="w-3.5 h-3.5 inline-icon" />
        </button>

        <!-- Maximize / Restore -->
        <button
          @click="toggleMaximizeWindow"
          class="w-11 h-9 flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors cursor-pointer"
          :title="isMaximized ? '还原' : '最大化'"
        >
          <Square class="w-3 h-3 inline-icon" />
        </button>

        <!-- Close (Windows 11 Red Hover) -->
        <button
          @click="closeWindow"
          class="w-11 h-9 flex items-center justify-center hover:bg-[#e81123] hover:text-white text-mim-light-muted dark:text-mim-dark-muted transition-colors cursor-pointer"
          title="关闭"
        >
          <X class="w-3.5 h-3.5 inline-icon" />
        </button>
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ShieldCheck, Search, Keyboard, Minus, Square, X, Snowflake } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useThemeStore } from '../stores/theme';
import searchMelody from '../assets/vodyanitsa/images/titlebar_search_melody.png';

defineEmits(['open-command-palette', 'open-shortcuts']);

const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();
const isMaximized = ref(false);

const isTauri = typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window;

const minimizeWindow = async () => {
  if (isTauri) {
    try {
      const { getCurrentWindow } = await import('@tauri-apps/api/window');
      getCurrentWindow().minimize();
    } catch (_) {}
  }
};

const toggleMaximizeWindow = async () => {
  if (isTauri) {
    try {
      const { getCurrentWindow } = await import('@tauri-apps/api/window');
      const win = getCurrentWindow();
      win.toggleMaximize();
      isMaximized.value = await win.isMaximized();
    } catch (_) {}
  } else {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch(() => {});
      isMaximized.value = true;
    } else {
      document.exitFullscreen().catch(() => {});
      isMaximized.value = false;
    }
  }
};

const closeWindow = async () => {
  if (isTauri) {
    try {
      const { getCurrentWindow } = await import('@tauri-apps/api/window');
      getCurrentWindow().close();
    } catch (_) {}
  }
};

onMounted(() => {
  if (!isTauri) {
    document.addEventListener('fullscreenchange', () => {
      isMaximized.value = !!document.fullscreenElement;
    });
  }
});
</script>
