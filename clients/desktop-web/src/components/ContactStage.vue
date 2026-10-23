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
          class="px-3 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium flex items-center gap-1.5 transition-colors shadow-sm"
        >
          <Users class="w-3.5 h-3.5 inline-icon" />
          <span>{{ langStore.t('contacts.newGroup') }}</span>
        </button>
      </div>
    </header>

    <!-- Main Content Grid -->
    <div class="flex-1 overflow-y-auto p-6 space-y-6">
      <!-- Section: Groups & Teams -->
      <div>
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-xs font-semibold text-mim-light-muted dark:text-mim-dark-muted uppercase tracking-wider">
            {{ langStore.t('contacts.teamGroups') }} (1)
          </h3>
        </div>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div
            @click="openGroupChat('group_101')"
            class="p-4 rounded-xl bg-mim-light-card dark:bg-mim-dark-card border border-black/5 dark:border-white/5 hover:border-blue-500/30 transition-all cursor-pointer flex items-center justify-between group shadow-sm"
          >
            <div class="flex items-center gap-3">
              <div class="w-11 h-11 rounded-xl bg-indigo-600/15 text-indigo-600 dark:text-indigo-400 font-semibold flex items-center justify-center text-sm">
                CI
              </div>
              <div>
                <h4 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
                  {{ langStore.currentLang === 'zh' ? '核心基础设施架构组' : 'Core Infrastructure Team' }}
                </h4>
                <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted">
                  4 {{ langStore.t('contacts.membersCount') }} · v1.4.2 canary deployed
                </p>
              </div>
            </div>
            <MessageSquare class="w-4 h-4 text-mim-light-muted dark:text-mim-dark-muted group-hover:text-blue-600 transition-colors inline-icon" />
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
                <div class="w-11 h-11 rounded-xl bg-blue-600/15 text-blue-600 dark:text-blue-400 font-semibold flex items-center justify-center text-sm">
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
                class="px-2.5 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-blue-600 hover:text-white text-xs font-medium transition-colors flex items-center gap-1"
                :title="langStore.t('contacts.chatBtn')"
              >
                <MessageSquare class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.t('contacts.chatBtn') }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Friend Modal -->
    <div
      v-if="showAddFriend"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <div class="w-full max-w-md rounded-2xl bg-mim-light-card dark:bg-mim-dark-card p-6 shadow-2xl border border-black/5 dark:border-white/10 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">{{ langStore.t('contacts.addFriendModalTitle') }}</h3>
          <button @click="showAddFriend = false" class="text-mim-light-muted hover:text-mim-light-text">
            <X class="w-4 h-4 inline-icon" />
          </button>
        </div>

        <div class="relative">
          <input
            v-model="searchKeyword"
            @keyup.enter="handleSearch"
            type="text"
            :placeholder="langStore.t('contacts.searchPlaceholder')"
            class="mim-input w-full h-10 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text border-none outline-none"
          />
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
            <button
              @click="sendFriendRequest(u.id)"
              class="px-3 h-7 rounded bg-blue-600 text-white text-xs font-medium hover:bg-blue-500"
            >
              {{ langStore.t('contacts.addBtn') }}
            </button>
          </div>
        </div>

        <div v-else-if="searched" class="text-center py-4 text-xs text-mim-light-muted">
          {{ langStore.t('contacts.noUserFound') }}
        </div>

        <div class="flex justify-end pt-2">
          <button
            @click="handleSearch"
            class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium"
          >
            {{ langStore.t('contacts.searchBtn') }}
          </button>
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
            class="px-4 h-8 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium"
          >
            {{ langStore.t('contacts.create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { UserPlus, Users, MessageSquare, X } from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';

const chatStore = useChatStore();
const langStore = useLangStore();
const friends = ref<any[]>([]);

const showAddFriend = ref(false);
const showCreateGroup = ref(false);
const searchKeyword = ref('');
const searchResults = ref<any[]>([]);
const searched = ref(false);

const newGroupTitle = ref('');
const newGroupNotice = ref('');

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
    await fetch(`/api/user/add-friend?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({ target_id: targetId, msg: 'Hello, let us connect on Mim!' }),
    });
    showAddFriend.value = false;
    fetchFriends();
  } catch (_) {}
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
        member_ids: [2, 3],
      }),
    });
    showCreateGroup.value = false;
    newGroupTitle.value = '';
    newGroupNotice.value = '';
    chatStore.fetchConversations();
  } catch (_) {}
};

const openDirectChat = (friendId: number) => {
  const u1 = Math.min(chatStore.currentUserId, friendId);
  const u2 = Math.max(chatStore.currentUserId, friendId);
  const convId = `${u1}_${u2}`;
  chatStore.selectConversation(convId);
  chatStore.activeTab = 'chats';
};

const openGroupChat = (convId: string) => {
  chatStore.selectConversation(convId);
  chatStore.activeTab = 'chats';
};

onMounted(() => {
  fetchFriends();
});
</script>
