<template>
  <Teleport to="body">
    <div
      v-if="dialogStore.visible"
      class="fixed inset-0 z-[10000] bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 select-none animate-in fade-in duration-150"
      @click.self="dialogStore.onCancel"
    >
      <div
        class="w-full max-w-sm rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/10 dark:border-white/10 p-6 flex flex-col gap-4 animate-in zoom-in-95 duration-150"
      >
        <!-- Icon & Header -->
        <div class="flex items-start gap-3.5">
          <!-- Type Icon Pill -->
          <div
            class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0 shadow-2xs"
            :class="dialogStore.isDanger || dialogStore.type === 'danger'
              ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border border-rose-500/25'
              : dialogStore.type === 'warning'
                ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/25'
                : dialogStore.type === 'success'
                  ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/25'
                  : 'bg-mim-accent-subtle text-mim-accent border border-mim-accent/25'"
          >
            <AlertTriangle v-if="dialogStore.isDanger || dialogStore.type === 'danger'" class="w-5 h-5 inline-icon" />
            <AlertCircle v-else-if="dialogStore.type === 'warning'" class="w-5 h-5 inline-icon" />
            <CheckCircle2 v-else-if="dialogStore.type === 'success'" class="w-5 h-5 inline-icon" />
            <HelpCircle v-else class="w-5 h-5 inline-icon" />
          </div>

          <div class="flex-1 min-w-0 pt-0.5">
            <h3 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text leading-tight">
              {{ dialogStore.title }}
            </h3>
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted mt-1.5 leading-relaxed break-words whitespace-pre-wrap">
              {{ dialogStore.message }}
            </p>
          </div>
        </div>

        <!-- Buttons -->
        <div class="flex items-center justify-end gap-2.5 pt-2 border-t border-black/5 dark:border-white/5">
          <button
            v-if="dialogStore.isConfirm"
            type="button"
            @click="dialogStore.onCancel"
            class="px-4 h-8 rounded-xl bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium text-mim-light-text dark:text-mim-dark-text transition-colors"
          >
            {{ dialogStore.cancelText }}
          </button>
          <button
            type="button"
            @click="dialogStore.onConfirm"
            class="px-5 h-8 rounded-xl text-xs font-semibold shadow-sm transition-all"
            :class="dialogStore.isDanger || dialogStore.type === 'danger'
              ? 'bg-rose-600 hover:bg-rose-500 text-white shadow-rose-600/20'
              : 'mim-btn-primary'"
          >
            {{ dialogStore.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import {
  AlertTriangle,
  AlertCircle,
  CheckCircle2,
  HelpCircle,
} from 'lucide-vue-next';
import { useDialogStore } from '../stores/dialog';

const dialogStore = useDialogStore();

const handleKeydown = (e: KeyboardEvent) => {
  if (!dialogStore.visible) return;
  if (e.key === 'Escape') {
    e.preventDefault();
    dialogStore.onCancel();
  } else if (e.key === 'Enter') {
    e.preventDefault();
    dialogStore.onConfirm();
  }
};

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown);
});
</script>
