import { defineStore } from 'pinia';
import { ref } from 'vue';

export type ThemeMode = 'light' | 'dark' | 'system';

export const useThemeStore = defineStore('theme', () => {
  const currentMode = ref<ThemeMode>((localStorage.getItem('mim_theme_mode') as ThemeMode) || 'system');
  const isDark = ref<boolean>(false);

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
  };

  const setTheme = (mode: ThemeMode) => {
    currentMode.value = mode;
    localStorage.setItem('mim_theme_mode', mode);
    updateActualTheme();
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
    isDark,
    setTheme,
  };
});
