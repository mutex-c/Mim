<template>
  <div class="flex-1 h-full flex flex-col bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Header -->
    <header class="h-16 px-6 flex items-center justify-between border-b border-black/[0.04] dark:border-white/[0.06] bg-mim-light-stage/80 dark:bg-mim-dark-stage/80 backdrop-blur-md">
      <div>
        <h2 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('calls.title') }}</h2>
        <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted">{{ langStore.t('calls.subtitle') }}</p>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="startQuickCall('audio')"
          class="px-3 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium text-mim-light-text dark:text-mim-dark-text flex items-center gap-1.5 transition-colors"
        >
          <Phone class="w-3.5 h-3.5 inline-icon text-emerald-500" />
          <span>{{ langStore.t('calls.newAudioCall') }}</span>
        </button>
        <button
          @click="startQuickCall('video')"
          class="px-3 h-8 rounded-lg mim-btn-primary text-xs font-medium flex items-center gap-1.5 shadow-sm"
        >
          <Video class="w-3.5 h-3.5 inline-icon" />
          <span>{{ langStore.t('calls.newVideoCall') }}</span>
        </button>
      </div>
    </header>

    <!-- Call Log List -->
    <div class="flex-1 overflow-y-auto p-6 space-y-3">
      <div v-if="callLogs.length === 0" class="text-center py-12 text-mim-light-muted text-sm">
        {{ langStore.t('calls.empty') }}
      </div>

      <div
        v-for="call in callLogs"
        :key="call.id"
        class="p-4 rounded-xl bg-mim-light-card dark:bg-mim-dark-card border border-black/5 dark:border-white/5 flex items-center justify-between shadow-sm"
      >
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-semibold"
            :class="call.call_type === 'video' ? 'bg-mim-accent-subtle text-mim-accent' : 'bg-emerald-600/15 text-emerald-600 dark:text-emerald-400'"
          >
            <Video v-if="call.call_type === 'video'" class="w-5 h-5 inline-icon" />
            <Phone v-else class="w-5 h-5 inline-icon" />
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">
                {{ call.is_caller ? call.to_user : (call.from_user || call.to_user) }}
              </h4>
              <span class="text-[10px] px-1.5 py-0.5 rounded bg-black/5 dark:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted font-mono uppercase">
                {{ call.call_type === 'video' ? langStore.t('calls.video') : langStore.t('calls.audio') }}
              </span>
              <span class="text-[10px] px-1.5 py-0.5 rounded font-mono" :class="call.is_caller ? 'text-sky-500 bg-sky-500/10' : 'text-emerald-500 bg-emerald-500/10'">
                {{ call.is_caller ? (langStore.currentLang === 'zh' ? '呼出' : 'Outgoing') : (langStore.currentLang === 'zh' ? '呼入' : 'Incoming') }}
              </span>
            </div>
            <div class="flex items-center gap-2 text-xs text-mim-light-muted dark:text-mim-dark-muted font-mono tabular-nums">
              <span>{{ call.created_at }}</span>
              <span>·</span>
              <span>{{ langStore.t('calls.duration') }}: {{ call.duration }}s</span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <span
            class="text-xs px-2.5 py-1 rounded-full font-medium"
            :class="call.status === 'Completed' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-red-500/10 text-red-500'"
          >
            {{ call.status === 'Completed' ? langStore.t('calls.completed') : langStore.t('calls.missed') }}
          </span>
          <button
            @click="redial(call)"
            class="w-8 h-8 rounded-lg flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted hover:text-mim-accent transition-colors"
            :title="langStore.t('calls.redial')"
          >
            <PhoneCall class="w-4 h-4 inline-icon" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Phone, Video, PhoneCall } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';

const router = useRouter();
const chatStore = useChatStore();
const langStore = useLangStore();

const callLogs = ref<any[]>([]);

const fetchCallHistory = async () => {
  if (!chatStore.currentUserId || !chatStore.userToken) {
    callLogs.value = [];
    return;
  }
  try {
    const res = await fetch(`/api/calls/history?user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0 && Array.isArray(json.data)) {
      callLogs.value = json.data;
    } else {
      callLogs.value = [];
    }
  } catch (_) {
    callLogs.value = [];
  }
};

const startQuickCall = (type: 'audio' | 'video') => {
  const targetConv = chatStore.activeConvId || (chatStore.conversations.length > 0 ? chatStore.conversations[0].id : '');
  if (!targetConv) return;
  router.push(`/chats/${targetConv}`);
  chatStore.initiateWebRTCCall(type);
};

const redial = (call: any) => {
  const targetConv = chatStore.activeConvId || (chatStore.conversations.length > 0 ? chatStore.conversations[0].id : '');
  if (!targetConv) return;
  router.push(`/chats/${targetConv}`);
  chatStore.initiateWebRTCCall(call.call_type === 'video' ? 'video' : 'audio');
};

watch(
  () => chatStore.currentUserId,
  () => {
    fetchCallHistory();
  },
  { immediate: true }
);

onMounted(() => {
  fetchCallHistory();
});
</script>
