<template>
  <div
    class="genshin-picker-container absolute bottom-12 left-0 z-50 w-[420px] max-w-[92vw] bg-[#161724]/95 backdrop-blur-xl border border-[#caa159]/60 rounded-2xl shadow-[0_12px_40px_rgba(0,0,0,0.6)] overflow-hidden flex flex-col transition-all duration-200 animate-in fade-in zoom-in-95 select-none"
  >
    <!-- Top Header -->
    <div class="px-4 py-2.5 border-b border-[#caa159]/30 flex items-center justify-between bg-gradient-to-r from-[#202135] via-[#282a42] to-[#202135]">
      <div class="flex items-center gap-2">
        <span class="text-amber-400 text-sm font-bold genshin-gold-glow">✦</span>
        <span class="text-xs font-bold text-[#f3eed8] tracking-wider">
          提瓦特表情包 · 派蒙的画作
        </span>
      </div>
      <button
        @click="handleClose"
        class="w-6 h-6 rounded-lg flex items-center justify-center text-[#caa159] hover:bg-white/10 transition-colors cursor-pointer text-xs"
      >
        ✕
      </button>
    </div>

    <!-- Mode Selector: "派蒙的画作 (彩色插画)" vs "经典台词名梗" -->
    <div class="px-3 py-2 flex items-center gap-2 bg-[#12131c] border-b border-[#caa159]/20">
      <button
        @click="mainTab = 'chibi'; playGenshinClick()"
        class="flex-1 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 cursor-pointer"
        :class="mainTab === 'chibi'
          ? 'bg-gradient-to-r from-[#d97706] to-[#b45309] text-white shadow-md'
          : 'text-[#d1d5db]/70 hover:text-white hover:bg-white/5'"
      >
        <span>🎨</span>
        <span>派蒙画作 ({{ chibiList.length }})</span>
      </button>

      <button
        @click="mainTab = 'quotes'; playGenshinClick()"
        class="flex-1 py-1.5 rounded-lg text-xs font-bold transition-all flex items-center justify-center gap-1.5 cursor-pointer"
        :class="mainTab === 'quotes'
          ? 'bg-gradient-to-r from-[#d97706] to-[#b45309] text-white shadow-md'
          : 'text-[#d1d5db]/70 hover:text-white hover:bg-white/5'"
      >
        <span>💬</span>
        <span>名梗台词 ({{ filteredQuotes.length }})</span>
      </button>
    </div>

    <!-- Tab 1: "派蒙的画作" Full-Color Chibi Illustrated Stickers Grid -->
    <div v-if="mainTab === 'chibi'" class="p-3 max-h-72 overflow-y-auto custom-scrollbar">
      <div class="grid grid-cols-4 gap-2.5">
        <div
          v-for="stk in chibiList"
          :key="stk.id"
          @click="selectChibi(stk)"
          class="group relative flex flex-col items-center p-2 rounded-xl bg-[#1e2030]/80 border border-[#caa159]/20 hover:border-amber-400 hover:bg-[#25283d] hover:scale-105 transition-all cursor-pointer shadow-sm"
          :title="stk.caption"
        >
          <!-- Illustration Artwork -->
          <div class="w-16 h-16 flex items-center justify-center overflow-hidden rounded-lg">
            <img :src="stk.imageSrc || stk.svgDataUri" :alt="stk.name" class="w-full h-full object-contain filter group-hover:drop-shadow-[0_0_8px_rgba(250,204,21,0.5)] transition-all" />
          </div>

          <!-- Name & Tag -->
          <span class="text-[10px] font-bold text-[#fef9c3] mt-1 truncate max-w-full text-center">
            {{ stk.name }}
          </span>
          <span class="text-[9px] text-[#9ca3af] truncate max-w-full scale-90">
            {{ stk.character }}
          </span>
        </div>
      </div>
    </div>

    <!-- Tab 2: "原神名梗台词" Classic Voice Quotes -->
    <div v-else class="flex flex-col">
      <!-- Quote Category Filter -->
      <div class="px-2.5 py-1.5 flex items-center gap-1 overflow-x-auto no-scrollbar border-b border-[#caa159]/20 bg-[#141520]">
        <button
          v-for="cat in categories"
          :key="cat.id"
          @click="activeCategory = cat.id; playGenshinClick()"
          class="px-2 py-0.5 rounded-md text-[10px] font-semibold whitespace-nowrap transition-all cursor-pointer shrink-0"
          :class="activeCategory === cat.id
            ? 'bg-amber-500 text-white'
            : 'text-[#d1d5db]/70 hover:text-white hover:bg-white/5'"
        >
          <span>{{ cat.icon }}</span>
          <span>{{ cat.nameZh }}</span>
        </button>
      </div>

      <!-- Quotes List -->
      <div class="p-2.5 max-h-64 overflow-y-auto space-y-1.5 custom-scrollbar">
        <div
          v-for="s in filteredQuotes"
          :key="s.id"
          @click="selectQuote(s)"
          class="p-2.5 rounded-xl border border-[#caa159]/20 bg-[#1e2030]/80 hover:bg-[#282a42] hover:border-amber-400/50 transition-all cursor-pointer group flex flex-col gap-1"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1.5">
              <span
                class="text-[10px] font-bold px-1.5 py-0.5 rounded text-white shadow-2xs"
                :style="{ backgroundColor: s.color }"
              >
                {{ s.character }}
              </span>
              <span class="text-[10px] text-[#9ca3af]">
                {{ s.badge }}
              </span>
            </div>
            <span class="text-[10px] text-amber-400 opacity-0 group-hover:opacity-100 transition-opacity font-semibold">
              点击发送 💬
            </span>
          </div>
          <p class="text-xs text-[#f3eed8] font-medium leading-relaxed group-hover:text-amber-200 transition-colors">
            {{ s.quote }}
          </p>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="px-3.5 py-1.5 bg-[#10111a] border-t border-[#caa159]/30 text-[10px] text-[#d1d5db]/70 flex items-center justify-between">
      <span>点击任意表情或台词即可直接发送</span>
      <span class="text-amber-400/90 font-mono">✦ Paimon's Art</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { GENSHIN_CATEGORIES, GENSHIN_STICKERS, CHIBI_STICKERS, type GenshinSticker, type ChibiSticker } from '../assets/genshin/stickers';
import { playGenshinClick, playPrimogemSparkle } from '../utils/genshinAudio';

const emit = defineEmits<{
  (e: 'select-sticker', sticker: GenshinSticker): void;
  (e: 'select-chibi', chibi: ChibiSticker): void;
  (e: 'close'): void;
}>();

const mainTab = ref<'chibi' | 'quotes'>('chibi');
const categories = GENSHIN_CATEGORIES;
const activeCategory = ref<string>('all');

const chibiList = CHIBI_STICKERS;

const filteredQuotes = computed(() => {
  if (activeCategory.value === 'all') {
    return GENSHIN_STICKERS;
  }
  return GENSHIN_STICKERS.filter((s) => s.category === activeCategory.value);
});

const selectChibi = (chibi: ChibiSticker) => {
  playPrimogemSparkle();
  emit('select-chibi', chibi);
};

const selectQuote = (sticker: GenshinSticker) => {
  playGenshinClick();
  emit('select-sticker', sticker);
};

const handleClose = () => {
  playGenshinClick();
  emit('close');
};
</script>
