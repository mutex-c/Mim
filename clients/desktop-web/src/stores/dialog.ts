import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface DialogOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'warning' | 'danger' | 'info' | 'success';
  danger?: boolean;
}

export const useDialogStore = defineStore('dialog', () => {
  const visible = ref(false);
  const isConfirm = ref(true);
  const title = ref('操作确认');
  const message = ref('');
  const confirmText = ref('确定');
  const cancelText = ref('取消');
  const type = ref<'warning' | 'danger' | 'info' | 'success'>('info');
  const isDanger = ref(false);

  let resolvePromise: ((val: boolean) => void) | null = null;

  const showConfirm = (options: string | DialogOptions, extraOpts?: Partial<DialogOptions>): Promise<boolean> => {
    return new Promise((resolve) => {
      resolvePromise = resolve;
      isConfirm.value = true;
      if (typeof options === 'string') {
        message.value = options;
        title.value = extraOpts?.title || '操作确认';
        confirmText.value = extraOpts?.confirmText || '确定';
        cancelText.value = extraOpts?.cancelText || '取消';
        type.value = extraOpts?.type || (extraOpts?.danger ? 'danger' : 'info');
        isDanger.value = !!extraOpts?.danger;
      } else {
        message.value = options.message;
        title.value = options.title || '操作确认';
        confirmText.value = options.confirmText || '确定';
        cancelText.value = options.cancelText || '取消';
        type.value = options.type || (options.danger ? 'danger' : 'info');
        isDanger.value = !!options.danger;
      }
      visible.value = true;
    });
  };

  const showAlert = (options: string | DialogOptions, extraOpts?: Partial<DialogOptions>): Promise<boolean> => {
    return new Promise((resolve) => {
      resolvePromise = resolve;
      isConfirm.value = false;
      if (typeof options === 'string') {
        message.value = options;
        title.value = extraOpts?.title || '提示';
        confirmText.value = extraOpts?.confirmText || '知道了';
        type.value = extraOpts?.type || 'info';
        isDanger.value = !!extraOpts?.danger;
      } else {
        message.value = options.message;
        title.value = options.title || '提示';
        confirmText.value = options.confirmText || '知道了';
        type.value = options.type || 'info';
        isDanger.value = !!options.danger;
      }
      visible.value = true;
    });
  };

  const onConfirm = () => {
    visible.value = false;
    if (resolvePromise) {
      resolvePromise(true);
      resolvePromise = null;
    }
  };

  const onCancel = () => {
    visible.value = false;
    if (resolvePromise) {
      resolvePromise(false);
      resolvePromise = null;
    }
  };

  return {
    visible,
    isConfirm,
    title,
    message,
    confirmText,
    cancelText,
    type,
    isDanger,
    showConfirm,
    showAlert,
    onConfirm,
    onCancel,
  };
});

export function confirmDialog(options: string | DialogOptions, extraOpts?: Partial<DialogOptions>): Promise<boolean> {
  const store = useDialogStore();
  return store.showConfirm(options, extraOpts);
}

export function alertDialog(options: string | DialogOptions, extraOpts?: Partial<DialogOptions>): Promise<boolean> {
  const store = useDialogStore();
  return store.showAlert(options, extraOpts);
}
