<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-md rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col max-h-[80vh]">
      <!-- Header Tabs -->
      <div class="p-4 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <button
            @click="activeTab = 'group'"
            class="text-sm font-semibold transition-colors pb-1 relative"
            :class="activeTab === 'group' ? 'text-mim-accent' : 'text-mim-light-muted hover:text-mim-light-text'"
          >
            发起群聊
            <span v-if="activeTab === 'group'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-mim-accent rounded-full"></span>
          </button>
          <button
            @click="activeTab = 'single'"
            class="text-sm font-semibold transition-colors pb-1 relative"
            :class="activeTab === 'single' ? 'text-mim-accent' : 'text-mim-light-muted hover:text-mim-light-text'"
          >
            发起单聊
            <span v-if="activeTab === 'single'" class="absolute bottom-0 left-0 right-0 h-0.5 bg-mim-accent rounded-full"></span>
          </button>
        </div>
        <button @click="$emit('close')" class="text-mim-light-muted hover:text-mim-light-text">
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Tab: 发起群聊 -->
      <div v-if="activeTab === 'group'" class="p-5 flex-1 overflow-y-auto space-y-4">
        <div>
          <label class="block text-xs font-medium text-mim-light-muted mb-1">群聊名称</label>
          <input
            v-model="groupTitle"
            type="text"
            placeholder="例如：分布式架构核心组"
            class="mim-input w-full h-9 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm"
          />
        </div>
        <div>
          <label class="block text-xs font-medium text-mim-light-muted mb-1">群公告 (选填)</label>
          <input
            v-model="groupNotice"
            type="text"
            placeholder="欢迎各位成员入群"
            class="mim-input w-full h-9 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-mim-light-muted mb-1.5">选择群成员 ({{ selectedFriendIds.length }})</label>
          <div class="space-y-1.5 max-h-48 overflow-y-auto pr-1">
            <label
              v-for="friend in friends"
              :key="friend.friend_id"
              class="flex items-center gap-3 p-2 rounded-xl hover:bg-black/[0.03] dark:hover:bg-white/[0.04] cursor-pointer transition-colors group select-none"
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
              <div class="min-w-0">
                <span class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text block truncate">{{ friend.nickname }}</span>
                <span class="text-[10px] text-mim-light-muted font-mono">@{{ friend.account }}</span>
              </div>
            </label>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t border-black/5 dark:border-white/5">
          <button
            @click="$emit('close')"
            class="px-4 h-8 rounded-lg bg-black/5 dark:bg-white/5 text-xs font-medium text-mim-light-muted"
          >
            取消
          </button>
          <button
            @click="handleCreateGroup"
            :disabled="!groupTitle.trim() || selectedFriendIds.length === 0"
            class="px-5 h-8 rounded-lg mim-btn-primary disabled:opacity-40 text-xs font-medium shadow-sm transition-colors cursor-pointer disabled:cursor-not-allowed"
          >
            立即创建
          </button>
        </div>
      </div>

      <!-- Tab: 发起单聊 -->
      <div v-else class="p-5 flex-1 overflow-y-auto space-y-3">
        <span class="text-xs font-medium text-mim-light-muted block">选择好友快速发起单聊：</span>
        <div class="space-y-1 max-h-64 overflow-y-auto">
          <div
            v-for="friend in friends"
            :key="friend.friend_id"
            @click="startDirectChat(friend.friend_id)"
            class="p-2.5 rounded-xl hover:bg-black/[0.04] dark:hover:bg-white/[0.05] cursor-pointer flex items-center justify-between transition-colors"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="w-9 h-9 rounded-lg font-bold text-xs flex items-center justify-center shrink-0 shadow-xs ring-1 ring-black/10 dark:ring-white/10"
                :style="{ backgroundColor: getAvatarColor(friend.friend_id || friend.account, false).bg, color: getAvatarColor(friend.friend_id || friend.account, false).text }"
              >
                {{ friend.avatar || friend.nickname.substring(0, 2).toUpperCase() }}
              </div>
              <div class="min-w-0">
                <span class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text block truncate">{{ friend.nickname }}</span>
                <span class="text-[10px] text-mim-light-muted font-mono">@{{ friend.account }} (ID: {{ friend.friend_id }})</span>
              </div>
            </div>
            <MessageSquare class="w-4 h-4 text-mim-light-muted hover:text-mim-accent inline-icon" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { X, MessageSquare, Check } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { getAvatarColor } from '../utils/avatar';

const emit = defineEmits(['close']);
const router = useRouter();
const chatStore = useChatStore();

const activeTab = ref<'group' | 'single'>('group');
const friends = ref<any[]>([]);
const selectedFriendIds = ref<number[]>([]);
const groupTitle = ref('');
const groupNotice = ref('');

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

const handleCreateGroup = async () => {
  if (!groupTitle.value.trim() || selectedFriendIds.value.length === 0) return;
  try {
    const res = await fetch(`/api/group/create?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        title: groupTitle.value,
        notice: groupNotice.value,
        member_ids: selectedFriendIds.value,
      }),
    });
    const json = await res.json();
    if (json.code === 0 && json.data) {
      const convId = `group_${json.data.id}`;
      await chatStore.fetchConversations();
      chatStore.selectConversation(convId);
      router.push(`/chats/${convId}`);
      emit('close');
    }
  } catch (_) {}
};

const startDirectChat = (friendId: number) => {
  const u1 = Math.min(chatStore.currentUserId, friendId);
  const u2 = Math.max(chatStore.currentUserId, friendId);
  const convId = `${u1}_${u2}`;
  chatStore.selectConversation(convId);
  router.push(`/chats/${convId}`);
  emit('close');
};

onMounted(() => {
  fetchFriends();
});
</script>
