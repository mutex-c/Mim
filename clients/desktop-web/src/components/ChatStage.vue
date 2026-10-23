<template>
  <main class="flex-1 h-full flex flex-col bg-mim-light-stage dark:bg-mim-dark-stage relative select-none">
    <!-- Top Stage Header -->
    <header class="h-16 px-6 flex items-center justify-between bg-mim-light-stage/80 dark:bg-mim-dark-stage/80 backdrop-blur-md z-10 shrink-0">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-blue-600/15 text-blue-600 dark:text-blue-400 font-semibold flex items-center justify-center text-sm">
          {{ chatStore.activeConversation?.avatarText || 'MI' }}
        </div>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">
              {{ chatStore.activeConversation?.title }}
            </h2>
            <ShieldCheck class="w-3.5 h-3.5 text-blue-500 inline-icon" />
          </div>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted flex items-center gap-1.5">
            <span 
              v-if="chatStore.activeConversation?.isOnline" 
              class="w-1.5 h-1.5 rounded-full bg-emerald-500 inline-block"
            ></span>
            <span>{{ chatStore.activeConversation?.isOnline ? 'Online (E2EE Session)' : 'Encrypted Session' }}</span>
          </p>
        </div>
      </div>

      <!-- Action Buttons: Audio Call, Video Call, Details (No Emoji) -->
      <div class="flex items-center gap-1">
        <button
          @click="initiateCall('audio')"
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          title="Audio Call"
        >
          <Phone class="w-4 h-4 inline-icon" />
        </button>
        <button
          @click="initiateCall('video')"
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          title="Video Call"
        >
          <Video class="w-4 h-4 inline-icon" />
        </button>
        <button
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          title="Conversation Info"
        >
          <MoreHorizontal class="w-4 h-4 inline-icon" />
        </button>
      </div>
    </header>

    <!-- Message Stream -->
    <div ref="scrollContainer" class="flex-1 overflow-y-auto px-6 py-4 space-y-4">
      <div
        v-for="msg in chatStore.activeMessages"
        :key="msg.id"
        class="flex flex-col group relative"
        :class="msg.isMe ? 'items-end' : 'items-start'"
      >
        <!-- Quoted Message Bubble (if any) -->
        <div
          v-if="msg.quote"
          class="max-w-[70%] mb-1 px-3 py-1.5 rounded-lg text-xs bg-black/[0.03] dark:bg-white/[0.04] text-mim-light-muted dark:text-mim-dark-muted border-l-2 border-blue-500 cursor-pointer hover:opacity-80 transition-opacity"
        >
          <span class="font-medium text-blue-600 dark:text-blue-400 block mb-0.5">{{ msg.quote.userName }}</span>
          <span class="truncate block">{{ msg.quote.content }}</span>
        </div>

        <!-- Message Body Bubble -->
        <div class="relative flex items-center gap-2 max-w-[75%]">
          <!-- Quick Action Context Menu (appears on hover) -->
          <div
            class="opacity-0 group-hover:opacity-100 transition-opacity flex items-center gap-1 bg-mim-light-card dark:bg-mim-dark-card shadow-card rounded-lg px-1.5 py-1 z-10 shrink-0"
            :class="msg.isMe ? 'order-first' : 'order-last'"
          >
            <!-- Rule 10: Feedback Self-Extinguishing Copy Checkmark (2.5s) -->
            <button
              @click="copyMessage(msg.id, msg.content)"
              class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted transition-colors relative"
              title="Copy"
            >
              <Check v-if="copiedMsgId === msg.id" class="w-3.5 h-3.5 text-emerald-500 inline-icon" />
              <Copy v-else class="w-3.5 h-3.5 inline-icon" />
            </button>
            <button
              @click="setQuote(msg)"
              class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted transition-colors"
              title="Reply / Quote"
            >
              <Reply class="w-3.5 h-3.5 inline-icon" />
            </button>
            <button
              v-if="msg.isMe && !msg.isRevoked"
              @click="chatStore.revokeMessage(msg.id)"
              class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-red-500 hover:text-red-600 transition-colors"
              title="Recall Message"
            >
              <Undo2 class="w-3.5 h-3.5 inline-icon" />
            </button>
          </div>

          <!-- Bubble Main -->
          <div
            class="px-4 py-2.5 rounded-2xl transition-all shadow-sm"
            :class="[
              msg.isRevoked
                ? 'bg-black/[0.04] dark:bg-white/[0.04] text-mim-light-muted dark:text-mim-dark-muted italic text-xs'
                : msg.isMe
                  ? 'bg-blue-600 text-white rounded-br-sm'
                  : 'bg-mim-light-bubble dark:bg-mim-dark-bubble text-mim-light-text dark:text-mim-dark-text rounded-bl-sm'
            ]"
          >
            <!-- Text Message (Rule 8: text-wrap: pretty) -->
            <p v-if="msg.msgType === 'text'" class="text-sm leading-relaxed pretty-wrap">
              {{ msg.content }}
            </p>

            <!-- Voice Note Message (Rule 6: tabular-nums for duration) -->
            <div v-else-if="msg.msgType === 'voice'" class="flex items-center gap-3 py-1">
              <button class="w-8 h-8 rounded-full bg-black/10 dark:bg-white/10 flex items-center justify-center hover:scale-105 transition-transform">
                <Play class="w-4 h-4 inline-icon ml-0.5" />
              </button>
              <div class="flex items-center gap-1 h-5">
                <span class="w-0.5 h-3 bg-current opacity-40 rounded-full"></span>
                <span class="w-0.5 h-5 bg-current opacity-80 rounded-full"></span>
                <span class="w-0.5 h-2 bg-current opacity-40 rounded-full"></span>
                <span class="w-0.5 h-4 bg-current opacity-90 rounded-full"></span>
                <span class="w-0.5 h-5 bg-current opacity-70 rounded-full"></span>
                <span class="w-0.5 h-3 bg-current opacity-50 rounded-full"></span>
              </div>
              <span class="text-xs font-mono tabular-nums opacity-90">
                {{ msg.duration }}s
              </span>
            </div>

            <!-- Revoke Message -->
            <span v-else-if="msg.isRevoked">
              {{ msg.content }}
            </span>
          </div>
        </div>

        <!-- Timestamp & Delivery Status (Rule 6: tabular-nums) -->
        <div
          class="flex items-center gap-1 mt-1 text-[10px] text-mim-light-muted dark:text-mim-dark-muted tabular-nums"
          :class="msg.isMe ? 'mr-1' : 'ml-1'"
        >
          <span>{{ msg.time }}</span>
          <span v-if="msg.isMe">
            <Clock v-if="msg.status === 'sending'" class="w-3 h-3 inline-icon" />
            <Check v-else-if="msg.status === 'sent'" class="w-3 h-3 inline-icon" />
            <CheckCheck v-else-if="msg.status === 'delivered' || msg.status === 'read'" class="w-3 h-3 inline-icon text-blue-400" />
          </span>
        </div>
      </div>
    </div>

    <!-- Active Quote Bar (Rule 7: Grid 0fr -> 1fr accordion animation) -->
    <div class="mim-accordion px-6" :class="{ open: !!activeQuote }">
      <div class="mim-accordion-inner">
        <div v-if="activeQuote || displayedQuote" class="pt-2">
          <div class="flex items-center justify-between p-2.5 rounded-lg bg-black/[0.03] dark:bg-white/[0.04] border-l-2 border-blue-500">
            <div class="text-xs truncate pr-4">
              <span class="font-medium text-blue-600 dark:text-blue-400">Replying to {{ (activeQuote || displayedQuote)?.userName }}:</span>
              <span class="text-mim-light-muted dark:text-mim-dark-muted ml-1">{{ (activeQuote || displayedQuote)?.content }}</span>
            </div>
            <button @click="clearQuote" class="text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text">
              <X class="w-3.5 h-3.5 inline-icon" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Input Bar (Rule 2: Bite Feel Focus Shadow) -->
    <div class="p-4 pt-2">
      <div class="rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-card p-3 flex flex-col gap-2">
        <!-- Message Textarea -->
        <textarea
          v-model="inputContent"
          @keydown.enter.exact.prevent="handleSend"
          rows="2"
          placeholder="Write a message... (Press Enter to send, Shift+Enter for newline)"
          class="mim-input w-full p-2 bg-transparent text-sm text-mim-light-text dark:text-mim-dark-text placeholder-mim-light-muted dark:placeholder-mim-dark-muted resize-none border-none outline-none pretty-wrap"
        ></textarea>

        <!-- Bottom Controls -->
        <div class="flex items-center justify-between pt-1">
          <!-- Multimedia & AI Action Buttons (Zero Emoji) -->
          <div class="flex items-center gap-1">
            <button class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors" title="Attach Image">
              <Image class="w-4 h-4 inline-icon" />
            </button>
            <button class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors" title="Attach Document">
              <Paperclip class="w-4 h-4 inline-icon" />
            </button>
            <button class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors" title="Audio Recording">
              <Mic class="w-4 h-4 inline-icon" />
            </button>
            <button @click="triggerAISummary" class="px-2 h-8 rounded-lg flex items-center gap-1 text-xs text-blue-600 dark:text-blue-400 hover:bg-blue-600/10 dark:hover:bg-blue-500/15 transition-colors font-medium" title="AI Chat Summary">
              <Sparkles class="w-3.5 h-3.5 inline-icon" />
              <span>AI Summary</span>
            </button>
          </div>

          <!-- Send Action Button -->
          <button
            @click="handleSend"
            :disabled="!inputContent.trim()"
            class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-40 disabled:hover:bg-blue-600 text-white text-xs font-medium flex items-center gap-1.5 transition-all shadow-sm cursor-pointer disabled:cursor-not-allowed"
          >
            <span>Send</span>
            <Send class="w-3 h-3 inline-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- Active WebRTC Call Overlay (if call is active) -->
    <div
      v-if="activeCallType"
      class="absolute inset-x-6 top-6 bottom-24 rounded-2xl bg-black/85 backdrop-blur-xl border border-white/10 flex flex-col items-center justify-between p-8 text-white z-50 shadow-2xl"
    >
      <div class="flex flex-col items-center gap-3">
        <div class="w-20 h-20 rounded-2xl bg-gradient-to-tr from-blue-500 to-indigo-600 flex items-center justify-center text-2xl font-bold">
          {{ chatStore.activeConversation?.avatarText }}
        </div>
        <h3 class="text-lg font-semibold">{{ chatStore.activeConversation?.title }}</h3>
        <!-- Rule 6: Monospace Call Duration -->
        <span class="px-3 py-1 rounded-full bg-white/10 text-xs font-mono tabular-nums">
          {{ callTimer }}
        </span>
      </div>

      <div class="flex items-center gap-6">
        <button
          @click="isMuted = !isMuted"
          class="w-12 h-12 rounded-full flex items-center justify-center transition-colors"
          :class="isMuted ? 'bg-amber-600 text-white' : 'bg-white/10 text-white hover:bg-white/20'"
        >
          <MicOff v-if="isMuted" class="w-5 h-5 inline-icon" />
          <Mic v-else class="w-5 h-5 inline-icon" />
        </button>
        <button
          @click="endCall"
          class="w-14 h-14 rounded-full bg-red-600 hover:bg-red-500 flex items-center justify-center text-white transition-colors shadow-lg shadow-red-600/30"
        >
          <PhoneOff class="w-6 h-6 inline-icon" />
        </button>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue';
import {
  Phone,
  Video,
  MoreHorizontal,
  ShieldCheck,
  Check,
  CheckCheck,
  Clock,
  Copy,
  Reply,
  Undo2,
  Play,
  X,
  Image,
  Paperclip,
  Mic,
  MicOff,
  Sparkles,
  Send,
  PhoneOff,
} from 'lucide-vue-next';
import { useChatStore, MessageItem } from '../stores/chat';

const chatStore = useChatStore();
const inputContent = ref('');
const activeQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const displayedQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const copiedMsgId = ref<string | null>(null);
const scrollContainer = ref<HTMLDivElement | null>(null);

// Call state
const activeCallType = ref<'audio' | 'video' | null>(null);
const callTimer = ref('00:00:12');
const isMuted = ref(false);

const clearQuote = () => {
  activeQuote.value = null;
  setTimeout(() => {
    displayedQuote.value = null;
  }, 250);
};

const handleSend = () => {
  if (!inputContent.value.trim()) return;

  chatStore.sendMessage(inputContent.value, activeQuote.value || undefined);
  inputContent.value = '';
  clearQuote();

  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight;
    }
  });
};

// Rule 10: Feedback Self-Extinguishing Copy Checkmark (2.5s)
const copyMessage = (msgId: string, content: string) => {
  navigator.clipboard.writeText(content);
  copiedMsgId.value = msgId;
  setTimeout(() => {
    if (copiedMsgId.value === msgId) {
      copiedMsgId.value = null;
    }
  }, 2500);
};

const setQuote = (msg: MessageItem) => {
  const quoteData = {
    msgId: msg.id,
    userName: msg.senderName,
    content: msg.content,
  };
  displayedQuote.value = quoteData;
  activeQuote.value = quoteData;
};

const initiateCall = (type: 'audio' | 'video') => {
  activeCallType.value = type;
};

const endCall = () => {
  activeCallType.value = null;
};

const triggerAISummary = () => {
  chatStore.sendMessage('AI Copilot: Summarizing the recent 5 key consensus items on the distributed SeqId engine...', undefined);
};
</script>
