<template>
  <div
    class="fixed inset-0 z-[80] bg-black/50 backdrop-blur-sm flex items-center justify-center p-4 select-none animate-in fade-in duration-150"
    @click.self="$emit('close')"
  >
    <div
      class="w-full max-w-lg rounded-2xl bg-mim-light-card dark:bg-mim-dark-card border border-black/10 dark:border-white/10 shadow-2xl p-6 flex flex-col space-y-5"
    >
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-xl bg-mim-accent text-white flex items-center justify-center shadow-sm">
            <Keyboard class="w-4 h-4 inline-icon" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">
              {{ langStore.currentLang === 'zh' ? '桌面端快捷键指南' : 'Desktop Shortcuts Guide' }}
            </h3>
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted">
              {{ langStore.currentLang === 'zh' ? '熟练使用快捷键让沟通效率倍增' : 'Boost your communication speed with hotkeys' }}
            </p>
          </div>
        </div>
        <button
          @click="$emit('close')"
          class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text dark:hover:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Shortcuts Grid -->
      <div class="space-y-4 max-h-[420px] overflow-y-auto pr-1 text-xs">
        <!-- Group 1: Navigation -->
        <div>
          <div class="text-[11px] font-semibold text-mim-accent uppercase tracking-wider mb-2">
            {{ langStore.currentLang === 'zh' ? '会话与导航' : 'Navigation & Search' }}
          </div>
          <div class="space-y-1.5">
            <div
              v-for="s in navShortcuts"
              :key="s.name"
              class="flex items-center justify-between py-1 px-2 rounded-lg bg-black/[0.02] dark:bg-white/[0.03]"
            >
              <span class="text-mim-light-text dark:text-mim-dark-text">{{ s.name }}</span>
              <div class="flex items-center gap-1">
                <kbd
                  v-for="k in s.keys"
                  :key="k"
                  class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10 font-mono text-[10px] text-mim-light-muted dark:text-mim-dark-muted border border-black/5 dark:border-white/10 shadow-2xs"
                >
                  {{ k }}
                </kbd>
              </div>
            </div>
          </div>
        </div>

        <!-- Group 2: Chat Input -->
        <div>
          <div class="text-[11px] font-semibold text-mim-accent uppercase tracking-wider mb-2">
            {{ langStore.currentLang === 'zh' ? '消息与输入' : 'Messaging & Input' }}
          </div>
          <div class="space-y-1.5">
            <div
              v-for="s in msgShortcuts"
              :key="s.name"
              class="flex items-center justify-between py-1 px-2 rounded-lg bg-black/[0.02] dark:bg-white/[0.03]"
            >
              <span class="text-mim-light-text dark:text-mim-dark-text">{{ s.name }}</span>
              <div class="flex items-center gap-1">
                <kbd
                  v-for="k in s.keys"
                  :key="k"
                  class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10 font-mono text-[10px] text-mim-light-muted dark:text-mim-dark-muted border border-black/5 dark:border-white/10 shadow-2xs"
                >
                  {{ k }}
                </kbd>
              </div>
            </div>
          </div>
        </div>

        <!-- Group 3: WebRTC Call -->
        <div>
          <div class="text-[11px] font-semibold text-mim-accent uppercase tracking-wider mb-2">
            {{ langStore.currentLang === 'zh' ? '音视频通话' : 'Calls & Audio' }}
          </div>
          <div class="space-y-1.5">
            <div
              v-for="s in callShortcuts"
              :key="s.name"
              class="flex items-center justify-between py-1 px-2 rounded-lg bg-black/[0.02] dark:bg-white/[0.03]"
            >
              <span class="text-mim-light-text dark:text-mim-dark-text">{{ s.name }}</span>
              <div class="flex items-center gap-1">
                <kbd
                  v-for="k in s.keys"
                  :key="k"
                  class="px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/10 font-mono text-[10px] text-mim-light-muted dark:text-mim-dark-muted border border-black/5 dark:border-white/10 shadow-2xs"
                >
                  {{ k }}
                </kbd>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="pt-2 border-t border-black/[0.04] dark:border-white/[0.04] flex items-center justify-between text-[11px] text-mim-light-muted dark:text-mim-dark-muted">
        <span>Windows 原生快捷键生态兼容</span>
        <button
          @click="$emit('close')"
          class="px-3 py-1 rounded-lg mim-btn-primary text-xs font-medium cursor-pointer"
        >
          知道了
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { Keyboard, X } from 'lucide-vue-next';
import { useLangStore } from '../stores/lang';

defineEmits(['close']);
const langStore = useLangStore();

const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform);
const modKey = isMac ? '⌘' : 'Ctrl';

const navShortcuts = computed(() => [
  { name: langStore.currentLang === 'zh' ? '全局指令与搜索' : 'Command Palette / Search', keys: [modKey, 'K'] },
  { name: langStore.currentLang === 'zh' ? '关闭当前弹窗 / 引用' : 'Close Modal / Cancel Reply', keys: ['Esc'] },
  { name: langStore.currentLang === 'zh' ? '呼出快捷键帮助' : 'Shortcuts Help', keys: [modKey, '/'] },
]);

const msgShortcuts = computed(() => [
  { name: langStore.currentLang === 'zh' ? '发送消息' : 'Send Message', keys: ['Enter'] },
  { name: langStore.currentLang === 'zh' ? '输入框换行' : 'Insert Newline', keys: ['Shift', 'Enter'] },
  { name: langStore.currentLang === 'zh' ? '搜索当前会话消息' : 'Search in Conversation', keys: [modKey, 'F'] },
]);

const callShortcuts = computed(() => [
  { name: langStore.currentLang === 'zh' ? '静音 / 取消静音麦克风' : 'Mute / Unmute Microphone', keys: [modKey, 'Shift', 'M'] },
  { name: langStore.currentLang === 'zh' ? '开启 / 关闭摄像头' : 'Toggle Camera Video', keys: [modKey, 'Shift', 'V'] },
  { name: langStore.currentLang === 'zh' ? '挂断通话' : 'Hang up Active Call', keys: [modKey, 'Shift', 'H'] },
]);
</script>
