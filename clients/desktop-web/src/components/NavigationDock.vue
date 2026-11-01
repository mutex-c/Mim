<template>
  <aside
    class="w-16 h-full flex flex-col items-center py-3.5 bg-mim-light-dock dark:bg-mim-dark-dock select-none z-20 relative transition-colors shrink-0"
    :class="themeStore.currentPalette === 'vodyanitsa' ? '!bg-[#060e22]/95 border-r border-cyan-400/20 shadow-[2px_0_20px_rgba(6,14,34,0.8)]' : ''"
  >
    <!-- User Avatar & Quick Account Switcher -->
    <div class="relative mb-4 flex flex-col items-center">
      <button
        @click="showAccountMenu = !showAccountMenu; if (themeStore.currentPalette === 'genshin') playGenshinClick();"
        class="relative flex items-center justify-center font-bold text-sm transition-all focus:outline-none cursor-pointer group"
        :class="[
          themeStore.currentPalette === 'vodyanitsa'
            ? 'w-12 h-12'
            : (themeStore.currentPalette === 'genshin'
                ? 'w-10 h-10 rounded-xl border-2 border-[#caa159] shadow-[0_0_10px_rgba(202,161,89,0.45)]'
                : (chatStore.isWsConnected 
                    ? 'w-10 h-10 rounded-xl ring-2 ring-emerald-500 ring-offset-2 ring-offset-mim-dock' 
                    : 'w-10 h-10 rounded-xl ring-2 ring-amber-500 ring-offset-2 ring-offset-mim-dock animate-pulse'))
        ]"
        :title="langStore.t('nav.switchAccount')"
      >
        <!-- Vodyanitsa Astrological Frost Dial Frame + Avatar -->
        <template v-if="themeStore.currentPalette === 'vodyanitsa'">
          <!-- Inner Circular Character Avatar -->
          <div class="w-8 h-8 rounded-full overflow-hidden relative z-10 border border-cyan-300/40 shadow-inner">
            <img
              :src="vodyanitsaAvatar"
              alt="沃雅妮莎"
              class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
            />
          </div>
          <!-- Outer Astrolabe Dial Wreath with lilies & compass points -->
          <img
            :src="avatarDial"
            alt="星盘花环"
            class="absolute inset-0 w-12 h-12 object-contain pointer-events-none drop-shadow-[0_0_8px_rgba(56,189,248,0.7)] group-hover:rotate-45 transition-transform duration-500 z-20"
          />
        </template>
        <template v-else>
          <img v-if="chatStore.currentUserAvatar && (chatStore.currentUserAvatar.startsWith('http') || chatStore.currentUserAvatar.startsWith('data:image'))" :src="chatStore.currentUserAvatar" class="w-full h-full object-cover rounded-xl" />
          <span v-else>{{ chatStore.currentUserAvatar || chatStore.currentUserName.substring(0, 2).toUpperCase() }}</span>
        </template>
      </button>

      <!-- Genshin AR 60 Badge -->
      <span
        v-if="themeStore.currentPalette === 'genshin'"
        class="mt-1 px-1 py-0.2 rounded-sm bg-[#caa159]/20 border border-[#caa159]/60 text-[9px] font-bold text-amber-300 font-mono tracking-tighter"
      >
        Lv.60
      </span>

      <!-- Status Indicator -->
      <span
        v-if="themeStore.currentPalette !== 'genshin' && themeStore.currentPalette !== 'vodyanitsa'"
        class="absolute -bottom-0.5 -right-0.5 w-3 h-3 border-2 border-mim-light-dock dark:border-mim-dark-dock rounded-full transition-colors"
        :class="chatStore.isWsConnected ? 'bg-emerald-500' : 'bg-amber-500 animate-pulse'"
        :title="chatStore.isWsConnected ? `${langStore.t('nav.backendWs')}: ${langStore.t('nav.online')}` : `${langStore.t('nav.backendWs')}: ${langStore.t('nav.offline')}`"
      ></span>

      <!-- Account Switcher Popover -->
      <Transition name="popover-dock">
        <div
          v-if="showAccountMenu"
          class="absolute left-16 top-0 w-64 rounded-xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 p-2.5 z-50 flex flex-col gap-1.5"
        >
          <div class="px-2 py-1.5 text-[11px] font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
            {{ langStore.t('nav.switchAccount') }}
          </div>
          <button
            v-for="item in chatStore.demoUsers"
            :key="item.user.id"
            @click="selectUser(item)"
            class="flex items-center gap-2.5 p-2 rounded-lg text-left text-xs transition-colors cursor-pointer"
            :class="chatStore.currentUserId === item.user.id
              ? 'bg-mim-accent-subtle/50 text-mim-accent font-semibold ring-1 ring-mim-accent/30'
              : 'text-mim-light-text dark:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5'"
          >
            <div
              class="w-7 h-7 rounded-lg font-bold flex items-center justify-center text-xs shrink-0"
              :class="chatStore.currentUserId === item.user.id ? 'ring-1.5 ring-emerald-500 ring-offset-1' : 'ring-1 ring-black/10 dark:ring-white/20'"
              :style="{ backgroundColor: getAvatarColor(item.user.id).bg, color: getAvatarColor(item.user.id).text }"
            >
              {{ item.user.avatar || item.user.nickname.substring(0, 2).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <div class="truncate">{{ item.user.nickname }}</div>
              <div class="text-[10px] text-mim-light-muted dark:text-mim-dark-muted font-mono truncate">ID: {{ item.user.id }} ({{ item.user.account }})</div>
            </div>
            <Check v-if="chatStore.currentUserId === item.user.id" class="w-3.5 h-3.5 text-mim-accent shrink-0 inline-icon" />
          </button>

          <!-- Extra Auth Actions -->
          <div class="mt-1 pt-1.5 border-t border-black/5 dark:border-white/10 flex flex-col gap-0.5">
            <button
              @click="openAuth"
              class="w-full flex items-center gap-2 p-1.5 rounded-lg text-xs text-mim-light-text dark:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
            >
              <UserPlus class="w-3.5 h-3.5 inline-icon text-mim-accent shrink-0" />
              <span>{{ langStore.t('auth.loginOtherAccount') }}</span>
            </button>
            <button
              @click="handleLogout"
              class="w-full flex items-center gap-2 p-1.5 rounded-lg text-xs text-rose-500 hover:bg-rose-500/10 transition-colors cursor-pointer"
            >
              <LogOut class="w-3.5 h-3.5 inline-icon text-rose-500 shrink-0" />
              <span>{{ langStore.t('auth.logoutBtn') }}</span>
            </button>
          </div>

          <div class="mt-1 pt-1.5 border-t border-black/5 dark:border-white/10 px-2 py-1 text-[10px] text-mim-light-muted dark:text-mim-dark-muted flex items-center justify-between">
            <span>{{ langStore.t('nav.backendWs') }}:</span>
            <span :class="chatStore.isWsConnected ? 'text-emerald-500' : 'text-amber-500 font-medium'">
              {{ chatStore.isWsConnected ? langStore.t('nav.online') : langStore.t('nav.offline') }}
            </span>
          </div>
        </div>
      </Transition>
    </div>

    <!-- Main Navigation Icons -->
    <nav class="flex-1 flex flex-col gap-3 w-full px-2">
      <button
        v-for="item in navItems"
        :key="item.id"
        @click="navigateTo(item.path)"
        class="w-full h-11 rounded-2xl flex items-center justify-center transition-[background-color,transform,box-shadow] duration-150 active:scale-95 relative group cursor-pointer"
        :class="isItemActive(item) 
          ? (themeStore.currentPalette === 'vodyanitsa'
              ? 'bg-gradient-to-r from-cyan-500/30 via-cyan-400/15 to-transparent shadow-[inset_0_1px_1px_rgba(255,255,255,0.2)]'
              : 'bg-mim-accent-subtle text-mim-accent font-medium') 
          : (themeStore.currentPalette === 'vodyanitsa'
              ? 'hover:bg-cyan-900/25'
              : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text')"
      >
        <!-- Vodyanitsa Haute-Craft Icons -->
        <template v-if="themeStore.currentPalette === 'vodyanitsa'">
          <img
            v-if="item.id === 'chats'"
            :src="iconWaveWing"
            alt="Chats"
            class="w-7 h-7 object-contain transition-all duration-200 group-hover:scale-110"
            :class="isItemActive(item) ? 'drop-shadow-[0_0_10px_rgba(56,189,248,0.9)] filter brightness-110 scale-105' : 'opacity-80 group-hover:opacity-100'"
          />
          <img
            v-else-if="item.id === 'contacts'"
            :src="iconLily"
            alt="Contacts"
            class="w-7 h-7 object-contain transition-all duration-200 group-hover:scale-110"
            :class="isItemActive(item) ? 'drop-shadow-[0_0_10px_rgba(56,189,248,0.9)] filter brightness-110 scale-105' : 'opacity-80 group-hover:opacity-100'"
          />
          <img
            v-else-if="item.id === 'calls'"
            :src="iconConch"
            alt="Calls"
            class="w-7 h-7 object-contain transition-all duration-200 group-hover:scale-110"
            :class="isItemActive(item) ? 'drop-shadow-[0_0_10px_rgba(56,189,248,0.9)] filter brightness-110 scale-105' : 'opacity-80 group-hover:opacity-100'"
          />
          <component v-else :is="item.icon" class="w-5 h-5 inline-icon text-cyan-200" />
        </template>
        <component v-else :is="item.icon" class="w-5 h-5 inline-icon" />

        <!-- Active Cyan Left Indicator Line -->
        <span 
          v-if="isItemActive(item)" 
          class="absolute left-0 top-2.5 bottom-2.5 w-1 rounded-r-full"
          :class="themeStore.currentPalette === 'vodyanitsa' ? 'bg-cyan-400 shadow-[0_0_10px_#22d3ee]' : 'bg-mim-accent'"
        ></span>

        <!-- Tooltip -->
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap border border-black/5 dark:border-white/10">
          {{ langStore.t(`nav.${item.id}`) }}
        </span>
      </button>
    </nav>

    <!-- Bottom Controls: Grimoire, Language Switcher, Display/Theme & Settings -->
    <div class="flex flex-col gap-2 w-full px-2">
      <!-- Vodyanitsa Grimoire (Notes / AI Logbook) -->
      <button
        v-if="themeStore.currentPalette === 'vodyanitsa'"
        @click="openGrimoire"
        class="w-full h-10 rounded-xl flex items-center justify-center transition-all duration-200 relative group cursor-pointer hover:bg-cyan-900/25"
        title="科洛列夫茨基剧团 · 剧目档案"
      >
        <img :src="iconGrimoire" alt="剧目档案" class="w-6 h-6 object-contain opacity-80 group-hover:opacity-100 group-hover:scale-110 transition-all drop-shadow-[0_0_6px_rgba(192,132,252,0.6)]" />
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-[#091530] text-purple-200 border border-purple-400/30 opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap shadow-xl">
          剧目档案 · 历史记录
        </span>
      </button>

      <!-- Language Toggle Button -->
      <button
        @click="langStore.toggleLang()"
        class="w-full h-10 rounded-xl flex items-center justify-center transition-all duration-200 relative group cursor-pointer"
        :class="themeStore.currentPalette === 'vodyanitsa' ? 'hover:bg-cyan-900/25' : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5'"
        :title="langStore.currentLang === 'zh' ? '切换为 English' : 'Switch to 简体中文'"
      >
        <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconLang" alt="语言" class="w-6 h-6 object-contain opacity-80 group-hover:opacity-100 group-hover:scale-110 transition-all drop-shadow-[0_0_6px_rgba(56,189,248,0.5)]" />
        <Languages v-else class="w-5 h-5 inline-icon text-mim-accent" />
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          {{ langStore.currentLang === 'zh' ? '语言: 简体中文 (点击切 EN)' : 'Language: English (Click for 中文)' }}
        </span>
      </button>

      <!-- Theme / Display Switcher Button -->
      <button
        @click="cycleTheme"
        class="w-full h-10 rounded-xl flex items-center justify-center transition-[background-color,transform] duration-150 active:scale-90 relative group cursor-pointer"
        :class="themeStore.currentPalette === 'vodyanitsa' ? 'hover:bg-cyan-900/25' : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5'"
        title="切换外观主题"
      >
        <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconScreen" alt="外观" class="w-6 h-6 object-contain opacity-80 group-hover:opacity-100 group-hover:scale-110 transition-all drop-shadow-[0_0_6px_rgba(56,189,248,0.5)]" />
        <template v-else>
          <Sun v-if="themeStore.currentMode === 'light'" class="w-5 h-5 inline-icon text-amber-500" />
          <Moon v-else-if="themeStore.currentMode === 'dark'" class="w-5 h-5 inline-icon text-mim-accent" />
          <Monitor v-else class="w-5 h-5 inline-icon text-emerald-500" />
        </template>
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          {{ langStore.t('nav.theme') }}: {{ themeStore.currentMode.toUpperCase() }}
        </span>
      </button>

      <!-- Settings Button -->
      <button
        @click="$emit('open-settings')"
        class="w-full h-10 rounded-xl flex items-center justify-center transition-[background-color,transform] duration-150 active:scale-90 relative group cursor-pointer"
        :class="themeStore.currentPalette === 'vodyanitsa' ? 'hover:bg-cyan-900/25' : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5'"
        :title="langStore.t('nav.settings')"
      >
        <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconGear" alt="设置" class="w-6 h-6 object-contain opacity-80 group-hover:opacity-100 group-hover:rotate-45 group-hover:scale-110 transition-all duration-300 drop-shadow-[0_0_8px_rgba(56,189,248,0.6)]" />
        <Settings v-else class="w-5 h-5 inline-icon" />
        <span class="absolute left-16 px-2.5 py-1 text-xs rounded-md bg-mim-light-card dark:bg-mim-dark-card shadow-card text-mim-light-text dark:text-mim-dark-text opacity-0 pointer-events-none group-hover:opacity-100 transition-opacity z-50 whitespace-nowrap">
          {{ langStore.t('nav.settings') }}
        </span>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessageSquare, Users, Phone, Settings, Sun, Moon, Monitor, Check, Languages, UserPlus, LogOut } from 'lucide-vue-next';
import { useThemeStore } from '../stores/theme';
import { useChatStore, DemoUser } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { getAvatarColor } from '../utils/avatar';
import { playGenshinClick } from '../utils/genshinAudio';
import vodyanitsaAvatar from '../assets/vodyanitsa/images/vodyanitsa_avatar.png';

// Vodyanitsa 100% Newly Generated Haute-Craft Art Assets
import iconWaveWing from '../assets/vodyanitsa/dock/nav_chats_wave_wing.png';
import iconLily from '../assets/vodyanitsa/dock/nav_contacts_lily.png';
import iconConch from '../assets/vodyanitsa/dock/nav_calls_conch.png';
import iconGrimoire from '../assets/vodyanitsa/dock/nav_grimoire.png';
import iconLang from '../assets/vodyanitsa/dock/nav_lang.png';
import iconScreen from '../assets/vodyanitsa/dock/nav_screen.png';
import iconGear from '../assets/vodyanitsa/dock/nav_settings_gear.png';
import avatarDial from '../assets/vodyanitsa/dock/nav_avatar_dial.png';

const route = useRoute();
const router = useRouter();
const themeStore = useThemeStore();
const chatStore = useChatStore();
const langStore = useLangStore();
const showAccountMenu = ref(false);

const emit = defineEmits(['open-settings', 'open-auth']);

const openAuth = () => {
  showAccountMenu.value = false;
  emit('open-auth');
};

const handleLogout = () => {
  showAccountMenu.value = false;
  chatStore.logout();
};

const openGrimoire = () => {
  emit('open-settings');
};

const navItems = [
  { id: 'chats', path: '/chats', icon: MessageSquare },
  { id: 'contacts', path: '/contacts', icon: Users },
  { id: 'calls', path: '/calls', icon: Phone },
];

const isItemActive = (item: { id: string; path: string }) => {
  return route.path.startsWith(item.path);
};

const navigateTo = (path: string) => {
  if (themeStore.currentPalette === 'genshin') {
    playGenshinClick();
  }
  if (path === '/chats') {
    if (chatStore.activeConvId) {
      router.push(`/chats/${chatStore.activeConvId}`);
    } else {
      router.push('/chats');
    }
  } else {
    router.push(path);
  }
};

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

<style scoped>
/* Account Popover Transition (overlays/popover spec) */
.popover-dock-enter-active {
  transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1),
              opacity 160ms cubic-bezier(0.16, 1, 0.3, 1);
}
.popover-dock-leave-active {
  transition: transform 140ms cubic-bezier(0.22, 1, 0.36, 1),
              opacity 120ms cubic-bezier(0.22, 1, 0.36, 1);
}
.popover-dock-enter-from {
  opacity: 0;
  transform: translateX(-6px) scale(0.96);
}
.popover-dock-leave-to {
  opacity: 0;
  transform: translateX(-4px) scale(0.98);
}
</style>
