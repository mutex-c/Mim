<template>
  <div
    :class="isFullscreen 
      ? 'w-full h-full flex items-center justify-center p-4 md:p-8' 
      : 'fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none'"
  >
    <!-- Horizontal Split Card Container -->
    <div
      class="w-full max-w-4xl rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/5 dark:border-white/10 overflow-hidden flex flex-col md:flex-row transition-all relative min-h-[500px]"
    >
      <!-- ================= LEFT COLUMN: Brand Hero & Highlights ================= -->
      <div
        class="hidden md:flex md:w-[340px] lg:w-[370px] shrink-0 flex-col justify-between p-7 bg-mim-light-list/60 dark:bg-mim-dark-dock/50 border-r border-black/[0.05] dark:border-white/[0.05] relative overflow-hidden"
      >
        <!-- Left Top: Logo & Title -->
        <div class="relative z-10 space-y-6">
          <div class="flex items-center gap-3">
            <img
              src="/logo.svg"
              alt="Mim"
              class="w-10 h-10 rounded-xl shadow-xs shrink-0 object-contain"
            />
            <div>
              <div class="flex items-center gap-2">
                <span class="font-semibold text-base text-mim-light-text dark:text-mim-dark-text tracking-tight">Mim</span>
                <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-mim-accent-subtle text-mim-accent font-medium">
                  v1.0.0
                </span>
              </div>
              <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted font-mono">
                M Instant Message
              </p>
            </div>
          </div>

          <!-- Headline -->
          <div class="space-y-1.5 pt-1">
            <h3 class="text-lg font-semibold text-mim-light-text dark:text-mim-dark-text tracking-tight leading-snug">
              {{ langStore.currentLang === 'zh' ? '安全 · 私密 · 毫秒级' : 'Secure · Private · Real-time' }}
            </h3>
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted leading-relaxed font-normal">
              {{ langStore.currentLang === 'zh' 
                ? '纯血自研跨平台企业即时通讯系统，专为高敏捷团队与私有集群打造。'
                : 'Pure native enterprise real-time messaging architecture, built for agile teams.' }}
            </p>
          </div>

          <!-- Feature Pillars -->
          <div class="space-y-4 pt-2">
            <div class="flex items-start gap-3">
              <div class="w-7 h-7 rounded-lg bg-black/[0.04] dark:bg-white/[0.06] flex items-center justify-center text-mim-accent shrink-0 mt-0.5">
                <Lock class="w-3.5 h-3.5 inline-icon" />
              </div>
              <div>
                <h4 class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text">
                  {{ langStore.currentLang === 'zh' ? '端到端集群加密' : 'Cluster End-to-End Encryption' }}
                </h4>
                <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 font-normal">
                  {{ langStore.currentLang === 'zh' ? '二进制自定义 Frame 协议，严防数据嗅探' : 'Custom binary protocol with end-to-end security' }}
                </p>
              </div>
            </div>

            <div class="flex items-start gap-3">
              <div class="w-7 h-7 rounded-lg bg-black/[0.04] dark:bg-white/[0.06] flex items-center justify-center text-mim-accent shrink-0 mt-0.5">
                <Zap class="w-3.5 h-3.5 inline-icon" />
              </div>
              <div>
                <h4 class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text">
                  {{ langStore.currentLang === 'zh' ? '毫秒级同步 & 秒传' : 'Instant Sync & Upload' }}
                </h4>
                <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 font-normal">
                  {{ langStore.currentLang === 'zh' ? 'SHA-256 去重瞬传，跨端断点续传' : 'Content deduplication with instant resume' }}
                </p>
              </div>
            </div>

            <div class="flex items-start gap-3">
              <div class="w-7 h-7 rounded-lg bg-black/[0.04] dark:bg-white/[0.06] flex items-center justify-center text-mim-accent shrink-0 mt-0.5">
                <Video class="w-3.5 h-3.5 inline-icon" />
              </div>
              <div>
                <h4 class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text">
                  {{ langStore.currentLang === 'zh' ? '点对点 WebRTC 通话' : 'P2P WebRTC Calls' }}
                </h4>
                <p class="text-[11px] text-mim-light-muted dark:text-mim-dark-muted mt-0.5 font-normal">
                  {{ langStore.currentLang === 'zh' ? '硬件加速音视频，支持画中画悬浮' : 'Hardware-accelerated calls with floating PiP' }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Left Bottom: Network Status -->
        <div class="relative z-10 pt-5 border-t border-black/[0.05] dark:border-white/[0.05] flex items-center justify-between text-[11px]">
          <div class="flex items-center gap-2 text-mim-light-muted dark:text-mim-dark-muted font-mono">
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>Gateway: 24ms 正常</span>
          </div>
          <span class="text-[10px] text-mim-light-muted/60 font-mono">TLS 1.3 / E2EE</span>
        </div>
      </div>

      <!-- ================= RIGHT COLUMN: Form & Action Area ================= -->
      <div class="flex-1 flex flex-col justify-between p-6 sm:p-7 overflow-y-auto">
        <!-- Right Header: Language, Theme & Close Controls -->
        <div class="flex items-center justify-between pb-3.5 border-b border-black/[0.04] dark:border-white/[0.04]">
          <!-- Mobile Brand Logo (Visible on mobile screens) -->
          <div class="flex md:hidden items-center gap-2">
            <img
              src="/logo.svg"
              alt="Mim"
              class="w-7 h-7 rounded-lg shadow-xs shrink-0 object-contain"
            />
            <span class="font-semibold text-sm text-mim-light-text dark:text-mim-dark-text">Mim</span>
          </div>

          <!-- Segmented Tab Switcher (Horizontal Pill) -->
          <div class="inline-flex p-1 rounded-xl bg-black/[0.04] dark:bg-white/[0.06] text-xs font-medium">
            <button
              @click="switchTab('login')"
              class="px-4 py-1.5 rounded-lg transition-all text-center cursor-pointer"
              :class="activeTab === 'login'
                ? 'bg-mim-light-card dark:bg-mim-dark-card text-mim-light-text dark:text-mim-dark-text shadow-xs font-medium'
                : 'text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text font-normal'"
            >
              {{ langStore.t('auth.tabLogin') }}
            </button>
            <button
              @click="switchTab('register')"
              class="px-4 py-1.5 rounded-lg transition-all text-center cursor-pointer"
              :class="activeTab === 'register'
                ? 'bg-mim-light-card dark:bg-mim-dark-card text-mim-light-text dark:text-mim-dark-text shadow-xs font-medium'
                : 'text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text font-normal'"
            >
              {{ langStore.t('auth.tabRegister') }}
            </button>
          </div>

          <!-- Utility Icons: Language, Theme, Close -->
          <div class="flex items-center gap-1">
            <!-- Language Toggle -->
            <button
              @click="langStore.toggleLang()"
              class="h-7 px-2 rounded-lg text-xs flex items-center gap-1 text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
              :title="langStore.currentLang === 'zh' ? 'Switch to English' : '切换为简体中文'"
            >
              <Languages class="w-3.5 h-3.5 inline-icon text-mim-accent" />
              <span class="font-mono text-[11px]">{{ langStore.currentLang.toUpperCase() }}</span>
            </button>

            <!-- Theme Toggle -->
            <button
              @click="cycleTheme"
              class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer"
              :title="langStore.t('nav.theme')"
            >
              <Sun v-if="themeStore.currentMode === 'light'" class="w-3.5 h-3.5 inline-icon text-amber-500" />
              <Moon v-else-if="themeStore.currentMode === 'dark'" class="w-3.5 h-3.5 inline-icon text-mim-accent" />
              <Monitor v-else class="w-3.5 h-3.5 inline-icon text-emerald-500" />
            </button>

            <!-- Server Connection Settings -->
            <button
              @click="openServerConfig"
              class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer relative"
              :title="langStore.currentLang === 'zh' ? '服务器连接配置' : 'Server Connection'"
            >
              <Server class="w-3.5 h-3.5 inline-icon" :class="isCustomServer ? 'text-mim-accent' : ''" />
              <span v-if="isCustomServer" class="absolute top-1 right-1 w-1.5 h-1.5 rounded-full bg-mim-accent"></span>
            </button>

            <!-- Close Button (if rendered in modal mode) -->
            <button
              v-if="!isFullscreen"
              @click="$emit('close')"
              class="w-7 h-7 rounded-lg flex items-center justify-center text-mim-light-muted hover:text-mim-light-text hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer ml-1"
            >
              <X class="w-4 h-4 inline-icon" />
            </button>
          </div>
        </div>

        <!-- Right Content Area -->
        <div class="py-3.5 space-y-3.5 flex-1">
          <!-- Title & Subtitle for Right Pane -->
          <div>
            <h2 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text tracking-tight">
              {{ activeTab === 'login' ? langStore.t('auth.loginTitle') : langStore.t('auth.registerTitle') }}
            </h2>
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted mt-0.5 font-normal">
              {{ activeTab === 'login' ? langStore.t('auth.loginSubtitle') : langStore.t('auth.registerSubtitle') }}
            </p>
          </div>

          <!-- Error Alert Banner -->
          <div
            v-if="errorMessage"
            class="p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-600 dark:text-rose-400 text-xs flex items-center gap-2"
          >
            <AlertCircle class="w-4 h-4 inline-icon shrink-0" />
            <span class="font-normal">{{ errorMessage }}</span>
          </div>

          <!-- Success Alert Banner -->
          <div
            v-if="successMessage"
            class="p-2.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs flex items-center gap-2"
          >
            <Check class="w-4 h-4 inline-icon shrink-0" />
            <span class="font-normal">{{ successMessage }}</span>
          </div>

          <!-- ================= TAB 1: 登录 (LOGIN) ================= -->
          <form v-if="activeTab === 'login'" @submit.prevent="handleLogin" class="space-y-3">
            <!-- Account -->
            <div class="space-y-1">
              <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                {{ langStore.t('auth.account') }}
              </label>
              <div class="relative flex items-center">
                <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                  <User class="w-4 h-4 inline-icon" />
                </span>
                <input
                  v-model="loginForm.account"
                  type="text"
                  required
                  autocomplete="username"
                  :placeholder="langStore.t('auth.accountPlaceholder')"
                  class="mim-input w-full h-9 pl-9 pr-3 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                />
              </div>
            </div>

            <!-- Password -->
            <div class="space-y-1">
              <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                {{ langStore.t('auth.password') }}
              </label>
              <div class="relative flex items-center">
                <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                  <Lock class="w-4 h-4 inline-icon" />
                </span>
                <input
                  v-model="loginForm.password"
                  :type="showPassword ? 'text' : 'password'"
                  required
                  autocomplete="current-password"
                  :placeholder="langStore.t('auth.passwordPlaceholder')"
                  class="mim-input w-full h-9 pl-9 pr-9 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute right-3 text-mim-light-muted hover:text-mim-light-text transition-colors cursor-pointer"
                >
                  <EyeOff v-if="showPassword" class="w-4 h-4 inline-icon" />
                  <Eye v-else class="w-4 h-4 inline-icon" />
                </button>
              </div>
            </div>

            <!-- Submit Button -->
            <button
              type="submit"
              :disabled="isLoading"
              class="w-full h-9 rounded-xl mim-btn-primary font-medium text-xs shadow-xs flex items-center justify-center gap-2 cursor-pointer transition-all disabled:opacity-50 mt-1"
            >
              <span v-if="isLoading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>{{ isLoading ? langStore.t('auth.loggingIn') : langStore.t('auth.loginBtn') }}</span>
              <ArrowRight v-if="!isLoading" class="w-3.5 h-3.5 inline-icon" />
            </button>
          </form>

          <!-- ================= TAB 2: 注册 (REGISTER) ================= -->
          <form v-else @submit.prevent="handleRegister" class="space-y-3">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Account -->
              <div class="space-y-1">
                <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                  {{ langStore.t('auth.account') }}
                </label>
                <div class="relative flex items-center">
                  <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                    <User class="w-3.5 h-3.5 inline-icon" />
                  </span>
                  <input
                    v-model="regForm.account"
                    type="text"
                    required
                    minlength="3"
                    autocomplete="username"
                    :placeholder="langStore.currentLang === 'zh' ? '至少3位字母或数字' : 'Min 3 characters'"
                    class="mim-input w-full h-9 pl-9 pr-3 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                  />
                </div>
              </div>

              <!-- Nickname with Header Avatar Preview -->
              <div class="space-y-1">
                <div class="flex items-center justify-between">
                  <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                    {{ langStore.t('auth.nickname') }}
                  </label>
                  <div class="flex items-center gap-1.5" :title="langStore.currentLang === 'zh' ? '头像预览' : 'Avatar Preview'">
                    <span class="text-[10px] text-mim-light-muted/70 font-mono">{{ langStore.currentLang === 'zh' ? '头像' : 'Avatar' }}</span>
                    <div
                      class="w-4.5 h-4.5 rounded font-bold text-[9px] flex items-center justify-center bg-mim-accent text-white shadow-2xs"
                    >
                      {{ avatarPreview }}
                    </div>
                  </div>
                </div>
                <div class="relative flex items-center">
                  <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                    <Smile class="w-3.5 h-3.5 inline-icon" />
                  </span>
                  <input
                    v-model="regForm.nickname"
                    type="text"
                    required
                    :placeholder="langStore.currentLang === 'zh' ? '你的昵称 (如: 张三)' : 'Your display name'"
                    class="mim-input w-full h-9 pl-9 pr-3 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                  />
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Password -->
              <div class="space-y-1">
                <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                  {{ langStore.t('auth.password') }}
                </label>
                <div class="relative flex items-center">
                  <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                    <Lock class="w-3.5 h-3.5 inline-icon" />
                  </span>
                  <input
                    v-model="regForm.password"
                    :type="showPassword ? 'text' : 'password'"
                    required
                    minlength="6"
                    autocomplete="new-password"
                    :placeholder="langStore.currentLang === 'zh' ? '至少6位密码' : 'Min 6 characters'"
                    class="mim-input w-full h-9 pl-9 pr-9 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                  />
                  <button
                    type="button"
                    @click="showPassword = !showPassword"
                    class="absolute right-2.5 text-mim-light-muted hover:text-mim-light-text transition-colors cursor-pointer"
                  >
                    <EyeOff v-if="showPassword" class="w-3.5 h-3.5 inline-icon" />
                    <Eye v-else class="w-3.5 h-3.5 inline-icon" />
                  </button>
                </div>
              </div>

              <!-- Confirm Password -->
              <div class="space-y-1">
                <label class="block text-xs font-normal text-mim-light-muted dark:text-mim-dark-muted">
                  {{ langStore.t('auth.confirmPassword') }}
                </label>
                <div class="relative flex items-center">
                  <span class="absolute left-3 text-mim-light-muted dark:text-mim-dark-muted pointer-events-none">
                    <ShieldCheck class="w-3.5 h-3.5 inline-icon" />
                  </span>
                  <input
                    v-model="regForm.confirmPassword"
                    :type="showPassword ? 'text' : 'password'"
                    required
                    autocomplete="new-password"
                    :placeholder="langStore.currentLang === 'zh' ? '请再次确认密码' : 'Confirm password'"
                    class="mim-input w-full h-9 pl-9 pr-3 rounded-xl bg-black/[0.03] dark:bg-white/[0.05] border border-black/5 dark:border-white/10 text-xs text-mim-light-text dark:text-mim-dark-text focus:outline-none focus:ring-2 focus:ring-mim-accent/30 focus:border-mim-accent transition-all font-normal"
                  />
                </div>
              </div>
            </div>

            <!-- Register Submit -->
            <button
              type="submit"
              :disabled="isLoading"
              class="w-full h-9 rounded-xl mim-btn-primary font-medium text-xs shadow-xs flex items-center justify-center gap-2 cursor-pointer transition-all disabled:opacity-50 mt-1"
            >
              <span v-if="isLoading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>{{ isLoading ? langStore.t('auth.registering') : langStore.t('auth.registerBtn') }}</span>
              <ArrowRight v-if="!isLoading" class="w-3.5 h-3.5 inline-icon" />
            </button>
          </form>
        </div>

        <!-- ================= Quick Demo Account Picker ================= -->
        <div class="pt-3 border-t border-black/[0.04] dark:border-white/[0.04] space-y-2">
          <div class="flex items-center justify-between text-[11px] text-mim-light-muted dark:text-mim-dark-muted font-normal">
            <span class="flex items-center gap-1.5">
              <Sparkles class="w-3 h-3 text-amber-500 inline-icon" />
              <span>{{ langStore.t('auth.quickDemoLogin') }}</span>
            </span>
            <span class="text-[10px] text-mim-light-muted/70 font-mono">密码: 123456</span>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="item in demoAccounts"
              :key="item.user.id"
              @click="quickLogin(item)"
              type="button"
              class="flex items-center gap-2 p-2 rounded-xl bg-black/[0.02] dark:bg-white/[0.03] hover:bg-black/[0.05] dark:hover:bg-white/[0.07] border border-black/[0.03] dark:border-white/[0.05] transition-all text-left group cursor-pointer"
            >
              <div
                class="w-7 h-7 rounded-lg font-bold flex items-center justify-center text-[11px] shrink-0"
                :style="{ backgroundColor: getAvatarColor(item.user.id).bg, color: getAvatarColor(item.user.id).text }"
              >
                {{ item.user.avatar || item.user.nickname.substring(0, 2).toUpperCase() }}
              </div>
              <div class="min-w-0 flex-1">
                <div class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text truncate group-hover:text-mim-accent transition-colors">
                  {{ item.user.nickname }}
                </div>
                <div class="text-[10px] text-mim-light-muted dark:text-mim-dark-muted font-mono truncate font-normal">
                  @{{ item.user.account }}
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Server Connection Config Modal -->
    <div
      v-if="showServerModal"
      class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 select-none"
      @click.self="showServerModal = false"
    >
      <div class="w-full max-w-sm rounded-2xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/10 dark:border-white/10 p-5 space-y-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Server class="w-4 h-4 text-mim-accent inline-icon" />
            <h3 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text">服务器连接配置</h3>
          </div>
          <button @click="showServerModal = false" class="text-mim-light-muted hover:text-mim-light-text cursor-pointer">
            <X class="w-4 h-4 inline-icon" />
          </button>
        </div>
        <p class="text-xs text-mim-light-muted leading-relaxed">
          默认为单机独立服务（127.0.0.1:8080）。如需局域网联机或远程协同聊天，请输入目标服务器 IP:端口。
        </p>
        <div class="space-y-1.5">
          <label class="text-[11px] font-medium text-mim-light-muted">服务器地址 (Host:Port)</label>
          <input
            v-model="tempServerHost"
            placeholder="127.0.0.1:8080"
            class="mim-input w-full h-8 px-3 rounded-lg bg-black/[0.03] dark:bg-white/[0.05] border border-black/10 dark:border-white/10 text-xs font-mono text-mim-light-text dark:text-mim-dark-text"
          />
        </div>
        <div class="flex items-center justify-end gap-2 pt-2">
          <button
            @click="resetServerHost"
            class="px-3 py-1.5 rounded-lg text-xs text-mim-light-muted hover:text-mim-light-text bg-black/5 dark:bg-white/5 cursor-pointer"
          >
            恢复默认
          </button>
          <button
            @click="saveServerHost"
            class="px-3 py-1.5 rounded-lg text-xs text-white bg-mim-accent hover:opacity-90 font-medium cursor-pointer"
          >
            保存并连接
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import {
  ShieldCheck,
  User,
  Lock,
  Eye,
  EyeOff,
  ArrowRight,
  AlertCircle,
  Check,
  Languages,
  Sun,
  Moon,
  Monitor,
  X,
  Sparkles,
  Zap,
  Video,
  Smile,
  Server,
} from 'lucide-vue-next';
import { useChatStore, DemoUser } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useThemeStore } from '../stores/theme';
import { getAvatarColor } from '../utils/avatar';
import { getServerHost, setServerHost } from '../utils/api';

const props = withDefaults(defineProps<{
  isFullscreen?: boolean;
  initialTab?: 'login' | 'register';
}>(), {
  isFullscreen: false,
  initialTab: 'login',
});

const emit = defineEmits(['close', 'tab-change', 'success']);

const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();

const showServerModal = ref(false);
const tempServerHost = ref(getServerHost());
const isCustomServer = computed(() => {
  const h = tempServerHost.value;
  return h !== '127.0.0.1:8080' && h !== 'localhost:8080' && h !== (typeof window !== 'undefined' ? window.location.host : '');
});

const openServerConfig = () => {
  tempServerHost.value = getServerHost();
  showServerModal.value = true;
};

const resetServerHost = () => {
  setServerHost('127.0.0.1:8080');
  tempServerHost.value = '127.0.0.1:8080';
  showServerModal.value = false;
  window.location.reload();
};

const saveServerHost = () => {
  if (tempServerHost.value) {
    setServerHost(tempServerHost.value);
  }
  showServerModal.value = false;
  window.location.reload();
};

const activeTab = ref<'login' | 'register'>(props.initialTab || 'login');

watch(() => props.initialTab, (newTab) => {
  if (newTab) activeTab.value = newTab;
});
const showPassword = ref(false);
const isLoading = ref(false);
const errorMessage = ref('');
const successMessage = ref('');

const loginForm = ref({
  account: '',
  password: '',
});

const regForm = ref({
  account: '',
  nickname: '',
  password: '',
  confirmPassword: '',
});

const avatarPreview = computed(() => {
  const name = regForm.value.nickname.trim() || regForm.value.account.trim() || 'MI';
  return name.substring(0, 2).toUpperCase();
});

const demoAccounts = computed(() => {
  if (chatStore.demoUsers.length > 0) {
    return chatStore.demoUsers;
  }
  // Fallback demo users metadata
  return [
    { user: { id: 1, account: 'alex', nickname: 'Alex Chen', avatar: 'AC', role: 1 }, token: { access_token: '', refresh_token: '', expires_in: 7200 } },
    { user: { id: 2, account: 'sarah', nickname: 'Sarah Jenkins', avatar: 'SJ', role: 2 }, token: { access_token: '', refresh_token: '', expires_in: 7200 } },
    { user: { id: 3, account: 'michael', nickname: 'Michael Zhang', avatar: 'MZ', role: 2 }, token: { access_token: '', refresh_token: '', expires_in: 7200 } },
    { user: { id: 4, account: 'david', nickname: 'David Miller', avatar: 'DM', role: 2 }, token: { access_token: '', refresh_token: '', expires_in: 7200 } },
  ];
});

const switchTab = (tab: 'login' | 'register') => {
  activeTab.value = tab;
  errorMessage.value = '';
  successMessage.value = '';
  emit('tab-change', tab);
};

const cycleTheme = () => {
  if (themeStore.currentMode === 'system') {
    themeStore.setTheme('dark');
  } else if (themeStore.currentMode === 'dark') {
    themeStore.setTheme('light');
  } else {
    themeStore.setTheme('system');
  }
};

const handleLogin = async () => {
  errorMessage.value = '';
  successMessage.value = '';

  if (!loginForm.value.account || !loginForm.value.password) {
    errorMessage.value = '请填写完整的用户名和密码';
    return;
  }

  isLoading.value = true;
  const res = await chatStore.login(loginForm.value.account.trim(), loginForm.value.password);
  isLoading.value = false;

  if (res.success) {
    successMessage.value = langStore.t('auth.loginSuccess');
    emit('success');
    emit('close');
  } else {
    errorMessage.value = res.error || '登录失败，请检查账号和密码';
  }
};

const handleRegister = async () => {
  errorMessage.value = '';
  successMessage.value = '';

  const account = regForm.value.account.trim();
  const nickname = regForm.value.nickname.trim();
  const password = regForm.value.password;
  const confirmPassword = regForm.value.confirmPassword;

  if (account.length < 3) {
    errorMessage.value = langStore.t('auth.accountMinLen');
    return;
  }
  if (password.length < 6) {
    errorMessage.value = langStore.t('auth.passwordMinLen');
    return;
  }
  if (password !== confirmPassword) {
    errorMessage.value = langStore.t('auth.passwordMismatch');
    return;
  }

  isLoading.value = true;
  const res = await chatStore.register(account, password, nickname, avatarPreview.value);
  isLoading.value = false;

  if (res.success) {
    successMessage.value = langStore.t('auth.registerSuccess');
    emit('success');
    emit('close');
  } else {
    errorMessage.value = res.error || '注册失败，可能账号已存在';
  }
};

const quickLogin = (item: DemoUser) => {
  if (item.token && item.token.access_token) {
    chatStore.switchAccount(item);
    emit('success');
    emit('close');
  } else {
    // Fill form and submit
    loginForm.value.account = item.user.account;
    loginForm.value.password = '123456';
    handleLogin();
  }
};
</script>
