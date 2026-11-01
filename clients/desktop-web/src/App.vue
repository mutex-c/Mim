<template>
  <div class="w-screen h-screen flex flex-col overflow-hidden bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Top Native Window Titlebar (Windows 11 / Desktop Frameless) -->
    <Titlebar
      @open-command-palette="showCommandPalette = true"
      @open-shortcuts="showShortcutsHelp = true"
    />

    <!-- Main Workspace Container -->
    <div class="flex-1 w-full flex overflow-hidden relative">
      <!-- Leftmost Navigation Dock (Only displayed when logged in and not on auth pages) -->
      <NavigationDock
        v-if="chatStore.isLoggedIn && !isAuthRoute"
        @open-settings="showSettings = true"
        @open-auth="chatStore.showAuthModal = true"
      />

      <!-- Dynamic Router View for Pages (/login, /register, /chats, /contacts, /calls) -->
      <router-view class="flex-1 h-full overflow-hidden" />
    </div>

    <!-- Modal Mode Auth (When Logged In and switching/adding accounts) -->
    <Transition name="app-modal">
      <AuthModal
        v-if="chatStore.isLoggedIn && chatStore.showAuthModal"
        :is-fullscreen="false"
        @close="chatStore.showAuthModal = false"
      />
    </Transition>

    <!-- Global Command Palette (Ctrl+K) -->
    <Transition name="app-modal">
      <CommandPalette
        v-if="showCommandPalette && chatStore.isLoggedIn"
        @close="showCommandPalette = false"
        @open-settings="showSettings = true"
      />
    </Transition>

    <!-- Shortcuts Help Modal (Ctrl+/) -->
    <Transition name="app-modal">
      <ShortcutsHelpModal
        v-if="showShortcutsHelp"
        @close="showShortcutsHelp = false"
      />
    </Transition>

    <!-- Global Settings Modal -->
    <Transition name="app-modal">
      <SettingsModal
        v-if="showSettings"
        @close="showSettings = false"
      />
    </Transition>

    <!-- Global Desktop Context Menu -->
    <ContextMenu />

    <!-- Global Desktop Dialog (Confirm / Alert) -->
    <GlobalDialog />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import {
  Copy,
  Clipboard,
  CheckSquare,
  RotateCcw,
  Search,
  Sun,
  Moon,
  Keyboard,
  Settings,
} from 'lucide-vue-next';
import Titlebar from './components/Titlebar.vue';
import NavigationDock from './components/NavigationDock.vue';
import CommandPalette from './components/CommandPalette.vue';
import ShortcutsHelpModal from './components/ShortcutsHelpModal.vue';
import SettingsModal from './components/SettingsModal.vue';
import AuthModal from './components/AuthModal.vue';
import ContextMenu from './components/ContextMenu.vue';
import GlobalDialog from './components/GlobalDialog.vue';
import { useThemeStore } from './stores/theme';
import { useChatStore } from './stores/chat';
import { useContextMenuStore, ContextMenuItem } from './stores/contextMenu';

const route = useRoute();
const isAuthRoute = computed(() => route.path === '/login' || route.path === '/register');

// Initialize Stores
const themeStore = useThemeStore();
const chatStore = useChatStore();
const contextMenuStore = useContextMenuStore();

const showSettings = ref(false);
const showCommandPalette = ref(false);
const showShortcutsHelp = ref(false);

const handleGlobalContextMenu = (e: MouseEvent) => {
  e.preventDefault();

  const selection = window.getSelection()?.toString();
  const target = e.target as HTMLElement | null;
  const isInput = target?.tagName === 'INPUT' || target?.tagName === 'TEXTAREA' || target?.isContentEditable;

  const items: ContextMenuItem[] = [];

  if (selection && selection.trim().length > 0) {
    items.push({
      id: 'copy-selection',
      label: '复制选中文本',
      icon: Copy,
      shortcut: 'Ctrl+C',
      action: () => {
        navigator.clipboard.writeText(selection);
      },
    });
  }

  if (isInput) {
    items.push({
      id: 'paste',
      label: '粘贴',
      icon: Clipboard,
      shortcut: 'Ctrl+V',
      action: async () => {
        try {
          const text = await navigator.clipboard.readText();
          if (target && 'value' in target) {
            const input = target as HTMLInputElement | HTMLTextAreaElement;
            const start = input.selectionStart || 0;
            const end = input.selectionEnd || 0;
            const val = input.value;
            input.value = val.substring(0, start) + text + val.substring(end);
            input.selectionStart = input.selectionEnd = start + text.length;
            input.dispatchEvent(new Event('input', { bubbles: true }));
          }
        } catch (_) {}
      },
    });

    items.push({
      id: 'select-all',
      label: '全选',
      icon: CheckSquare,
      shortcut: 'Ctrl+A',
      action: () => {
        if (target && 'select' in target) {
          (target as HTMLInputElement).select();
        }
      },
    });
  }

  items.push({
    id: 'reload',
    label: '刷新界面',
    icon: RotateCcw,
    shortcut: 'Ctrl+R',
    divided: items.length > 0,
    action: () => {
      window.location.reload();
    },
  });

  if (chatStore.isLoggedIn) {
    items.push({
      id: 'command-palette',
      label: '快捷搜索与命令',
      icon: Search,
      shortcut: 'Ctrl+K',
      action: () => {
        showCommandPalette.value = true;
      },
    });
  }

  items.push({
    id: 'toggle-theme',
    label: themeStore.isDark ? '切换为亮色模式' : '切换为深色模式',
    icon: themeStore.isDark ? Sun : Moon,
    action: () => {
      themeStore.setTheme(themeStore.isDark ? 'light' : 'dark');
    },
  });

  items.push({
    id: 'shortcuts',
    label: '快捷键列表',
    icon: Keyboard,
    shortcut: 'Ctrl+/',
    action: () => {
      showShortcutsHelp.value = true;
    },
  });

  if (chatStore.isLoggedIn) {
    items.push({
      id: 'settings',
      label: '软件设置',
      icon: Settings,
      divided: true,
      action: () => {
        showSettings.value = true;
      },
    });
  }

  contextMenuStore.show(e, items);
};

const handleGlobalKeyDown = (e: KeyboardEvent) => {
  const isCtrl = e.ctrlKey || e.metaKey;

  // 1. Ctrl + K: Command Palette
  if (isCtrl && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    showCommandPalette.value = !showCommandPalette.value;
    return;
  }

  // 2. Ctrl + /: Shortcuts Help
  if (isCtrl && e.key === '/') {
    e.preventDefault();
    showShortcutsHelp.value = !showShortcutsHelp.value;
    return;
  }

  // 3. Escape: Close top modal
  if (e.key === 'Escape') {
    if (showCommandPalette.value) {
      showCommandPalette.value = false;
      return;
    }
    if (showShortcutsHelp.value) {
      showShortcutsHelp.value = false;
      return;
    }
    if (showSettings.value) {
      showSettings.value = false;
      return;
    }
  }

  // 4. Hotkeys for active call (Ctrl+Shift+M, Ctrl+Shift+V, Ctrl+Shift+H)
  if (chatStore.activeCall && isCtrl && e.shiftKey) {
    if (e.key.toLowerCase() === 'm') {
      e.preventDefault();
      chatStore.toggleCallAudio();
    } else if (e.key.toLowerCase() === 'v') {
      e.preventDefault();
      chatStore.toggleCallVideo();
    } else if (e.key.toLowerCase() === 'h') {
      e.preventDefault();
      chatStore.endActiveCall();
    }
  }
};

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeyDown);
  window.addEventListener('contextmenu', handleGlobalContextMenu);

  // Request browser desktop notification permission smoothly
  if (typeof Notification !== 'undefined' && Notification.permission === 'default') {
    Notification.requestPermission().catch(() => {});
  }
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeyDown);
  window.removeEventListener('contextmenu', handleGlobalContextMenu);
});
</script>

<style>
/* App Modal Transitions (overlays/dialog spec) */
.app-modal-enter-active {
  transition: opacity 240ms cubic-bezier(0.16, 1, 0.3, 1);
}
.app-modal-leave-active {
  transition: opacity 160ms cubic-bezier(0.22, 1, 0.36, 1);
}
.app-modal-enter-active > div {
  transition: transform 320ms cubic-bezier(0.16, 1, 0.3, 1),
              opacity 240ms cubic-bezier(0.16, 1, 0.3, 1);
}
.app-modal-leave-active > div {
  transition: transform 160ms cubic-bezier(0.22, 1, 0.36, 1),
              opacity 140ms cubic-bezier(0.22, 1, 0.36, 1);
}

.app-modal-enter-from {
  opacity: 0;
}
.app-modal-enter-from > div {
  opacity: 0;
  transform: translateY(8px) scale(0.96);
}

.app-modal-leave-to {
  opacity: 0;
}
.app-modal-leave-to > div {
  opacity: 0;
  transform: translateY(4px) scale(0.98);
}
</style>
