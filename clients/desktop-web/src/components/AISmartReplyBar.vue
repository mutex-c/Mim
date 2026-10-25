<template>
  <div
    v-if="visible"
    class="rounded-xl p-3 bg-white/95 dark:bg-[#1a2332]/95 border border-black/[0.06] dark:border-white/[0.08] shadow-sm flex flex-col gap-2.5 mb-2 backdrop-blur-md transition-all select-none"
  >
    <!-- Top Bar Header -->
    <div class="flex items-center justify-between px-0.5">
      <div class="flex items-center gap-2">
        <div class="w-6 h-6 rounded-lg bg-mim-accent-subtle flex items-center justify-center text-mim-accent">
          <Sparkles class="w-3.5 h-3.5 inline-icon" />
        </div>
        <div class="flex items-center gap-1.5">
          <span class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text tracking-tight">AI 智能帮回</span>
          <span class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted hidden sm:inline">根据上下文推荐 3 种沟通策略</span>
        </div>
      </div>

      <div class="flex items-center gap-1">
        <!-- Regenerate Button -->
        <button
          type="button"
          @click="$emit('regenerate')"
          :disabled="loading"
          class="px-2 py-1 rounded-lg flex items-center gap-1 text-[11px] font-medium text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer disabled:opacity-40"
          title="重新生成建议"
        >
          <RotateCw class="w-3.5 h-3.5 inline-icon" :class="{ 'animate-spin': loading }" />
          <span>换一批</span>
        </button>

        <!-- Close Button -->
        <button
          type="button"
          @click="$emit('close')"
          class="w-6 h-6 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
          title="关闭 (Esc)"
        >
          <X class="w-3.5 h-3.5 inline-icon" />
        </button>
      </div>
    </div>

    <!-- Loading Skeleton State -->
    <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
      <div
        v-for="i in 3"
        :key="i"
        class="flex flex-col gap-2.5 p-3 rounded-xl bg-black/[0.02] dark:bg-white/[0.02] animate-pulse"
      >
        <div class="flex items-center justify-between">
          <div class="h-4 w-16 rounded-full bg-black/5 dark:bg-white/10"></div>
          <div class="h-5 w-5 rounded-md bg-black/5 dark:bg-white/10"></div>
        </div>
        <div class="space-y-1.5 pt-1">
          <div class="h-3.5 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '90%' : (i === 2 ? '95%' : '85%') }"></div>
          <div class="h-3.5 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '70%' : (i === 2 ? '80%' : '65%') }"></div>
        </div>
      </div>
    </div>

    <!-- Suggestion Cards Grid -->
    <div v-else-if="suggestions && suggestions.length > 0" class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
      <div
        v-for="(suggestion, sIdx) in suggestions"
        :key="suggestion.type + '-' + sIdx"
        @click="$emit('select', suggestion.text)"
        :aria-label="suggestion.text"
        class="ai-suggestion-card group relative flex flex-col justify-between p-3 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] hover:bg-black/[0.04] dark:hover:bg-white/[0.06] transition-all duration-150 cursor-pointer"
        title="点击填入输入框"
      >
        <!-- Card Top: Badge & Quick Send -->
        <div class="flex items-center justify-between gap-2 mb-2">
          <span
            class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium shrink-0"
            :class="getBadgeStyle(suggestion.type)"
          >
            <component :is="getBadgeIcon(suggestion.type)" class="w-2.5 h-2.5 inline-icon" />
            <span>{{ suggestion.label || getFallbackLabel(suggestion.type) }}</span>
          </span>

          <button
            type="button"
            @click.stop="$emit('send', suggestion.text)"
            class="w-6 h-6 rounded-md flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-accent hover:bg-black/5 dark:hover:bg-white/10 active:scale-95 transition-all cursor-pointer opacity-70 group-hover:opacity-100"
            title="直接发送"
          >
            <Send class="w-3.5 h-3.5 inline-icon" />
          </button>
        </div>

        <!-- Suggestion Text with 90ms blur-in typewriter effect -->
        <div
          class="text-xs text-mim-light-text dark:text-mim-dark-text leading-relaxed font-normal select-text break-words whitespace-pre-wrap"
          :key="suggestion.text"
        >
          <span
            v-for="(char, cIdx) in suggestion.text"
            :key="cIdx"
            aria-hidden="true"
            class="ai-blur-char"
            :style="{ '--char-index': cIdx }"
          >{{ char === ' ' ? '\u00A0' : char }}</span>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="py-6 text-center text-xs text-mim-light-muted dark:text-mim-dark-muted flex flex-col items-center gap-2">
      <span>暂未获取到回复建议</span>
      <button
        type="button"
        @click="$emit('regenerate')"
        class="text-xs text-mim-accent hover:underline flex items-center gap-1 cursor-pointer font-medium"
      >
        <RotateCw class="w-3 h-3 inline-icon" />
        <span>重新生成建议</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Sparkles,
  RotateCw,
  X,
  Send,
  Check,
  MessageSquare,
} from 'lucide-vue-next';

export interface SmartReplySuggestion {
  type: string;
  label: string;
  text: string;
}

const props = withDefaults(
  defineProps<{
    suggestions?: SmartReplySuggestion[];
    loading?: boolean;
    visible?: boolean;
  }>(),
  {
    suggestions: () => [],
    loading: false,
    visible: true,
  }
);

defineEmits<{
  (e: 'select', text: string): void;
  (e: 'send', text: string): void;
  (e: 'regenerate'): void;
  (e: 'close'): void;
}>();

const getBadgeStyle = (type: string) => {
  switch (type) {
    case 'positive':
      return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 dark:bg-emerald-500/20';
    case 'alternative':
      return 'bg-sky-500/10 text-sky-600 dark:text-sky-400 dark:bg-sky-500/20';
    case 'clarify':
      return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 dark:bg-purple-500/20';
    default:
      return 'bg-black/5 dark:bg-white/10 text-mim-light-muted dark:text-mim-dark-muted';
  }
};

const getBadgeIcon = (type: string) => {
  switch (type) {
    case 'positive':
      return Check;
    case 'alternative':
      return Sparkles;
    case 'clarify':
      return MessageSquare;
    default:
      return Sparkles;
  }
};

const getFallbackLabel = (type: string) => {
  switch (type) {
    case 'positive':
      return '高效确认';
    case 'alternative':
      return '委婉沟通';
    case 'clarify':
      return '追问细节';
    default:
      return '智能回复';
  }
};
</script>

<style scoped>
@keyframes aiBlurFadeIn {
  0% {
    filter: blur(6px);
    opacity: 0;
    transform: translateY(2px);
  }
  100% {
    filter: blur(0);
    opacity: 1;
    transform: translateY(0);
  }
}

.ai-blur-char {
  display: inline-block;
  will-change: filter, transform, opacity;
  transform: translateZ(0);
  animation: aiBlurFadeIn 90ms cubic-bezier(.215, .61, .355, 1) both;
  animation-delay: calc(var(--char-index) * 30ms);
}

.ai-suggestion-card {
  box-shadow: inset 0 0 0 1px transparent;
}

.ai-suggestion-card:hover {
  box-shadow: inset 0 0 0 1px var(--mim-accent, rgba(16, 185, 129, 0.35));
}
</style>
