<template>
  <Transition name="summary-drawer">
    <div
      v-if="visible"
      class="absolute inset-0 z-30 flex justify-end overflow-hidden"
      tabindex="-1"
      @keydown.esc="$emit('close')"
    >
      <!-- Backdrop overlay -->
      <div
        class="absolute inset-0 bg-black/20 dark:bg-black/40 backdrop-blur-[2px] transition-opacity cursor-pointer"
        @click="$emit('close')"
      ></div>

      <!-- Slide-out Drawer Panel -->
      <aside
        class="relative w-full sm:w-[420px] max-w-full h-full bg-mim-light-card dark:bg-mim-dark-card border-l border-black/[0.06] dark:border-white/[0.08] shadow-2xl flex flex-col z-10 select-none overflow-hidden"
      >
        <!-- Top Bar Header -->
        <div class="h-16 px-4 sm:px-5 flex items-center justify-between border-b border-black/[0.04] dark:border-white/[0.04] shrink-0 bg-mim-light-card/80 dark:bg-mim-dark-card/80 backdrop-blur-md">
          <div class="flex items-center gap-3 min-w-0">
            <div class="w-9 h-9 rounded-xl bg-mim-accent-subtle flex items-center justify-center text-mim-accent shrink-0">
              <Sparkles class="w-4 h-4 inline-icon" />
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text truncate leading-tight">
                会话智能纪要
              </h3>
              <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted truncate mt-0.5">
                <template v-if="messageCount > 0 && participants > 0">
                  基于近 {{ messageCount }} 条消息提炼 · {{ participants }} 位成员
                </template>
                <template v-else-if="messageCount > 0">
                  基于近 {{ messageCount }} 条消息提炼
                </template>
                <template v-else>
                  基于当前会话上下文提炼
                </template>
              </p>
            </div>
          </div>

          <div class="flex items-center gap-1 shrink-0">
            <!-- Refresh Button -->
            <button
              type="button"
              @click="$emit('refresh')"
              :disabled="loading"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer disabled:opacity-40 focus:outline-none focus:shadow-[inset_0_0_0_1px_var(--mim-accent)] active:shadow-[inset_0_0_0_1px_var(--mim-accent)]"
              title="重新生成纪要"
            >
              <RotateCw class="w-4 h-4 inline-icon" :class="{ 'animate-spin': loading }" />
            </button>

            <!-- Close Button -->
            <button
              type="button"
              @click="$emit('close')"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer focus:outline-none focus:shadow-[inset_0_0_0_1px_var(--mim-accent)] active:shadow-[inset_0_0_0_1px_var(--mim-accent)]"
              title="关闭 (Esc)"
            >
              <X class="w-4 h-4 inline-icon" />
            </button>
          </div>
        </div>

        <!-- Drawer Body -->
        <div class="flex-1 overflow-y-auto p-4 space-y-3.5">
          <!-- Loading State: Pulsing Skeleton Cards -->
          <template v-if="loading">
            <div
              v-for="i in 3"
              :key="i"
              class="rounded-xl bg-black/[0.02] dark:bg-white/[0.02] p-4 space-y-3 animate-pulse border-l-2"
              :class="i === 1 ? 'border-amber-500/40' : (i === 2 ? 'border-sky-500/40' : 'border-mim-accent/40')"
            >
              <div class="flex items-center gap-2">
                <div class="w-5 h-5 rounded-md bg-black/5 dark:bg-white/10"></div>
                <div class="h-4 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '6.5rem' : (i === 2 ? '8rem' : '6rem') }"></div>
              </div>
              <div class="space-y-2 pt-1 pl-1">
                <div class="h-3 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '92%' : (i === 2 ? '86%' : '78%') }"></div>
                <div class="h-3 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '75%' : (i === 2 ? '94%' : '84%') }"></div>
                <div class="h-3 rounded bg-black/5 dark:bg-white/10" :style="{ width: i === 1 ? '60%' : (i === 2 ? '68%' : '52%') }"></div>
              </div>
            </div>
          </template>

          <!-- Empty State -->
          <template v-else-if="!summary.trim()">
            <div class="h-full min-h-[280px] flex flex-col items-center justify-center text-center p-6 space-y-3">
              <div class="w-12 h-12 rounded-2xl bg-black/5 dark:bg-white/5 flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted">
                <FileText class="w-6 h-6 inline-icon opacity-70" />
              </div>
              <div class="space-y-1">
                <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">暂无可提炼的纪要内容</h4>
                <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted max-w-xs">
                  当前会话中消息记录不足，请在会话中交流几条消息后再试。
                </p>
              </div>
              <button
                type="button"
                @click="$emit('refresh')"
                class="mt-2 px-3 py-1.5 rounded-lg text-xs font-medium text-mim-accent bg-mim-accent-subtle hover:bg-mim-accent-subtle/80 transition-colors cursor-pointer"
              >
                刷新提炼
              </button>
            </div>
          </template>

          <!-- Loaded State: Structured Markdown Sections -->
          <template v-else>
            <div
              v-for="section in parsedSections"
              :key="section.id"
              class="rounded-xl p-4 bg-black/[0.02] dark:bg-white/[0.02] border-l-2 transition-all space-y-2.5"
              :class="getSectionBorderClass(section.type)"
            >
              <!-- Section Header -->
              <div class="flex items-center gap-2">
                <div
                  class="w-5 h-5 rounded-md flex items-center justify-center shrink-0"
                  :class="getSectionIconBgClass(section.type)"
                >
                  <component :is="getSectionIcon(section.type)" class="w-3.5 h-3.5 inline-icon" :class="getSectionIconColorClass(section.type)" />
                </div>
                <h4 class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text tracking-tight uppercase">
                  {{ section.title }}
                </h4>
              </div>

              <!-- Section Content Items -->
              <div class="space-y-2 pl-0.5 text-xs text-mim-light-text dark:text-mim-dark-text leading-relaxed">
                <div
                  v-for="item in section.items"
                  :key="item.id"
                  class="flex items-start gap-2 group"
                >
                  <!-- Checkbox for Task / Action Items -->
                  <div
                    v-if="item.type === 'task'"
                    role="checkbox"
                    :aria-checked="isTaskChecked(item)"
                    tabindex="0"
                    @click="toggleTaskChecked(item)"
                    @keydown.space.prevent="toggleTaskChecked(item)"
                    class="mt-0.5 w-4 h-4 rounded-md flex items-center justify-center shrink-0 cursor-pointer transition-all duration-150 select-none focus:outline-none focus:shadow-[inset_0_0_0_1px_var(--mim-accent)]"
                    :class="isTaskChecked(item)
                      ? 'bg-mim-accent text-white shadow-2xs'
                      : 'border border-black/25 dark:border-white/25 hover:border-mim-accent group-hover:border-mim-accent/70'"
                  >
                    <Check v-if="isTaskChecked(item)" class="w-3 h-3 stroke-[2.5] inline-icon" />
                  </div>

                  <!-- Bullet Dot for Bullet Items -->
                  <span
                    v-else-if="item.type === 'bullet'"
                    class="mt-1.5 w-1.5 h-1.5 rounded-full shrink-0"
                    :class="getSectionBulletDotClass(section.type)"
                  ></span>

                  <!-- Item Text with Safe Inline Markdown -->
                  <div
                    class="flex-1 select-text break-words pretty-wrap"
                    :class="{
                      'line-through text-mim-light-muted dark:text-mim-dark-muted opacity-70': item.type === 'task' && isTaskChecked(item)
                    }"
                    v-html="formatInlineMarkdown(item.text)"
                  ></div>
                </div>
              </div>
            </div>
          </template>
        </div>

        <!-- Bottom Action Bar -->
        <div class="p-4 border-t border-black/[0.04] dark:border-white/[0.04] bg-mim-light-card/80 dark:bg-mim-dark-card/80 backdrop-blur-md flex items-center gap-2.5 shrink-0">
          <!-- Copy Markdown Button -->
          <button
            type="button"
            @click="handleCopy"
            :disabled="loading || !summary.trim()"
            class="flex-1 h-10 rounded-xl px-3 flex items-center justify-center gap-1.5 text-xs font-medium bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-mim-light-text dark:text-mim-dark-text transition-all duration-150 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed focus:outline-none focus:shadow-[inset_0_0_0_1px_var(--mim-accent)] active:shadow-[inset_0_0_0_1px_var(--mim-accent)]"
            :title="isCopied ? '已复制到剪贴板' : '复制为 Markdown 格式'"
          >
            <Check v-if="isCopied" class="w-4 h-4 text-emerald-500 inline-icon" />
            <Copy v-else class="w-4 h-4 inline-icon" />
            <span :class="{ 'text-emerald-500 font-semibold': isCopied }">
              {{ isCopied ? '已复制' : '复制为 Markdown' }}
            </span>
          </button>

          <!-- Send to Chat Button -->
          <button
            type="button"
            @click="handleSend"
            :disabled="loading || !summary.trim()"
            class="flex-1 h-10 rounded-xl px-3 flex items-center justify-center gap-1.5 text-xs font-semibold bg-mim-accent hover:bg-mim-accent-hover text-white transition-all duration-150 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed focus:outline-none focus:shadow-[inset_0_0_0_1px_rgba(255,255,255,0.4)] active:shadow-[inset_0_0_0_1px_rgba(255,255,255,0.4)]"
            title="发送纪要到当前聊天会话"
          >
            <Send class="w-4 h-4 inline-icon" />
            <span>发送到聊天</span>
          </button>
        </div>
      </aside>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue';
import {
  Sparkles,
  FileText,
  Check,
  Copy,
  Send,
  RotateCw,
  X,
  ListTodo,
  Lightbulb,
  Pin,
  Users
} from 'lucide-vue-next';

interface Props {
  visible: boolean;
  loading: boolean;
  summary: string;
  messageCount?: number;
  participants?: number;
  conversationTitle?: string;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  loading: false,
  summary: '',
  messageCount: 0,
  participants: 0,
  conversationTitle: '',
});

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'refresh'): void;
  (e: 'copy'): void;
  (e: 'send', summaryText: string): void;
}>();

// Copy state
const isCopied = ref(false);
let copyTimer: number | null = null;

const handleCopy = async () => {
  if (!props.summary) return;
  try {
    await navigator.clipboard.writeText(props.summary);
    isCopied.value = true;
    if (copyTimer) clearTimeout(copyTimer);
    copyTimer = window.setTimeout(() => {
      isCopied.value = false;
    }, 2000);
    emit('copy');
  } catch (err) {
    console.error('Failed to copy summary to clipboard:', err);
  }
};

const handleSend = () => {
  if (!props.summary) return;
  emit('send', props.summary);
};

// Interactive task item state
const taskCheckedState = reactive<Record<string, boolean>>({});

const isTaskChecked = (item: SummaryItem): boolean => {
  if (taskCheckedState[item.id] !== undefined) {
    return taskCheckedState[item.id];
  }
  return item.checked;
};

const toggleTaskChecked = (item: SummaryItem) => {
  const current = isTaskChecked(item);
  taskCheckedState[item.id] = !current;
};

// Section & Item structures
export interface SummaryItem {
  id: string;
  type: 'task' | 'bullet' | 'text';
  text: string;
  checked: boolean;
}

export type SectionType = 'topic' | 'consensus' | 'action' | 'participants' | 'general';

export interface SummarySection {
  id: string;
  title: string;
  type: SectionType;
  items: SummaryItem[];
}

// Helper: Strip emojis for strictly emoji-free UI text
const stripEmojis = (str: string): string => {
  return str
    .replace(/[\u{1F300}-\u{1F9FF}\u{2600}-\u{27BF}\u{1F600}-\u{1F64F}\u{1F680}-\u{1F6FF}\u{1F1E0}-\u{1F1FF}\u{FE00}-\u{FE0F}\u{1FA00}-\u{1FAFF}]/gu, '')
    .trim();
};

// Helper: Sanitize and format bold / inline code
const formatInlineMarkdown = (text: string): string => {
  if (!text) return '';
  let escaped = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');

  // Bold **bold text**
  escaped = escaped.replace(/\*\*(.+?)\*\*/g, '<strong class="font-semibold text-mim-light-text dark:text-mim-dark-text">$1</strong>');

  // Inline `code`
  escaped = escaped.replace(/`([^`]+)`/g, '<code class="px-1 py-0.5 rounded bg-black/5 dark:bg-white/10 font-mono text-[11px]">$1</code>');

  return escaped;
};

// Helper: Detect section type based on title keywords
const detectSectionType = (title: string): SectionType => {
  const lower = title.toLowerCase();
  if (/议题|讨论|脉络|timeline|highlight|topic|交流/.test(lower)) {
    return 'topic';
  }
  if (/共识|决策|结论|takeaway|synthesis|consensus|decision|萃取/.test(lower)) {
    return 'consensus';
  }
  if (/待办|行动|清单|action|todo|task/.test(lower)) {
    return 'action';
  }
  if (/参与|成员|人员|contributor|participant/.test(lower)) {
    return 'participants';
  }
  return 'general';
};

// Structured Section Parser
const parsedSections = computed<SummarySection[]>(() => {
  const raw = props.summary.trim();
  if (!raw) return [];

  const lines = raw.split(/\r?\n/);
  const sections: SummarySection[] = [];
  let currentSection: SummarySection | null = null;
  let sectionIndex = 0;
  let itemIndex = 0;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;

    // Check if line is a section header:
    // 1. Markdown heading (# ... or ## ... or ### ...)
    // 2. Numbered list with bold title: 1. **核心议题** or 1. 📌 **核心议题**
    // 3. Bullet list with bold title: • **核心交流脉络提炼**： or • **发言参与人员**：
    // 4. Standalone bold header: **核心议题**
    let headerTitle = '';
    const headingMatch = line.match(/^#{1,4}\s*(.+)$/);
    const numberedMatch = line.match(/^(?:\d+[\.、]|\([0-9]+\))\s*(?:[^\w\s]\s*)?\*\*?([^*]+)\*\*?(.*)$/);
    const bulletHeaderMatch = line.match(/^[•\-\*]\s*\*\*?([^*]+)\*\*?(?::|：)?(.*)$/);

    if (headingMatch) {
      headerTitle = stripEmojis(headingMatch[1]);
    } else if (numberedMatch) {
      const main = numberedMatch[1].trim();
      const rest = numberedMatch[2] ? numberedMatch[2].replace(/^[（(][^）)]*[）)]/, '').trim() : '';
      headerTitle = stripEmojis(main + (rest ? ' ' + rest : ''));
    } else if (bulletHeaderMatch && (bulletHeaderMatch[1].length < 30 || bulletHeaderMatch[2])) {
      headerTitle = stripEmojis(bulletHeaderMatch[1]);
    }

    // Clean up header title
    headerTitle = headerTitle.replace(/^[:：\s]+|[:：\s]+$/g, '').trim();

    // If title is recognized as a new section or first line of doc
    if (headerTitle && (detectSectionType(headerTitle) !== 'general' || headingMatch || numberedMatch)) {
      sectionIndex++;
      currentSection = {
        id: `sec-${sectionIndex}-${headerTitle.substring(0, 8)}`,
        title: headerTitle,
        type: detectSectionType(headerTitle),
        items: [],
      };
      sections.push(currentSection);

      // If bullet header had inline content after title, e.g. "• **发言参与人员**：Alex, Sarah"
      if (bulletHeaderMatch && bulletHeaderMatch[2] && bulletHeaderMatch[2].replace(/^[:：\s]+/, '').trim()) {
        const trailingText = stripEmojis(bulletHeaderMatch[2].replace(/^[:：\s]+/, '').trim());
        if (trailingText) {
          itemIndex++;
          currentSection.items.push({
            id: `item-${sectionIndex}-${itemIndex}`,
            type: 'text',
            text: trailingText,
            checked: false,
          });
        }
      }
      continue;
    }

    // If line is not a header, append to currentSection or create default section
    if (!currentSection) {
      sectionIndex++;
      currentSection = {
        id: `sec-${sectionIndex}-overview`,
        title: '会话提炼与要点',
        type: 'general',
        items: [],
      };
      sections.push(currentSection);
    }

    // Line item analysis
    const cleanLine = stripEmojis(line);
    if (!cleanLine) continue;

    itemIndex++;
    // Case A: Markdown task item: - [ ] Task or - [x] Task
    const taskMatch = cleanLine.match(/^[-*•]\s*\[([ xX])\]\s*(.*)$/);
    if (taskMatch) {
      currentSection.items.push({
        id: `item-${sectionIndex}-${itemIndex}`,
        type: 'task',
        text: taskMatch[2].trim(),
        checked: taskMatch[1].toLowerCase() === 'x',
      });
      continue;
    }

    // Case B: In action items section, bullet points or numbered lists default to interactive tasks
    if (currentSection.type === 'action') {
      const actionBulletMatch = cleanLine.match(/^(?:[-*•]|\d+[\.、])\s*(.*)$/);
      if (actionBulletMatch) {
        currentSection.items.push({
          id: `item-${sectionIndex}-${itemIndex}`,
          type: 'task',
          text: actionBulletMatch[1].trim(),
          checked: false,
        });
        continue;
      }
    }

    // Case C: Normal bullet item: - item or • item
    const bulletMatch = cleanLine.match(/^[-*•]\s*(.*)$/);
    if (bulletMatch) {
      currentSection.items.push({
        id: `item-${sectionIndex}-${itemIndex}`,
        type: 'bullet',
        text: bulletMatch[1].trim(),
        checked: false,
      });
      continue;
    }

    // Case D: Plain text paragraph
    currentSection.items.push({
      id: `item-${sectionIndex}-${itemIndex}`,
      type: 'text',
      text: cleanLine,
      checked: false,
    });
  }

  return sections;
});

// Section Style Helpers
const getSectionBorderClass = (type: SectionType): string => {
  switch (type) {
    case 'topic':
      return 'border-l-amber-500';
    case 'consensus':
      return 'border-l-sky-500';
    case 'action':
      return 'border-l-mim-accent';
    case 'participants':
      return 'border-l-indigo-500';
    case 'general':
    default:
      return 'border-l-mim-accent/70';
  }
};

const getSectionIcon = (type: SectionType) => {
  switch (type) {
    case 'topic':
      return Pin;
    case 'consensus':
      return Lightbulb;
    case 'action':
      return ListTodo;
    case 'participants':
      return Users;
    case 'general':
    default:
      return FileText;
  }
};

const getSectionIconBgClass = (type: SectionType): string => {
  switch (type) {
    case 'topic':
      return 'bg-amber-500/10 dark:bg-amber-500/20';
    case 'consensus':
      return 'bg-sky-500/10 dark:bg-sky-500/20';
    case 'action':
      return 'bg-mim-accent-subtle';
    case 'participants':
      return 'bg-indigo-500/10 dark:bg-indigo-500/20';
    case 'general':
    default:
      return 'bg-mim-accent-subtle';
  }
};

const getSectionIconColorClass = (type: SectionType): string => {
  switch (type) {
    case 'topic':
      return 'text-amber-500';
    case 'consensus':
      return 'text-sky-500';
    case 'action':
      return 'text-mim-accent';
    case 'participants':
      return 'text-indigo-500';
    case 'general':
    default:
      return 'text-mim-accent';
  }
};

const getSectionBulletDotClass = (type: SectionType): string => {
  switch (type) {
    case 'topic':
      return 'bg-amber-500';
    case 'consensus':
      return 'bg-sky-500';
    case 'action':
      return 'bg-mim-accent';
    case 'participants':
      return 'bg-indigo-500';
    case 'general':
    default:
      return 'bg-mim-accent';
  }
};
</script>

<style scoped>
.summary-drawer-enter-active,
.summary-drawer-leave-active {
  transition: opacity 0.25s ease;
}

.summary-drawer-enter-active aside,
.summary-drawer-leave-active aside {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.summary-drawer-enter-from,
.summary-drawer-leave-to {
  opacity: 0;
}

.summary-drawer-enter-from aside,
.summary-drawer-leave-to aside {
  transform: translateX(100%);
}
</style>
