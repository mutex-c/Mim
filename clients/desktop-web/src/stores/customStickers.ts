import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface CustomSticker {
  id: string;
  name: string;
  url: string; // Base64 data URI or HTTP URL
  createdAt: number;
}

const STORAGE_KEY = 'mim_custom_stickers_v1';

export const useCustomStickersStore = defineStore('customStickers', () => {
  const stickers = ref<CustomSticker[]>([]);

  const loadFromStorage = () => {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (raw) {
        stickers.value = JSON.parse(raw);
      } else {
        stickers.value = [];
      }
    } catch (e) {
      console.warn('Failed to parse custom stickers from localStorage:', e);
      stickers.value = [];
    }
  };

  const saveToStorage = () => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(stickers.value));
    } catch (e) {
      console.warn('Failed to save custom stickers to localStorage:', e);
    }
  };

  const addSticker = (name: string, url: string): CustomSticker => {
    const id = 'custom_' + Date.now() + '_' + Math.random().toString(36).substring(2, 7);
    const newSticker: CustomSticker = {
      id,
      name: name.trim() || '自定义表情',
      url,
      createdAt: Date.now(),
    };
    stickers.value.unshift(newSticker);
    saveToStorage();
    return newSticker;
  };

  const deleteSticker = (id: string) => {
    stickers.value = stickers.value.filter((s) => s.id !== id);
    saveToStorage();
  };

  const clearAll = () => {
    stickers.value = [];
    saveToStorage();
  };

  const getStickerById = (id: string): CustomSticker | undefined => {
    return stickers.value.find((s) => s.id === id);
  };

  // Initialize immediately
  loadFromStorage();

  return {
    stickers,
    addSticker,
    deleteSticker,
    clearAll,
    getStickerById,
    reload: loadFromStorage,
  };
});
