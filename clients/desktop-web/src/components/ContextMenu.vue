<template>
  <Teleport to="body">
    <!-- Menu Floating Card with Snappy Spring Entrance & Dynamic Origin Clamp -->
    <Transition name="cm-pop">
      <div
        v-if="menuStore.visible"
        ref="menuEl"
        class="cm-menu fixed z-[9999] w-52 p-1.5 rounded-xl bg-mim-light-card/95 dark:bg-mim-dark-card/95 backdrop-blur-xl border border-black/10 dark:border-white/10 shadow-2xl flex flex-col gap-0.5 text-xs text-mim-light-text dark:text-mim-dark-text select-none"
        :style="computedStyle"
        @click.stop
        @contextmenu.stop.prevent
        @mouseleave="handleMouseLeave"
      >
        <!-- Gliding Highlight Pill (Single instance, spring follow) -->
        <span
          class="cm-highlight"
          :class="highlight.visible ? 'opacity-100' : 'opacity-0'"
          :data-tone="highlight.tone"
          :style="{
            transform: `translateY(${highlight.top}px)`,
            height: `${highlight.height}px`,
          }"
          aria-hidden="true"
        />

        <template v-for="(item, index) in menuStore.items" :key="item.id || index">
          <!-- Divider -->
          <div
            v-if="item.divided"
            class="my-1 border-t border-black/[0.06] dark:border-white/[0.06] relative z-10"
          ></div>

          <!-- Menu Item Button -->
          <button
            v-if="item.label"
            ref="itemEls"
            type="button"
            :disabled="item.disabled"
            :style="{ '--i': index }"
            @mouseenter="handleItemMouseEnter($event, item, index)"
            @click="handleItemClick(item)"
            class="cm-item w-full h-8 px-2.5 rounded-lg flex items-center justify-between text-left cursor-pointer group disabled:opacity-40 disabled:cursor-not-allowed relative z-10 bg-transparent"
            :class="item.danger ? 'text-rose-500' : 'text-mim-light-text dark:text-mim-dark-text'"
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
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, reactive, watch, nextTick, onMounted, onUnmounted } from 'vue';
import { useContextMenuStore, ContextMenuItem } from '../stores/contextMenu';

const menuStore = useContextMenuStore();
const menuEl = ref<HTMLElement | null>(null);
const itemEls = ref<HTMLElement[]>([]);

// Gliding highlight state
const highlight = reactive({
  top: 0,
  height: 32,
  tone: 'normal',
  visible: false,
});

const activeIndex = ref<number>(-1);
let leaveTimeout: any = null;

const computedStyle = computed(() => {
  const menuWidth = 208; // w-52
  const approxItemHeight = 34;
  const menuHeight = Math.max(60, menuStore.items.length * approxItemHeight + 16);

  const winW = typeof window !== 'undefined' ? window.innerWidth : 1000;
  const winH = typeof window !== 'undefined' ? window.innerHeight : 800;

  // Viewport clamp
  const left = Math.max(8, Math.min(menuStore.x, winW - menuWidth - 8));
  const top = Math.max(8, Math.min(menuStore.y, winH - menuHeight - 8));

  // Dynamic origin calculation from cursor hit-point
  const originX = Math.max(0, Math.min(menuWidth, menuStore.x - left));
  const originY = Math.max(0, Math.min(menuHeight, menuStore.y - top));

  return {
    left: `${left}px`,
    top: `${top}px`,
    transformOrigin: `${originX}px ${originY}px`,
  };
});

// Update highlight capsule position
const handleItemMouseEnter = (event: MouseEvent, item: ContextMenuItem, index: number) => {
  if (leaveTimeout) {
    clearTimeout(leaveTimeout);
    leaveTimeout = null;
  }
  const el = event.currentTarget as HTMLElement;
  if (!el) return;

  highlight.top = el.offsetTop;
  highlight.height = el.offsetHeight;
  highlight.tone = item.danger ? 'danger' : 'normal';
  highlight.visible = true;
  activeIndex.value = index;
};

// 70ms debounce before fading out highlight on mouse leave
const handleMouseLeave = () => {
  leaveTimeout = setTimeout(() => {
    highlight.visible = false;
    activeIndex.value = -1;
  }, 70);
};

const handleItemClick = (item: ContextMenuItem) => {
  if (item.disabled) return;
  menuStore.close();
  try {
    item.action();
  } catch (err) {
    console.error('[ContextMenu] Action error:', err);
  }
};

// Arrow key navigation
const getEnabledIndices = () => {
  return menuStore.items
    .map((item, idx) => (item.label && !item.disabled ? idx : -1))
    .filter(idx => idx !== -1);
};

const moveActive = (delta: number) => {
  const enabled = getEnabledIndices();
  if (enabled.length === 0) return;

  const currentPos = enabled.indexOf(activeIndex.value);
  let nextPos = currentPos + delta;
  if (currentPos === -1) {
    nextPos = delta > 0 ? 0 : enabled.length - 1;
  } else if (nextPos < 0) {
    nextPos = enabled.length - 1;
  } else if (nextPos >= enabled.length) {
    nextPos = 0;
  }

  const targetIndex = enabled[nextPos];
  activeIndex.value = targetIndex;

  nextTick(() => {
    const el = itemEls.value[targetIndex];
    if (el) {
      highlight.top = el.offsetTop;
      highlight.height = el.offsetHeight;
      const targetItem = menuStore.items[targetIndex];
      highlight.tone = targetItem?.danger ? 'danger' : 'normal';
      highlight.visible = true;
    }
  });
};

const handleKeyDown = (e: KeyboardEvent) => {
  if (!menuStore.visible) return;

  if (e.key === 'Escape') {
    e.preventDefault();
    menuStore.close();
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    moveActive(1);
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    moveActive(-1);
  } else if (e.key === 'Enter') {
    e.preventDefault();
    if (activeIndex.value >= 0 && activeIndex.value < menuStore.items.length) {
      const item = menuStore.items[activeIndex.value];
      if (item && !item.disabled) {
        handleItemClick(item);
      }
    }
  }
};

const handleOutsidePointer = (e: MouseEvent) => {
  if (menuStore.visible && menuEl.value && !menuEl.value.contains(e.target as Node)) {
    menuStore.close();
  }
};

// Reset highlight on menu reopen
watch(() => menuStore.visible, (newVal) => {
  if (!newVal) {
    highlight.visible = false;
    activeIndex.value = -1;
  }
});

onMounted(() => {
  window.addEventListener('mousedown', handleOutsidePointer);
  window.addEventListener('keydown', handleKeyDown);
  window.addEventListener('resize', menuStore.close);
  window.addEventListener('blur', menuStore.close);
});

onUnmounted(() => {
  if (leaveTimeout) clearTimeout(leaveTimeout);
  window.removeEventListener('mousedown', handleOutsidePointer);
  window.removeEventListener('keydown', handleKeyDown);
  window.removeEventListener('resize', menuStore.close);
  window.removeEventListener('blur', menuStore.close);
});
</script>

<style scoped>
/* Snappy Spring In & Fast Exit */
.cm-pop-enter-active {
  transition: transform 160ms cubic-bezier(0.16, 1, 0.3, 1),
              opacity 140ms cubic-bezier(0.16, 1, 0.3, 1);
}
.cm-pop-leave-active {
  transition: transform 120ms cubic-bezier(0.22, 1, 0.36, 1),
              opacity 100ms cubic-bezier(0.22, 1, 0.36, 1);
}
.cm-pop-enter-from {
  opacity: 0;
  transform: scale(0.96);
}
.cm-pop-leave-to {
  opacity: 0;
  transform: scale(0.98);
}

/* Gliding highlight capsule */
.cm-highlight {
  position: absolute;
  left: 6px;
  right: 6px;
  top: 0;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.05);
  pointer-events: none;
  will-change: transform, height, opacity;
  transition: transform 180ms cubic-bezier(0.22, 1, 0.36, 1),
              height 180ms cubic-bezier(0.22, 1, 0.36, 1),
              opacity 120ms cubic-bezier(0.22, 1, 0.36, 1),
              background-color 160ms cubic-bezier(0.22, 1, 0.36, 1);
}
:root.dark .cm-highlight {
  background: rgba(255, 255, 255, 0.08);
}
.cm-highlight[data-tone="danger"] {
  background: rgba(244, 63, 94, 0.14) !important;
}

/* Staggered Item Entrance: min(i, 4) * 35ms */
.cm-item {
  animation: cmItemIn 200ms cubic-bezier(0.16, 1, 0.3, 1) both;
  animation-delay: calc(min(var(--i, 0), 4) * 35ms);
}
@keyframes cmItemIn {
  from {
    opacity: 0;
    transform: translateY(-2px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
