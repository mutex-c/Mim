<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-lg rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col">
      <!-- Header (Aligned with ui_invite.png) -->
      <div class="p-5 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
        <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">
          {{ langStore.t('group.inviteModalTitle') }}
        </h3>
        <button
          @click="$emit('close')"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Two-column Body (ui_invite.png) -->
      <div class="grid grid-cols-2 divide-x divide-black/5 dark:divide-white/5 min-h-[300px] max-h-[420px] overflow-hidden">
        <!-- Left: 我的好友 -->
        <div class="p-4 flex flex-col gap-2 overflow-y-auto">
          <span class="text-xs font-semibold text-mim-light-muted uppercase tracking-wider block mb-1">
            {{ langStore.t('group.myFriends') }}
          </span>

          <div v-if="availableFriends.length === 0" class="text-xs text-mim-light-muted py-8 text-center">
            暂无可邀请的好友
          </div>

          <label
            v-for="friend in availableFriends"
            :key="friend.friend_id"
            class="flex items-center gap-2.5 p-2 rounded-xl hover:bg-black/[0.03] dark:hover:bg-white/[0.04] cursor-pointer transition-colors group select-none"
          >
            <input
              type="checkbox"
              :value="friend.friend_id"
              v-model="selectedFriendIds"
              class="sr-only"
            />
            <div
              class="w-4 h-4 rounded border flex items-center justify-center transition-all shrink-0"
              :class="selectedFriendIds.includes(friend.friend_id)
                ? 'border-mim-accent bg-mim-accent text-white shadow-xs'
                : 'border-black/25 dark:border-white/25 bg-black/[0.02] dark:bg-white/[0.04] group-hover:border-mim-accent'"
            >
              <Check v-if="selectedFriendIds.includes(friend.friend_id)" class="w-3 h-3 stroke-[3]" />
            </div>
            <div
              class="w-8 h-8 rounded-lg font-bold text-xs flex items-center justify-center shrink-0 shadow-xs ring-1 ring-black/10 dark:ring-white/10"
              :style="{ backgroundColor: getAvatarColor(friend.friend_id || friend.account, false).bg, color: getAvatarColor(friend.friend_id || friend.account, false).text }"
            >
              {{ friend.avatar || friend.nickname.substring(0, 2).toUpperCase() }}
            </div>
            <span class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text truncate">
              {{ friend.nickname }}
            </span>
          </label>
        </div>

        <!-- Right: 已选联系人 -->
        <div class="p-4 flex flex-col gap-2 overflow-y-auto">
          <span class="text-xs font-semibold text-mim-light-muted uppercase tracking-wider block mb-1">
            {{ langStore.t('group.selectedContacts') }} ({{ selectedFriendIds.length }})
          </span>

          <div v-if="selectedFriendIds.length === 0" class="text-xs text-mim-light-muted py-8 text-center">
            请从左侧勾选好友
          </div>

          <div
            v-for="friend in selectedFriends"
            :key="friend.friend_id"
            class="flex items-center justify-between p-2 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <div
                class="w-8 h-8 rounded-lg font-bold text-xs flex items-center justify-center shrink-0 shadow-xs"
                :style="{ backgroundColor: getAvatarColor(friend.friend_id || friend.account, false).bg, color: getAvatarColor(friend.friend_id || friend.account, false).text }"
              >
                {{ friend.avatar || friend.nickname.substring(0, 2).toUpperCase() }}
              </div>
              <span class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text truncate">
                {{ friend.nickname }}
              </span>
            </div>
            <button
              @click="removeSelected(friend.friend_id)"
              class="text-mim-light-muted hover:text-red-500 p-1"
            >
              <X class="w-3.5 h-3.5 inline-icon" />
            </button>
          </div>
        </div>
      </div>

      <!-- Footer Buttons (ui_invite.png) -->
      <div class="p-4 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.02] flex justify-end gap-2.5">
        <button
          @click="$emit('close')"
          class="px-4 h-8 rounded-lg border border-black/10 dark:border-white/10 hover:bg-black/5 dark:hover:bg-white/5 text-xs font-medium text-mim-light-text dark:text-mim-dark-text transition-colors"
        >
          {{ langStore.t('group.cancelInvite') }}
        </button>
        <button
          @click="confirmInvite"
          :disabled="selectedFriendIds.length === 0 || isSubmitting"
          class="px-5 h-8 rounded-lg mim-btn-primary disabled:opacity-40 text-xs font-medium shadow-sm cursor-pointer disabled:cursor-not-allowed"
        >
          {{ langStore.t('group.confirmInvite') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { X, Check } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { getAvatarColor } from '../utils/avatar';

const props = defineProps<{
  groupId: number;
  existingMemberIds?: number[];
}>();

const emit = defineEmits(['close', 'invited']);

const chatStore = useChatStore();
const langStore = useLangStore();

const friends = ref<any[]>([]);
const selectedFriendIds = ref<number[]>([]);
const isSubmitting = ref(false);

const availableFriends = computed(() => {
  const existingSet = new Set(props.existingMemberIds || []);
  return friends.value.filter((f) => !existingSet.has(f.friend_id));
});

const selectedFriends = computed(() => {
  const set = new Set(selectedFriendIds.value);
  return friends.value.filter((f) => set.has(f.friend_id));
});

const removeSelected = (id: number) => {
  selectedFriendIds.value = selectedFriendIds.value.filter((x) => x !== id);
};

const fetchFriends = async () => {
  try {
    const res = await fetch(`/api/user/friends?user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0 && Array.isArray(json.data)) {
      friends.value = json.data;
    }
  } catch (_) {}
};

const confirmInvite = async () => {
  if (selectedFriendIds.value.length === 0) return;
  isSubmitting.value = true;
  try {
    const res = await fetch(`/api/group/invite?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        member_ids: selectedFriendIds.value,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      emit('invited');
      emit('close');
      chatStore.fetchConversations();
    }
  } catch (_) {
  } finally {
    isSubmitting.value = false;
  }
};

onMounted(() => {
  fetchFriends();
});
</script>
