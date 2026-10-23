<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-lg rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col">
      <!-- Modal Header -->
      <div class="p-5 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <Settings class="w-5 h-5 text-blue-600 dark:text-blue-400 inline-icon" />
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('settings.title') }}</h3>
        </div>
        <button @click="$emit('close')" class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors">
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Settings Content -->
      <div class="p-6 space-y-5 overflow-y-auto max-h-[70vh]">
        <!-- Item 1: Protocol & Transport Architecture -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.transportTitle') }}</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono font-medium">
              24-Byte Binary Frame (Big-Endian)
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            {{ langStore.t('settings.transportDesc') }}
          </p>
        </div>

        <!-- Item 2: Security & DFA Filter -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.dfaTitle') }}</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-blue-500/10 text-blue-600 dark:text-blue-400 font-mono font-medium">
              DFA Active
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            {{ langStore.t('settings.dfaDesc') }}
          </p>
        </div>

        <!-- Item 3: File Instant Upload (秒传) -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.dedupTitle') }}</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-purple-500/10 text-purple-600 dark:text-purple-400 font-mono font-medium">
              SHA-256 Active
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            {{ langStore.t('settings.dedupDesc') }}
          </p>
        </div>

        <!-- Item 4: WebRTC Signaling -->
        <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.webrtcTitle') }}</span>
            <span class="text-[10px] px-2 py-0.5 rounded bg-amber-500/10 text-amber-600 dark:text-amber-400 font-mono font-medium">
              Memory Only
            </span>
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed">
            {{ langStore.t('settings.webrtcDesc') }}
          </p>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="p-4 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.02] flex justify-end">
        <button
          @click="$emit('close')"
          class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors"
        >
          {{ langStore.t('settings.done') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Settings, X } from 'lucide-vue-next';
import { useLangStore } from '../stores/lang';

const langStore = useLangStore();
defineEmits(['close']);
</script>
