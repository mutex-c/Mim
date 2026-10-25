<template>
  <div class="flex items-end gap-2 group/filecard" :class="message.isMe ? 'flex-row-reverse' : 'flex-row'">
    <!-- The File Bubble Card -->
    <div
      @click="handleCardClick"
      class="relative p-3.5 rounded-2xl border transition-all cursor-pointer select-none shadow-2xs hover:shadow-md flex items-center justify-between gap-3.5 min-w-[240px] max-w-[300px]"
      :class="[
        message.isMe
          ? 'bg-white dark:bg-[#23272e] text-mim-light-text dark:text-mim-dark-text border-black/10 dark:border-white/10 rounded-br-sm'
          : 'bg-white dark:bg-[#23272e] text-mim-light-text dark:text-mim-dark-text border-black/10 dark:border-white/10 rounded-bl-sm'
      ]"
    >
      <!-- Left Column: File Info -->
      <div class="flex-1 min-w-0 pr-1 flex flex-col justify-between py-0.5">
        <div
          class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text truncate leading-snug select-text"
          :title="message.fileName || '未知文件'"
        >
          {{ formatMiddleEllipsis(message.fileName || '未知文件', 22) }}
        </div>

        <!-- Subtitle: Size + Download Status -->
        <div class="text-xs text-mim-light-muted dark:text-mim-dark-muted font-normal mt-2.5 flex items-center gap-2 tabular-nums">
          <span>{{ message.fileSize || '文件' }}</span>

          <span v-if="isDownloading" class="text-mim-accent text-[11px] font-medium flex items-center gap-1">
            <Loader2 class="w-3 h-3 animate-spin inline-icon" />
            下载中...
          </span>

          <span v-else-if="isDownloaded" class="text-mim-light-muted/80 dark:text-mim-dark-muted/80 text-[11px]">
            已下载
          </span>
        </div>
      </div>

      <!-- Right Column: Stylized File Document Sheet Icon -->
      <div class="relative w-12 h-14 rounded-lg bg-black/[0.02] dark:bg-white/[0.04] border border-black/[0.08] dark:border-white/[0.08] flex flex-col items-center justify-center shrink-0 overflow-hidden shadow-2xs">
        <!-- Dog-ear top-right corner fold -->
        <div class="absolute top-0 right-0 w-3 h-3 bg-black/10 dark:bg-white/15 rounded-bl-xs"></div>

        <!-- File Type Graphic Emblem -->
        <div class="flex flex-col items-center justify-center transition-transform group-hover/filecard:scale-105">
          <!-- PDF Acrobat Symbol -->
          <svg
            v-if="fileMeta.type === 'pdf'"
            class="w-7 h-7 text-red-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <path d="M9 13h6"/>
            <path d="M9 17h6"/>
            <path d="M9 9h1"/>
          </svg>

          <!-- Word / DOC -->
          <svg
            v-else-if="fileMeta.type === 'word'"
            class="w-7 h-7 text-blue-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <path d="M9 13l2 4 2-4"/>
          </svg>

          <!-- Excel / Sheet -->
          <svg
            v-else-if="fileMeta.type === 'excel'"
            class="w-7 h-7 text-emerald-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <path d="M9 11l6 6"/>
            <path d="M15 11l-6 6"/>
          </svg>

          <!-- PPT / Slides -->
          <svg
            v-else-if="fileMeta.type === 'ppt'"
            class="w-7 h-7 text-amber-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <circle cx="12" cy="14" r="3"/>
          </svg>

          <!-- Compressed Archive (ZIP/RAR) -->
          <svg
            v-else-if="fileMeta.type === 'archive'"
            class="w-7 h-7 text-purple-500"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <circle cx="10" cy="12" r="1"/>
            <circle cx="10" cy="16" r="1"/>
          </svg>

          <!-- General File Fallback -->
          <FileText v-else class="w-6 h-6 text-slate-400 inline-icon" />

          <!-- Tag Badge -->
          <span
            class="text-[7.5px] font-black uppercase font-mono tracking-tighter -mt-0.5 leading-none"
            :style="{ color: fileMeta.color }"
          >
            {{ fileMeta.label }}
          </span>
        </div>

        <!-- Download Overlay (Shown when NOT downloaded, exactly matching Image 1) -->
        <div
          v-if="!isDownloaded"
          class="absolute inset-0 bg-black/40 dark:bg-black/55 backdrop-blur-[0.5px] flex items-center justify-center transition-all group-hover/filecard:bg-black/50"
          title="点击下载文件"
        >
          <!-- Circular Badge with Down Arrow (Image 1) -->
          <div class="w-7 h-7 rounded-full bg-black/55 border border-white/50 flex items-center justify-center text-white shadow-sm transition-transform group-hover/filecard:scale-110">
            <Loader2 v-if="isDownloading" class="w-4 h-4 animate-spin text-white inline-icon" />
            <ArrowDown v-else class="w-4 h-4 text-white inline-icon stroke-[2.5]" />
          </div>
        </div>
      </div>
    </div>

    <!-- Folder Button (Shown ONLY when downloaded, exactly matching Image 2) -->
    <button
      v-if="isDownloaded"
      type="button"
      @click.stop="handleOpenFolder"
      class="w-7 h-7 rounded-lg border border-black/10 dark:border-white/10 bg-white/90 dark:bg-[#23272e] hover:bg-black/5 dark:hover:bg-white/10 text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text shadow-2xs flex items-center justify-center transition-all shrink-0 cursor-pointer mb-1 hover:scale-105"
      title="打开对应文件夹"
    >
      <Folder class="w-3.5 h-3.5 inline-icon" />
    </button>
  </div>
</template>

<script lang="ts">
export const formatMiddleEllipsis = (name: string, maxLen = 22): string => {
  if (!name || name.length <= maxLen) return name;
  const lastDot = name.lastIndexOf('.');
  const ext = lastDot > 0 ? name.slice(lastDot) : '';
  const base = lastDot > 0 ? name.slice(0, lastDot) : name;

  const maxBaseLen = Math.max(4, maxLen - ext.length - 3);
  const leftLen = Math.ceil(maxBaseLen / 2);
  const rightLen = Math.floor(maxBaseLen / 2);

  if (base.length <= maxBaseLen) return base + '...' + ext;
  return base.slice(0, leftLen) + '...' + base.slice(-rightLen) + ext;
};
</script>

<script setup lang="ts">
import { ref, computed } from 'vue';
import {
  FileText,
  ArrowDown,
  Loader2,
  Folder,
} from 'lucide-vue-next';
import { MessageItem } from '../stores/chat';
import { alertDialog } from '../stores/dialog';
import { resolveApiUrl } from '../utils/api';

const props = defineProps<{
  message: MessageItem;
}>();

const isDownloading = ref(false);

const getSavedPathMap = (): Record<string, string> => {
  try {
    const raw = localStorage.getItem('mim_file_paths');
    if (raw) return JSON.parse(raw);
  } catch (_) {}
  return {};
};

const saveFilePath = (id: string, path: string) => {
  try {
    const map = getSavedPathMap();
    map[id] = path;
    localStorage.setItem('mim_file_paths', JSON.stringify(map));
  } catch (_) {}
};

// Local persistence for downloaded file messages
const getDownloadedSet = (): Set<string> => {
  try {
    const raw = localStorage.getItem('mim_downloaded_files');
    if (raw) return new Set(JSON.parse(raw));
  } catch (_) {}
  return new Set();
};

const saveDownloadedId = (id: string) => {
  try {
    const s = getDownloadedSet();
    s.add(id);
    localStorage.setItem('mim_downloaded_files', JSON.stringify([...s]));
  } catch (_) {}
};

const downloadedSet = ref<Set<string>>(getDownloadedSet());

// A file is considered downloaded if:
// 1. Current user is the sender (msg.isMe), since it originated on this local machine
// 2. OR it was downloaded previously and exists in downloadedSet
const isDownloaded = computed(() => {
  return props.message.isMe || downloadedSet.value.has(props.message.id);
});

// File Meta parser
const fileMeta = computed(() => {
  const name = props.message.fileName || '';
  const ext = name.split('.').pop()?.toLowerCase() || '';

  if (ext === 'pdf') {
    return { type: 'pdf', label: 'PDF', color: '#ef4444' };
  }
  if (['doc', 'docx'].includes(ext)) {
    return { type: 'word', label: 'DOC', color: '#3b82f6' };
  }
  if (['xls', 'xlsx', 'csv'].includes(ext)) {
    return { type: 'excel', label: 'XLS', color: '#10b981' };
  }
  if (['ppt', 'pptx'].includes(ext)) {
    return { type: 'ppt', label: 'PPT', color: '#f59e0b' };
  }
  if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
    return { type: 'archive', label: 'ZIP', color: '#a855f7' };
  }
  return {
    type: 'file',
    label: ext ? ext.substring(0, 4).toUpperCase() : 'FILE',
    color: '#94a3b8',
  };
});

// Trigger file download
const startDownload = async () => {
  if (isDownloading.value) return;
  isDownloading.value = true;

  const downloadUrl = resolveApiUrl(props.message.content);
  const targetFileName = props.message.fileName || 'file';

  // 1. In Tauri Desktop: use native download_file command directly into Downloads
  if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
    try {
      const { invoke } = await import('@tauri-apps/api/core');
      const savedPath = await invoke<string>('download_file', {
        url: downloadUrl,
        fileName: targetFileName,
      });
      saveDownloadedId(props.message.id);
      saveFilePath(props.message.id, savedPath);
      downloadedSet.value = getDownloadedSet();
      isDownloading.value = false;
      return;
    } catch (e) {
      console.warn('Tauri download_file fallback to web fetch:', e);
    }
  }

  // 2. Browser fallback
  try {
    const res = await fetch(downloadUrl);
    if (!res.ok) throw new Error('Download failed');
    const blob = await res.blob();
    const blobUrl = window.URL.createObjectURL(blob);

    const a = document.createElement('a');
    a.href = blobUrl;
    a.download = targetFileName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(blobUrl);

    // Save state
    saveDownloadedId(props.message.id);
    downloadedSet.value = getDownloadedSet();
  } catch (err) {
    // Fallback: direct browser anchor click
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = targetFileName;
    a.target = '_blank';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);

    saveDownloadedId(props.message.id);
    downloadedSet.value = getDownloadedSet();
  } finally {
    isDownloading.value = false;
  }
};

// Handle clicking on the card
const handleCardClick = async () => {
  if (!isDownloaded.value) {
    await startDownload();
  } else {
    // Already downloaded: open/preview the file with native association in desktop
    if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
      try {
        const { invoke } = await import('@tauri-apps/api/core');
        const savedPaths = getSavedPathMap();
        const path = props.message.filePath || savedPaths[props.message.id] || null;
        await invoke('open_file', {
          filePath: path,
          fileName: props.message.fileName || null,
        });
        return;
      } catch (e) {
        console.warn('open_file fallback to browser open:', e);
      }
    }
    window.open(resolveApiUrl(props.message.content), '_blank');
  }
};

// Handle clicking "Open Folder" button
const handleOpenFolder = async () => {
  // 1. In Tauri Desktop Environment
  if (typeof window !== 'undefined' && '__TAURI_INTERNALS__' in window) {
    try {
      const { invoke } = await import('@tauri-apps/api/core');
      const savedPaths = getSavedPathMap();
      const path = props.message.filePath || savedPaths[props.message.id] || null;
      await invoke('open_downloads_folder', {
        filePath: path,
        fileName: props.message.fileName || null,
      });
      return;
    } catch (e: any) {
      console.warn('Tauri open_downloads_folder call error:', e);
      alertDialog({
        title: '打开文件目录',
        message: `检测到桌面客户端底层服务已升级。\n请重新启动 Mim 桌面客户端即可直接唤起 Windows 资源管理器并定位 "${props.message.fileName || '文件'}"。`,
        type: 'info',
      });
      return;
    }
  }

  // 2. Web browser environment
  // Open the file directly and show information dialog
  window.open(resolveApiUrl(props.message.content), '_blank');
  alertDialog({
    title: '打开文件目录',
    message: `文件 "${props.message.fileName || '文件'}" 已保存至系统默认【下载】文件夹 (Downloads)。\n已为您在新标签页中打开文件。`,
    type: 'info',
  });
};
</script>
