<template>
  <div class="flex-1 h-full flex flex-col bg-mim-light-stage dark:bg-mim-dark-stage select-none">
    <!-- Header -->
    <header class="h-16 px-6 flex items-center justify-between border-b border-black/[0.04] dark:border-white/[0.06] bg-mim-light-stage/80 dark:bg-mim-dark-stage/80 backdrop-blur-md">
      <div>
        <h2 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('contacts.title') }}</h2>
        <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted">{{ langStore.t('contacts.subtitle') }}</p>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="showAddFriend = true"
          class="px-3 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium text-mim-light-text dark:text-mim-dark-text flex items-center gap-1.5 transition-colors"
        >
          <UserPlus class="w-3.5 h-3.5 inline-icon" />
          <span>{{ langStore.t('contacts.addFriend') }}</span>
        </button>
        <button
          @click="showCreateGroup = true"
          class="px-3 h-8 rounded-lg mim-btn-primary text-xs font-medium flex items-center gap-1.5 shadow-sm"
        >
          <Users class="w-3.5 h-3.5 inline-icon" />
          <span>{{ langStore.t('contacts.newGroup') }}</span>
        </button>
      </div>
    </header>

    <!-- Main Content Grid -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6">
      <!-- Section: Friend Requests (好友申请) -->
      <div v-if="friendRequests.length > 0">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <h3 class="text-xs font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
              {{ langStore.currentLang === 'zh' ? '好友申请 / 新的朋友' : 'Friend Requests' }}
            </h3>
            <span
              v-if="pendingRequestCount > 0"
              class="px-2 py-0.5 rounded-full bg-rose-500 text-white text-[10px] font-bold leading-none shadow-xs animate-pulse"
            >
              {{ pendingRequestCount }} 条待处理
            </span>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div
            v-for="req in friendRequests"
            :key="req.id"
            class="p-4 rounded-xl bg-mim-light-card dark:bg-mim-dark-card border border-black/5 dark:border-white/5 flex items-center justify-between shadow-sm"
          >
            <div class="flex items-center gap-3 min-w-0">
              <div
                class="w-11 h-11 rounded-xl font-bold flex items-center justify-center text-sm shadow-xs shrink-0"
                :style="{ backgroundColor: getAvatarColor(req.sender_id || req.account).bg, color: getAvatarColor(req.sender_id || req.account).text }"
              >
                {{ req.avatar || req.nickname.substring(0, 2).toUpperCase() }}
              </div>
              <div class="min-w-0">
                <div class="flex items-center gap-2">
                  <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text truncate">
                    {{ req.nickname }}
                  </h4>
                  <span class="text-[10px] text-mim-light-muted dark:text-mim-dark-muted font-mono shrink-0">
                    @{{ req.account }}
                  </span>
                </div>
                <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted truncate mt-0.5">
                  {{ req.verify_msg || (langStore.currentLang === 'zh' ? '请求添加你为好友' : 'Wants to add you as friend') }}
                </p>
                <div class="text-[10px] text-mim-light-muted/70 font-mono mt-0.5">
                  {{ req.created_at }}
                </div>
              </div>
            </div>

            <div class="flex items-center gap-1.5 shrink-0 ml-3">
              <template v-if="req.status === 1">
                <button
                  @click="handleProcessRequest(req.id, 2)"
                  class="px-3 h-7 rounded-lg mim-btn-primary text-xs font-medium shadow-sm cursor-pointer"
                >
                  {{ langStore.currentLang === 'zh' ? '通过' : 'Accept' }}
                </button>
                <button
                  @click="handleProcessRequest(req.id, 3)"
                  class="px-2.5 h-7 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium text-mim-light-muted hover:text-mim-light-text transition-colors cursor-pointer"
                >
                  {{ langStore.currentLang === 'zh' ? '拒绝' : 'Decline' }}
                </button>
              </template>
              <template v-else-if="req.status === 2">
                <span class="text-xs text-emerald-600 dark:text-emerald-400 font-medium px-2 py-1 rounded bg-emerald-500/10">
                  {{ langStore.currentLang === 'zh' ? '已同意' : 'Accepted' }}
                </span>
                <button
                  @click="openDirectChat(req.sender_id)"
                  class="px-2.5 h-7 rounded-lg mim-btn-primary text-xs font-medium ml-1"
                >
                  <MessageSquare class="w-3 h-3 inline-icon mr-1" />
                  <span>{{ langStore.currentLang === 'zh' ? '发消息' : 'Chat' }}</span>
                </button>
              </template>
              <template v-else>
                <span class="text-xs text-mim-light-muted px-2 py-1 rounded bg-black/5 dark:bg-white/5">
                  {{ langStore.currentLang === 'zh' ? '已拒绝' : 'Declined' }}
                </span>
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Section: Groups & Teams -->
      <div>
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
            {{ langStore.t('contacts.teamGroups') }} ({{ groupList.length }})
          </h3>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div
            v-for="grp in groupList"
            :key="grp.id"
            @click="openGroupChat(grp.id)"
            class="p-4 rounded-xl bg-mim-light-card dark:bg-mim-dark-card border border-black/5 dark:border-white/5 hover:border-mim-accent/30 transition-all cursor-pointer flex items-center justify-between group shadow-sm"
          >
            <div class="flex items-center gap-3">
              <div
                class="w-11 h-11 rounded-xl font-bold flex items-center justify-center text-sm shadow-xs transition-all"
                :class="getStatusRingClass('group')"
                :style="{ backgroundColor: getAvatarColor(grp.id, true).bg, color: getAvatarColor(grp.id, true).text }"
              >
                {{ grp.avatarText || '群' }}
              </div>
              <div>
                <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text group-hover:text-mim-accent transition-colors">
                  {{ grp.title }}
                </h4>
                <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted">
                  {{ grp.lastMessage || '团队协作群组' }}
                </p>
              </div>
            </div>
            <MessageSquare class="w-4 h-4 text-mim-light-muted dark:text-mim-dark-muted group-hover:text-mim-accent transition-colors inline-icon" />
          </div>
        </div>
      </div>

      <!-- Section: Direct Contacts -->
      <div>
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
            {{ langStore.t('contacts.directContacts') }} ({{ friends.length }})
          </h3>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div
            v-for="friend in friends"
            :key="friend.friend_id"
            class="p-4 rounded-xl bg-mim-light-card dark:bg-mim-dark-card border border-black/5 dark:border-white/5 flex items-center justify-between shadow-sm"
          >
            <div class="flex items-center gap-3">
              <div class="relative">
                <div
                  class="w-11 h-11 rounded-xl font-bold flex items-center justify-center text-sm shadow-xs transition-all"
                  :class="getStatusRingClass(friend.is_online ? 'online' : 'offline')"
                  :style="{ backgroundColor: getAvatarColor(friend.friend_id || friend.account).bg, color: getAvatarColor(friend.friend_id || friend.account).text }"
                >
                  {{ friend.avatar || friend.nickname.substring(0, 2).toUpperCase() }}
                </div>
                <span
                  v-if="friend.is_online"
                  class="absolute -bottom-0.5 -right-0.5 w-3 h-3 bg-emerald-500 border-2 border-mim-light-card dark:border-mim-dark-card rounded-full"
                ></span>
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">
                    {{ friend.nickname }}
                  </h4>
                  <span v-if="friend.is_online" class="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono">
                    {{ langStore.t('contacts.online') }}
                  </span>
                </div>
                <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted font-mono">
                  @{{ friend.account }} · ID: {{ friend.friend_id }}
                </p>
              </div>
            </div>

            <div class="flex items-center gap-1">
              <button
                @click="openDirectChat(friend.friend_id)"
                class="px-2.5 h-8 rounded-lg mim-btn-primary text-xs font-medium flex items-center gap-1 shadow-sm"
                :title="langStore.t('contacts.chatBtn')"
              >
                <MessageSquare class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.t('contacts.chatBtn') }}</span>
              </button>
              <button
                @click="selectedContactProfile = friend"
                class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-mim-light-muted hover:text-mim-light-text flex items-center justify-center transition-colors"
                title="查看资料"
              >
                <User class="w-3.5 h-3.5 inline-icon" />
              </button>
              <button
                @click="handleDeleteFriend(friend.friend_id, friend.nickname)"
                class="w-8 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-red-500/10 text-mim-light-muted hover:text-red-500 flex items-center justify-center transition-colors"
                title="删除好友"
              >
                <UserMinus class="w-3.5 h-3.5 inline-icon" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- User Profile Modal -->
    <UserProfileModal
      v-if="selectedContactProfile"
      :user="selectedContactProfile"
      @close="selectedContactProfile = null"
    />

    <!-- Add Friend Modal -->
    <div
      v-if="showAddFriend"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <div class="w-full max-w-md rounded-2xl bg-mim-light-card dark:bg-mim-dark-card p-6 shadow-2xl border border-black/5 dark:border-white/10 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('contacts.addFriendModalTitle') }}</h3>
          <button @click="showAddFriend = false" class="text-mim-light-muted hover:text-mim-light-text cursor-pointer">
            <X class="w-4 h-4 inline-icon" />
          </button>
        </div>

        <div class="flex gap-2">
          <input
            v-model="searchKeyword"
            @keyup.enter="handleSearch"
            type="text"
            :placeholder="langStore.t('contacts.searchPlaceholder')"
            class="mim-input flex-1 h-10 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text border-none outline-none"
          />
          <button
            @click="handleSearch"
            class="px-4 h-10 rounded-lg mim-btn-primary text-xs font-medium cursor-pointer shrink-0"
          >
            {{ langStore.t('contacts.searchBtn') }}
          </button>
        </div>

        <div>
          <label class="block text-xs font-medium text-mim-light-muted mb-1">
            {{ langStore.currentLang === 'zh' ? '申请附言' : 'Verification Message' }}
          </label>
          <input
            v-model="addFriendMsg"
            type="text"
            :placeholder="`我是 ${chatStore.currentUserName || 'Mim用户'}`"
            class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
          />
        </div>

        <div v-if="addFriendStatusMsg" class="p-2.5 rounded-lg bg-mim-accent-subtle/50 text-mim-accent text-xs font-medium text-center">
          {{ addFriendStatusMsg }}
        </div>

        <div v-if="searchResults.length > 0" class="space-y-2 max-h-48 overflow-y-auto">
          <div
            v-for="u in searchResults"
            :key="u.id"
            class="p-2.5 rounded-lg bg-black/[0.02] dark:bg-white/[0.04] flex items-center justify-between"
          >
            <div>
              <div class="text-sm font-medium">{{ u.nickname }}</div>
              <div class="text-xs text-mim-light-muted">@{{ u.account }} (ID: {{ u.id }})</div>
            </div>
            <div>
              <span v-if="u.id === chatStore.currentUserId" class="text-xs text-mim-light-muted px-2.5 py-1 bg-black/5 dark:bg-white/5 rounded">
                自己
              </span>
              <span v-else-if="friends.some(f => f.friend_id === u.id)" class="text-xs text-emerald-600 dark:text-emerald-400 px-2.5 py-1 bg-emerald-500/10 rounded font-medium">
                已是好友
              </span>
              <button
                v-else
                @click="sendFriendRequest(u.id)"
                class="px-3 h-7 rounded mim-btn-primary text-xs font-medium cursor-pointer"
              >
                {{ langStore.t('contacts.addBtn') }}
              </button>
            </div>
          </div>
        </div>

        <div v-else-if="searched" class="text-center py-4 text-xs text-mim-light-muted">
          {{ langStore.t('contacts.noUserFound') }}
        </div>
      </div>
    </div>

    <!-- Create Group Modal -->
    <div
      v-if="showCreateGroup"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <div class="w-full max-w-md rounded-2xl bg-mim-light-card dark:bg-mim-dark-card p-6 shadow-2xl border border-black/5 dark:border-white/10 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('contacts.createGroupModalTitle') }}</h3>
          <button @click="showCreateGroup = false" class="text-mim-light-muted hover:text-mim-light-text">
            <X class="w-4 h-4 inline-icon" />
          </button>
        </div>

        <div class="space-y-3">
          <div>
            <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('contacts.groupTitleLabel') }}</label>
            <input
              v-model="newGroupTitle"
              type="text"
              :placeholder="langStore.t('contacts.groupTitlePlaceholder')"
              class="mim-input w-full h-9 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm"
            />
          </div>
          <div>
            <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('contacts.groupNoticeLabel') }}</label>
            <textarea
              v-model="newGroupNotice"
              rows="2"
              :placeholder="langStore.t('contacts.groupNoticePlaceholder')"
              class="mim-input w-full p-2.5 rounded-lg bg-black/5 dark:bg-white/5 text-sm resize-none"
            ></textarea>
          </div>

          <!-- Select Initial Members from Friends -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="text-xs font-medium text-mim-light-muted">选择初始成员</label>
              <span class="text-[11px] text-mim-accent font-medium">已选 {{ selectedFriendIds.length }} 人</span>
            </div>
            <div class="max-h-36 overflow-y-auto space-y-1 p-2 rounded-lg bg-black/5 dark:bg-white/5 border border-black/5 dark:border-white/5">
              <div
                v-for="f in friends"
                :key="f.friend_id"
                @click="toggleSelectFriend(f.friend_id)"
                class="flex items-center justify-between p-1.5 rounded-md hover:bg-black/5 dark:hover:bg-white/10 cursor-pointer transition-colors"
              >
                <div class="flex items-center gap-2">
                  <div
                    class="w-6 h-6 rounded-md font-bold text-[10px] flex items-center justify-center shadow-xs"
                    :style="{ backgroundColor: getAvatarColor(f.friend_id, false).bg, color: getAvatarColor(f.friend_id, false).text }"
                  >
                    {{ f.avatar || f.nickname.substring(0, 1) }}
                  </div>
                  <span class="text-xs text-mim-light-text dark:text-mim-dark-text font-medium">{{ f.nickname }}</span>
                </div>
                <input
                  type="checkbox"
                  :checked="selectedFriendIds.includes(f.friend_id)"
                  class="rounded text-mim-accent pointer-events-none"
                />
              </div>
              <div v-if="friends.length === 0" class="text-center py-3 text-xs text-mim-light-muted">
                暂无好友可供选择
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button
            @click="showCreateGroup = false"
            class="px-3 h-8 rounded-lg bg-black/5 dark:bg-white/5 text-xs font-medium text-mim-light-muted"
          >
            {{ langStore.t('contacts.cancel') }}
          </button>
          <button
            @click="handleCreateGroup"
            class="px-4 h-8 rounded-lg mim-btn-primary text-xs font-medium"
          >
            {{ langStore.t('contacts.create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { UserPlus, Users, MessageSquare, X, User, UserMinus } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { confirmDialog } from '../stores/dialog';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';
import UserProfileModal from './UserProfileModal.vue';

const router = useRouter();
const chatStore = useChatStore();
const langStore = useLangStore();
const friends = ref<any[]>([]);
const friendRequests = ref<any[]>([]);
const pendingRequestCount = computed(() => friendRequests.value.filter((r) => r.status === 1).length);

const showAddFriend = ref(false);
const showCreateGroup = ref(false);
const searchKeyword = ref('');
const searchResults = ref<any[]>([]);
const searched = ref(false);
const addFriendMsg = ref('');
const addFriendStatusMsg = ref('');
const selectedContactProfile = ref<any | null>(null);

const groupList = computed(() => {
  return chatStore.conversations.filter((c) => c.type === 'group');
});

const newGroupTitle = ref('');
const newGroupNotice = ref('');
const selectedFriendIds = ref<number[]>([]);

const toggleSelectFriend = (id: number) => {
  if (selectedFriendIds.value.includes(id)) {
    selectedFriendIds.value = selectedFriendIds.value.filter((x) => x !== id);
  } else {
    selectedFriendIds.value.push(id);
  }
};

const fetchFriends = async () => {
  if (!chatStore.currentUserId || !chatStore.userToken) return;
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

const fetchFriendRequests = async () => {
  if (!chatStore.currentUserId || !chatStore.userToken) return;
  try {
    const res = await fetch(`/api/user/friend-requests?user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0 && Array.isArray(json.data)) {
      friendRequests.value = json.data;
    }
  } catch (_) {}
};

const handleProcessRequest = async (requestId: number, action: number) => {
  try {
    const res = await fetch(`/api/user/handle-friend-request?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({ request_id: requestId, action }),
    });
    const json = await res.json();
    if (json.code === 0) {
      await fetchFriendRequests();
      await fetchFriends();
      await chatStore.fetchConversations();
    }
  } catch (_) {}
};

const handleSearch = async () => {
  if (!searchKeyword.value.trim()) return;
  try {
    const res = await fetch(`/api/user/search?keyword=${encodeURIComponent(searchKeyword.value)}&user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0) {
      searchResults.value = json.data || [];
      searched.value = true;
    }
  } catch (_) {}
};

const sendFriendRequest = async (targetId: number) => {
  try {
    const res = await fetch(`/api/user/add-friend?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        target_id: targetId,
        msg: addFriendMsg.value.trim() || `我是 ${chatStore.currentUserName || 'Mim用户'}`,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      addFriendStatusMsg.value = json.data?.msg || '申请已发送！';
      setTimeout(() => {
        showAddFriend.value = false;
        addFriendStatusMsg.value = '';
      }, 1500);
      fetchFriends();
      fetchFriendRequests();
      chatStore.fetchConversations();
    } else {
      addFriendStatusMsg.value = json.msg || '添加失败';
    }
  } catch (err: any) {
    addFriendStatusMsg.value = err.message || '网络错误';
  }
};

const handleCreateGroup = async () => {
  if (!newGroupTitle.value.trim()) return;
  try {
    await fetch(`/api/group/create?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        title: newGroupTitle.value,
        notice: newGroupNotice.value,
        member_ids: selectedFriendIds.value,
      }),
    });
    showCreateGroup.value = false;
    newGroupTitle.value = '';
    newGroupNotice.value = '';
    selectedFriendIds.value = [];
    chatStore.fetchConversations();
  } catch (_) {}
};

const openDirectChat = (friendId: number) => {
  const u1 = Math.min(chatStore.currentUserId, friendId);
  const u2 = Math.max(chatStore.currentUserId, friendId);
  const convId = `${u1}_${u2}`;
  chatStore.selectConversation(convId);
  router.push(`/chats/${convId}`);
};

const openGroupChat = (convId: string) => {
  chatStore.selectConversation(convId);
  router.push(`/chats/${convId}`);
};

const handleDeleteFriend = async (friendId: number, nickname: string) => {
  const ok = await confirmDialog({
    title: '解除好友关系',
    message: `确定要解除与 ${nickname} 的好友关系吗？`,
    confirmText: '解除好友',
    danger: true,
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/user/delete-friend?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({ friend_id: friendId }),
    });
    const json = await res.json();
    if (json.code === 0) {
      friends.value = friends.value.filter((f) => f.friend_id !== friendId);
    }
  } catch (_) {}
};

watch(
  () => chatStore.currentUserId,
  () => {
    fetchFriends();
    fetchFriendRequests();
  },
  { immediate: true }
);

onMounted(() => {
  fetchFriends();
  fetchFriendRequests();
});
</script>
