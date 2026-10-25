import { defineStore } from 'pinia';
import { ref } from 'vue';

export type ThemeMode = 'light' | 'dark' | 'system';
export type ColorPalette = 'pine' | 'petrol' | 'obsidian' | 'genshin' | 'vodyanitsa';

export interface PaletteOption {
  id: ColorPalette;
  nameZh: string;
  nameEn: string;
  previewColor: string;
  descZh: string;
  descEn: string;
}

export const PALETTES: PaletteOption[] = [
  {
    id: 'pine',
    nameZh: '松风墨翠',
    nameEn: 'Nordic Forest Pine',
    previewColor: '#1b4d3e',
    descZh: '北欧松柏墨绿与暖石白，自然温润、沉稳耐看',
    descEn: 'Nordic Alpine Pine with warm bone canvas, calm and organic',
  },
  {
    id: 'petrol',
    nameZh: '沧浪深黛',
    nameEn: 'Baltic Sea Petrol',
    previewColor: '#1e404f',
    descZh: '深海深黛与冰岩冷灰，建筑工业级理性克制',
    descEn: 'Baltic Sea petrol with glacial slate, precise and architectural',
  },
  {
    id: 'obsidian',
    nameZh: '黑曜暖琥',
    nameEn: 'Obsidian & Amber',
    previewColor: '#b45309',
    descZh: '黑曜玄岩与暖琥珀金，现代极简高阶质感',
    descEn: 'Monolith charcoal with warm amber bronze highlights',
  },
  {
    id: 'genshin',
    nameZh: '提瓦特星穹 · 原神',
    nameEn: 'Teyvat Starlight · Genshin',
    previewColor: '#e5b044',
    descZh: '提瓦特深邃星穹、原石晶蓝与愿望流金，沉浸式幻想质感',
    descEn: 'Teyvat celestial deep blue with Primogem cyan and 5-Star Wish Gold',
  },
  {
    id: 'vodyanitsa',
    nameZh: '幽歌萦渊 · 沃雅妮莎',
    nameEn: 'Opera Siren · Vodyanitsa',
    previewColor: '#0284c7',
    descZh: '至冬科洛列夫茨基剧团首席女高音，深渊曜夜与水妖冰泉流金',
    descEn: 'Snezhnaya Grand Opera diva, abyssal starlight with sirens crystal azure',
  },
];

export const useThemeStore = defineStore('theme', () => {
  const currentMode = ref<ThemeMode>((localStorage.getItem('mim_theme_mode') as ThemeMode) || 'system');
  const currentPalette = ref<ColorPalette>((localStorage.getItem('mim_palette') as ColorPalette) || 'pine');
  const customBgUrl = ref<string | null>(localStorage.getItem('mim_custom_bg') || null);
  const customBgOpacity = ref<number>(parseFloat(localStorage.getItem('mim_custom_bg_opacity') || '1'));
  const customBgBlur = ref<number>(parseFloat(localStorage.getItem('mim_custom_bg_blur') || '0'));
  const isDark = ref<boolean>(false);

  const applyCustomBg = () => {
    if (customBgUrl.value) {
      document.documentElement.style.setProperty('--mim-custom-bg', `url("${customBgUrl.value}")`);
      document.documentElement.style.setProperty('--mim-custom-bg-opacity', customBgOpacity.value.toString());
      document.documentElement.style.setProperty('--mim-custom-bg-blur', `${customBgBlur.value}px`);
      document.documentElement.classList.add('has-custom-bg');
    } else {
      document.documentElement.style.removeProperty('--mim-custom-bg');
      document.documentElement.style.removeProperty('--mim-custom-bg-opacity');
      document.documentElement.style.removeProperty('--mim-custom-bg-blur');
      document.documentElement.classList.remove('has-custom-bg');
    }
  };

  const updateActualTheme = () => {
    if (currentMode.value === 'system') {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      isDark.value = prefersDark;
    } else {
      isDark.value = currentMode.value === 'dark';
    }

    if (isDark.value) {
      document.documentElement.classList.add('dark');
      document.documentElement.dataset.theme = 'dark';
    } else {
      document.documentElement.classList.remove('dark');
      document.documentElement.dataset.theme = 'light';
    }

    document.documentElement.dataset.palette = currentPalette.value;
    applyCustomBg();
  };

  const setTheme = (mode: ThemeMode) => {
    currentMode.value = mode;
    localStorage.setItem('mim_theme_mode', mode);
    updateActualTheme();
  };

  const setPalette = (palette: ColorPalette) => {
    currentPalette.value = palette;
    localStorage.setItem('mim_palette', palette);
    updateActualTheme();
  };

  const setCustomBg = (url: string | null) => {
    customBgUrl.value = url;
    if (url) {
      localStorage.setItem('mim_custom_bg', url);
    } else {
      localStorage.removeItem('mim_custom_bg');
    }
    applyCustomBg();
  };

  const setCustomBgOpacity = (opacity: number) => {
    customBgOpacity.value = opacity;
    localStorage.setItem('mim_custom_bg_opacity', opacity.toString());
    applyCustomBg();
  };

  const setCustomBgBlur = (blur: number) => {
    customBgBlur.value = blur;
    localStorage.setItem('mim_custom_bg_blur', blur.toString());
    applyCustomBg();
  };

  // Initialize and watch system changes
  updateActualTheme();
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (currentMode.value === 'system') {
      updateActualTheme();
    }
  });

  return {
    currentMode,
    currentPalette,
    customBgUrl,
    customBgOpacity,
    customBgBlur,
    isDark,
    setTheme,
    setPalette,
    setCustomBg,
    setCustomBgOpacity,
    setCustomBgBlur,
    palettes: PALETTES,
  };
});
