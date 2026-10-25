<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-2xl rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col max-h-[85vh]">
      <!-- Header Tabs & My Role Status Bar -->
      <div class="px-6 pt-5 pb-3 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
        <div class="flex items-center gap-6">
          <nav class="flex items-center gap-6">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              @click="activeTab = tab.key"
              class="text-sm pb-2 font-medium transition-all relative"
              :class="activeTab === tab.key
                ? 'text-mim-accent font-semibold'
                : 'text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text'"
            >
              {{ tab.label }}
              <span
                v-if="activeTab === tab.key"
                class="absolute bottom-0 left-0 right-0 h-0.5 bg-mim-accent rounded-full"
              ></span>
            </button>
          </nav>

          <!-- Current User Role Badge -->
          <div class="hidden sm:flex items-center gap-1.5 pb-2">
            <span class="text-[11px] text-mim-light-muted">{{ langStore.t('group.myRole') }}:</span>
            <span
              class="text-[11px] px-2 py-0.5 rounded-full font-semibold flex items-center gap-1 shadow-2xs"
              :class="isOwner
                ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30'
                : isAdmin
                  ? 'bg-indigo-500/15 text-indigo-600 dark:text-indigo-400 border border-indigo-500/30'
                  : 'bg-black/5 dark:bg-white/10 text-mim-light-muted border border-black/5'"
            >
              <Crown v-if="isOwner" class="w-3 h-3 inline-icon text-amber-500" />
              <ShieldCheck v-else-if="isAdmin" class="w-3 h-3 inline-icon text-indigo-500" />
              <User v-else class="w-3 h-3 inline-icon" />
              <span>{{ isOwner ? langStore.t('group.roleOwner') : isAdmin ? langStore.t('group.roleAdmin') : langStore.t('group.roleMember') }}</span>
            </span>
          </div>
        </div>

        <button
          @click="$emit('close')"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Tab Content Area -->
      <div class="p-6 overflow-y-auto flex-1 space-y-6">
        <!-- Toast Notification -->
        <div
          v-if="toastMsg"
          class="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs font-medium flex items-center gap-2"
        >
          <Check class="w-4 h-4 inline-icon shrink-0" />
          <span>{{ toastMsg }}</span>
        </div>

        <!-- TAB 1: 群资料 -->
        <div v-if="activeTab === 'info'" class="space-y-4">
          <!-- Non-admin read-only tip -->
          <div
            v-if="!canManage"
            class="p-3 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 text-xs flex items-center gap-2"
          >
            <Info class="w-4 h-4 inline-icon shrink-0" />
            <span>{{ langStore.t('group.memberReadOnlyTip') }}（{{ langStore.t('group.onlyAdminCanEdit') }}）</span>
          </div>

          <div class="flex items-center gap-4">
            <div
              class="w-14 h-14 rounded-2xl font-bold text-lg flex items-center justify-center shadow-xs transition-all"
              :class="getStatusRingClass('group')"
              :style="{ backgroundColor: getAvatarColor(groupId, true).bg, color: getAvatarColor(groupId, true).text }"
            >
              {{ groupForm.avatar || '群' }}
            </div>
            <div class="flex-1">
              <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('group.groupAvatar') }}</label>
              <input
                v-model="groupForm.avatar"
                :disabled="!canManage"
                type="text"
                placeholder="群头像文字或URL"
                class="mim-input w-full h-8 px-2.5 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text disabled:opacity-70 disabled:cursor-not-allowed"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('group.groupTitle') }}</label>
              <input
                v-model="groupForm.title"
                :disabled="!canManage"
                type="text"
                class="mim-input w-full h-9 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text disabled:opacity-70 disabled:cursor-not-allowed"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('group.groupId') }}</label>
              <input
                :value="groupId"
                disabled
                class="mim-input w-full h-9 px-3 rounded-lg bg-black/[0.02] dark:bg-white/[0.03] text-sm text-mim-light-muted cursor-not-allowed font-mono opacity-80"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('group.notice') }}</label>
            <textarea
              v-model="groupForm.notice"
              :disabled="!canManage"
              rows="3"
              placeholder="填写群公告..."
              class="mim-input w-full p-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text resize-none disabled:opacity-70 disabled:cursor-not-allowed"
            ></textarea>
          </div>

          <div v-if="canManage" class="pt-2">
            <button
              @click="saveGroupInfo"
              class="px-5 h-8 rounded-lg mim-btn-primary text-xs font-medium shadow-sm flex items-center gap-1.5"
            >
              <Save class="w-3.5 h-3.5 inline-icon" />
              <span>保存群资料</span>
            </button>
          </div>
        </div>

        <!-- TAB 2: 群成员 -->
        <div v-else-if="activeTab === 'members'" class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <span class="text-xs text-mim-light-muted">共 {{ members.length }} 名群成员</span>
              <span class="text-[11px] text-mim-light-muted">
                (群主 1 人，管理员 {{ members.filter(m => m.role === 2).length }} 人)
              </span>
            </div>
            <input
              v-model="memberKeyword"
              type="text"
              :placeholder="langStore.t('group.searchMember')"
              class="mim-input w-48 h-8 px-2.5 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
            />
          </div>

          <div class="space-y-2 max-h-80 overflow-y-auto">
            <div
              v-for="m in filteredMembers"
              :key="m.user_id"
              class="p-3 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 flex items-center justify-between gap-3"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="w-9 h-9 rounded-xl font-bold text-xs flex items-center justify-center shadow-xs shrink-0"
                  :class="getStatusRingClass(m.role === 1 ? 'busy' : 'offline')"
                  :style="{ backgroundColor: getAvatarColor(m.user_id || m.nickname, false).bg, color: getAvatarColor(m.user_id || m.nickname, false).text }"
                >
                  {{ m.avatar || m.nickname.substring(0, 2).toUpperCase() }}
                </div>
                <div class="min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text truncate max-w-[120px]">{{ m.nickname }}</span>
                    <!-- Role Pill -->
                    <span
                      class="text-[10px] px-1.5 py-0.2 rounded font-medium flex items-center gap-1"
                      :class="m.role === 1
                        ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 font-semibold border border-amber-500/30'
                        : m.role === 2
                          ? 'bg-indigo-500/15 text-indigo-600 dark:text-indigo-400 font-semibold border border-indigo-500/30'
                          : 'bg-black/5 dark:bg-white/5 text-mim-light-muted'"
                    >
                      <Crown v-if="m.role === 1" class="w-3 h-3 text-amber-500 inline-icon" />
                      <ShieldCheck v-else-if="m.role === 2" class="w-3 h-3 text-indigo-500 inline-icon" />
                      <span>{{ m.role === 1 ? langStore.t('group.roleOwner') : m.role === 2 ? langStore.t('group.roleAdmin') : langStore.t('group.roleMember') }}</span>
                    </span>
                    <span v-if="m.is_muted" class="text-[10px] px-1.5 py-0.2 rounded bg-red-500/10 text-red-500 font-medium">
                      禁言中
                    </span>
                    <span v-if="m.user_id === chatStore.currentUserId" class="text-[10px] px-1.5 py-0.2 rounded bg-emerald-500/10 text-emerald-600 font-medium">
                      自己
                    </span>
                  </div>
                  <span class="text-[11px] text-mim-light-muted font-mono block truncate">@{{ m.account }} (ID: {{ m.user_id }})</span>
                </div>
              </div>

              <!-- Permission Based Member Operations -->
              <div v-if="m.user_id !== chatStore.currentUserId" class="flex items-center gap-1.5 shrink-0 flex-wrap justify-end">
                <!-- OWNER (Role 1) Controls -->
                <template v-if="isOwner">
                  <!-- Set / Revoke Admin -->
                  <button
                    v-if="m.role === 3"
                    @click="setRole(m, 2)"
                    class="px-2 h-6 rounded-md bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-600 text-[11px] font-medium transition-colors"
                    title="将此成员设为群管理员"
                  >
                    设为管理
                  </button>
                  <button
                    v-else-if="m.role === 2"
                    @click="setRole(m, 3)"
                    class="px-2 h-6 rounded-md bg-black/5 dark:bg-white/10 hover:bg-black/10 text-mim-light-muted hover:text-mim-light-text text-[11px] font-medium transition-colors"
                    title="取消其管理员身份"
                  >
                    取消管理
                  </button>

                  <!-- Transfer Ownership -->
                  <button
                    @click="transferOwner(m)"
                    class="px-2 h-6 rounded-md bg-amber-500/10 hover:bg-amber-500/20 text-amber-600 text-[11px] font-medium transition-colors"
                    title="将群主头衔转让给该成员"
                  >
                    转让群主
                  </button>

                  <!-- Mute / Unmute -->
                  <button
                    @click="toggleMute(m)"
                    class="px-2 h-6 rounded-md text-[11px] font-medium transition-colors"
                    :class="m.is_muted ? 'bg-black/5 text-mim-light-text hover:bg-black/10' : 'bg-amber-500/10 text-amber-600 hover:bg-amber-500/20'"
                  >
                    {{ m.is_muted ? '解禁' : '禁言' }}
                  </button>

                  <!-- Kick -->
                  <button
                    @click="kickMember(m.user_id)"
                    class="px-2 h-6 rounded-md bg-red-500/10 hover:bg-red-500/20 text-red-600 text-[11px] font-medium transition-colors"
                  >
                    踢出
                  </button>
                </template>

                <!-- ADMIN (Role 2) Controls (Can only manage regular members, NOT Owner or other Admins) -->
                <template v-else-if="isAdmin">
                  <template v-if="m.role === 3">
                    <button
                      @click="toggleMute(m)"
                      class="px-2 h-6 rounded-md text-[11px] font-medium transition-colors"
                      :class="m.is_muted ? 'bg-black/5 text-mim-light-text hover:bg-black/10' : 'bg-amber-500/10 text-amber-600 hover:bg-amber-500/20'"
                    >
                      {{ m.is_muted ? '解禁' : '禁言' }}
                    </button>
                    <button
                      @click="kickMember(m.user_id)"
                      class="px-2 h-6 rounded-md bg-red-500/10 hover:bg-red-500/20 text-red-600 text-[11px] font-medium transition-colors"
                    >
                      踢出
                    </button>
                  </template>
                  <span v-else class="text-[10px] text-mim-light-muted italic pr-1">无权操作</span>
                </template>
              </div>
            </div>
          </div>
        </div>

        <!-- TAB 3: 群设置 (Modern Card Group Form) -->
        <div v-else-if="activeTab === 'settings'" class="space-y-5">
          <!-- Non-admin read-only tip -->
          <div
            v-if="!canManage"
            class="p-3 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 text-xs flex items-center gap-2"
          >
            <Info class="w-4 h-4 inline-icon shrink-0" />
            <span>{{ langStore.t('group.memberReadOnlyTip') }}（{{ langStore.t('group.onlyAdminCanEdit') }}）</span>
          </div>

          <!-- 1. 查找方式 -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                <Search class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                <span>{{ langStore.t('group.searchableMode') }}</span>
              </h4>
              <p class="text-[11px] text-mim-light-muted mt-0.5">控制此群聊是否对全网成员公开可查</p>
            </div>

            <div
              @click="canManage && (groupForm.isSearchable = !groupForm.isSearchable)"
              class="p-3.5 rounded-xl border border-black/5 dark:border-white/5 bg-white/60 dark:bg-black/20 flex items-center justify-between transition-all select-none shadow-2xs"
              :class="canManage ? 'cursor-pointer hover:border-black/15 dark:hover:border-white/15' : 'cursor-not-allowed opacity-80'"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-lg bg-blue-500/10 text-blue-600 dark:bg-blue-500/20 dark:text-blue-400 flex items-center justify-center shrink-0">
                  <Search class="w-4 h-4 inline-icon" />
                </div>
                <div>
                  <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text">
                    {{ langStore.t('group.searchableCheckbox') }}
                  </div>
                  <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5">
                    {{ langStore.t('group.searchableTip') }}
                  </p>
                </div>
              </div>
              <div
                class="mim-toggle-track ml-4"
                :class="groupForm.isSearchable ? 'mim-toggle-track-on' : 'mim-toggle-track-off'"
              >
                <div
                  class="mim-toggle-knob"
                  :class="{ 'mim-toggle-knob-on': groupForm.isSearchable }"
                ></div>
              </div>
            </div>
          </div>

          <!-- 2. 加群方式 (Modern Card Group Form) -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                <ShieldCheck class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                <span>{{ langStore.t('group.joinMode') }}</span>
              </h4>
              <p class="text-[11px] text-mim-light-muted mt-0.5">设置外部成员加入此群聊的审核与准入策略</p>
            </div>

            <div class="grid grid-cols-1 gap-2 pt-1">
              <div
                v-for="jOpt in joinModeOptions"
                :key="jOpt.value"
                @click="canManage && (groupForm.joinMode = jOpt.value)"
                class="p-3.5 rounded-xl border transition-all flex items-center justify-between select-none shadow-2xs"
                :class="[
                  groupForm.joinMode === jOpt.value
                    ? 'border-mim-accent bg-mim-accent-subtle/30 ring-1 ring-mim-accent/40 shadow-xs'
                    : 'border-black/5 dark:border-white/5 bg-white/60 dark:bg-black/20',
                  canManage ? 'cursor-pointer hover:border-black/15 dark:hover:border-white/15' : 'cursor-not-allowed opacity-80'
                ]"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-4 h-4 rounded-full border flex items-center justify-center transition-all shrink-0"
                    :class="groupForm.joinMode === jOpt.value
                      ? 'border-mim-accent bg-mim-accent'
                      : 'border-black/30 dark:border-white/30'"
                  >
                    <div v-if="groupForm.joinMode === jOpt.value" class="w-1.5 h-1.5 rounded-full bg-white"></div>
                  </div>
                  <div>
                    <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-2">
                      <span>{{ jOpt.label }}</span>
                      <span v-if="jOpt.badge" class="text-[10px] px-1.5 py-0.2 rounded bg-mim-accent-subtle text-mim-accent font-medium">{{ jOpt.badge }}</span>
                    </div>
                    <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 leading-snug">
                      {{ jOpt.desc }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 3. 邀请方式 (Modern Card Toggle) -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                <UserPlus class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                <span>{{ langStore.t('group.inviteMode') }}</span>
              </h4>
              <p class="text-[11px] text-mim-light-muted mt-0.5">控制普通群成员是否有权限向群内邀请好友</p>
            </div>

            <div
              @click="canManage && (groupForm.allowMemberInvite = !groupForm.allowMemberInvite)"
              class="p-3.5 rounded-xl border border-black/5 dark:border-white/5 bg-white/60 dark:bg-black/20 flex items-center justify-between transition-all select-none shadow-2xs"
              :class="canManage ? 'cursor-pointer hover:border-black/15 dark:hover:border-white/15' : 'cursor-not-allowed opacity-80'"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-lg bg-indigo-500/10 text-indigo-600 dark:bg-indigo-500/20 dark:text-indigo-400 flex items-center justify-center shrink-0">
                  <UserPlus class="w-4 h-4 inline-icon" />
                </div>
                <div>
                  <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text">
                    {{ langStore.t('group.allowMemberInvite') }}
                  </div>
                  <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5">
                    开启后所有群成员可直接邀请外部好友；关闭时仅群主与管理员可发起邀请
                  </p>
                </div>
              </div>
              <div
                class="mim-toggle-track ml-4"
                :class="groupForm.allowMemberInvite ? 'mim-toggle-track-on' : 'mim-toggle-track-off'"
              >
                <div
                  class="mim-toggle-knob"
                  :class="{ 'mim-toggle-knob-on': groupForm.allowMemberInvite }"
                ></div>
              </div>
            </div>
          </div>

          <!-- 4. 会话与发言权限 (全员禁言) -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                <MessageSquare class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                <span>{{ langStore.t('group.chatPerm') }}</span>
              </h4>
              <p class="text-[11px] text-mim-light-muted mt-0.5">控制全员禁言与群发言管控（群主和管理员不受禁言限制）</p>
            </div>

            <div class="space-y-2 pt-1">
              <!-- 全员禁言 -->
              <div
                @click="canManage && (groupForm.isAllMuted = !groupForm.isAllMuted)"
                class="p-3.5 rounded-xl border transition-all flex items-center justify-between select-none shadow-2xs"
                :class="[
                  groupForm.isAllMuted
                    ? 'border-amber-500/40 bg-amber-500/[0.06] shadow-xs'
                    : 'border-black/5 dark:border-white/5 bg-white/60 dark:bg-black/20',
                  canManage ? 'cursor-pointer hover:border-black/15 dark:hover:border-white/15' : 'cursor-not-allowed opacity-80'
                ]"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 transition-colors"
                    :class="groupForm.isAllMuted ? 'bg-amber-500/20 text-amber-600 dark:text-amber-400' : 'bg-black/5 text-mim-light-muted dark:bg-white/10 dark:text-mim-dark-muted'"
                  >
                    <VolumeX class="w-4 h-4 inline-icon" />
                  </div>
                  <div>
                    <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-2">
                      <span>{{ langStore.t('group.allMuteSwitch') }}</span>
                      <span v-if="groupForm.isAllMuted" class="text-[10px] px-1.5 py-0.2 rounded bg-amber-500/20 text-amber-700 dark:text-amber-400 font-medium">全员禁言中</span>
                    </div>
                    <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5">
                      开启后除群主和管理员外，所有普通群成员将被禁止在群内发言
                    </p>
                  </div>
                </div>
                <div
                  class="mim-toggle-track ml-4"
                  :class="groupForm.isAllMuted ? 'mim-toggle-track-warning' : 'mim-toggle-track-off'"
                >
                  <div
                    class="mim-toggle-knob"
                    :class="{ 'mim-toggle-knob-on': groupForm.isAllMuted }"
                  ></div>
                </div>
              </div>
            </div>
          </div>

          <!-- 5. 底部操作按钮 -->
          <div v-if="canManage" class="pt-4 flex items-center justify-between border-t border-black/5 dark:border-white/5">
            <span class="text-xs text-mim-light-muted">修改后点击保存立即同步到服务端生效</span>
            <button
              @click="saveGroupSettings"
              class="px-5 h-8 rounded-lg mim-btn-primary text-xs font-medium shadow-sm flex items-center gap-1.5"
            >
              <Save class="w-3.5 h-3.5 inline-icon" />
              <span>{{ langStore.t('group.modifyGroupSettings') }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Footer With Danger / Leave Actions -->
      <div class="p-4 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.02] flex items-center justify-between">
        <div>
          <!-- Owner: Disband Group -->
          <button
            v-if="isOwner"
            @click="disbandGroup"
            class="px-4 h-8 rounded-lg bg-red-600/10 hover:bg-red-600/20 text-red-600 text-xs font-medium transition-colors flex items-center gap-1.5"
          >
            <Trash2 class="w-3.5 h-3.5 inline-icon" />
            <span>{{ langStore.t('group.disbandGroup') }}</span>
          </button>
          <!-- Admin or Member: Leave Group -->
          <button
            v-else
            @click="quitGroup"
            class="px-4 h-8 rounded-lg bg-red-500/10 hover:bg-red-500/20 text-red-600 text-xs font-medium transition-colors flex items-center gap-1.5"
          >
            <LogOut class="w-3.5 h-3.5 inline-icon" />
            <span>{{ langStore.t('group.leaveGroup') }}</span>
          </button>
        </div>

        <button
          @click="$emit('close')"
          class="px-5 h-8 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium text-mim-light-text dark:text-mim-dark-text transition-colors"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import {
  X,
  Check,
  Search,
  ShieldCheck,
  UserPlus,
  MessageSquare,
  VolumeX,
  Trash2,
  Save,
  Crown,
  User,
  Info,
  LogOut,
} from 'lucide-vue-next';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useDialogStore, confirmDialog } from '../stores/dialog';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';

const props = defineProps<{
  groupId: number;
}>();

const emit = defineEmits(['close', 'updated', 'disbanded']);

const chatStore = useChatStore();
const langStore = useLangStore();

const activeTab = ref<'info' | 'members' | 'settings'>('members');
const toastMsg = ref('');
const memberKeyword = ref('');
const members = ref<any[]>([]);

const groupForm = ref({
  title: '',
  avatar: '',
  notice: '',
  isSearchable: true,
  joinMode: 3,
  allowMemberInvite: true,
  allowTempChat: false,
  isAllMuted: false,
  ownerId: 0,
});

const joinModeOptions = computed(() => [
  { value: 1, label: '不允许任何人加群', desc: '关闭加群通道，仅群主与管理员可手动邀请成员' },
  { value: 2, label: '允许任何人加群', desc: '公开团队群，无需任何审核即可直接加入' },
  { value: 3, label: '需要验证消息', desc: '常规团队群，申请人附带验证说明，由管理员审核确认', badge: '推荐' },
  { value: 4, label: '需要回答问题', desc: '申请人需如实解答管理员设定的入群问卷' },
  { value: 5, label: '需要正确回答问题', desc: '申请人输入的答案必须与预设答案精准一致才可加入' },
]);

const currentMember = computed(() => members.value.find((m) => m.user_id === chatStore.currentUserId));
const isOwner = computed(() => groupForm.value.ownerId === chatStore.currentUserId || currentMember.value?.role === 1);
const isAdmin = computed(() => !isOwner.value && currentMember.value?.role === 2);
const canManage = computed(() => isOwner.value || isAdmin.value);

const tabs = computed(() => [
  { key: 'info' as const, label: langStore.t('group.tabInfo') },
  { key: 'members' as const, label: `${langStore.t('group.tabMembers')} (${members.value.length})` },
  { key: 'settings' as const, label: langStore.t('group.tabSettings') },
]);

const filteredMembers = computed(() => {
  if (!memberKeyword.value.trim()) {
    // Sort: Owner (1) first, then Admins (2), then Members (3)
    return [...members.value].sort((a, b) => (a.role || 3) - (b.role || 3));
  }
  const kw = memberKeyword.value.toLowerCase();
  return members.value
    .filter((m) => m.nickname.toLowerCase().includes(kw) || m.account.toLowerCase().includes(kw))
    .sort((a, b) => (a.role || 3) - (b.role || 3));
});

const showToast = (msg: string) => {
  toastMsg.value = msg;
  setTimeout(() => {
    toastMsg.value = '';
  }, 2500);
};

const fetchGroupInfo = async () => {
  try {
    const res = await fetch(`/api/group/info?group_id=${props.groupId}&user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0 && json.data) {
      const g = json.data;
      groupForm.value.title = g.title;
      groupForm.value.avatar = g.avatar;
      groupForm.value.notice = g.notice;
      groupForm.value.isSearchable = g.is_searchable === 1;
      groupForm.value.joinMode = g.join_mode || 3;
      groupForm.value.isAllMuted = g.is_all_muted === 1;
      groupForm.value.ownerId = g.owner_id;
    }
  } catch (_) {}
};

const fetchMembers = async () => {
  try {
    const res = await fetch(`/api/group/members?group_id=${props.groupId}&user_id=${chatStore.currentUserId}`, {
      headers: { Authorization: `Bearer ${chatStore.userToken}` },
    });
    const json = await res.json();
    if (json.code === 0 && Array.isArray(json.data)) {
      members.value = json.data;
    }
  } catch (_) {}
};

const saveGroupInfo = async () => {
  if (!canManage.value) return;
  try {
    const res = await fetch(`/api/group/update?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        title: groupForm.value.title,
        avatar: groupForm.value.avatar,
        notice: groupForm.value.notice,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      showToast('群资料已保存');
      emit('updated');
      chatStore.fetchConversations();
    }
  } catch (_) {}
};

const saveGroupSettings = async () => {
  if (!canManage.value) return;
  try {
    const res = await fetch(`/api/group/update?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        is_searchable: groupForm.value.isSearchable ? 1 : 0,
        join_mode: groupForm.value.joinMode,
        is_all_muted: groupForm.value.isAllMuted ? 1 : 0,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      showToast('群设置修改成功');
      emit('updated');
    }
  } catch (_) {}
};

const setRole = async (member: any, newRole: number) => {
  const roleName = newRole === 2 ? '管理员' : '普通成员';
  const ok = await confirmDialog({
    title: '调整群权限',
    message: `确定将 ${member.nickname} 调整为 ${roleName} 吗？`,
    confirmText: '确定调整',
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/group/set-role?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        target_id: member.user_id,
        role: newRole,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      member.role = newRole;
      showToast(`已成功将 ${member.nickname} 设置为 ${roleName}`);
      emit('updated');
    }
  } catch (_) {}
};

const transferOwner = async (member: any) => {
  const ok = await confirmDialog({
    title: '转让群主',
    message: `确定要将群主身份转让给 ${member.nickname} 吗？\n转让后您将降级为群管理员。`,
    confirmText: '确认转让',
    danger: true,
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/group/transfer-owner?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        new_owner_id: member.user_id,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      showToast(`已成功转让群主给 ${member.nickname}`);
      groupForm.value.ownerId = member.user_id;
      await fetchMembers();
      emit('updated');
    }
  } catch (_) {}
};

const toggleMute = async (member: any) => {
  const duration = member.is_muted ? 0 : 600; // 10 minutes or unmute
  try {
    const res = await fetch(`/api/group/mute-member?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        target_id: member.user_id,
        duration,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      member.is_muted = !member.is_muted;
      showToast(member.is_muted ? `已对 ${member.nickname} 开启禁言` : `已解除 ${member.nickname} 的禁言`);
      emit('updated');
    }
  } catch (_) {}
};

const kickMember = async (targetId: number) => {
  const target = members.value.find((m) => m.user_id === targetId);
  const name = target ? target.nickname : '该成员';
  const ok = await confirmDialog({
    title: '移出群聊',
    message: `确定要将 ${name} 移出本群吗？`,
    confirmText: '移出',
    danger: true,
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/group/kick-member?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
        target_id: targetId,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      members.value = members.value.filter((m) => m.user_id !== targetId);
      showToast('成员已被移出群聊');
      emit('updated');
    }
  } catch (_) {}
};

const disbandGroup = async () => {
  const ok = await confirmDialog({
    title: '解散群聊',
    message: '确定要解散该群聊吗？所有群记录将被清除，此操作无法撤销。',
    confirmText: '确认解散',
    danger: true,
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/group/disband?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      emit('disbanded');
      emit('close');
      chatStore.fetchConversations();
    }
  } catch (_) {}
};

const quitGroup = async () => {
  const ok = await confirmDialog({
    title: '退出群聊',
    message: '确定要退出该群聊吗？退出后将无法接收此群消息。',
    confirmText: '退出',
    danger: true,
  });
  if (!ok) return;
  try {
    const res = await fetch(`/api/group/quit?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: props.groupId,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      emit('disbanded');
      emit('close');
      chatStore.fetchConversations();
    }
  } catch (_) {}
};

onMounted(() => {
  fetchGroupInfo();
  fetchMembers();
});
</script>
