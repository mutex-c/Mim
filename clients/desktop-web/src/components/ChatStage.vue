<template>
  <main
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="handleFileDrop"
    class="flex-1 h-full flex flex-col bg-mim-light-stage dark:bg-mim-dark-stage relative select-none"
  >
    <!-- Hidden File Inputs -->
    <input type="file" ref="imageInputRef" @change="handleImageSelected" accept="image/*" class="hidden" />
    <input type="file" ref="fileInputRef" @change="handleFileSelected" class="hidden" />

    <!-- Drag & Drop Overlay -->
    <div
      v-if="isDragging"
      class="absolute inset-4 rounded-2xl border-2 border-dashed border-blue-500 bg-blue-500/10 backdrop-blur-sm z-40 flex flex-col items-center justify-center pointer-events-none text-blue-600 dark:text-blue-400 gap-2"
    >
      <UploadCloud class="w-10 h-10 animate-bounce inline-icon" />
      <span class="text-sm font-semibold">{{ langStore.t('chat.dropOverlay') }}</span>
    </div>

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
            <span>{{ chatStore.activeConversation?.isOnline ? langStore.t('chat.onlineE2EE') : langStore.t('chat.clusterEncrypted') }}</span>
          </p>
        </div>
      </div>

      <!-- Action Buttons: Audio Call, Video Call, Details (No Emoji) -->
      <div class="flex items-center gap-1">
        <button
          @click="startCall('audio')"
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          :title="langStore.t('chat.audioCall')"
        >
          <Phone class="w-4 h-4 inline-icon" />
        </button>
        <button
          @click="startCall('video')"
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          :title="langStore.t('chat.videoCall')"
        >
          <Video class="w-4 h-4 inline-icon" />
        </button>
        <button
          @click="showConvInfo = !showConvInfo"
          class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
          :title="langStore.t('chat.convInfo')"
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
              :title="langStore.t('chat.copy')"
            >
              <Check v-if="copiedMsgId === msg.id" class="w-3.5 h-3.5 text-emerald-500 inline-icon" />
              <Copy v-else class="w-3.5 h-3.5 inline-icon" />
            </button>
            <button
              @click="setQuote(msg)"
              class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted transition-colors"
              :title="langStore.t('chat.reply')"
            >
              <Reply class="w-3.5 h-3.5 inline-icon" />
            </button>
            <button
              v-if="msg.isMe && !msg.isRevoked"
              @click="chatStore.revokeMessage(msg.id)"
              class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-red-500 hover:text-red-600 transition-colors"
              :title="langStore.t('chat.recall')"
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

            <!-- Image Message -->
            <div v-else-if="msg.msgType === 'image'" class="py-1">
              <img
                :src="msg.content"
                alt="Image"
                class="max-w-xs max-h-60 rounded-xl object-cover cursor-pointer hover:opacity-95 shadow-sm transition-opacity"
                @click="openLightbox(msg.content)"
              />
            </div>

            <!-- File / Document Message -->
            <div v-else-if="msg.msgType === 'file'" class="flex items-center gap-3 py-1 min-w-[220px]">
              <div class="w-10 h-10 rounded-xl bg-black/10 dark:bg-white/10 flex items-center justify-center shrink-0">
                <FileText class="w-5 h-5 inline-icon" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="text-xs font-semibold truncate">{{ msg.fileName || 'Document' }}</div>
                <div class="text-[10px] opacity-75 font-mono tabular-nums">{{ msg.fileSize || 'Instant Upload' }}</div>
              </div>
              <a
                :href="msg.content"
                target="_blank"
                download
                class="w-7 h-7 rounded-lg flex items-center justify-center bg-black/10 dark:bg-white/10 hover:bg-black/20 dark:hover:bg-white/20 transition-colors shrink-0"
                title="Download"
              >
                <Download class="w-3.5 h-3.5 inline-icon" />
              </a>
            </div>

            <!-- Voice Note Message (Rule 6: tabular-nums for duration) -->
            <div v-else-if="msg.msgType === 'voice'" class="flex items-center gap-3 py-1">
              <button
                @click="playVoice(msg.id)"
                class="w-8 h-8 rounded-full bg-black/10 dark:bg-white/10 flex items-center justify-center hover:scale-105 transition-transform"
              >
                <Pause v-if="playingVoiceId === msg.id" class="w-4 h-4 inline-icon" />
                <Play v-else class="w-4 h-4 inline-icon ml-0.5" />
              </button>
              <div class="flex items-center gap-1 h-5">
                <span class="w-0.5 h-3 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                <span class="w-0.5 h-5 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-80'"></span>
                <span class="w-0.5 h-2 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                <span class="w-0.5 h-4 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-90'"></span>
                <span class="w-0.5 h-5 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-70'"></span>
                <span class="w-0.5 h-3 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-50'"></span>
              </div>
              <span class="text-xs font-mono tabular-nums opacity-90">
                {{ msg.duration || 6 }}s
              </span>
            </div>

            <!-- AI Summary Card (Rule 4: Stream blur typography) -->
            <div v-else-if="msg.msgType === 'ai_summary'" class="space-y-2 py-1 max-w-md">
              <div class="flex items-center gap-1.5 text-xs font-semibold text-blue-600 dark:text-blue-400">
                <Sparkles class="w-4 h-4 inline-icon" />
                <span>{{ langStore.t('chat.aiSynthesisTitle') }}</span>
              </div>
              <div class="text-xs leading-relaxed pretty-wrap">
                <span
                  v-for="(char, idx) in msg.content"
                  :key="idx"
                  class="stream-char"
                  :style="{ animationDelay: `${idx * 20}ms` }"
                >
                  {{ char }}
                </span>
              </div>
            </div>

            <!-- Revoke Message -->
            <span v-else-if="msg.isRevoked">
              {{ msg.isMe ? langStore.t('chat.recalledSelf') : langStore.t('chat.recalledPeer') }}
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
              <span class="font-medium text-blue-600 dark:text-blue-400">{{ langStore.t('chat.replyingTo') }} {{ (activeQuote || displayedQuote)?.userName }}:</span>
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
          :placeholder="langStore.t('chat.inputPlaceholder')"
          class="mim-input w-full p-2 bg-transparent text-sm text-mim-light-text dark:text-mim-dark-text placeholder-mim-light-muted dark:placeholder-mim-dark-muted resize-none border-none outline-none pretty-wrap"
        ></textarea>

        <!-- Bottom Controls -->
        <div class="flex items-center justify-between pt-1">
          <!-- Multimedia & AI Action Buttons (Zero Emoji) -->
          <div class="flex items-center gap-1">
            <button
              @click="triggerImageInput"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
              :title="langStore.t('chat.attachImage')"
            >
              <Image class="w-4 h-4 inline-icon" />
            </button>
            <button
              @click="triggerFileInput"
              class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
              :title="langStore.t('chat.attachFile')"
            >
              <Paperclip class="w-4 h-4 inline-icon" />
            </button>
            <button
              @click="toggleVoiceRecording"
              class="w-8 h-8 rounded-lg flex items-center justify-center transition-colors"
              :class="isRecording ? 'bg-red-500/20 text-red-500 animate-pulse' : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
              :title="langStore.t('chat.voiceRecording')"
            >
              <Mic class="w-4 h-4 inline-icon" />
            </button>
            <button
              @click="triggerAISummary"
              class="px-2 h-8 rounded-lg flex items-center gap-1 text-xs text-blue-600 dark:text-blue-400 hover:bg-blue-600/10 dark:hover:bg-blue-500/15 transition-colors font-medium"
              :title="langStore.t('chat.aiSummary')"
            >
              <Sparkles class="w-3.5 h-3.5 inline-icon" />
              <span>{{ langStore.t('chat.aiSummary') }}</span>
            </button>
          </div>

          <!-- Send Action Button -->
          <button
            @click="handleSend"
            :disabled="!inputContent.trim()"
            class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-40 disabled:hover:bg-blue-600 text-white text-xs font-medium flex items-center gap-1.5 transition-all shadow-sm cursor-pointer disabled:cursor-not-allowed"
          >
            <span>{{ langStore.t('chat.send') }}</span>
            <Send class="w-3 h-3 inline-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- Active WebRTC Call Overlay -->
    <div
      v-if="chatStore.activeCall"
      class="absolute inset-x-6 top-6 bottom-24 rounded-2xl bg-black/90 backdrop-blur-xl border border-white/10 flex flex-col items-center justify-between p-8 text-white z-50 shadow-2xl"
    >
      <div class="flex flex-col items-center gap-3">
        <div class="w-20 h-20 rounded-2xl bg-gradient-to-tr from-blue-500 to-indigo-600 flex items-center justify-center text-2xl font-bold">
          {{ chatStore.activeCall.peerName.substring(0, 2).toUpperCase() }}
        </div>
        <h3 class="text-lg font-semibold">{{ chatStore.activeCall.peerName }}</h3>
        <span class="px-3 py-1 rounded-full bg-white/10 text-xs font-mono tabular-nums">
          {{ formattedCallDuration }}
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
          @click="chatStore.endActiveCall"
          class="w-14 h-14 rounded-full bg-red-600 hover:bg-red-500 flex items-center justify-center text-white transition-colors shadow-lg shadow-red-600/30"
          :title="langStore.t('chat.hangup')"
        >
          <PhoneOff class="w-6 h-6 inline-icon" />
        </button>
      </div>
    </div>

    <!-- Incoming Call Overlay -->
    <div
      v-if="chatStore.incomingCall"
      class="fixed inset-0 bg-black/70 backdrop-blur-md z-50 flex items-center justify-center p-4 select-none"
    >
      <div class="w-full max-w-sm rounded-2xl bg-mim-light-card dark:bg-mim-dark-card p-6 shadow-2xl border border-black/10 dark:border-white/10 flex flex-col items-center gap-4 text-center">
        <div class="w-16 h-16 rounded-2xl bg-gradient-to-tr from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-xl animate-pulse">
          {{ chatStore.incomingCall.callerName.substring(0, 2).toUpperCase() }}
        </div>
        <div>
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">
            {{ chatStore.incomingCall.callerName }}
          </h3>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted mt-0.5">
            {{ langStore.t('chat.incomingCallDesc') }}
          </p>
        </div>
        <div class="flex items-center gap-6 mt-2">
          <button
            @click="chatStore.rejectIncomingCall"
            class="w-12 h-12 rounded-full bg-red-600 hover:bg-red-500 text-white flex items-center justify-center transition-colors shadow-lg shadow-red-600/30"
            :title="langStore.t('chat.decline')"
          >
            <PhoneOff class="w-5 h-5 inline-icon" />
          </button>
          <button
            @click="chatStore.acceptIncomingCall"
            class="w-12 h-12 rounded-full bg-emerald-600 hover:bg-emerald-500 text-white flex items-center justify-center transition-colors shadow-lg shadow-emerald-600/30 animate-bounce"
            :title="langStore.t('chat.accept')"
          >
            <Phone class="w-5 h-5 inline-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- Lightbox Modal for Images -->
    <div
      v-if="lightboxImage"
      @click="lightboxImage = null"
      class="fixed inset-0 bg-black/90 backdrop-blur-md z-50 flex items-center justify-center p-6 cursor-zoom-out"
    >
      <img :src="lightboxImage" class="max-w-full max-h-full rounded-xl object-contain shadow-2xl" />
    </div>
  </main>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from 'vue';
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
  Pause,
  X,
  Image,
  Paperclip,
  Mic,
  MicOff,
  Sparkles,
  Send,
  PhoneOff,
  UploadCloud,
  FileText,
  Download,
} from 'lucide-vue-next';
import { useChatStore, MessageItem } from '../stores/chat';
import { useLangStore } from '../stores/lang';

const chatStore = useChatStore();
const langStore = useLangStore();

const inputContent = ref('');
const activeQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const displayedQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const copiedMsgId = ref<string | null>(null);
const scrollContainer = ref<HTMLDivElement | null>(null);

const imageInputRef = ref<HTMLInputElement | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const isDragging = ref(false);
const isRecording = ref(false);
const isMuted = ref(false);
const playingVoiceId = ref<string | null>(null);
const lightboxImage = ref<string | null>(null);
const showConvInfo = ref(false);

const formattedCallDuration = computed(() => {
  const d = chatStore.callDuration;
  const mins = Math.floor(d / 60).toString().padStart(2, '0');
  const secs = (d % 60).toString().padStart(2, '0');
  return `00:${mins}:${secs}`;
});

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

  scrollToBottom();
};

const scrollToBottom = () => {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight;
    }
  });
};

// Copy Checkmark Feedback (Rule 10: 2.5s self-extinguishing)
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

// Media & File Upload
const triggerImageInput = () => {
  imageInputRef.value?.click();
};

const triggerFileInput = () => {
  fileInputRef.value?.click();
};

const handleImageSelected = async (e: Event) => {
  const files = (e.target as HTMLInputElement).files;
  if (!files || files.length === 0) return;
  await uploadAndSendFile(files[0], 'image');
};

const handleFileSelected = async (e: Event) => {
  const files = (e.target as HTMLInputElement).files;
  if (!files || files.length === 0) return;
  await uploadAndSendFile(files[0], 'file');
};

const handleFileDrop = async (e: DragEvent) => {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (!files || files.length === 0) return;

  const file = files[0];
  const isImg = file.type.startsWith('image/');
  await uploadAndSendFile(file, isImg ? 'image' : 'file');
};

const uploadAndSendFile = async (file: File, type: 'image' | 'file') => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('name', file.name);

  try {
    const res = await fetch(`/api/file/upload?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: formData,
    });
    const json = await res.json();
    if (json.code === 0 && json.data) {
      const fileUrl = json.data.file_url;
      const formattedSize = (file.size / 1024 / 1024).toFixed(2) + ' MB';

      chatStore.sendMessage(fileUrl, undefined, type, {
        fileName: file.name,
        fileSize: formattedSize,
      });
      scrollToBottom();
    }
  } catch (err) {
    console.error('File upload failed:', err);
  }
};

// Voice Note Simulation
const toggleVoiceRecording = () => {
  if (isRecording.value) {
    isRecording.value = false;
    const duration = Math.floor(Math.random() * 8) + 3;
    chatStore.sendMessage(langStore.t('chat.voiceMessage'), undefined, 'voice', { duration });
    scrollToBottom();
  } else {
    isRecording.value = true;
    setTimeout(() => {
      if (isRecording.value) {
        toggleVoiceRecording();
      }
    }, 4000);
  }
};

const playVoice = (msgId: string) => {
  if (playingVoiceId.value === msgId) {
    playingVoiceId.value = null;
  } else {
    playingVoiceId.value = msgId;
    setTimeout(() => {
      if (playingVoiceId.value === msgId) {
        playingVoiceId.value = null;
      }
    }, 4000);
  }
};

// AI Summary Generation (Rule 4: Stream Blur Typography)
const triggerAISummary = () => {
  const summaryText = langStore.currentLang === 'zh'
    ? 'Mim 智能协同提炼总结：\n• 单调递增 SeqId 架构已通过 Redis 原子计数器完成闭环验证。\n• 确认网络集群分区恢复期间零消息丢失保障。\n• Q3 架构评审顺利通过，已批准金丝雀集群上线。'
    : 'Mim AI Intelligence Synthesis:\n• Monotonic SeqId generator proposal verified with Redis atomic counter.\n• Zero message loss guaranteed during cluster network partition.\n• Q3 Architecture review finalized and approved for canary deployment.';

  const list = chatStore.activeMessages;
  const nextSeq = list.length > 0 ? list[list.length - 1].seqId + 1 : 1;

  const now = new Date();
  const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

  const aiMsg: MessageItem = {
    id: 'ai_' + Date.now(),
    clientMsgId: 'c_ai_' + Date.now(),
    seqId: nextSeq,
    senderId: 0,
    senderName: langStore.t('chat.aiCopilot'),
    content: summaryText,
    msgType: 'ai_summary',
    time: timeStr,
    isMe: false,
    status: 'read',
  };

  list.push(aiMsg);
  scrollToBottom();
};

// WebRTC Calls
const startCall = (type: 'audio' | 'video') => {
  chatStore.initiateWebRTCCall(type);
};

const openLightbox = (url: string) => {
  lightboxImage.value = url;
};
</script>
