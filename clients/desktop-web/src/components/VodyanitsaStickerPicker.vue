<template>
  <div
    class="vodyanitsa-picker-container absolute bottom-12 left-0 z-50 w-[380px] max-w-[92vw] bg-[#091124]/95 backdrop-blur-xl border border-[#38bdf8]/60 rounded-2xl shadow-[0_12px_40px_rgba(0,0,0,0.7)] overflow-hidden flex flex-col transition-all duration-200 animate-in fade-in zoom-in-95 select-none"
  >
    <!-- Top Header -->
    <div class="px-4 py-2.5 border-b border-[#38bdf8]/30 flex items-center justify-between bg-gradient-to-r from-[#0d1b38] via-[#10234a] to-[#0d1b38]">
      <div class="flex items-center gap-2">
        <span class="text-cyan-400 text-sm font-bold">💧</span>
        <span class="text-xs font-bold text-[#e0f2fe] tracking-wider">
          幽歌萦渊 · 沃雅妮莎微表情
        </span>
      </div>
      <button
        @click="handleClose"
        class="w-6 h-6 rounded-lg flex items-center justify-center text-[#7dd3fc] hover:bg-white/10 transition-colors cursor-pointer text-xs"
      >
        ✕
      </button>
    </div>

    <!-- Category Header Subtitle -->
    <div class="px-3.5 py-1.5 bg-[#060c1c] border-b border-[#38bdf8]/20 flex items-center justify-between">
      <span class="text-[11px] text-cyan-300/80 font-medium">科洛列夫茨基剧团 · 首席女高音日常 ({{ VODYANITSA_STICKERS.length }})</span>
      <span class="text-[10px] text-amber-300 font-bold">金鳍座 ★★★★★</span>
    </div>

    <!-- Stickers Grid -->
    <div class="p-3 max-h-72 overflow-y-auto custom-scrollbar">
      <div class="grid grid-cols-3 gap-2.5">
        <div
          v-for="stk in VODYANITSA_STICKERS"
          :key="stk.id"
          @click="selectSticker(stk)"
          class="group relative flex flex-col items-center p-2 rounded-xl bg-[#0c1833]/80 border border-[#38bdf8]/25 hover:border-cyan-400 hover:bg-[#12234a] hover:scale-105 transition-all cursor-pointer shadow-sm"
          :title="stk.caption"
        >
          <div class="w-20 h-20 flex items-center justify-center overflow-hidden rounded-lg">
            <img :src="stk.imageSrc" :alt="stk.name" class="w-full h-full object-contain filter group-hover:drop-shadow-[0_0_8px_rgba(56,189,248,0.7)] transition-all" />
          </div>
          <span class="text-[11px] font-bold text-[#e0f2fe] mt-1.5 text-center truncate max-w-full">
            {{ stk.name.replace('沃雅妮莎·', '') }}
          </span>
          <span class="text-[9px] text-[#7dd3fc]/70 text-center truncate max-w-full scale-95">
            {{ stk.caption.replace(/[「」]/g, '') }}
          </span>
        </div>
      </div>
    </div>

    <!-- Footer Note -->
    <div class="px-3 py-2 bg-[#060c1c]/90 border-t border-[#38bdf8]/20 flex items-center justify-between text-[11px] text-[#7dd3fc]/80">
      <span>点击直接发送至当前剧场会话</span>
      <span class="text-cyan-400 font-mono">✦ 离群的水妖</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { VODYANITSA_STICKERS, VodyanitsaSticker } from '../assets/vodyanitsa/vodyanitsaStickers';

const emit = defineEmits<{
  (e: 'select', code: string): void;
  (e: 'close'): void;
}>();

const handleClose = () => {
  emit('close');
};

const selectSticker = (stk: VodyanitsaSticker) => {
  emit('select', `[VODYANITSA_STICKER]:${stk.id}`);
  emit('close');
};
</script>
