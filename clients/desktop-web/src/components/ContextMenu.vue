<template>
  <Teleport to="body">
    <!-- Menu Floating Card (Native Desktop Non-Blocking Overlay) -->
    <div
      v-if="menuStore.visible"
      ref="menuEl"
      class="fixed z-[9999] w-52 p-1.5 rounded-xl bg-mim-light-card/95 dark:bg-mim-dark-card/95 backdrop-blur-xl border border-black/10 dark:border-white/10 shadow-2xl flex flex-col gap-0.5 text-xs text-mim-light-text dark:text-mim-dark-text animate-in fade-in zoom-in-95 duration-100 select-none"
      :style="computedPosition"
      @click.stop
      @contextmenu.stop.prevent
    >
      <template v-for="(item, index) in menuStore.items" :key="item.id || index">
        <!-- Divider -->
        <div
          v-if="item.divided"
          class="my-1 border-t border-black/[0.06] dark:border-white/[0.06]"
        ></div>

        <!-- Menu Item Button -->
        <button
          v-if="item.label"
          type="button"
          :disabled="item.disabled"
          @click="handleItemClick(item)"
          class="w-full h-8 px-2.5 rounded-lg flex items-center justify-between text-left transition-colors cursor-pointer group disabled:opacity-40 disabled:cursor-not-allowed"
          :class="item.danger
            ? 'text-rose-500 hover:bg-rose-500/10'
            : 'text-mim-light-text dark:text-mim-dark-text hover:bg-black/5 dark:hover:bg-white/5'"
        >
          <!-- Left: Icon + Label -->
          <div class="flex items-center gap-2 min-w-0">
            <component
              :is="item.icon"
              v-if="item.icon"
              class="w-3.5 h-3.5 shrink-0 inline-icon"
              :class="item.danger ? 'text-rose-500' : 'text-mim-light-muted dark:text-mim-dark-muted group-hover:text-mim-light-text dark:group-hover:text-mim-dark-text'"
            />
            <span class="truncate">{{ item.label }}</span>
          </div>

          <!-- Right: Shortcut Badge -->
          <span
            v-if="item.shortcut"
            class="text-[10px] font-mono text-mim-light-muted/60 dark:text-mim-dark-muted/60 ml-2 shrink-0"
          >
            {{ item.shortcut }}
          </span>
        </button>
      </template>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { useContextMenuStore, ContextMenuItem } from '../stores/contextMenu';

const menuStore = useContextMenuStore();
const menuEl = ref<HTMLElement | null>(null);

const computedPosition = computed(() => {
  const menuWidth = 208; // 52 * 4 = 208px
  const approxItemHeight = 34;
  const menuHeight = Math.max(80, menuStore.items.length * approxItemHeight + 20);

  const winW = typeof window !== 'undefined' ? window.innerWidth : 1000;
  const winH = typeof window !== 'undefined' ? window.innerHeight : 800;

  let left = menuStore.x;
  let top = menuStore.y;

  // Horizontal clamp
  if (left + menuWidth > winW - 12) {
    left = Math.max(12, winW - menuWidth - 12);
  }

  // Vertical clamp
  if (top + menuHeight > winH - 12) {
    top = Math.max(12, winH - menuHeight - 12);
  }

  return {
    left: `${left}px`,
    top: `${top}px`,
  };
});

const handleItemClick = (item: ContextMenuItem) => {
  if (item.disabled) return;
  menuStore.close();
  try {
    item.action();
  } catch (err) {
    console.error('[ContextMenu] Action error:', err);
  }
};

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && menuStore.visible) {
    menuStore.close();
  }
};

const handleOutsidePointer = (e: MouseEvent) => {
  if (menuStore.visible && menuEl.value && !menuEl.value.contains(e.target as Node)) {
    menuStore.close();
  }
};

onMounted(() => {
  window.addEventListener('mousedown', handleOutsidePointer);
  window.addEventListener('keydown', handleKeyDown);
  window.addEventListener('resize', menuStore.close);
  window.addEventListener('blur', menuStore.close);
});

onUnmounted(() => {
  window.removeEventListener('mousedown', handleOutsidePointer);
  window.removeEventListener('keydown', handleKeyDown);
  window.removeEventListener('resize', menuStore.close);
  window.removeEventListener('blur', menuStore.close);
});
</script>
