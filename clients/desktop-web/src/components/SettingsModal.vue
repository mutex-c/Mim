<template>
  <div class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none">
    <div class="w-full max-w-2xl rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col max-h-[85vh]">
      <!-- Header Tabs (Aligned with ui_perm.png & Reference Images) -->
      <div class="px-6 pt-5 pb-3 border-b border-black/5 dark:border-white/5 flex items-center justify-between">
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
            <!-- Active Underline -->
            <span
              v-if="activeTab === tab.key"
              class="absolute bottom-0 left-0 right-0 h-0.5 bg-mim-accent rounded-full"
            ></span>
          </button>
        </nav>
        <button
          @click="$emit('close')"
          class="w-8 h-8 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors"
        >
          <X class="w-4 h-4 inline-icon" />
        </button>
      </div>

      <!-- Tab Content Area -->
      <div class="p-6 overflow-y-auto flex-1 space-y-6">
        <!-- Toast Feedback -->
        <div
          v-if="feedbackMsg"
          class="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs font-medium flex items-center gap-2 transition-all"
        >
          <Check class="w-4 h-4 inline-icon shrink-0" />
          <span>{{ feedbackMsg }}</span>
        </div>

        <!-- TAB 1: 个人资料 -->
        <div v-if="activeTab === 'profile'" class="space-y-5">
          <div class="flex items-center gap-4">
            <div
              class="w-16 h-16 rounded-2xl font-bold text-xl flex items-center justify-center shadow-sm transition-all overflow-hidden"
              :class="getStatusRingClass('online')"
              :style="{
                backgroundColor: getAvatarColor(chatStore.currentUserId).bg,
                color: getAvatarColor(chatStore.currentUserId).text
              }"
            >
              <img
                v-if="profileForm.avatar && (profileForm.avatar.startsWith('http') || profileForm.avatar.startsWith('data:image'))"
                :src="profileForm.avatar"
                class="w-full h-full object-cover"
              />
              <span v-else>{{ profileForm.avatar || profileForm.nickname.substring(0, 2).toUpperCase() }}</span>
            </div>
            <div>
              <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('settings.avatar') }}</label>
              <input
                v-model="profileForm.avatar"
                type="text"
                placeholder="AC / URL / SVG"
                class="mim-input w-48 h-8 px-2.5 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text truncate"
              />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-mim-light-muted mb-1.5">{{ langStore.t('settings.nickname') }}</label>
              <input
                v-model="profileForm.nickname"
                type="text"
                class="mim-input w-full h-9 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-mim-light-muted mb-1.5">{{ langStore.t('settings.account') }}</label>
              <input
                :value="chatStore.userProfile?.account ? `${chatStore.userProfile.account} (ID: ${chatStore.currentUserId})` : `User ${chatStore.currentUserId}`"
                disabled
                class="mim-input w-full h-9 px-3 rounded-lg bg-black/[0.02] dark:bg-white/[0.03] text-sm text-mim-light-muted cursor-not-allowed opacity-75 font-mono"
              />
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium text-mim-light-muted mb-1.5">{{ langStore.t('settings.signature') }}</label>
            <textarea
              v-model="profileForm.signature"
              rows="3"
              placeholder="分享一段个性签名..."
              class="mim-input w-full p-3 rounded-lg bg-black/5 dark:bg-white/5 text-sm text-mim-light-text dark:text-mim-dark-text resize-none"
            ></textarea>
          </div>

          <!-- Genshin Character Persona Quick Presets -->
          <div class="p-3.5 rounded-xl border border-amber-500/25 bg-amber-500/5 dark:bg-amber-500/10 flex flex-col gap-2.5">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-1.5">
                <span class="text-amber-500 font-bold text-xs genshin-gold-glow">✦</span>
                <span class="text-xs font-bold text-mim-light-text dark:text-mim-dark-text">原神角色预设一键换装</span>
              </div>
              <span class="text-[10px] text-amber-600 dark:text-amber-400 font-medium">点击套用角色头像与名台词</span>
            </div>

            <div class="grid grid-cols-4 sm:grid-cols-8 gap-2">
              <button
                v-for="char in genshinCharacters"
                :key="char.id"
                @click="applyGenshinPersona(char)"
                class="flex flex-col items-center gap-1 p-1.5 rounded-xl border border-black/5 dark:border-white/5 hover:border-amber-500/50 hover:bg-amber-500/15 transition-all cursor-pointer group"
                :class="profileForm.nickname === char.name ? 'ring-2 ring-amber-500 bg-amber-500/20' : ''"
                :title="`${char.name} · ${char.title}`"
              >
                <img :src="char.avatarSvg" class="w-10 h-10 rounded-lg shadow-2xs group-hover:scale-105 transition-transform" />
                <span class="text-[10px] font-semibold text-mim-light-text dark:text-mim-dark-text truncate w-full text-center">
                  {{ char.name }}
                </span>
              </button>
            </div>
          </div>

          <div class="pt-4 border-t border-black/5 dark:border-white/5 flex items-center justify-between">
            <button
              @click="saveProfile"
              class="px-5 h-8 rounded-lg mim-btn-primary text-xs font-medium shadow-sm cursor-pointer"
            >
              {{ langStore.t('settings.saveProfile') }}
            </button>

            <button
              @click="handleLogout"
              class="px-4 h-8 rounded-lg border border-rose-500/30 text-rose-500 hover:bg-rose-500/10 text-xs font-medium transition-colors flex items-center gap-1.5 cursor-pointer"
            >
              <LogOut class="w-3.5 h-3.5 inline-icon" />
              <span>{{ langStore.t('auth.logoutBtn') }}</span>
            </button>
          </div>
        </div>

        <!-- TAB 2: 基本信息 -->
        <div v-else-if="activeTab === 'basic'" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5">
              <span class="text-xs text-mim-light-muted block mb-1">{{ langStore.t('settings.email') }}</span>
              <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text font-mono">
                {{ chatStore.userProfile?.email || 'alex.chen@mim.enterprise.internal' }}
              </span>
            </div>
            <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5">
              <span class="text-xs text-mim-light-muted block mb-1">{{ langStore.t('settings.phone') }}</span>
              <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text font-mono">
                {{ chatStore.userProfile?.phone || '+86 138-0000-8888' }}
              </span>
            </div>
            <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5">
              <span class="text-xs text-mim-light-muted block mb-1">{{ langStore.t('settings.regTime') }}</span>
              <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text font-mono">
                2024-04-15 10:24:00
              </span>
            </div>
            <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5">
              <span class="text-xs text-mim-light-muted block mb-1">{{ langStore.t('settings.clientVer') }}</span>
              <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text font-mono">
                v2.4.0 Canary (Release 2026.10)
              </span>
            </div>
          </div>
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5">
            <span class="text-xs text-mim-light-muted block mb-1">{{ langStore.t('settings.platform') }}</span>
            <span class="text-sm font-medium text-mim-light-text dark:text-mim-dark-text font-mono">
              Desktop Web & Electron Embedded (Windows x86_64)
            </span>
          </div>
        </div>

        <!-- TAB 3: 安全设置 -->
        <div v-else-if="activeTab === 'security'" class="space-y-5">
          <div class="space-y-3">
            <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">修改登录密码</h4>
            <div class="space-y-3 max-w-sm">
              <div>
                <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('settings.oldPassword') }}</label>
                <input
                  v-model="pwdForm.oldPassword"
                  type="password"
                  placeholder="默认密码: 123456"
                  class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('settings.newPassword') }}</label>
                <input
                  v-model="pwdForm.newPassword"
                  type="password"
                  placeholder="输入新密码"
                  class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-mim-light-muted mb-1">{{ langStore.t('settings.confirmPassword') }}</label>
                <input
                  v-model="pwdForm.confirmPassword"
                  type="password"
                  placeholder="确认新密码"
                  class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
                />
              </div>
              <button
                @click="changePassword"
                class="px-4 h-8 rounded-lg mim-btn-primary text-xs font-medium"
              >
                {{ langStore.t('settings.changePasswordBtn') }}
              </button>
            </div>
          </div>

          <div class="pt-4 border-t border-black/5 dark:border-white/5 space-y-3">
            <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.sessionTitle') }}</h4>
            <div class="p-3.5 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 flex items-center justify-between">
              <div>
                <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                  <span>Windows 11 / Chrome Client</span>
                </div>
                <div class="text-[11px] text-mim-light-muted font-mono mt-0.5">IP: 127.0.0.1 · 活跃中</div>
              </div>
              <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono font-medium">当前会话</span>
            </div>
          </div>
        </div>

        <!-- TAB 4: 权限与防骚扰设置 (Modern Card Group Form) -->
        <div v-else-if="activeTab === 'permission'" class="space-y-6">
          <!-- 1. 查找方式与防骚扰 -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div class="flex items-center justify-between">
              <div>
                <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                  <Search class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                  <span>{{ langStore.t('settings.searchTitle') }}</span>
                </h4>
                <p class="text-[11px] text-mim-light-muted mt-0.5">控制他人如何在 Mim 中通过账号或姓名发现你</p>
              </div>
            </div>

            <div class="grid grid-cols-1 gap-2 pt-1">
              <div
                v-for="opt in searchModeOptions"
                :key="opt.value"
                @click="permForm.searchMode = opt.value"
                class="p-3 rounded-xl border transition-all cursor-pointer flex items-center justify-between group"
                :class="permForm.searchMode === opt.value
                  ? 'border-mim-accent bg-mim-accent-subtle/30 ring-1 ring-mim-accent/40 shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:border-black/15 dark:hover:border-white/15 bg-black/[0.01] dark:bg-white/[0.02]'"
              >
                <div class="flex items-center gap-3">
                  <!-- Custom Sleek Radio Indicator -->
                  <div
                    class="w-4 h-4 rounded-full border flex items-center justify-center transition-all shrink-0"
                    :class="permForm.searchMode === opt.value
                      ? 'border-mim-accent bg-mim-accent'
                      : 'border-black/30 dark:border-white/30 group-hover:border-mim-accent'"
                  >
                    <div v-if="permForm.searchMode === opt.value" class="w-1.5 h-1.5 rounded-full bg-white"></div>
                  </div>
                  <div>
                    <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-2">
                      <span>{{ opt.label }}</span>
                      <span v-if="opt.badge" class="text-[10px] px-1.5 py-0.2 rounded bg-mim-accent-subtle text-mim-accent font-medium">{{ opt.badge }}</span>
                    </div>
                    <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 leading-snug">
                      {{ opt.desc }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 2. 好友添加验证方式 -->
          <div class="p-4 rounded-2xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted flex items-center gap-1.5">
                <Shield class="w-3.5 h-3.5 inline-icon text-mim-accent" />
                <span>{{ langStore.t('settings.verifyTitle') }}</span>
              </h4>
              <p class="text-[11px] text-mim-light-muted mt-0.5">设定当他人请求添加你为好友时的验证策略与通行条件</p>
            </div>

            <div class="grid grid-cols-1 gap-2 pt-1">
              <div
                v-for="vOpt in verifyModeOptions"
                :key="vOpt.value"
                @click="permForm.verifyMode = vOpt.value"
                class="p-3 rounded-xl border transition-all cursor-pointer flex items-center justify-between group"
                :class="permForm.verifyMode === vOpt.value
                  ? 'border-mim-accent bg-mim-accent-subtle/30 ring-1 ring-mim-accent/40 shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:border-black/15 dark:hover:border-white/15 bg-black/[0.01] dark:bg-white/[0.02]'"
              >
                <div class="flex items-center gap-3">
                  <div
                    class="w-4 h-4 rounded-full border flex items-center justify-center transition-all shrink-0"
                    :class="permForm.verifyMode === vOpt.value
                      ? 'border-mim-accent bg-mim-accent'
                      : 'border-black/30 dark:border-white/30 group-hover:border-mim-accent'"
                  >
                    <div v-if="permForm.verifyMode === vOpt.value" class="w-1.5 h-1.5 rounded-full bg-white"></div>
                  </div>
                  <div>
                    <div class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-2">
                      <span>{{ vOpt.label }}</span>
                      <span v-if="vOpt.badge" class="text-[10px] px-1.5 py-0.2 rounded bg-mim-accent-subtle text-mim-accent font-medium">{{ vOpt.badge }}</span>
                    </div>
                    <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 leading-snug">
                      {{ vOpt.desc }}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <!-- Conditional Question / Answer Fields -->
            <div
              v-if="permForm.verifyMode === 4 || permForm.verifyMode === 5"
              class="mt-3 p-3.5 rounded-xl bg-black/[0.03] dark:bg-white/[0.04] border border-black/5 dark:border-white/5 space-y-3"
            >
              <div>
                <label class="block text-xs font-semibold text-mim-light-text dark:text-mim-dark-text mb-1">
                  {{ langStore.t('settings.verifyQuestionLabel') }}
                </label>
                <input
                  v-model="permForm.question"
                  type="text"
                  placeholder="例如：您所在的团队名称是？"
                  class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
                />
              </div>
              <div v-if="permForm.verifyMode === 5">
                <label class="block text-xs font-semibold text-mim-light-text dark:text-mim-dark-text mb-1">
                  {{ langStore.t('settings.verifyAnswerLabel') }}
                </label>
                <input
                  v-model="permForm.answer"
                  type="text"
                  placeholder="设定标准正确答案（申请人必须完全匹配）"
                  class="mim-input w-full h-8 px-3 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text"
                />
              </div>
            </div>
          </div>

          <!-- Save Button -->
          <div class="pt-2 flex items-center justify-between">
            <span class="text-xs text-mim-light-muted">修改后点击保存立即同步到服务端生效</span>
            <button
              @click="savePermissions"
              class="px-6 h-8 rounded-lg mim-btn-primary text-xs font-medium shadow-sm flex items-center gap-1.5"
            >
              <Check class="w-3.5 h-3.5 inline-icon" />
              <span>{{ langStore.t('settings.saveConfig') }}</span>
            </button>
          </div>
        </div>

        <!-- TAB 5: 系统设置 -->
        <div v-else-if="activeTab === 'system'" class="space-y-5">
          <!-- 1. Designer Aesthetic Palette Selector (Anti-AI Slop) -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted block">
                {{ langStore.t('settings.paletteTitle') }}
              </span>
              <span class="text-[10px] text-mim-accent font-medium px-2 py-0.5 rounded-full bg-mim-accent-subtle">
                {{ themeStore.palettes.length }} Sets Available
              </span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
              <button
                v-for="p in themeStore.palettes"
                :key="p.id"
                @click="handleSelectPalette(p.id)"
                class="p-3 rounded-xl border text-left transition-all flex flex-col justify-between gap-2 cursor-pointer"
                :class="themeStore.currentPalette === p.id
                  ? 'border-mim-accent bg-mim-accent-subtle/25 ring-1 ring-mim-accent shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:bg-black/5 dark:hover:bg-white/5'"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <span
                      class="w-3.5 h-3.5 rounded-full shadow-xs shrink-0 border border-black/10 dark:border-white/10"
                      :style="{ backgroundColor: p.previewColor }"
                    ></span>
                    <span class="text-xs font-semibold text-mim-light-text dark:text-mim-dark-text flex items-center gap-1">
                      <span>{{ langStore.currentLang === 'zh' ? p.nameZh : p.nameEn }}</span>
                      <span v-if="p.id === 'genshin'" class="text-amber-500 font-bold genshin-gold-glow text-xs">✦</span>
                    </span>
                  </div>
                  <Check
                    v-if="themeStore.currentPalette === p.id"
                    class="w-3.5 h-3.5 text-mim-accent inline-icon"
                  />
                </div>
                <p class="text-[11px] text-mim-light-muted leading-tight line-clamp-2">
                  {{ langStore.currentLang === 'zh' ? p.descZh : p.descEn }}
                </p>
              </button>
            </div>
          </div>

          <!-- 2. Theme Mode Setting (Light, Dark, System Auto) -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted block">{{ langStore.t('settings.themeTitle') }}</span>
            <div class="grid grid-cols-3 gap-2">
              <button
                @click="themeStore.setTheme('light')"
                class="h-9 rounded-lg border text-xs font-medium flex items-center justify-center gap-1.5 transition-all"
                :class="themeStore.currentMode === 'light'
                  ? 'border-mim-accent bg-mim-accent-subtle text-mim-accent-text font-semibold shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted'"
              >
                <Sun class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.t('settings.themeLight') }}</span>
              </button>
              <button
                @click="themeStore.setTheme('dark')"
                class="h-9 rounded-lg border text-xs font-medium flex items-center justify-center gap-1.5 transition-all"
                :class="themeStore.currentMode === 'dark'
                  ? 'border-mim-accent bg-mim-accent-subtle text-mim-accent-text font-semibold shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted'"
              >
                <Moon class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.t('settings.themeDark') }}</span>
              </button>
              <button
                @click="themeStore.setTheme('system')"
                class="h-9 rounded-lg border text-xs font-medium flex items-center justify-center gap-1.5 transition-all"
                :class="themeStore.currentMode === 'system'
                  ? 'border-mim-accent bg-mim-accent-subtle text-mim-accent-text font-semibold shadow-xs'
                  : 'border-black/5 dark:border-white/5 hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted'"
              >
                <Monitor class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.t('settings.themeAuto') }}</span>
              </button>
            </div>
          </div>

          <!-- 3. Custom Chat Wallpaper & Background -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted block">
                {{ langStore.currentLang === 'zh' ? '聊天背景与自定义壁纸' : 'Custom Chat Wallpaper' }}
              </span>
              <span v-if="themeStore.customBgUrl" class="text-[10px] text-cyan-400 font-medium px-2 py-0.5 rounded-full bg-cyan-500/10">
                已应用自定义壁纸
              </span>
            </div>
            <div class="flex items-center justify-between">
              <p class="text-xs text-mim-light-muted max-w-sm">
                {{ langStore.currentLang === 'zh' ? '支持上传本地壁纸图片、官方画风预设背景，以及背景透明度与毛玻璃虚化深度调节' : 'Upload custom wallpaper or choose presets with opacity and blur controls' }}
              </p>
              <button
                @click="showBgModal = true"
                class="px-3.5 h-8 rounded-lg mim-btn-primary text-xs font-medium flex items-center gap-1.5 shrink-0 shadow-xs cursor-pointer ml-3"
              >
                <Palette class="w-3.5 h-3.5 inline-icon" />
                <span>{{ langStore.currentLang === 'zh' ? '配置背景图' : 'Customize' }}</span>
              </button>
            </div>
          </div>

          <!-- 3. Language Setting -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 flex items-center justify-between">
            <div>
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted block">{{ langStore.t('settings.langTitle') }}</span>
              <p class="text-xs text-mim-light-muted mt-0.5">中文 (Chinese) / English</p>
            </div>
            <div class="flex items-center gap-1 bg-black/5 dark:bg-white/5 p-1 rounded-lg">
              <button
                @click="langStore.setLang('zh')"
                class="px-2.5 py-1 rounded text-xs font-medium transition-all"
                :class="langStore.currentLang === 'zh' ? 'mim-btn-primary shadow-xs font-semibold' : 'text-mim-light-muted hover:text-mim-light-text'"
              >
                中文
              </button>
              <button
                @click="langStore.setLang('en')"
                class="px-2.5 py-1 rounded text-xs font-medium transition-all"
                :class="langStore.currentLang === 'en' ? 'mim-btn-primary shadow-xs font-semibold' : 'text-mim-light-muted hover:text-mim-light-text'"
              >
                English
              </button>
            </div>
          </div>

          <!-- 4. Notification Chime Preview -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 flex items-center justify-between">
            <div>
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted block">{{ langStore.t('settings.soundTitle') }}</span>
              <p class="text-xs text-mim-light-muted mt-0.5">内置低延迟 Web Audio 消息到达提醒音</p>
            </div>
            <button
              @click="testSound"
              class="px-3 h-7 rounded-lg bg-black/5 dark:bg-white/5 hover:bg-black/10 dark:hover:bg-white/10 text-xs font-medium flex items-center gap-1.5 transition-colors"
            >
              <Volume2 class="w-3.5 h-3.5 inline-icon" />
              <span>{{ langStore.t('settings.soundTest') }}</span>
            </button>
          </div>

          <!-- 5. Architecture & Protocol Telemetry -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.transportTitle') }}</span>
              <span class="text-[10px] px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-mono font-medium">
                24-Byte Big-Endian Aligned
              </span>
            </div>
            <p class="text-xs text-mim-light-muted leading-relaxed">
              {{ langStore.t('settings.transportDesc') }}
            </p>
          </div>

          <!-- 6. DFA Filter Telemetry -->
          <div class="p-4 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] border border-black/5 dark:border-white/5 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-xs font-semibold uppercase tracking-wider text-mim-light-muted">{{ langStore.t('settings.dfaTitle') }}</span>
              <span class="text-[10px] px-2 py-0.5 rounded bg-mim-accent-subtle text-mim-accent-text font-mono font-medium">
                DFA Active
              </span>
            </div>
            <p class="text-xs text-mim-light-muted leading-relaxed">
              {{ langStore.t('settings.dfaDesc') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="p-4 border-t border-black/5 dark:border-white/5 bg-black/[0.01] dark:bg-white/[0.02] flex justify-end">
        <button
          @click="$emit('close')"
          class="px-5 h-8 rounded-lg mim-btn-primary text-xs font-medium shadow-sm"
        >
          {{ langStore.t('settings.done') }}
        </button>
      </div>
    </div>

    <!-- Custom Background Modal -->
    <CustomBgModal
      v-if="showBgModal"
      @close="showBgModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { X, Check, Sun, Moon, Monitor, Volume2, Search, Shield, LogOut, Palette } from 'lucide-vue-next';
import CustomBgModal from './CustomBgModal.vue';
import { useChatStore } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useThemeStore } from '../stores/theme';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';
import { GENSHIN_CHARACTERS, type GenshinCharacter } from '../assets/genshin/characters';
import { playGenshinClick, playPrimogemSparkle } from '../utils/genshinAudio';
import type { ColorPalette } from '../stores/theme';

const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();
const genshinCharacters = GENSHIN_CHARACTERS;

const handleSelectPalette = (id: ColorPalette) => {
  themeStore.setPalette(id);
  if (id === 'genshin') {
    playPrimogemSparkle();
  } else {
    playGenshinClick();
  }
};

const applyGenshinPersona = (char: GenshinCharacter) => {
  playPrimogemSparkle();
  profileForm.value.nickname = char.name;
  profileForm.value.signature = char.quote;
  profileForm.value.avatar = char.avatarSvg;
  showToast(`已套用「${char.name}」专属预设，点击下方保存即可应用`);
};

const emit = defineEmits(['close']);

const handleLogout = () => {
  emit('close');
  chatStore.logout();
};

const activeTab = ref<'profile' | 'basic' | 'security' | 'permission' | 'system'>('permission');
const feedbackMsg = ref('');
const showBgModal = ref(false);

const tabs = computed(() => [
  { key: 'profile' as const, label: langStore.t('settings.tabProfile') },
  { key: 'basic' as const, label: langStore.t('settings.tabBasic') },
  { key: 'security' as const, label: langStore.t('settings.tabSecurity') },
  { key: 'permission' as const, label: langStore.t('settings.tabPermission') },
  { key: 'system' as const, label: langStore.t('settings.tabSystem') },
]);

const searchModeOptions = computed(() => [
  {
    value: 0,
    label: langStore.t('settings.searchModeNone'),
    desc: '完全隐身，他人无法通过搜索发现你，仅可通过群聊或已有会话发起互动',
  },
  {
    value: 1,
    label: langStore.t('settings.searchModeAccount'),
    desc: '高隐私模式，他人必须输入精准的用户号 (Account ID) 才能找到你',
  },
  {
    value: 2,
    label: langStore.t('settings.searchModeNickname'),
    desc: '便捷模式，同事与好友可通过姓名、昵称或账号快速查找并联系你',
    badge: '推荐',
  },
]);

const verifyModeOptions = computed(() => [
  {
    value: 1,
    label: langStore.t('settings.verifyModeForbid'),
    desc: '拒绝任何人添加，系统将自动拦截所有好友申请',
  },
  {
    value: 2,
    label: langStore.t('settings.verifyModeAllowAll'),
    desc: '开放模式，任何人发起申请无需确认直接成为好友',
  },
  {
    value: 3,
    label: langStore.t('settings.verifyModeNeedMsg'),
    desc: '常规验证，申请人需附带验证说明信息，由您手动审核确认',
    badge: '默认推荐',
  },
  {
    value: 4,
    label: langStore.t('settings.verifyModeNeedQuestion'),
    desc: '问答验证，申请人需填写对预设问题的解答，由您审核内容',
  },
  {
    value: 5,
    label: langStore.t('settings.verifyModeNeedCorrect'),
    desc: '精准问答，申请人必须输入与预设答案完全一致的内容才可通过',
  },
]);

// Profile Form
const profileForm = ref({
  nickname: chatStore.currentUserName,
  avatar: chatStore.currentUserAvatar,
  signature: '专注于高性能企业级 IM 架构研发与微服务实践。',
});

// Password Form
const pwdForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: '',
});

// Permissions Form (ui_perm.png)
const permForm = ref({
  searchMode: 1, // 0: 不允许, 1: 账号, 2: 昵称
  verifyMode: 3, // 1: 拒绝, 2: 允许, 3: 验证消息, 4: 问答, 5: 精确问答
  question: '',
  answer: '',
});

const showToast = (msg: string) => {
  feedbackMsg.value = msg;
  setTimeout(() => {
    feedbackMsg.value = '';
  }, 2500);
};

const saveProfile = async () => {
  try {
    const res = await fetch(`/api/user/update-profile?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        nickname: profileForm.value.nickname,
        avatar: profileForm.value.avatar,
        signature: profileForm.value.signature,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      chatStore.currentUserName = profileForm.value.nickname;
      chatStore.currentUserAvatar = profileForm.value.avatar;
      showToast(langStore.t('settings.profileSaved'));
    }
  } catch (_) {}
};

const changePassword = async () => {
  if (!pwdForm.value.oldPassword || !pwdForm.value.newPassword) return;
  if (pwdForm.value.newPassword !== pwdForm.value.confirmPassword) {
    showToast(langStore.t('settings.passwordMismatch'));
    return;
  }
  try {
    const res = await fetch(`/api/user/change-password?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        old_password: pwdForm.value.oldPassword,
        new_password: pwdForm.value.newPassword,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      showToast(langStore.t('settings.passwordChanged'));
      pwdForm.value.oldPassword = '';
      pwdForm.value.newPassword = '';
      pwdForm.value.confirmPassword = '';
    } else {
      showToast(json.msg || '密码错误');
    }
  } catch (_) {}
};

const savePermissions = async () => {
  try {
    const res = await fetch(`/api/user/config?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        user_id: chatStore.currentUserId,
        search_by_account: permForm.value.searchMode >= 1 ? 1 : 0,
        search_by_nickname: permForm.value.searchMode === 2 ? 1 : 0,
        friend_verify_mode: permForm.value.verifyMode,
        friend_verify_question: permForm.value.question,
        friend_verify_answer: permForm.value.answer,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      showToast(langStore.t('settings.configSaved'));
    }
  } catch (_) {}
};

// Web Audio sound test
const testSound = () => {
  try {
    const ctx = new (window.AudioContext || (window as any).webkitAudioContext)();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.type = 'sine';
    osc.frequency.setValueAtTime(880, ctx.currentTime); // A5 note
    osc.frequency.exponentialRampToValueAtTime(1320, ctx.currentTime + 0.12);
    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.25);
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.start();
    osc.stop(ctx.currentTime + 0.25);
  } catch (_) {}
};

onMounted(async () => {
  await chatStore.fetchUserProfile();
  if (chatStore.userProfile) {
    profileForm.value.nickname = chatStore.userProfile.nickname;
    profileForm.value.avatar = chatStore.userProfile.avatar;
    if (chatStore.userProfile.signature) {
      profileForm.value.signature = chatStore.userProfile.signature;
    }
  }
  if (chatStore.userConfig) {
    if (chatStore.userConfig.search_by_nickname === 1) {
      permForm.value.searchMode = 2;
    } else if (chatStore.userConfig.search_by_account === 1) {
      permForm.value.searchMode = 1;
    } else {
      permForm.value.searchMode = 0;
    }
    permForm.value.verifyMode = chatStore.userConfig.friend_verify_mode || 3;
    permForm.value.question = chatStore.userConfig.friend_verify_question || '';
    permForm.value.answer = chatStore.userConfig.friend_verify_answer || '';
  }
});
</script>
