<template>
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[80] flex items-center justify-center p-4 select-none animate-in fade-in duration-150"
    @click.self="$emit('close')"
  >
    <div
      class="w-full max-w-lg rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/10 dark:border-white/10 p-5 space-y-4 relative"
      :class="themeStore.currentPalette === 'vodyanitsa'
        ? '!bg-[#091326]/95 border-cyan-400/50 shadow-[0_0_30px_rgba(56,189,248,0.25)] text-white'
        : ''"
    >
      <!-- Header -->
      <div class="flex items-center justify-between border-b border-black/5 dark:border-white/10 pb-3">
        <div class="flex items-center gap-2">
          <span class="text-base" :class="themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-400' : 'text-mim-accent'">🖼</span>
          <h3 class="text-sm font-semibold tracking-wide">
            {{ langStore.currentLang === 'zh' ? '自定义聊天背景图' : 'Custom Chat Wallpaper' }}
          </h3>
        </div>
        <button
          @click="$emit('close')"
          class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Upload Local Image -->
      <div class="space-y-2">
        <label class="block text-xs font-medium text-mim-light-muted dark:text-mim-dark-muted">
          {{ langStore.currentLang === 'zh' ? '上传本地壁纸文件' : 'Upload Local Wallpaper File' }}
        </label>
        <div class="flex items-center gap-3">
          <input
            type="file"
            ref="fileInputRef"
            accept="image/*"
            class="hidden"
            @change="handleFileUpload"
          />
          <button
            @click="fileInputRef?.click()"
            class="flex-1 h-9 rounded-xl mim-btn-secondary text-xs font-medium flex items-center justify-center gap-2 cursor-pointer transition-all border border-dashed border-cyan-400/40 hover:border-cyan-300"
          >
            <UploadCloud class="w-4 h-4 inline-icon text-cyan-400" />
            <span>{{ langStore.currentLang === 'zh' ? '选择本地图片 (PNG / JPG / WebP)' : 'Select Image File' }}</span>
          </button>
          <button
            v-if="themeStore.customBgUrl"
            @click="clearCustomBg"
            class="h-9 px-3 rounded-xl bg-red-500/10 hover:bg-red-500/20 text-red-500 text-xs font-medium transition-colors cursor-pointer"
            :title="langStore.currentLang === 'zh' ? '恢复默认背景' : 'Reset to Default'"
          >
            {{ langStore.currentLang === 'zh' ? '恢复默认' : 'Reset' }}
          </button>
        </div>
      </div>

      <!-- Presets Grid -->
      <div class="space-y-2 pt-1">
        <label class="block text-xs font-medium text-mim-light-muted dark:text-mim-dark-muted">
          {{ langStore.currentLang === 'zh' ? '精选官方画风预设背景' : 'Featured Preset Wallpapers' }}
        </label>
        <div class="grid grid-cols-3 gap-2.5">
          <div
            v-for="p in presets"
            :key="p.id"
            @click="selectPreset(p.url)"
            class="group relative rounded-xl overflow-hidden border-2 cursor-pointer transition-all aspect-video flex flex-col justify-end p-1.5 shadow-sm"
            :class="themeStore.customBgUrl === p.url
              ? 'border-cyan-400 shadow-[0_0_12px_rgba(56,189,248,0.4)] scale-102'
              : 'border-transparent hover:border-cyan-400/50 hover:scale-102'"
          >
            <img :src="p.url" :alt="p.name" class="absolute inset-0 w-full h-full object-cover filter brightness-90 group-hover:brightness-100 transition-all" />
            <div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent pointer-events-none"></div>
            <span class="relative z-10 text-[11px] font-bold text-white drop-shadow truncate">
              {{ p.name }}
            </span>
          </div>
        </div>
      </div>

      <!-- Sliders: Opacity & Blur -->
      <div class="space-y-3 pt-2 border-t border-black/5 dark:border-white/10">
        <!-- Opacity Slider -->
        <div class="space-y-1">
          <div class="flex items-center justify-between text-xs text-mim-light-muted dark:text-mim-dark-muted">
            <span>{{ langStore.currentLang === 'zh' ? '背景不透明度' : 'Wallpaper Opacity' }}</span>
            <span class="font-mono text-cyan-300 font-bold">{{ Math.round(themeStore.customBgOpacity * 100) }}%</span>
          </div>
          <input
            type="range"
            min="0.1"
            max="1"
            step="0.05"
            :value="themeStore.customBgOpacity"
            @input="onOpacityChange"
            class="w-full accent-cyan-400 cursor-pointer h-1.5 bg-black/10 dark:bg-white/10 rounded-lg"
          />
        </div>

        <!-- Blur Slider -->
        <div class="space-y-1">
          <div class="flex items-center justify-between text-xs text-mim-light-muted dark:text-mim-dark-muted">
            <span>{{ langStore.currentLang === 'zh' ? '背景虚化模糊度' : 'Wallpaper Blur' }}</span>
            <span class="font-mono text-cyan-300 font-bold">{{ themeStore.customBgBlur }}px</span>
          </div>
          <input
            type="range"
            min="0"
            max="20"
            step="1"
            :value="themeStore.customBgBlur"
            @input="onBlurChange"
            class="w-full accent-cyan-400 cursor-pointer h-1.5 bg-black/10 dark:bg-white/10 rounded-lg"
          />
        </div>
      </div>

      <!-- Footer Done Button -->
      <div class="pt-2 flex justify-end">
        <button
          @click="$emit('close')"
          class="px-5 h-8.5 rounded-xl mim-btn-primary text-xs font-semibold shadow-sm cursor-pointer"
        >
          {{ langStore.currentLang === 'zh' ? '完成' : 'Done' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { X, UploadCloud } from 'lucide-vue-next';
import { useThemeStore } from '../stores/theme';
import { useLangStore } from '../stores/lang';

import bgSiren from '../assets/vodyanitsa/images/vodyanitsa_stage_bg.png';
import bgSplash from '../assets/vodyanitsa/images/splash_vodyanitsa.png';
import bgFurina from '../assets/genshin/images/splash_furina.png';
import bgZhongli from '../assets/genshin/images/splash_zhongli.png';
import bgRaiden from '../assets/genshin/images/splash_raiden.png';
import bgHutao from '../assets/genshin/images/splash_hutao.png';

defineEmits(['close']);

const themeStore = useThemeStore();
const langStore = useLangStore();
const fileInputRef = ref<HTMLInputElement | null>(null);

const presets = [
  { id: 'siren', name: '极光天鹅 · 水妖幻境', url: bgSiren },
  { id: 'splash', name: '幽歌萦渊 · 5星立绘', url: bgSplash },
  { id: 'furina', name: '不休独舞 · 欧庇克莱', url: bgFurina },
  { id: 'zhongli', name: '尘世闲游 · 帝君岩金', url: bgZhongli },
  { id: 'raiden', name: '一心净土 · 鸣神雷光', url: bgRaiden },
  { id: 'hutao', name: '雪霁梅香 · 生死幽蝶', url: bgHutao },
];

const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement;
  if (!target.files || target.files.length === 0) return;
  const file = target.files[0];
  const reader = new FileReader();
  reader.onload = (event) => {
    const dataUrl = event.target?.result as string;
    if (dataUrl) {
      themeStore.setCustomBg(dataUrl);
    }
  };
  reader.readAsDataURL(file);
};

const selectPreset = (url: string) => {
  themeStore.setCustomBg(url);
};

const clearCustomBg = () => {
  themeStore.setCustomBg(null);
};

const onOpacityChange = (e: Event) => {
  const val = parseFloat((e.target as HTMLInputElement).value);
  themeStore.setCustomBgOpacity(val);
};

const onBlurChange = (e: Event) => {
  const val = parseFloat((e.target as HTMLInputElement).value);
  themeStore.setCustomBgBlur(val);
};
</script>
