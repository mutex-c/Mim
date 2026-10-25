<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[70] flex items-center justify-center p-4 select-none">
    
    <!-- GENSHIN TEYVAT ADVENTURE HANDBOOK CARD -->
    <div
      v-if="themeStore.currentPalette === 'genshin'"
      class="w-full max-w-md rounded-3xl bg-gradient-to-b from-[#1e2030] via-[#161724] to-[#0e0f18] shadow-[0_15px_50px_rgba(0,0,0,0.7)] border-2 border-[#caa159]/70 overflow-hidden flex flex-col p-6 space-y-4 relative"
    >
      <!-- Gold Corner Brackets -->
      <div class="absolute top-3 left-3 w-6 h-6 border-t-2 border-l-2 border-[#caa159]" />
      <div class="absolute top-3 right-3 w-6 h-6 border-t-2 border-r-2 border-[#caa159]" />
      <div class="absolute bottom-3 left-3 w-6 h-6 border-b-2 border-l-2 border-[#caa159]" />
      <div class="absolute bottom-3 right-3 w-6 h-6 border-b-2 border-r-2 border-[#caa159]" />

      <!-- Header -->
      <div class="flex items-center justify-between border-b border-[#caa159]/30 pb-3">
        <div class="flex items-center gap-2">
          <span class="text-amber-400 font-bold text-sm genshin-gold-glow">✦</span>
          <h3 class="text-xs font-bold tracking-widest text-[#f5eed6] uppercase">
            冒险家协会 · 旅行者档案
          </h3>
        </div>
        <button
          @click="handleClose"
          class="w-6 h-6 rounded-lg flex items-center justify-center text-[#caa159] hover:bg-white/10 transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Main Profile Banner -->
      <div class="flex items-center gap-4 bg-gradient-to-r from-[#25283d] to-[#1c1d2c] p-3.5 rounded-2xl border border-[#caa159]/40 relative">
        <!-- Avatar with Gold Frame -->
        <div class="relative">
          <div
            class="w-16 h-16 rounded-2xl font-bold text-xl flex items-center justify-center shadow-lg border-2 border-[#caa159] overflow-hidden bg-[#2d1b22]"
            :style="{
              backgroundColor: getAvatarColor(user?.id || user?.user_id || user?.nickname).bg,
              color: getAvatarColor(user?.id || user?.user_id || user?.nickname).text
            }"
          >
            <img
              v-if="user?.avatar && (user.avatar.startsWith('http') || user.avatar.startsWith('data:image'))"
              :src="user.avatar"
              class="w-full h-full object-cover"
            />
            <span v-else>{{ user?.avatar || (user?.nickname ? user.nickname.substring(0, 2).toUpperCase() : 'MI') }}</span>
          </div>

          <!-- Vision Element Badge -->
          <div
            class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black text-white border border-[#caa159] shadow-sm"
            :style="{ backgroundColor: getVisionColor(user?.nickname || '') }"
          >
            {{ getVisionElement(user?.nickname || '') }}
          </div>
        </div>

        <!-- Name & Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h4 class="text-base font-bold text-[#fef9c3] truncate">
              {{ user?.nickname || '旅行者' }}
            </h4>
            <span class="text-amber-400 text-xs">★★★★★</span>
          </div>
          
          <div class="flex items-center gap-2 mt-1 text-[11px] font-mono">
            <span class="px-1.5 py-0.2 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30">
              UID: 1000{{ String(user?.id || user?.user_id || 10086).padStart(5, '0') }}
            </span>
          </div>

          <div class="flex items-center gap-3 mt-1.5 text-[10px] text-[#d1d5db]">
            <span>冒险等阶 <strong class="text-amber-300 font-bold font-mono">Lv.60</strong></span>
            <span>世界等级 <strong class="text-amber-300 font-bold font-mono">8</strong></span>
          </div>
        </div>
      </div>

      <!-- Character Showcase Row (角色展柜) -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-[11px] text-[#caa159] font-semibold">
          <span>角色展柜 (5★ 伙伴)</span>
          <span class="text-[10px] text-[#9ca3af]">好感度 10</span>
        </div>
        <div class="grid grid-cols-4 gap-2 bg-[#12131c] p-2.5 rounded-xl border border-[#caa159]/25">
          <div
            v-for="char in showcaseCharacters"
            :key="char.id"
            class="flex flex-col items-center p-1 rounded-lg bg-[#1a1b28] border border-[#caa159]/30"
          >
            <div class="w-10 h-10 rounded-full border border-amber-400 overflow-hidden flex items-center justify-center bg-black/40">
              <img :src="char.avatar" class="w-full h-full object-cover" :alt="char.name" />
            </div>
            <span class="text-[10px] font-bold text-[#fef9c3] mt-1">{{ char.name }}</span>
            <span class="text-[8px] text-amber-300 font-mono">Lv.90</span>
          </div>
        </div>
      </div>

      <!-- Signature Quote -->
      <div class="px-3.5 py-2 rounded-xl bg-[#141522] border border-[#caa159]/30 text-xs text-[#e5e7eb] italic text-center leading-relaxed">
        “{{ user?.signature || '欲买桂花同载酒……只可惜故人，何日再见呢。' }}”
      </div>

      <!-- Actions -->
      <div class="pt-1 flex gap-2.5">
        <button
          @click="startDirectChat"
          class="flex-1 h-9 rounded-xl bg-gradient-to-r from-[#d97706] to-[#b45309] hover:from-[#f59e0b] hover:to-[#d97706] text-white text-xs font-bold flex items-center justify-center gap-1.5 shadow-lg border border-[#caa159]/60 cursor-pointer"
        >
          <MessageSquare class="w-3.5 h-3.5 inline-icon" />
          <span>发起提瓦特通讯</span>
        </button>
      </div>
    </div>

    <!-- VODYANITSA GRAND OPERA PERSONA PROFILE CARD -->
    <div
      v-else-if="themeStore.currentPalette === 'vodyanitsa'"
      class="w-full max-w-md rounded-3xl bg-gradient-to-b from-[#0e172e] via-[#091124] to-[#040817] shadow-[0_15px_50px_rgba(0,0,0,0.8)] border-2 border-[#38bdf8]/70 overflow-hidden flex flex-col p-6 space-y-4 relative"
    >
      <!-- Aqua Corner Brackets -->
      <div class="absolute top-3 left-3 w-6 h-6 border-t-2 border-l-2 border-[#38bdf8]" />
      <div class="absolute top-3 right-3 w-6 h-6 border-t-2 border-r-2 border-[#38bdf8]" />
      <div class="absolute bottom-3 left-3 w-6 h-6 border-b-2 border-l-2 border-[#38bdf8]" />
      <div class="absolute bottom-3 right-3 w-6 h-6 border-b-2 border-r-2 border-[#38bdf8]" />

      <!-- Header -->
      <div class="flex items-center justify-between border-b border-[#38bdf8]/30 pb-3">
        <div class="flex items-center gap-2">
          <Sparkles class="w-4 h-4 text-cyan-400 inline-icon" />
          <h3 class="text-xs font-bold tracking-widest text-[#e0f2fe] uppercase">
            科洛列夫茨基剧团 · 首席档案
          </h3>
        </div>
        <button
          @click="handleClose"
          class="w-6 h-6 rounded-lg flex items-center justify-center text-[#7dd3fc] hover:bg-white/10 transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Main Profile Banner -->
      <div class="flex items-center gap-4 bg-gradient-to-r from-[#12234a] to-[#0d1833] p-3.5 rounded-2xl border border-[#38bdf8]/40 relative">
        <div class="relative">
          <div
            class="w-16 h-16 rounded-2xl font-bold text-xl flex items-center justify-center shadow-lg border-2 border-[#38bdf8] overflow-hidden bg-[#061124]"
            :style="{
              backgroundColor: getAvatarColor(user?.id || user?.user_id || user?.nickname).bg,
              color: getAvatarColor(user?.id || user?.user_id || user?.nickname).text
            }"
          >
            <img
              :src="vodyanitsaAvatar"
              alt="沃雅妮莎"
              class="w-full h-full object-cover"
            />
          </div>

          <!-- Vision Element Badge (Hydro) -->
          <div
            class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full flex items-center justify-center text-[10px] font-black text-white bg-[#0284c7] border border-[#7dd3fc] shadow-sm"
          >
            水
          </div>
        </div>

        <!-- Name & Info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <h4 class="text-base font-bold text-[#f0f9ff] truncate">
              {{ user?.nickname || '沃雅妮莎' }}
            </h4>
            <span class="text-[9px] font-bold text-amber-300">★★★★★</span>
          </div>
          <p class="text-xs text-[#7dd3fc] font-medium mt-0.5">
            ✦ 离群的水妖 · 金鳍座
          </p>
          <p class="text-[11px] text-cyan-200/60 font-mono">
            科洛列夫茨基剧团 · 首席女高音
          </p>
        </div>
      </div>

      <!-- Vodyanitsa Character Showcase & Battle Attributes -->
      <div class="space-y-2">
        <div class="flex items-center justify-between text-xs font-bold text-[#7dd3fc]">
          <span>命之座与战斗特权</span>
          <span class="text-[10px] text-cyan-300/80 font-normal">金鳍座 · 第六幕</span>
        </div>
        <div class="grid grid-cols-3 gap-2">
          <div class="p-2 rounded-xl bg-[#09142b] border border-[#38bdf8]/20 text-center">
            <span class="block text-[9px] text-[#94a3b8]">生命值上限</span>
            <span class="text-xs font-bold text-cyan-300 font-mono">65,280</span>
          </div>
          <div class="p-2 rounded-xl bg-[#09142b] border border-[#38bdf8]/20 text-center">
            <span class="block text-[9px] text-[#94a3b8]">治疗加成</span>
            <span class="text-xs font-bold text-emerald-400 font-mono">+75.4%</span>
          </div>
          <div class="p-2 rounded-xl bg-[#09142b] border border-[#38bdf8]/20 text-center">
            <span class="block text-[9px] text-[#94a3b8]">水/冰抗性削减</span>
            <span class="text-xs font-bold text-blue-400 font-mono">-35.0%</span>
          </div>
        </div>
      </div>

      <!-- Splash Art Vignette Preview -->
      <div class="relative rounded-2xl overflow-hidden border border-[#38bdf8]/40 h-28 flex items-center justify-center bg-black/40">
        <img :src="splashVodyanitsa" alt="沃雅妮莎" class="w-full h-full object-cover object-top opacity-85 hover:opacity-100 transition-opacity" />
        <div class="absolute inset-0 bg-gradient-to-t from-[#040817] via-transparent to-transparent"></div>
        <span class="absolute bottom-2 left-3 text-[10px] text-cyan-200 font-serif italic">
          「水流过冰霜的台阶，舞台的帷幕即将升起。」
        </span>
      </div>

      <!-- Action Button -->
      <div class="pt-2 flex gap-2.5">
        <button
          @click="startDirectChat"
          class="flex-1 h-9 rounded-xl bg-gradient-to-r from-sky-500 to-blue-600 hover:from-sky-400 hover:to-blue-500 text-white text-xs font-bold flex items-center justify-center gap-1.5 shadow-md cursor-pointer transition-all"
        >
          <MessageSquare class="w-3.5 h-3.5 inline-icon" />
          <span>进入剧场对谈</span>
        </button>
      </div>
    </div>

    <!-- STANDARD DEFAULT THEME USER PROFILE MODAL -->
    <div
      v-else
      class="w-full max-w-sm rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col p-6 space-y-5"
    >
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">用户资料</h3>
        <button
          @click="$emit('close')"
          class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <div class="flex flex-col items-center gap-3 text-center">
        <div
          class="w-20 h-20 rounded-2xl font-bold text-2xl flex items-center justify-center shadow-md transition-all overflow-hidden"
          :class="getStatusRingClass('online')"
          :style="{
            backgroundColor: getAvatarColor(user?.id || user?.user_id || user?.nickname).bg,
            color: getAvatarColor(user?.id || user?.user_id || user?.nickname).text
          }"
        >
          <img
            v-if="user?.avatar && (user.avatar.startsWith('http') || user.avatar.startsWith('data:image'))"
            :src="user.avatar"
            class="w-full h-full object-cover"
          />
          <span v-else>{{ user?.avatar || (user?.nickname ? user.nickname.substring(0, 2).toUpperCase() : 'MI') }}</span>
        </div>
        <div>
          <h4 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">{{ user?.nickname || '未知用户' }}</h4>
          <p class="text-xs text-mim-light-muted font-mono mt-0.5">@{{ user?.account }} · ID: {{ user?.id || user?.user_id }}</p>
        </div>
        <div class="px-3 py-1.5 rounded-xl bg-black/[0.03] dark:bg-white/[0.04] text-xs text-mim-light-muted max-w-xs text-center">
          {{ user?.signature || '这个用户很懒，什么都没写。' }}
        </div>
      </div>

      <div class="pt-2 flex gap-2.5">
        <button
          @click="startDirectChat"
          class="flex-1 h-9 rounded-xl mim-btn-primary text-xs font-medium flex items-center justify-center gap-1.5 shadow-sm"
        >
          <MessageSquare class="w-3.5 h-3.5 inline-icon" />
          <span>发消息</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { X, MessageSquare, Sparkles } from 'lucide-vue-next';
import { useRouter } from 'vue-router';
import { useChatStore } from '../stores/chat';
import { useThemeStore } from '../stores/theme';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';
import { playGenshinClick } from '../utils/genshinAudio';

import splashZhongli from '../assets/genshin/images/splash_zhongli.png';
import splashFurina from '../assets/genshin/images/splash_furina.png';
import splashRaiden from '../assets/genshin/images/splash_raiden.png';
import splashHutao from '../assets/genshin/images/splash_hutao.png';
import splashVodyanitsa from '../assets/vodyanitsa/images/splash_vodyanitsa.png';
import vodyanitsaAvatar from '../assets/vodyanitsa/images/vodyanitsa_avatar.png';

const props = defineProps<{
  user: any;
}>();

const emit = defineEmits(['close']);
const router = useRouter();
const chatStore = useChatStore();
const themeStore = useThemeStore();

const showcaseCharacters = [
  { id: 'zhongli', name: '钟离', avatar: splashZhongli },
  { id: 'furina', name: '芙宁娜', avatar: splashFurina },
  { id: 'raiden', name: '雷电将军', avatar: splashRaiden },
  { id: 'hutao', name: '胡桃', avatar: splashHutao },
];

const VISION_ELEMENTS = [
  { name: '岩', color: '#d97706' },
  { name: '水', color: '#0284c7' },
  { name: '雷', color: '#7c3aed' },
  { name: '草', color: '#16a34a' },
  { name: '风', color: '#0d9488' },
  { name: '火', color: '#dc2626' },
  { name: '冰', color: '#0ea5e9' },
];

function getVisionElement(name: string): string {
  if (name.includes('钟离')) return '岩';
  if (name.includes('芙宁娜') || name.includes('水')) return '水';
  if (name.includes('雷') || name.includes('影')) return '雷';
  if (name.includes('纳西妲') || name.includes('草')) return '草';
  if (name.includes('温迪') || name.includes('风')) return '风';
  if (name.includes('胡桃') || name.includes('可莉') || name.includes('火')) return '火';
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash += name.charCodeAt(i);
  return VISION_ELEMENTS[Math.abs(hash) % VISION_ELEMENTS.length].name;
}

function getVisionColor(name: string): string {
  const el = getVisionElement(name);
  return VISION_ELEMENTS.find(v => v.name === el)?.color || '#caa159';
}

const handleClose = () => {
  playGenshinClick();
  emit('close');
};

const startDirectChat = () => {
  if (themeStore.currentPalette === 'genshin') {
    playGenshinClick();
  }
  const targetId = props.user.id || props.user.user_id;
  if (!targetId) return;
  const u1 = Math.min(chatStore.currentUserId, targetId);
  const u2 = Math.max(chatStore.currentUserId, targetId);
  const convId = `${u1}_${u2}`;
  chatStore.selectConversation(convId);
  router.push(`/chats/${convId}`);
  emit('close');
};
</script>
