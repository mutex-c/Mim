<template>
  <div
    class="vodyanitsa-picker-container absolute bottom-[calc(100%+12px)] left-0 z-50 w-[420px] max-w-[94vw] bg-[#091124]/95 backdrop-blur-xl border border-[#38bdf8]/60 rounded-2xl shadow-[0_16px_50px_rgba(0,0,0,0.75)] overflow-hidden flex flex-col transition-all duration-200 animate-in fade-in zoom-in-95 select-none"
  >
    <!-- Hidden File Input for Custom Stickers -->
    <input
      type="file"
      ref="fileInputRef"
      @change="handleFileUpload"
      accept="image/*"
      multiple
      class="hidden"
    />

    <!-- Top Header -->
    <div class="px-4 py-2.5 border-b border-[#38bdf8]/30 flex items-center justify-between bg-gradient-to-r from-[#0d1b38] via-[#10234a] to-[#0d1b38]">
      <div class="flex items-center gap-2">
        <img :src="iconLily" alt="" class="w-4 h-4 object-contain inline-block shrink-0" />
        <span class="text-xs font-bold text-[#e0f2fe] tracking-wider">
          科洛列夫茨基剧团 · 表情工坊
        </span>
      </div>
      <button
        @click="handleClose"
        class="w-6 h-6 rounded-lg flex items-center justify-center text-[#7dd3fc] hover:bg-white/10 hover:text-white transition-colors cursor-pointer"
        title="关闭"
      >
        <X class="w-3.5 h-3.5 inline-icon" />
      </button>
    </div>

    <!-- Tab Bar -->
    <div class="px-3 pt-2 bg-[#060c1c] border-b border-[#38bdf8]/20 flex items-center justify-between">
      <div class="flex items-center gap-1.5">
        <button
          @click="activeTab = 'vodyanitsa'"
          class="px-3 py-1.5 rounded-t-lg text-xs font-medium transition-all flex items-center gap-1.5 cursor-pointer border-t border-x"
          :class="activeTab === 'vodyanitsa'
            ? 'bg-[#0f1f3d] text-cyan-200 font-bold border-[#38bdf8]/40 shadow-xs'
            : 'text-cyan-400/60 hover:text-cyan-200 border-transparent hover:bg-white/5'"
        >
          <img :src="iconWing" alt="" class="w-3 h-3 object-contain inline-block shrink-0" />
          <span>沃雅妮莎 ({{ VODYANITSA_STICKERS.length }})</span>
        </button>

        <button
          @click="activeTab = 'custom'"
          class="px-3 py-1.5 rounded-t-lg text-xs font-medium transition-all flex items-center gap-1.5 cursor-pointer border-t border-x"
          :class="activeTab === 'custom'
            ? 'bg-[#0f1f3d] text-cyan-200 font-bold border-[#38bdf8]/40 shadow-xs'
            : 'text-cyan-400/60 hover:text-cyan-200 border-transparent hover:bg-white/5'"
        >
          <FolderPlus class="w-3 h-3 text-cyan-400 inline-icon" />
          <span>自定义表情 ({{ customStore.stickers.length }})</span>
        </button>
      </div>

      <!-- Action buttons for Custom Tab -->
      <div v-if="activeTab === 'custom'" class="flex items-center gap-1.5 pb-1.5">
        <button
          @click="triggerAddSticker"
          class="px-2 py-1 rounded bg-cyan-600/30 hover:bg-cyan-500/40 text-cyan-200 hover:text-white border border-cyan-400/40 text-[11px] font-medium flex items-center gap-1 transition-all cursor-pointer"
          title="从本地添加表情包图片 (支持 PNG/JPG/GIF/WebP)"
        >
          <Plus class="w-3 h-3 inline-icon" />
          <span>添加</span>
        </button>
        <button
          v-if="customStore.stickers.length > 0"
          @click="isManaging = !isManaging"
          class="px-2 py-1 rounded text-[11px] font-medium transition-all cursor-pointer flex items-center gap-1"
          :class="isManaging
            ? 'bg-rose-600/40 text-rose-200 border border-rose-400/50'
            : 'bg-white/5 text-cyan-300 hover:bg-white/10 border border-cyan-400/20'"
        >
          <span>{{ isManaging ? '完成' : '管理' }}</span>
        </button>
      </div>
    </div>

    <!-- Content Area: Vodyanitsa Official Tab -->
    <div v-if="activeTab === 'vodyanitsa'" class="p-3 max-h-80 overflow-y-auto custom-scrollbar">
      <div class="grid grid-cols-3 gap-2.5">
        <div
          v-for="stk in VODYANITSA_STICKERS"
          :key="stk.id"
          @click="selectSticker(stk)"
          class="group relative flex flex-col items-center p-2 rounded-xl bg-[#0c1833]/80 border border-[#38bdf8]/25 hover:border-cyan-400 hover:bg-[#12234a] hover:scale-105 transition-all cursor-pointer shadow-sm"
          :title="stk.caption"
        >
          <div class="w-20 h-20 flex items-center justify-center overflow-hidden rounded-lg">
            <img
              :src="stk.imageSrc"
              :alt="stk.name"
              class="w-full h-full object-contain filter group-hover:drop-shadow-[0_0_8px_rgba(56,189,248,0.7)] transition-all"
            />
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

    <!-- Content Area: Custom Stickers Tab -->
    <div
      v-else-if="activeTab === 'custom'"
      @dragover.prevent="isDragOver = true"
      @dragleave.prevent="isDragOver = false"
      @drop.prevent="handleDrop"
      class="p-3 max-h-80 overflow-y-auto custom-scrollbar relative min-h-[160px]"
    >
      <!-- Drag & Drop Overlay -->
      <div
        v-if="isDragOver"
        class="absolute inset-2 z-20 rounded-xl border-2 border-dashed border-cyan-400 bg-cyan-950/80 backdrop-blur-sm flex flex-col items-center justify-center text-cyan-200 pointer-events-none gap-1"
      >
        <Plus class="w-8 h-8 animate-bounce inline-icon text-cyan-400" />
        <span class="text-xs font-bold">释放图片以添加至自定义表情</span>
      </div>

      <!-- Empty State -->
      <div
        v-if="customStore.stickers.length === 0"
        class="py-8 flex flex-col items-center justify-center text-center text-cyan-300/60 gap-2.5"
      >
        <div class="w-12 h-12 rounded-2xl bg-cyan-950/40 border border-cyan-500/20 flex items-center justify-center">
          <FolderPlus class="w-6 h-6 text-cyan-400/60 inline-icon" />
        </div>
        <p class="text-xs font-medium">暂无自定义表情</p>
        <p class="text-[11px] text-cyan-400/40 max-w-[240px]">
          点击上方「添加」或直接拖拽图片到此处，也可右键聊天中的图片「添加到表情」
        </p>
        <button
          @click="triggerAddSticker"
          class="mt-1 px-3 py-1.5 rounded-lg bg-cyan-600/30 hover:bg-cyan-500/40 text-cyan-200 border border-cyan-400/40 text-xs font-semibold flex items-center gap-1.5 cursor-pointer transition-all"
        >
          <Plus class="w-3.5 h-3.5 inline-icon" />
          <span>立即添加本地表情</span>
        </button>
      </div>

      <!-- Custom Stickers Grid -->
      <div v-else class="grid grid-cols-4 gap-2">
        <!-- Quick Add Tile -->
        <button
          @click="triggerAddSticker"
          class="aspect-square rounded-xl border border-dashed border-cyan-400/40 hover:border-cyan-300 bg-cyan-950/30 hover:bg-cyan-900/40 flex flex-col items-center justify-center text-cyan-300 gap-1 transition-all cursor-pointer"
          title="添加新表情"
        >
          <Plus class="w-5 h-5 inline-icon" />
          <span class="text-[10px] font-medium">添加</span>
        </button>

        <!-- Sticker Items -->
        <div
          v-for="stk in customStore.stickers"
          :key="stk.id"
          class="group relative aspect-square rounded-xl bg-[#0c1833]/80 border border-[#38bdf8]/25 hover:border-cyan-400 hover:bg-[#12234a] p-1.5 flex flex-col items-center justify-center cursor-pointer transition-all"
          :class="isManaging ? 'animate-pulse' : 'hover:scale-105'"
          @click="handleCustomClick(stk)"
          :title="isManaging ? '点击删除此表情' : (stk.name || '自定义表情')"
        >
          <img
            :src="stk.url"
            :alt="stk.name"
            class="w-full h-full object-contain rounded-lg"
          />

          <!-- Manage / Delete Badge -->
          <button
            v-if="isManaging"
            @click.stop="deleteCustomSticker(stk.id)"
            class="absolute -top-1.5 -right-1.5 w-5 h-5 rounded-full bg-rose-600 text-white flex items-center justify-center shadow-md hover:bg-rose-500 transition-colors z-10 cursor-pointer"
            title="删除表情"
          >
            <X class="w-3 h-3 inline-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- Footer Note -->
    <div class="px-3 py-2 bg-[#060c1c]/90 border-t border-[#38bdf8]/20 flex items-center justify-between text-[11px] text-[#7dd3fc]/80">
      <span>{{ activeTab === 'vodyanitsa' ? '点击直接发送至当前剧场会话' : (isManaging ? '管理模式：点击右上角红色按钮删除表情' : '点击发送自定义表情 · 支持本地收藏') }}</span>
      <span class="text-cyan-400 font-mono">✦ 科洛列夫茨基</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { X, Plus, FolderPlus } from 'lucide-vue-next';
import { VODYANITSA_STICKERS, VodyanitsaSticker } from '../assets/vodyanitsa/vodyanitsaStickers';
import { useCustomStickersStore, CustomSticker } from '../stores/customStickers';
import iconLily from '../assets/vodyanitsa/dock/nav_contacts_lily.png';
import iconWing from '../assets/vodyanitsa/dock/nav_chats_wave_wing.png';

const emit = defineEmits<{
  (e: 'select', code: string): void;
  (e: 'select-custom', sticker: CustomSticker): void;
  (e: 'close'): void;
}>();

const customStore = useCustomStickersStore();
const activeTab = ref<'vodyanitsa' | 'custom'>('vodyanitsa');
const isManaging = ref(false);
const isDragOver = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

const handleClose = () => {
  emit('close');
};

const selectSticker = (stk: VodyanitsaSticker) => {
  emit('select', `[VODYANITSA_STICKER]:${stk.id}`);
  emit('close');
};

const handleCustomClick = (stk: CustomSticker) => {
  if (isManaging.value) {
    deleteCustomSticker(stk.id);
  } else {
    emit('select-custom', stk);
    emit('close');
  }
};

const deleteCustomSticker = (id: string) => {
  customStore.deleteSticker(id);
  if (customStore.stickers.length === 0) {
    isManaging.value = false;
  }
};

const triggerAddSticker = () => {
  fileInputRef.value?.click();
};

const readFileAsDataUrl = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as string);
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
};

const processFiles = async (files: FileList | File[]) => {
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    if (!file.type.startsWith('image/')) continue;
    try {
      const dataUrl = await readFileAsDataUrl(file);
      const name = file.name.replace(/\.[^/.]+$/, '');
      customStore.addSticker(name, dataUrl);
    } catch (err) {
      console.warn('Failed to read sticker file:', err);
    }
  }
  activeTab.value = 'custom';
};

const handleFileUpload = (e: Event) => {
  const target = e.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    processFiles(target.files);
    target.value = '';
  }
};

const handleDrop = (e: DragEvent) => {
  isDragOver.value = false;
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    processFiles(e.dataTransfer.files);
  }
};
</script>
