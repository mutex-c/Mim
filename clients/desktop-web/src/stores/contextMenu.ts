import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface ContextMenuItem {
  id: string;
  label: string;
  icon?: any;
  shortcut?: string;
  danger?: boolean;
  disabled?: boolean;
  divided?: boolean;
  action: () => void;
}

export const useContextMenuStore = defineStore('contextMenu', () => {
  const visible = ref<boolean>(false);
  const x = ref<number>(0);
  const y = ref<number>(0);
  const items = ref<ContextMenuItem[]>([]);

  const show = (event: MouseEvent, menuItems: ContextMenuItem[]) => {
    event.preventDefault();
    event.stopPropagation();
    x.value = event.clientX;
    y.value = event.clientY;
    items.value = menuItems;
    visible.value = true;
  };

  const close = () => {
    visible.value = false;
  };

  return {
    visible,
    x,
    y,
    items,
    show,
    close,
  };
});
