<template>
  <main
    @dragover.prevent="isDragging = true"
    @dragleave.prevent="isDragging = false"
    @drop.prevent="handleFileDrop"
    class="flex-1 h-full flex bg-mim-light-stage dark:bg-mim-dark-stage genshin-stage-bg relative select-none overflow-hidden"
  >
    <!-- Hidden File Inputs -->
    <input type="file" ref="imageInputRef" @change="handleImageSelected" accept="image/*" class="hidden" />
    <input type="file" ref="fileInputRef" @change="handleFileSelected" class="hidden" />

    <!-- Drag & Drop Overlay -->
    <div
      v-if="isDragging"
      class="absolute inset-4 rounded-2xl border-2 border-dashed border-mim-accent bg-mim-accent-subtle/30 backdrop-blur-sm z-40 flex flex-col items-center justify-center pointer-events-none text-mim-accent gap-2"
    >
      <UploadCloud class="w-10 h-10 animate-bounce inline-icon" />
      <span class="text-sm font-semibold">{{ langStore.t('chat.dropOverlay') }}</span>
    </div>

    <!-- Center Chat Column -->
    <div
      v-if="chatStore.activeConversation"
      class="flex-1 h-full flex flex-col min-w-0"
      :class="themeStore.currentPalette === 'genshin' ? 'genshin-stage-bg' : (themeStore.currentPalette === 'vodyanitsa' ? 'vodyanitsa-stage-bg' : '')"
    >
      <!-- Top Stage Header -->
      <header class="h-16 px-6 flex items-center justify-between bg-mim-light-stage/80 dark:bg-mim-dark-stage/80 backdrop-blur-md z-10 shrink-0 border-b border-black/[0.04] dark:border-white/[0.04]">
        <div class="flex items-center gap-3">
          <div
            class="w-10 h-10 rounded-xl font-bold flex items-center justify-center text-sm shadow-xs transition-all overflow-hidden"
            :class="getStatusRingClass(isGroup ? 'group' : (chatStore.activeConversation?.isOnline ? 'online' : 'offline'))"
            :style="{
              backgroundColor: getAvatarColor(chatStore.activeConversation?.targetId || chatStore.activeConversation?.id, isGroup).bg,
              color: getAvatarColor(chatStore.activeConversation?.targetId || chatStore.activeConversation?.id, isGroup).text
            }"
          >
            <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="vodyanitsaAvatar" alt="沃雅妮莎" class="w-full h-full object-cover" />
            <img v-else-if="chatStore.activeConversation?.avatarText && (chatStore.activeConversation.avatarText.startsWith('http') || chatStore.activeConversation.avatarText.startsWith('data:image'))" :src="chatStore.activeConversation.avatarText" class="w-full h-full object-cover" />
            <span v-else>{{ chatStore.activeConversation?.avatarText || 'MI' }}</span>
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-sm font-semibold text-mim-light-text dark:text-mim-dark-text truncate flex items-center gap-1.5">
                <span>{{ chatStore.activeConversation?.title }}</span>
                <span v-if="themeStore.currentPalette === 'genshin'" class="text-amber-500 font-bold genshin-gold-glow text-xs" title="原神星穹认证">✦</span>
                <img v-else-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconCrest" class="w-4 h-4 object-contain inline-block shrink-0 drop-shadow-[0_0_6px_rgba(56,189,248,0.7)]" title="科洛列夫茨基剧团认证" />
              </h2>
              <span v-if="isGroup" class="text-xs text-mim-light-muted font-normal">
                ({{ groupMembers.length }}人)
              </span>
              <span
                v-if="isGroup && myGroupRole"
                class="text-[10px] px-1.5 py-0.2 rounded font-medium flex items-center gap-0.5"
                :class="myGroupRole === 1
                  ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 font-semibold border border-amber-500/25'
                  : myGroupRole === 2
                    ? 'bg-indigo-500/15 text-indigo-600 dark:text-indigo-400 font-semibold border border-indigo-500/25'
                    : 'bg-black/5 dark:bg-white/5 text-mim-light-muted'"
              >
                <Crown v-if="myGroupRole === 1" class="w-2.5 h-2.5 text-amber-500 inline-icon" />
                <ShieldCheck v-else-if="myGroupRole === 2" class="w-2.5 h-2.5 text-indigo-500 inline-icon" />
                <span>{{ myGroupRole === 1 ? '群主' : myGroupRole === 2 ? '管理员' : '成员' }}</span>
              </span>
              <ShieldCheck class="w-3.5 h-3.5 text-mim-accent inline-icon" />
            </div>
            <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted flex items-center gap-1.5">
              <span 
                v-if="chatStore.activeConversation?.isOnline" 
                class="w-1.5 h-1.5 rounded-full bg-emerald-500 inline-block"
              ></span>
              <span>{{ isGroup ? (groupDetails?.notice ? `公告: ${groupDetails.notice.substring(0, 24)}...` : '团队架构协作加密群') : (chatStore.activeConversation?.isOnline ? langStore.t('chat.onlineE2EE') : langStore.t('chat.clusterEncrypted')) }}</span>
            </p>
          </div>
        </div>

        <!-- Action Buttons: Wallpaper, Audio Call, Video Call, Details / Settings -->
        <div class="flex items-center gap-1">
          <!-- Custom Chat Wallpaper -->
          <button
            @click="showBgModal = true"
            class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors cursor-pointer"
            :class="themeStore.customBgUrl
              ? 'text-cyan-400 bg-cyan-500/15 border border-cyan-400/30'
              : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
            :title="langStore.currentLang === 'zh' ? '自定义聊天背景图与壁纸' : 'Custom Chat Wallpaper'"
          >
            <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="iconPalette" class="w-5 h-5 object-contain inline-block shrink-0 drop-shadow-[0_0_6px_rgba(56,189,248,0.5)]" />
            <Palette v-else class="w-4 h-4 inline-icon" />
          </button>

          <!-- AI Summary Header Button -->
          <button
            @click="openSummaryDrawer"
            class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors cursor-pointer"
            :class="showSummaryDrawer
              ? 'bg-mim-accent-subtle text-mim-accent'
              : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
            :title="langStore.t('chat.aiSummary') || '会话智能纪要'"
          >
            <Sparkles class="w-4 h-4 inline-icon" />
          </button>

          <button
            @click="startCall('audio')"
            class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
            :title="langStore.t('chat.audioCall')"
          >
            <Phone class="w-4 h-4 inline-icon" />
          </button>
          <button
            @click="startCall('video')"
            class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
            :title="langStore.t('chat.videoCall')"
          >
            <Video class="w-4 h-4 inline-icon" />
          </button>
          <button
            v-if="isGroup"
            @click="showGroupSettingsModal = true"
            class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
            :title="langStore.t('group.groupSettings')"
          >
            <MoreHorizontal class="w-4 h-4 inline-icon" />
          </button>
          <button
            v-else
            @click="openPeerProfile"
            class="w-9 h-9 rounded-lg flex items-center justify-center text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text transition-colors"
            :title="langStore.t('chat.convInfo')"
          >
            <MoreHorizontal class="w-4 h-4 inline-icon" />
          </button>

          <!-- Drawer Toggle Button (for group chat) -->
          <button
            v-if="isGroup"
            @click="showGroupDrawer = !showGroupDrawer"
            class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors"
            :class="showGroupDrawer ? 'bg-mim-accent-subtle text-mim-accent' : 'text-mim-light-muted hover:bg-black/5 dark:hover:bg-white/5'"
            title="群成员与公告"
          >
            <Users class="w-4 h-4 inline-icon" />
          </button>
        </div>
      </header>

      <!-- Message Stream Area with Floating Jump Pill (规范 2) -->
      <div class="flex-1 min-h-0 relative z-1">
        <div ref="scrollContainer" @scroll="onMessagesScroll" class="w-full h-full overflow-y-auto px-6 py-4 space-y-4">
          <div
            v-for="msg in chatStore.activeMessages"
            :key="msg.id"
            class="flex flex-col group relative"
            :class="msg.isRevoked || msg.msgType === 'revoke' ? 'items-center my-2' : (msg.isMe ? 'items-end' : 'items-start')"
          >
          <!-- Revoked Notice Pill (Centered, minimal, sleek WeChat/Slack-style capsule) -->
          <div
            v-if="msg.isRevoked || msg.msgType === 'revoke'"
            class="w-full flex items-center justify-center my-1 select-none"
          >
            <span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs text-mim-light-muted dark:text-mim-dark-muted bg-black/[0.04] dark:bg-white/[0.05] border border-black/[0.04] dark:border-white/[0.06] shadow-2xs font-normal">
              <Undo2 class="w-3 h-3 opacity-60 inline-icon" />
              <span>{{ msg.isMe ? langStore.t('chat.recalledSelf') : (langStore.currentLang === 'zh' ? `${msg.senderName || '对方'} 撤回了一条消息` : `${msg.senderName || 'Partner'} recalled a message`) }}</span>
            </span>
          </div>

          <!-- Normal Non-Revoked Message Container -->
          <template v-else>
            <!-- In Group Chat: Sender Nickname above bubble if not me -->
            <div
              v-if="isGroup && !msg.isMe && themeStore.currentPalette !== 'genshin' && themeStore.currentPalette !== 'vodyanitsa'"
              class="text-[11px] text-mim-light-muted font-medium mb-1 pl-10"
            >
              {{ msg.senderName }}
            </div>

            <!-- Message Row with Avatar for Peers & Traveler -->
            <div class="flex items-start gap-2.5 max-w-[80%]" :class="msg.isMe ? 'flex-row-reverse' : 'flex-row'">
              <!-- Peer Avatar in Group or Direct -->
              <div
                v-if="!msg.isMe"
                @click="openUserProfileById(msg.senderId, msg.senderName)"
                class="w-8 h-8 rounded-xl font-bold text-xs flex items-center justify-center shrink-0 cursor-pointer hover:opacity-85 transition-all mt-0.5 shadow-xs"
                :class="[
                  getStatusRingClass(isGroup ? 'offline' : (chatStore.activeConversation?.isOnline ? 'online' : 'offline')),
                  themeStore.currentPalette === 'genshin' ? 'genshin-avatar-frame' : (themeStore.currentPalette === 'vodyanitsa' ? 'vodyanitsa-avatar-frame' : '')
                ]"
                :style="{
                  backgroundColor: getAvatarColor(msg.senderId || getSpeakerName(msg), false).bg,
                  color: getAvatarColor(msg.senderId || getSpeakerName(msg), false).text
                }"
              >
                <img v-if="themeStore.currentPalette === 'vodyanitsa'" :src="vodyanitsaAvatar" alt="沃雅妮莎" class="w-full h-full object-cover rounded-xl" />
                <span v-else>{{ getSpeakerName(msg).substring(0, 2).toUpperCase() }}</span>
              </div>

              <!-- Traveler Avatar for Me when in Genshin or Vodyanitsa theme -->
              <div
                v-else-if="msg.isMe && (themeStore.currentPalette === 'genshin' || themeStore.currentPalette === 'vodyanitsa')"
                class="w-8 h-8 rounded-xl font-bold text-xs flex items-center justify-center shrink-0 mt-0.5 shadow-xs select-none"
                :class="themeStore.currentPalette === 'genshin' ? 'genshin-avatar-frame' : 'vodyanitsa-avatar-frame'"
                :style="themeStore.currentPalette === 'genshin'
                  ? 'background: linear-gradient(135deg, #b45309, #d97706); color: #fef08a;'
                  : 'background: linear-gradient(135deg, #0284c7, #0369a1); color: #e0f2fe;'"
                title="旅行者"
              >
                旅
              </div>

              <!-- Bubble Body Container -->
              <div class="flex flex-col" :class="msg.isMe ? 'items-end' : 'items-start'">
                <!-- Quoted Message Bubble (if any) -->
                <div
                  v-if="msg.quote"
                  class="mb-1 px-3 py-1.5 rounded-lg text-xs cursor-pointer hover:opacity-80 transition-opacity max-w-full"
                  :class="themeStore.currentPalette === 'genshin'
                    ? 'bg-[#181a28]/90 text-amber-100/90 border-l-2 border-[#caa159] shadow-xs'
                    : 'bg-black/[0.03] dark:bg-white/[0.04] text-mim-light-muted dark:text-mim-dark-muted border-l-2 border-mim-accent'"
                >
                  <span class="font-medium block mb-0.5" :class="themeStore.currentPalette === 'genshin' ? 'text-amber-400' : 'text-mim-accent'">{{ msg.quote.userName }}</span>
                  <span class="truncate block">{{ msg.quote.content }}</span>
                </div>

                <!-- Content Bubble with Actions -->
                <div class="relative flex items-center gap-2 group/bubble">
                  <!-- Quick Action Context Menu -->
                  <div
                    class="opacity-0 group-hover/bubble:opacity-100 transition-opacity flex items-center gap-1 shadow-card rounded-lg px-1.5 py-1 z-10 shrink-0"
                    :class="[
                      msg.isMe ? 'order-first' : 'order-last',
                      themeStore.currentPalette === 'genshin'
                        ? 'bg-[#1b1e2c] border border-[#caa159]/40 text-amber-200'
                        : 'bg-mim-light-card dark:bg-mim-dark-card'
                    ]"
                  >
                    <button
                      @click="copyMessage(msg.id, msg.content)"
                      class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted transition-colors relative"
                      :title="langStore.t('chat.copy')"
                    >
                      <Check v-if="copiedMsgId === msg.id" class="w-3.5 h-3.5 text-emerald-500 inline-icon" />
                      <Copy v-else class="w-3.5 h-3.5 inline-icon" />
                    </button>
                    <button
                      @click="setQuote(msg)"
                      class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted dark:text-mim-dark-muted transition-colors"
                      :title="langStore.t('chat.reply')"
                    >
                      <Reply class="w-3.5 h-3.5 inline-icon" />
                    </button>
                    <button
                      v-if="msg.isMe"
                      @click="chatStore.revokeMessage(msg.id)"
                      class="p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 text-red-500 hover:text-red-600 transition-colors"
                      :title="langStore.t('chat.recall')"
                    >
                      <Undo2 class="w-3.5 h-3.5 inline-icon" />
                    </button>
                  </div>

                  <!-- File Message Card (QQ Style: Document Card + Download Badge + Open Folder Button) -->
                  <div
                    v-if="msg.msgType === 'file'"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <FileMessageCard :message="msg" />
                  </div>

                  <!-- Paimon Full-Color Chibi Sticker Display -->
                  <div
                    v-else-if="isPaimonChibi(msg.content)"
                    class="flex flex-col items-center p-2 rounded-2xl bg-[#181926]/90 border border-[#caa159]/40 shadow-lg max-w-[200px] select-none hover:scale-105 transition-transform"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <img
                      :src="getChibiById(parseChibiId(msg.content))?.svgDataUri"
                      :alt="getChibiById(parseChibiId(msg.content))?.name || '派蒙画作'"
                      class="w-32 h-32 object-contain"
                    />
                    <span class="text-xs font-bold text-[#fef9c3] mt-1">
                      {{ getChibiById(parseChibiId(msg.content))?.name }}
                    </span>
                    <span class="text-[10px] text-[#9ca3af] text-center px-1">
                      {{ getChibiById(parseChibiId(msg.content))?.caption }}
                    </span>
                  </div>

                  <!-- Vodyanitsa Bespoke Sticker Display -->
                  <div
                    v-else-if="isVodyanitsaSticker(msg.content)"
                    class="flex flex-col items-center p-2.5 rounded-2xl bg-[#091124]/95 border border-[#38bdf8]/50 shadow-xl max-w-[200px] select-none hover:scale-105 transition-transform"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <img
                      :src="getVodyanitsaStickerById(parseVodyanitsaStickerId(msg.content))?.imageSrc"
                      :alt="getVodyanitsaStickerById(parseVodyanitsaStickerId(msg.content))?.name || '沃雅妮莎表情'"
                      class="w-32 h-32 object-contain"
                    />
                    <span class="text-xs font-bold text-[#7dd3fc] mt-1">
                      {{ getVodyanitsaStickerById(parseVodyanitsaStickerId(msg.content))?.name }}
                    </span>
                    <span class="text-[10px] text-[#bae6fd]/80 text-center px-1">
                      {{ getVodyanitsaStickerById(parseVodyanitsaStickerId(msg.content))?.caption }}
                    </span>
                  </div>

                  <!-- Custom User Added Sticker Display -->
                  <div
                    v-else-if="isCustomSticker(msg.content)"
                    class="flex flex-col items-center p-2 rounded-2xl bg-[#091124]/95 border border-[#38bdf8]/50 shadow-xl max-w-[200px] select-none hover:scale-105 transition-transform"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <img
                      :src="parseCustomSticker(msg.content)?.url"
                      :alt="parseCustomSticker(msg.content)?.name || '自定义表情'"
                      class="w-32 h-32 object-contain rounded-lg"
                    />
                    <span v-if="parseCustomSticker(msg.content)?.name" class="text-xs font-bold text-[#7dd3fc] mt-1 truncate max-w-full">
                      {{ parseCustomSticker(msg.content)?.name }}
                    </span>
                  </div>

                  <!-- Wish Interactive Result Card (Genshin or Vodyanitsa) -->
                  <div
                    v-else-if="isGenshinWish(msg.content)"
                    class="p-4 space-y-3 rounded-2xl select-none"
                    :class="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')
                      ? 'bg-gradient-to-b from-[#091736]/95 to-[#050c1f]/95 border border-[#38bdf8]/60 shadow-[0_4px_24px_rgba(56,189,248,0.25)]'
                      : 'genshin-wish-card'"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <div
                      class="flex items-center justify-between pb-2 border-b"
                      :class="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')
                        ? 'border-cyan-400/30'
                        : 'border-[#caa159]/30'"
                    >
                      <div class="flex items-center gap-1.5">
                        <img
                          v-if="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')"
                          :src="iconBtnHorn"
                          alt=""
                          class="w-4 h-4 object-contain inline-block shrink-0"
                        />
                        <span v-else class="text-amber-400 text-sm">✦</span>
                        <span
                          class="text-xs font-bold"
                          :class="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')
                            ? 'text-cyan-200'
                            : 'text-[#fef9c3]'"
                        >
                          {{ (parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')) ? '唤春角笛 · 剧院特约祈愿战报' : '提瓦特祈愿战报' }} · {{ parseWishData(msg.content).bannerName }}
                        </span>
                      </div>
                      <span
                        class="text-[10px] font-mono"
                        :class="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')
                          ? 'text-cyan-300'
                          : 'text-amber-300'"
                      >祈愿成功</span>
                    </div>

                    <p
                      class="text-xs font-bold leading-relaxed"
                      :class="parseWishData(msg.content).bannerName.includes('幽歌') || parseWishData(msg.content).summaryText.includes('沃雅妮莎')
                        ? 'text-cyan-100'
                        : 'text-amber-200'"
                    >
                      {{ parseWishData(msg.content).summaryText }}
                    </p>

                    <div class="grid grid-cols-5 gap-1.5 bg-black/40 p-2 rounded-xl border border-white/5">
                      <div
                        v-for="(item, pIdx) in (parseWishData(msg.content).pulls || []).slice(0, 10)"
                        :key="pIdx"
                        class="flex flex-col items-center p-1 rounded-lg border text-center"
                        :class="{
                          'bg-amber-950/70 border-amber-400': item.rarity === 5,
                          'bg-purple-950/70 border-purple-400': item.rarity === 4,
                          'bg-slate-900/60 border-slate-700': item.rarity === 3,
                        }"
                      >
                        <span class="text-[8px] font-bold" :class="item.rarity === 5 ? 'text-amber-300' : item.rarity === 4 ? 'text-purple-300' : 'text-slate-400'">
                          {{ '★'.repeat(item.rarity) }}
                        </span>
                        <span class="text-[9px] font-medium text-white truncate max-w-full">
                          {{ item.name }}
                        </span>
                      </div>
                    </div>

                    <div class="flex items-center justify-between pt-1">
                      <span class="text-[10px] text-slate-400">来自旅行者的祈愿分享</span>
                      <button
                        @click="showWishModal = true; playPrimogemSparkle()"
                        class="px-2.5 py-1 rounded-lg bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-white text-[10px] font-bold shadow transition-all cursor-pointer"
                      >
                        ✦ 我也要抽卡
                      </button>
                    </div>
                  </div>

                  <!-- GENSHIN IMPACT HAUTE-CRAFT DIALOGUE PLATE -->
                  <div
                    v-else-if="themeStore.currentPalette === 'genshin'"
                    class="genshin-dialogue-plate select-text shadow-xl"
                    :class="msg.isMe ? 'genshin-plate-self' : 'genshin-plate-peer'"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <!-- Authentic Celtic Corner Brackets -->
                    <span class="genshin-corner-trim tl"></span>
                    <span class="genshin-corner-trim tr"></span>
                    <span class="genshin-corner-trim bl"></span>
                    <span class="genshin-corner-trim br"></span>

                    <!-- Integrated Speaker Header Plaque -->
                    <div class="genshin-integrated-header">
                      <div class="flex items-center gap-2">
                        <!-- Cut-Gemstone Vision Rhombus -->
                        <div
                          class="genshin-vision-rhombus"
                          :style="{ '--gem-color': getVisionColor(getSpeakerName(msg)) }"
                        >
                          <span class="vision-text">
                            {{ getVisionElement(getSpeakerName(msg)) }}
                          </span>
                        </div>
                        <!-- Speaker Title -->
                        <span class="genshin-speaker-title">
                          {{ getSpeakerName(msg) }}
                        </span>
                        <!-- 5-Star Metallic Rating -->
                        <span class="genshin-stars-metallic">★★★★★</span>
                      </div>

                      <span class="genshin-role-badge">
                        {{ getGenshinCharacterTitle(getSpeakerName(msg)) }}
                      </span>
                    </div>

                    <!-- Ornamental Gold Divider Bar -->
                    <div class="genshin-divider-strip">
                      <div class="divider-line"></div>
                      <div class="divider-pip">◆</div>
                      <div class="divider-line"></div>
                    </div>

                    <!-- Plate Dialogue Body -->
                    <div class="genshin-plate-content">
                      <!-- Text Message -->
                      <p v-if="msg.msgType === 'text'" class="genshin-speech-text pretty-wrap">
                        {{ msg.content }}
                      </p>

                      <!-- Image Message -->
                      <div v-else-if="msg.msgType === 'image'" class="py-1">
                        <img
                          :src="resolveApiUrl(msg.content)"
                          alt="Image"
                          class="max-w-xs max-h-60 rounded-xl object-cover cursor-pointer hover:opacity-95 shadow-sm transition-opacity"
                          @click="openLightbox(resolveApiUrl(msg.content))"
                        />
                      </div>

                      <!-- Voice Note Message -->
                      <div v-else-if="msg.msgType === 'voice'" class="flex items-center gap-3 py-1 text-amber-200">
                        <button
                          @click="playVoice(msg.id)"
                          class="w-8 h-8 rounded-full bg-amber-500/20 border border-amber-400/40 text-amber-300 flex items-center justify-center hover:scale-105 transition-transform"
                        >
                          <Pause v-if="playingVoiceId === msg.id" class="w-4 h-4 inline-icon" />
                          <Play v-else class="w-4 h-4 inline-icon ml-0.5" />
                        </button>
                        <div class="flex items-center gap-1 h-5">
                          <span class="w-0.5 h-3 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                          <span class="w-0.5 h-5 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-80'"></span>
                          <span class="w-0.5 h-2 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                          <span class="w-0.5 h-4 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-90'"></span>
                          <span class="w-0.5 h-5 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-70'"></span>
                          <span class="w-0.5 h-3 bg-amber-400 rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-50'"></span>
                        </div>
                        <span class="text-xs font-mono tabular-nums text-amber-300">
                          {{ msg.duration || 6 }}s
                        </span>
                      </div>

                      <!-- AI Summary Card -->
                      <div v-else-if="msg.msgType === 'ai_summary'" class="space-y-2 py-1 max-w-md">
                        <div class="flex items-center gap-1.5 text-xs font-semibold text-amber-400">
                          <Sparkles class="w-4 h-4 inline-icon" />
                          <span>{{ langStore.t('chat.aiSynthesisTitle') }}</span>
                        </div>
                        <div class="text-xs leading-relaxed pretty-wrap text-[#f7f5ed]">
                          <span
                            v-for="(char, idx) in msg.content"
                            :key="idx"
                            class="stream-char"
                            :style="{ animationDelay: `${idx * 20}ms` }"
                          >
                            {{ char }}
                          </span>
                        </div>
                      </div>
                    </div>

                    <!-- Animated Teyvat Primogem Waiting Cursor -->
                    <div class="genshin-dialogue-cursor">
                      <span class="cursor-glow-star">✦</span>
                    </div>
                  </div>

                  <!-- VODYANITSA HAUTE-CRAFT OPERA CARTOUCHE DIALOGUE PLATE (Exact match to Mockup) -->
                  <div
                    v-else-if="themeStore.currentPalette === 'vodyanitsa'"
                    class="vodyanitsa-dialogue-plate select-text shadow-md"
                    :class="msg.isMe ? 'vodyanitsa-plate-self' : 'vodyanitsa-plate-peer'"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <!-- Text Message -->
                    <p v-if="msg.msgType === 'text'" class="text-[13px] leading-relaxed pretty-wrap">
                      {{ msg.content }}
                    </p>

                    <!-- Image Message -->
                    <div v-else-if="msg.msgType === 'image'" class="py-1">
                      <img
                        :src="resolveApiUrl(msg.content)"
                        alt="Image"
                        class="max-w-xs max-h-60 rounded-xl object-cover cursor-pointer hover:opacity-95 shadow-sm transition-opacity"
                        @click="openLightbox(resolveApiUrl(msg.content))"
                      />
                    </div>

                    <!-- Voice Note Message -->
                    <div v-else-if="msg.msgType === 'voice'" class="flex items-center gap-3 py-1" :class="msg.isMe ? 'text-cyan-100' : 'text-slate-800'">
                      <button
                        @click="playVoice(msg.id)"
                        class="w-8 h-8 rounded-full flex items-center justify-center hover:scale-105 transition-transform"
                        :class="msg.isMe ? 'bg-white/20 text-white' : 'bg-black/10 text-slate-900'"
                      >
                        <Pause v-if="playingVoiceId === msg.id" class="w-4 h-4 inline-icon" />
                        <Play v-else class="w-4 h-4 inline-icon ml-0.5" />
                      </button>
                      <div class="flex items-center gap-1 h-5">
                        <span class="w-0.5 h-3 rounded-full" :class="[msg.isMe ? 'bg-cyan-200' : 'bg-slate-700', playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40']"></span>
                        <span class="w-0.5 h-5 rounded-full" :class="[msg.isMe ? 'bg-cyan-200' : 'bg-slate-700', playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-80']"></span>
                        <span class="w-0.5 h-2 rounded-full" :class="[msg.isMe ? 'bg-cyan-200' : 'bg-slate-700', playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40']"></span>
                        <span class="w-0.5 h-4 rounded-full" :class="[msg.isMe ? 'bg-cyan-200' : 'bg-slate-700', playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-90']"></span>
                        <span class="w-0.5 h-3 rounded-full" :class="[msg.isMe ? 'bg-cyan-200' : 'bg-slate-700', playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-50']"></span>
                      </div>
                      <span class="text-xs font-mono tabular-nums" :class="msg.isMe ? 'text-cyan-100' : 'text-slate-700'">
                        {{ msg.duration || 6 }}s
                      </span>
                    </div>

                    <!-- AI Summary Card -->
                    <div v-else-if="msg.msgType === 'ai_summary'" class="space-y-2 py-1 max-w-md">
                      <div class="flex items-center gap-1.5 text-xs font-semibold" :class="msg.isMe ? 'text-cyan-200' : 'text-sky-700'">
                        <Sparkles class="w-4 h-4 inline-icon" />
                        <span>{{ langStore.t('chat.aiSynthesisTitle') }}</span>
                      </div>
                      <div class="text-xs leading-relaxed pretty-wrap">
                        <span
                          v-for="(char, idx) in msg.content"
                          :key="idx"
                          class="stream-char"
                          :style="{ animationDelay: `${idx * 20}ms` }"
                        >
                          {{ char }}
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- Standard Bubble Main for Text, Image, Voice, AI Summary -->
                  <div
                    v-else
                    class="px-4 py-2.5 rounded-2xl transition-all cursor-default select-text"
                    :class="[
                      msg.isMe
                        ? 'bg-mim-bubble-self text-mim-bubble-self-text rounded-br-sm shadow-sm'
                        : 'bg-mim-bubble-peer text-mim-bubble-peer-text rounded-bl-sm border border-black/[0.08] dark:border-white/[0.08] shadow-sm'
                    ]"
                    @contextmenu.prevent.stop="handleMessageContextMenu($event, msg)"
                  >
                    <!-- Text Message -->
                    <p v-if="msg.msgType === 'text'" class="text-sm leading-relaxed pretty-wrap">
                      {{ msg.content }}
                    </p>

                    <!-- Image Message -->
                    <div v-else-if="msg.msgType === 'image'" class="py-1">
                      <img
                        :src="resolveApiUrl(msg.content)"
                        alt="Image"
                        class="max-w-xs max-h-60 rounded-xl object-cover cursor-pointer hover:opacity-95 shadow-sm transition-opacity"
                        @click="openLightbox(resolveApiUrl(msg.content))"
                      />
                    </div>

                    <!-- Voice Note Message -->
                    <div v-else-if="msg.msgType === 'voice'" class="flex items-center gap-3 py-1">
                      <button
                        @click="playVoice(msg.id)"
                        class="w-8 h-8 rounded-full bg-black/10 dark:bg-white/10 flex items-center justify-center hover:scale-105 transition-transform"
                      >
                        <Pause v-if="playingVoiceId === msg.id" class="w-4 h-4 inline-icon" />
                        <Play v-else class="w-4 h-4 inline-icon ml-0.5" />
                      </button>
                      <div class="flex items-center gap-1 h-5">
                        <span class="w-0.5 h-3 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                        <span class="w-0.5 h-5 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-80'"></span>
                        <span class="w-0.5 h-2 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-40'"></span>
                        <span class="w-0.5 h-4 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-90'"></span>
                        <span class="w-0.5 h-5 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-70'"></span>
                        <span class="w-0.5 h-3 bg-current rounded-full" :class="playingVoiceId === msg.id ? 'animate-pulse' : 'opacity-50'"></span>
                      </div>
                      <span class="text-xs font-mono tabular-nums opacity-90">
                        {{ msg.duration || 6 }}s
                      </span>
                    </div>

                    <!-- AI Summary Card -->
                    <div v-else-if="msg.msgType === 'ai_summary'" class="space-y-2 py-1 max-w-md">
                      <div class="flex items-center gap-1.5 text-xs font-semibold text-mim-accent">
                        <Sparkles class="w-4 h-4 inline-icon" />
                        <span>{{ langStore.t('chat.aiSynthesisTitle') }}</span>
                      </div>
                      <div class="text-xs leading-relaxed pretty-wrap">
                        <span
                          v-for="(char, idx) in msg.content"
                          :key="idx"
                          class="stream-char"
                          :style="{ animationDelay: `${idx * 20}ms` }"
                        >
                          {{ char }}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Timestamp & Delivery Status -->
                <div
                  class="flex items-center gap-1 mt-1 text-[10px] text-mim-light-muted dark:text-mim-dark-muted tabular-nums"
                  :class="msg.isMe ? 'mr-1' : 'ml-1'"
                >
                  <span>{{ msg.time }}</span>
                  <span v-if="themeStore.currentPalette === 'genshin'" class="text-amber-400 font-bold ml-0.5 text-[10px]" title="提瓦特印记">◆</span>
                  <span v-if="msg.isMe">
                    <Clock v-if="msg.status === 'sending'" class="w-3 h-3 inline-icon" />
                    <Check v-else-if="msg.status === 'sent'" class="w-3 h-3 inline-icon" />
                    <CheckCheck v-else-if="msg.status === 'delivered' || msg.status === 'read'" class="w-3 h-3 inline-icon text-mim-accent" />
                  </span>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- Floating Jump to Bottom Pill (规范 2) -->
      <Transition name="fade">
        <button
          v-if="showJumpToBottom"
          @click="scrollToBottom(true)"
          class="absolute bottom-4 right-6 z-30 px-3.5 py-1.5 rounded-full bg-white/90 dark:bg-[#1a2332]/90 backdrop-blur-md shadow-lg border border-black/[0.08] dark:border-white/[0.1] text-xs font-medium text-mim-light-text dark:text-mim-dark-text hover:text-mim-accent transition-all flex items-center gap-1.5 hover:scale-105 active:scale-95 cursor-pointer"
        >
          <ArrowDown class="w-3.5 h-3.5" />
          <span>跳至最新</span>
        </button>
      </Transition>
    </div>

      <!-- Active Quote Bar -->
      <div class="mim-accordion px-6" :class="{ open: !!activeQuote }">
        <div class="mim-accordion-inner">
          <div v-if="activeQuote || displayedQuote" class="pt-2">
            <div class="flex items-center justify-between p-2.5 rounded-lg bg-black/[0.03] dark:bg-white/[0.04] border-l-2 border-mim-accent">
              <div class="text-xs truncate pr-4">
                <span class="font-medium text-mim-accent">{{ langStore.t('chat.replyingTo') }} {{ (activeQuote || displayedQuote)?.userName }}:</span>
                <span class="text-mim-light-muted dark:text-mim-dark-muted ml-1">{{ (activeQuote || displayedQuote)?.content }}</span>
              </div>
              <button @click="clearQuote" class="text-mim-light-muted dark:text-mim-dark-muted hover:text-mim-light-text dark:hover:text-mim-dark-text">
                <X class="w-3.5 h-3.5 inline-icon" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Mute Notification Banner (if muted) -->
      <div
        v-if="isMutedInGroup"
        class="mx-4 mb-2 p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-600 dark:text-amber-400 text-xs font-medium text-center flex items-center justify-center gap-1.5"
      >
        <VolumeX class="w-4 h-4 inline-icon" />
        <span>{{ isAllMuted ? langStore.t('group.allMutedBanner') : langStore.t('group.userMutedBanner') }}</span>
      </div>

      <!-- Bottom Input Bar -->
      <div class="p-4 pt-2 relative z-10">
        <div
          class="rounded-[20px] shadow-card p-3 flex flex-col gap-2 relative transition-all border border-black/[0.08] dark:border-white/[0.08] focus-within:border-transparent focus-within:shadow-[inset_0_0_0_1.5px_var(--mim-accent)]"
          :class="themeStore.currentPalette === 'vodyanitsa'
            ? 'vodyanitsa-input-deck !m-0 !p-3.5'
            : 'bg-mim-light-card dark:bg-mim-dark-card'"
        >
          <!-- Snowflake corner pips in Vodyanitsa mode -->
          <template v-if="themeStore.currentPalette === 'vodyanitsa'">
            <span class="corner-snowflake tl">✦</span>
            <span class="corner-snowflake tr">✦</span>
            <span class="corner-snowflake bl">✦</span>
            <span class="corner-snowflake br">✦</span>
          </template>

          <!-- AI Smart Reply Bar (above textarea) -->
          <AISmartReplyBar
            v-if="showSmartReply"
            :suggestions="replySuggestions"
            :loading="isReplying"
            :visible="showSmartReply"
            @select="onSmartReplySelect"
            @send="onSmartReplySend"
            @regenerate="fetchSmartReplies"
            @close="showSmartReply = false"
          />

          <!-- Input Toolbar (above textarea) -->
          <div class="flex items-center justify-between pb-1 border-b border-black/[0.04] dark:border-white/[0.04]">
            <div class="flex items-center gap-1.5">
              <button
                @click="toggleSmartReply"
                :disabled="isMutedInGroup"
                type="button"
                class="px-2.5 h-7 rounded-lg flex items-center gap-1.5 text-xs transition-all cursor-pointer font-medium disabled:opacity-40 border-0"
                :class="showSmartReply
                  ? 'bg-mim-accent text-white shadow-2xs'
                  : (themeStore.currentPalette === 'vodyanitsa'
                    ? 'text-cyan-300 hover:bg-cyan-500/20'
                    : 'text-mim-accent hover:bg-mim-accent-subtle')"
                title="AI 智能帮回"
              >
                <Sparkles class="w-3.5 h-3.5 inline-icon" />
                <span>AI 帮回</span>
              </button>
            </div>
          </div>

          <!-- Message Textarea -->
          <textarea
            ref="textareaRef"
            v-model="inputContent"
            :disabled="isMutedInGroup"
            @compositionstart="onCompositionStart"
            @compositionend="onCompositionEnd"
            @keydown="onInputKeyDown"
            @keydown.esc="showSmartReply = false"
            @paste="handlePaste"
            rows="2"
            :placeholder="isMutedInGroup ? '当前禁言中，无法发送消息' : (themeStore.currentPalette === 'vodyanitsa' ? '输入消息... (Enter 发送, Shift+Enter 换行)' : langStore.t('chat.inputPlaceholder'))"
            class="w-full p-2 bg-transparent text-sm resize-none border-none outline-none shadow-none focus:outline-none focus:ring-0 focus:shadow-none pretty-wrap disabled:cursor-not-allowed disabled:opacity-50"
            :class="themeStore.currentPalette === 'vodyanitsa'
              ? 'text-white placeholder-cyan-200/60'
              : 'text-mim-light-text dark:text-mim-dark-text placeholder-mim-light-muted dark:placeholder-mim-dark-muted'"
          ></textarea>

          <!-- Bottom Controls -->
          <div class="flex items-center justify-between pt-1">
            <!-- Multimedia & Action Buttons -->
            <div class="flex items-center gap-1.5 flex-wrap">
              <button
                @click="triggerImageInput"
                :disabled="isMutedInGroup"
                class="w-8 h-8 rounded-lg flex items-center justify-center disabled:opacity-40 transition-colors border-0"
                :class="themeStore.currentPalette === 'vodyanitsa'
                  ? 'text-cyan-300 hover:bg-cyan-500/20'
                  : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
                :title="langStore.t('chat.attachImage')"
              >
                <Image class="w-4 h-4 inline-icon" />
              </button>
              <button
                @click="triggerFileInput"
                :disabled="isMutedInGroup"
                class="w-8 h-8 rounded-lg flex items-center justify-center disabled:opacity-40 transition-colors border-0"
                :class="themeStore.currentPalette === 'vodyanitsa'
                  ? 'text-cyan-300 hover:bg-cyan-500/20'
                  : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text'"
                :title="langStore.t('chat.attachFile')"
              >
                <Paperclip class="w-4 h-4 inline-icon" />
              </button>
              <button
                @click="toggleVoiceRecording"
                :disabled="isMutedInGroup"
                class="px-2 h-8 rounded-lg flex items-center gap-1.5 transition-colors disabled:opacity-40 cursor-pointer border-0"
                :class="isRecording
                  ? 'bg-red-500/20 text-red-500 animate-pulse font-medium'
                  : (themeStore.currentPalette === 'vodyanitsa' ? 'text-cyan-300 hover:bg-cyan-500/20' : 'text-mim-light-muted dark:text-mim-dark-muted hover:bg-black/5 dark:hover:bg-white/5 hover:text-mim-light-text dark:hover:text-mim-dark-text')"
                :title="isRecording ? '点击停止并发送语音' : langStore.t('chat.voiceRecording')"
              >
                <Mic class="w-4 h-4 inline-icon" />
                <span v-if="isRecording" class="text-xs font-mono font-medium">00:0{{ recordingSeconds }}s 停止并发送</span>
              </button>
              <button
                @click="triggerAISummary"
                :disabled="isMutedInGroup"
                class="px-2.5 h-8 rounded-lg flex items-center gap-1 text-xs transition-colors font-medium cursor-pointer disabled:opacity-50 border-0"
                :class="showSummaryDrawer
                  ? 'bg-mim-accent text-white'
                  : (themeStore.currentPalette === 'vodyanitsa'
                    ? 'bg-cyan-950/40 border border-cyan-400/40 text-cyan-200 hover:border-cyan-300 hover:bg-cyan-900/50'
                    : 'text-mim-accent hover:bg-mim-accent-subtle')"
                :title="langStore.t('chat.aiSummary')"
              >
                <Loader2 v-if="isSummarizing" class="w-3.5 h-3.5 inline-icon animate-spin" />
                <Sparkles v-else class="w-3.5 h-3.5 inline-icon" />
                <span>{{ isSummarizing ? '提炼中...' : (themeStore.currentPalette === 'vodyanitsa' ? '✦ AI 会话摘要' : langStore.t('chat.aiSummary')) }}</span>
              </button>

              <!-- Vodyanitsa Haute-Craft Stickers Drawer Button -->
              <div class="relative" v-if="themeStore.currentPalette === 'vodyanitsa'">
                <button
                  @click="showVodyanitsaStickers = !showVodyanitsaStickers; showGenshinStickers = false"
                  :disabled="isMutedInGroup"
                  class="px-2.5 h-8 rounded-lg flex items-center gap-1.5 text-xs transition-all disabled:opacity-40 cursor-pointer"
                  :class="showVodyanitsaStickers
                    ? 'bg-cyan-500/25 text-white font-bold border border-cyan-300 shadow-[0_0_12px_rgba(56,189,248,0.4)]'
                    : 'bg-cyan-950/40 text-cyan-200 border border-cyan-400/40 hover:border-cyan-300 hover:bg-cyan-900/50'"
                  title="科洛列夫茨基剧团 · 沃雅妮莎专属微表情"
                >
                  <img :src="iconBtnSticker" alt="" class="w-4 h-4 object-contain inline-block shrink-0" @error="(e: any) => e.target.style.display = 'none'" />
                  <span class="font-medium text-[11px]">沃雅贴纸</span>
                </button>
                <VodyanitsaStickerPicker
                  v-if="showVodyanitsaStickers"
                  @select="handleSelectVodyanitsaSticker"
                  @select-custom="handleSelectCustomSticker"
                  @close="showVodyanitsaStickers = false"
                />
              </div>

              <!-- Vodyanitsa Conch Horn (唤春角笛) Button -->
              <button
                v-if="themeStore.currentPalette === 'vodyanitsa'"
                @click="showWishModal = true; playPrimogemSparkle()"
                :disabled="isMutedInGroup"
                class="px-3 h-8 rounded-lg flex items-center gap-1.5 text-xs transition-all disabled:opacity-40 cursor-pointer bg-gradient-to-r from-amber-600/40 via-amber-500/30 to-amber-700/40 text-amber-200 font-bold border border-amber-400/50 hover:border-amber-300 hover:shadow-[0_0_12px_rgba(234,179,8,0.4)]"
                title="「唤春角笛」- 沃雅妮莎角色祈愿"
              >
                <img :src="iconBtnHorn" alt="" class="w-4 h-4 object-contain inline-block shrink-0" @error="(e: any) => e.target.style.display = 'none'" />
                <span class="font-bold text-[11px] tracking-wider">唤春角笛</span>
              </button>

              <!-- Genshin Quotes & Stickers Drawer Button -->
              <div class="relative" v-if="themeStore.currentPalette !== 'vodyanitsa'">
                <button
                  @click="showGenshinStickers = !showGenshinStickers; showVodyanitsaStickers = false; if (themeStore.currentPalette === 'genshin') playGenshinClick();"
                  :disabled="isMutedInGroup"
                  class="px-2 h-8 rounded-lg flex items-center gap-1 text-xs transition-all disabled:opacity-40 cursor-pointer border-0"
                  :class="showGenshinStickers
                    ? 'bg-amber-500/20 text-amber-600 dark:text-amber-400 font-bold'
                    : 'text-amber-600/80 dark:text-amber-400/80 hover:bg-amber-500/10 hover:text-amber-600 dark:hover:text-amber-300'"
                  title="提瓦特派蒙画作与经典台词"
                >
                  <Sparkles class="w-3.5 h-3.5 inline-icon text-amber-500" />
                  <span class="font-medium text-[11px]">派蒙贴纸</span>
                </button>
                <GenshinStickerPicker
                  v-if="showGenshinStickers"
                  @select-sticker="handleSelectGenshinSticker"
                  @select-chibi="handleSelectChibi"
                  @close="showGenshinStickers = false"
                />
              </div>

              <!-- Genshin Wish Button (Non-Vodyanitsa) -->
              <button
                v-if="themeStore.currentPalette !== 'vodyanitsa'"
                @click="showWishModal = true; playPrimogemSparkle()"
                :disabled="isMutedInGroup"
                class="px-2.5 h-8 rounded-lg flex items-center gap-1 text-xs transition-all disabled:opacity-40 cursor-pointer bg-gradient-to-r from-amber-500/15 via-orange-500/15 to-purple-500/15 text-amber-600 dark:text-amber-300 font-bold hover:shadow-xs border-0"
                title="提瓦特祈愿抽卡系统"
              >
                <Sparkles class="w-3.5 h-3.5 inline-icon text-amber-400" />
                <span class="font-bold text-[11px] tracking-wider">祈愿</span>
              </button>
            </div>

            <!-- Send Action Button -->
            <button
              @click="handleSend"
              :disabled="!inputContent.trim() || isMutedInGroup"
              class="px-4 h-8 rounded-lg disabled:opacity-40 text-xs font-semibold flex items-center gap-1.5 shadow-sm cursor-pointer disabled:cursor-not-allowed transition-all border-0"
              :class="themeStore.currentPalette === 'vodyanitsa'
                ? 'bg-gradient-to-r from-sky-500 to-cyan-400 hover:from-sky-400 hover:to-cyan-300 text-white shadow-[0_0_12px_rgba(56,189,248,0.5)] border border-cyan-200/50'
                : 'mim-btn-primary !border-0'"
            >
              <span>{{ langStore.t('chat.send') }}</span>
              <Send class="w-3 h-3 inline-icon" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State when no conversation is selected -->
    <div v-else class="flex-1 h-full flex flex-col items-center justify-center text-center p-8 select-none text-mim-light-muted dark:text-mim-dark-muted">
      <div class="w-16 h-16 rounded-2xl bg-black/[0.03] dark:bg-white/[0.04] flex items-center justify-center mb-4 text-mim-light-muted dark:text-mim-dark-muted">
        <MessageSquare class="w-8 h-8 opacity-40 inline-icon" />
      </div>
      <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text mb-1">
        {{ langStore.currentLang === 'zh' ? '暂无选中的会话' : 'No Conversation Selected' }}
      </h3>
      <p class="text-xs max-w-xs text-mim-light-muted dark:text-mim-dark-muted">
        {{ langStore.currentLang === 'zh' ? '从左侧选择一个联系人或发起新对话开始交流' : 'Choose a contact or start a new chat from the sidebar' }}
      </p>
    </div>

    <!-- RIGHT GROUP DRAWER (Aligned with ui_group.png & proto_2.png) -->
    <aside
      v-if="isGroup && showGroupDrawer"
      class="w-72 h-full bg-mim-light-list dark:bg-mim-dark-list border-l border-black/[0.04] dark:border-white/[0.04] flex flex-col shrink-0 select-none z-10"
    >
      <!-- Drawer Section 1: 群通知 (ui_group.png) -->
      <div class="p-4 border-b border-black/[0.04] dark:border-white/[0.04] space-y-2">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-mim-light-muted uppercase tracking-wider">
            {{ langStore.t('group.notice') }}
          </span>
          <button
            v-if="isGroupAdmin"
            @click="isEditingNotice = !isEditingNotice"
            class="text-[11px] text-mim-accent hover:underline"
          >
            {{ isEditingNotice ? '取消' : '编辑' }}
          </button>
        </div>

        <!-- Editing Notice Box -->
        <div v-if="isEditingNotice" class="space-y-2">
          <textarea
            v-model="noticeEditContent"
            rows="3"
            class="mim-input w-full p-2 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text resize-none"
          ></textarea>
          <div class="flex justify-end">
            <button
              @click="saveNotice"
              class="px-3 h-6 rounded mim-btn-primary text-[11px] font-medium"
            >
              发布
            </button>
          </div>
        </div>

        <!-- Display Notice -->
        <div v-else>
          <div v-if="groupDetails?.notice" class="text-xs text-mim-light-text dark:text-mim-dark-text leading-relaxed p-2 rounded-lg bg-black/[0.02] dark:bg-white/[0.03]">
            {{ groupDetails.notice }}
          </div>
          <div v-else class="py-6 flex flex-col items-center justify-center text-mim-light-muted gap-1.5">
            <Inbox class="w-8 h-8 opacity-40 inline-icon" />
            <span class="text-xs">{{ langStore.t('group.noNotice') }}</span>
          </div>
        </div>
      </div>

      <!-- Drawer Section 2: 群成员 (ui_group.png) -->
      <div class="flex-1 flex flex-col p-4 overflow-hidden space-y-3">
        <div class="flex items-center justify-between">
          <span class="text-xs font-semibold text-mim-light-muted uppercase tracking-wider">
            {{ langStore.t('group.members') }} {{ groupMembers.length }}/500
          </span>
          <div class="flex items-center gap-1">
            <button
              @click="showInviteModal = true"
              class="w-7 h-7 rounded-lg flex items-center justify-center hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted hover:text-mim-accent transition-colors"
              :title="langStore.t('group.inviteBtn')"
            >
              <UserPlus class="w-4 h-4 inline-icon" />
            </button>
          </div>
        </div>

        <!-- Search Members Input -->
        <div class="relative">
          <Search class="w-3.5 h-3.5 absolute left-2.5 top-2.5 text-mim-light-muted pointer-events-none inline-icon" />
          <input
            v-model="drawerMemberKeyword"
            type="text"
            :placeholder="langStore.t('group.searchMember')"
            class="mim-input w-full h-8 pl-8 pr-2 rounded-lg bg-black/5 dark:bg-white/5 text-xs text-mim-light-text dark:text-mim-dark-text border-none"
          />
        </div>

        <!-- Member List -->
        <div class="flex-1 overflow-y-auto space-y-1 pr-1">
          <div
            v-for="m in filteredDrawerMembers"
            :key="m.user_id"
            @click.stop="openMemberMenu(m, $event)"
            class="w-full p-2 rounded-xl flex items-center justify-between hover:bg-black/[0.04] dark:hover:bg-white/[0.05] cursor-pointer transition-colors relative group"
          >
            <div class="flex items-center gap-2.5 min-w-0">
              <div
                class="w-8 h-8 rounded-lg font-bold text-xs flex items-center justify-center shrink-0 shadow-xs"
                :class="getStatusRingClass(m.role === 1 ? 'busy' : 'offline')"
                :style="{
                  backgroundColor: getAvatarColor(m.user_id || m.nickname, false).bg,
                  color: getAvatarColor(m.user_id || m.nickname, false).text
                }"
              >
                {{ m.avatar || m.nickname.substring(0, 2).toUpperCase() }}
              </div>
              <div class="min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="text-xs font-medium text-mim-light-text dark:text-mim-dark-text truncate">
                    {{ m.nickname }}
                  </span>
                  <span v-if="m.role === 1" class="text-[9px] px-1 py-0.2 rounded bg-amber-500/10 text-amber-600 dark:text-amber-400 font-medium">
                    群主
                  </span>
                  <span v-else-if="m.role === 2" class="text-[9px] px-1 py-0.2 rounded bg-mim-accent-subtle text-mim-accent font-medium">
                    管理员
                  </span>
                </div>
                <span v-if="m.is_muted" class="text-[10px] text-red-500 block">已禁言</span>
              </div>
            </div>

            <!-- Member Action Dots -->
            <MoreHorizontal class="w-3.5 h-3.5 text-mim-light-muted opacity-0 group-hover:opacity-100 transition-opacity inline-icon" />
          </div>
        </div>
      </div>
    </aside>

    <!-- Member Context Popup Menu (ui_group.png) -->
    <div
      v-if="activeMemberMenu"
      @click.stop
      :style="{ top: `${activeMemberMenu.y}px`, right: '290px' }"
      class="fixed z-50 w-36 rounded-xl bg-mim-light-card dark:bg-mim-dark-card shadow-2xl border border-black/10 dark:border-white/10 p-1.5 space-y-1 text-xs"
    >
      <button
        @click="directMessageMember(activeMemberMenu.member)"
        class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-text dark:text-mim-dark-text flex items-center gap-2"
      >
        <MessageSquare class="w-3.5 h-3.5 inline-icon" />
        <span>{{ langStore.t('group.dmUser') }}</span>
      </button>

      <button
        @click="openMemberProfile(activeMemberMenu.member)"
        class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-text dark:text-mim-dark-text flex items-center gap-2"
      >
        <User class="w-3.5 h-3.5 inline-icon" />
        <span>{{ langStore.t('group.viewProfile') }}</span>
      </button>

      <!-- Owner Only: Promote/Demote Admin & Transfer Owner -->
      <template v-if="isGroupOwner && activeMemberMenu.member.user_id !== chatStore.currentUserId">
        <button
          v-if="activeMemberMenu.member.role === 3"
          @click="setGroupRole(activeMemberMenu.member, 2)"
          class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-indigo-500/10 text-indigo-600 flex items-center gap-2"
        >
          <ShieldCheck class="w-3.5 h-3.5 inline-icon" />
          <span>设为管理员</span>
        </button>
        <button
          v-else-if="activeMemberMenu.member.role === 2"
          @click="setGroupRole(activeMemberMenu.member, 3)"
          class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-mim-light-muted hover:text-mim-light-text flex items-center gap-2"
        >
          <ShieldAlert class="w-3.5 h-3.5 inline-icon" />
          <span>取消管理员</span>
        </button>
        <button
          @click="transferGroupOwner(activeMemberMenu.member)"
          class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-amber-500/10 text-amber-600 flex items-center gap-2"
        >
          <Crown class="w-3.5 h-3.5 inline-icon" />
          <span>转让群主</span>
        </button>
      </template>

      <!-- Mute & Kick (Owner can manage anyone; Admin can only manage regular members) -->
      <template v-if="canManageTarget(activeMemberMenu.member)">
        <button
          @click="toggleMuteMember(activeMemberMenu.member)"
          class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-black/5 dark:hover:bg-white/5 text-amber-600 flex items-center gap-2"
        >
          <VolumeX class="w-3.5 h-3.5 inline-icon" />
          <span>{{ activeMemberMenu.member.is_muted ? langStore.t('group.unmuteUser') : langStore.t('group.muteUser') }}</span>
        </button>

        <button
          @click="kickMemberFromGroup(activeMemberMenu.member)"
          class="w-full px-2.5 py-1.5 text-left rounded-lg hover:bg-red-500/10 text-red-600 flex items-center gap-2"
        >
          <UserMinus class="w-3.5 h-3.5 inline-icon" />
          <span>{{ langStore.t('group.kickUser') }}</span>
        </button>
      </template>
    </div>

    <!-- Minimized Floating Call Card (PiP Capsule) -->
    <!-- Shown when call is active AND (manually minimized OR user navigated to a different conversation) -->
    <div
      v-if="chatStore.activeCall && (isCallMinimized || chatStore.activeConvId !== chatStore.activeCall.convId)"
      class="absolute top-20 right-6 z-40 rounded-2xl bg-black/90 dark:bg-black/90 backdrop-blur-xl border border-white/15 p-3 shadow-2xl flex items-center gap-3 text-white select-none transition-all duration-200"
    >
      <div
        class="w-10 h-10 rounded-xl bg-mim-accent text-white flex items-center justify-center text-sm font-bold shrink-0 cursor-pointer shadow-md relative"
        @click="returnToCallConversation"
        :title="langStore.currentLang === 'zh' ? '点击返回通话' : 'Click to return to call'"
      >
        {{ chatStore.activeCall.peerName.substring(0, 2).toUpperCase() }}
        <span class="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-emerald-500 border-2 border-black"></span>
      </div>

      <div class="flex flex-col min-w-[90px] cursor-pointer" @click="returnToCallConversation">
        <div class="flex items-center gap-1.5">
          <span class="text-xs font-semibold truncate max-w-[110px]">{{ chatStore.activeCall.peerName }}</span>
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
        </div>
        <div class="flex items-center gap-1.5 text-[11px] font-mono opacity-80 tabular-nums">
          <Phone class="w-3 h-3 text-emerald-400 inline-icon" />
          <span>{{ formattedCallDuration }}</span>
        </div>
      </div>

      <div class="flex items-center gap-1.5 shrink-0 pl-1 border-l border-white/10">
        <!-- Mic Toggle Button -->
        <button
          @click.stop="chatStore.toggleCallAudio()"
          class="w-8 h-8 rounded-lg flex items-center justify-center transition-colors cursor-pointer"
          :class="chatStore.isCallAudioMuted ? 'bg-amber-600 text-white' : 'bg-white/10 hover:bg-white/20 text-white'"
          :title="chatStore.isCallAudioMuted ? '取消静音' : '静音'"
        >
          <MicOff v-if="chatStore.isCallAudioMuted" class="w-3.5 h-3.5 inline-icon" />
          <Mic v-else class="w-3.5 h-3.5 inline-icon" />
        </button>

        <!-- Video Camera Toggle Button -->
        <button
          @click.stop="chatStore.toggleCallVideo()"
          class="w-8 h-8 rounded-lg flex items-center justify-center transition-colors cursor-pointer"
          :class="chatStore.isCallVideoStopped ? 'bg-amber-600 text-white' : 'bg-white/10 hover:bg-white/20 text-white'"
          :title="chatStore.isCallVideoStopped ? '开启摄像头' : '关闭摄像头'"
        >
          <VideoOff v-if="chatStore.isCallVideoStopped" class="w-3.5 h-3.5 inline-icon" />
          <Video v-else class="w-3.5 h-3.5 inline-icon" />
        </button>

        <!-- Hangup Button -->
        <button
          @click.stop="chatStore.endActiveCall"
          class="w-8 h-8 rounded-lg bg-red-600 hover:bg-red-500 flex items-center justify-center text-white transition-colors cursor-pointer shadow-sm"
          :title="langStore.t('chat.hangup')"
        >
          <PhoneOff class="w-3.5 h-3.5 inline-icon" />
        </button>

        <!-- Return to Call / Expand Button -->
        <button
          @click.stop="returnToCallConversation"
          class="px-2.5 h-8 rounded-lg bg-mim-accent hover:opacity-90 flex items-center gap-1 text-white text-xs font-medium transition-opacity cursor-pointer shadow-sm ml-0.5"
          :title="langStore.currentLang === 'zh' ? '返回全屏通话' : 'Return to call'"
        >
          <Maximize2 class="w-3.5 h-3.5 inline-icon" />
          <span>{{ chatStore.activeConvId === chatStore.activeCall.convId ? (langStore.currentLang === 'zh' ? '展开' : 'Expand') : (langStore.currentLang === 'zh' ? '返回通话' : 'Return') }}</span>
        </button>
      </div>
    </div>

    <!-- Active WebRTC Call Overlay (Expanded Stage) -->
    <!-- Kept in DOM via v-show so video streams and WebRTC connections are NEVER lost or recreated -->
    <div
      v-if="chatStore.activeCall"
      v-show="!isCallMinimized && chatStore.activeConvId === chatStore.activeCall.convId"
      class="absolute inset-x-6 top-20 bottom-24 rounded-2xl bg-black/95 backdrop-blur-xl border border-white/10 flex flex-col justify-between p-6 text-white z-40 shadow-2xl overflow-hidden select-none transition-all duration-200"
    >
      <!-- Top Call Card Bar -->
      <div class="flex items-center justify-between w-full z-20 shrink-0">
        <div class="flex items-center gap-2">
          <span class="px-3 py-1 rounded-full bg-white/10 text-xs font-medium flex items-center gap-1.5 border border-white/10">
            <ShieldCheck class="w-3.5 h-3.5 text-mim-accent inline-icon" />
            <span>{{ chatStore.activeCall.callType.toUpperCase() }} · WebRTC P2P</span>
          </span>
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
        </div>

        <div class="flex items-center gap-2">
          <!-- View User Profile Button -->
          <button
            @click="openPeerProfile"
            class="px-3 py-1.5 rounded-xl bg-white/10 hover:bg-white/20 text-xs flex items-center gap-1.5 transition-colors cursor-pointer text-white/90"
            :title="langStore.t('chat.convInfo')"
          >
            <User class="w-3.5 h-3.5 inline-icon" />
            <span>{{ langStore.currentLang === 'zh' ? '查看资料' : 'User Info' }}</span>
          </button>

          <!-- Minimize Call Button -->
          <button
            @click="isCallMinimized = true"
            class="p-2 rounded-xl bg-white/10 hover:bg-white/20 text-xs flex items-center justify-center transition-colors cursor-pointer text-white/90"
            :title="langStore.currentLang === 'zh' ? '缩小为小窗' : 'Minimize'"
          >
            <Minimize2 class="w-4 h-4 inline-icon" />
          </button>
        </div>
      </div>

      <!-- Dedicated Remote Audio Player (Ensures crystal-clear audio during both video & audio calls) -->
      <audio
        ref="remoteAudioRef"
        autoplay
        playsinline
        class="hidden"
      ></audio>

      <!-- Main Background Video Stream (Swappable between Remote & Local) -->
      <video
        :ref="(el) => bindRemoteVideo(el as HTMLVideoElement)"
        autoplay
        playsinline
        muted
        class="w-full h-full object-cover rounded-2xl absolute inset-0 z-0 bg-neutral-900 transition-transform duration-200"
        :class="[
          { 'hidden': chatStore.activeCall.callType.toLowerCase() === 'audio' || (!isSwappedVideo && (!chatStore.remoteMediaStream || chatStore.isPeerCameraOff)) || (isSwappedVideo && (!chatStore.localMediaStream || chatStore.isCallVideoStopped)) },
          isSwappedVideo ? { '-scale-x-100': isLocalMirrored } : { '-scale-x-100': isRemoteMirrored }
        ]"
      ></video>

      <!-- Video Controls on Top Bar -->
      <div
        v-if="chatStore.activeCall.callType.toLowerCase() !== 'audio'"
        class="absolute top-16 left-6 z-20 flex items-center gap-2"
      >
        <!-- WeChat Style Swap Views Button -->
        <button
          @click="toggleVideoSwap"
          class="px-3 py-1.5 rounded-xl bg-black/60 hover:bg-black/80 backdrop-blur-md text-white text-xs flex items-center gap-1.5 transition-all cursor-pointer border border-white/15 hover:border-mim-accent shadow-lg"
          :title="langStore.currentLang === 'zh' ? '点击互换大小画面位置 (微信式切换)' : 'Swap big and small videos'"
        >
          <ArrowLeftRight class="w-3.5 h-3.5 inline-icon text-mim-accent" />
          <span>{{ isSwappedVideo ? (langStore.currentLang === 'zh' ? '恢复默认布局 (对方大屏)' : 'Reset Layout') : (langStore.currentLang === 'zh' ? '画面互换 (自己大屏)' : 'Swap Views') }}</span>
        </button>

        <!-- Video Mirror Flip Button -->
        <button
          v-if="(!isSwappedVideo && chatStore.remoteMediaStream && !chatStore.isPeerCameraOff) || (isSwappedVideo && chatStore.localMediaStream && !chatStore.isCallVideoStopped)"
          @click="isSwappedVideo ? toggleMirror() : (isRemoteMirrored = !isRemoteMirrored)"
          class="px-3 py-1.5 rounded-xl bg-black/50 hover:bg-black/70 backdrop-blur-md text-white/80 hover:text-white text-xs flex items-center gap-1.5 transition-colors cursor-pointer border border-white/10 shadow-lg"
          :title="langStore.currentLang === 'zh' ? '点击翻转大屏画面左右镜像' : 'Flip mirror'"
        >
          <FlipHorizontal class="w-3.5 h-3.5 inline-icon" />
          <span>{{ (isSwappedVideo ? isLocalMirrored : isRemoteMirrored) ? (langStore.currentLang === 'zh' ? '大屏镜像: 已翻转' : 'Mirrored') : (langStore.currentLang === 'zh' ? '大屏镜像: 正常' : 'Normal') }}</span>
        </button>
      </div>

      <!-- PiP Floating Window (Click to swap views like WeChat!) -->
      <div
        class="w-36 h-48 rounded-2xl border-2 border-white/20 shadow-2xl absolute top-16 right-6 z-20 overflow-hidden bg-black/70 group/localvid select-none transition-all duration-200 hover:scale-105 hover:border-mim-accent active:scale-95"
        :class="{ 'hidden': chatStore.activeCall.callType.toLowerCase() === 'audio' }"
      >
        <video
          :ref="(el) => bindLocalVideo(el as HTMLVideoElement)"
          autoplay
          playsinline
          :muted="!isSwappedVideo"
          class="w-full h-full object-cover transition-transform duration-200"
          :class="[
            { 'hidden': (!isSwappedVideo && chatStore.isCallVideoStopped) || (isSwappedVideo && chatStore.isPeerCameraOff) },
            isSwappedVideo ? { '-scale-x-100': isRemoteMirrored } : { '-scale-x-100': isLocalMirrored }
          ]"
        ></video>

        <!-- Fallback inside PiP if video stream is not yet ready or camera is off -->
        <div
          v-if="(!isSwappedVideo && (!chatStore.localMediaStream || chatStore.isCallVideoStopped)) || (isSwappedVideo && (!chatStore.remoteMediaStream || chatStore.isPeerCameraOff))"
          class="absolute inset-0 flex flex-col items-center justify-center bg-neutral-900/90 text-white p-2 text-center pointer-events-none"
        >
          <div class="w-10 h-10 rounded-full bg-white/10 flex items-center justify-center text-xs font-bold mb-1">
            {{ isSwappedVideo ? chatStore.activeCall.peerName.substring(0, 2).toUpperCase() : 'ME' }}
          </div>
          <span class="text-[10px] text-white/60">
            {{ isSwappedVideo ? (chatStore.isPeerCameraOff ? (langStore.currentLang === 'zh' ? '对方已关摄像头' : 'Camera off') : (langStore.currentLang === 'zh' ? '等待视频...' : 'Waiting...')) : (chatStore.isCallVideoStopped ? (langStore.currentLang === 'zh' ? '摄像头已关闭' : 'Camera off') : (langStore.currentLang === 'zh' ? '开启摄像头...' : 'Starting...')) }}
          </span>
        </div>

        <!-- Full Click Interceptor Layer (100% reliable click to swap views) -->
        <div
          @click="toggleVideoSwap"
          class="absolute inset-0 z-10 cursor-pointer"
          :title="langStore.currentLang === 'zh' ? '点击与大屏画面互换位置 (微信式切换)' : 'Click to swap video positions'"
        ></div>

        <!-- Floating Swap Indicator Badge -->
        <div class="absolute bottom-2 inset-x-2 z-20 flex items-center justify-center pointer-events-none">
          <span class="px-2 py-0.5 rounded-md bg-black/80 backdrop-blur-md text-[10px] text-white font-medium flex items-center gap-1 border border-white/20 group-hover/localvid:border-mim-accent group-hover/localvid:text-mim-accent transition-colors shadow">
            <ArrowLeftRight class="w-2.5 h-2.5 inline-icon" />
            <span>{{ isSwappedVideo ? (langStore.currentLang === 'zh' ? '对方 (点击互换)' : 'Peer') : (langStore.currentLang === 'zh' ? '自己 (点击互换)' : 'Me') }}</span>
          </span>
        </div>

        <!-- Quick flip button on hover -->
        <button
          @click.stop="isSwappedVideo ? (isRemoteMirrored = !isRemoteMirrored) : toggleMirror()"
          class="absolute top-2 right-2 z-20 p-1 rounded-md bg-black/70 hover:bg-black/90 text-white/90 opacity-0 group-hover/localvid:opacity-100 transition-opacity cursor-pointer text-[10px]"
          :title="langStore.currentLang === 'zh' ? '点击翻转小窗画面镜像' : 'Flip mirror'"
        >
          <FlipHorizontal class="w-3.5 h-3.5 inline-icon" />
        </button>
      </div>

      <!-- Audio Avatar & Call Information (When Audio Call or No Video on main canvas) -->
      <div
        v-if="chatStore.activeCall.callType.toLowerCase() === 'audio' || (!isSwappedVideo && (!chatStore.remoteMediaStream || chatStore.isPeerCameraOff)) || (isSwappedVideo && (!chatStore.localMediaStream || chatStore.isCallVideoStopped))"
        class="flex flex-col items-center gap-3 z-10 my-auto text-center"
      >
        <div class="relative cursor-pointer group" @click="openPeerProfile" title="点击查看用户资料">
          <div class="w-24 h-24 rounded-3xl bg-mim-accent text-white flex items-center justify-center text-3xl font-bold shadow-2xl ring-4 ring-white/10 group-hover:scale-105 transition-transform">
            {{ (isSwappedVideo ? '我' : chatStore.activeCall.peerName).substring(0, 2).toUpperCase() }}
          </div>
          <span class="absolute -bottom-1 -right-1 w-6 h-6 rounded-full bg-emerald-500 border-2 border-black flex items-center justify-center">
            <Check class="w-3.5 h-3.5 text-white inline-icon" />
          </span>
        </div>
        <div>
          <h3
            @click="openPeerProfile"
            class="text-xl font-bold tracking-tight hover:text-mim-accent cursor-pointer transition-colors"
            title="点击查看用户资料"
          >
            {{ isSwappedVideo ? (langStore.currentLang === 'zh' ? '我的本地画面' : 'My Local Video') : chatStore.activeCall.peerName }}
          </h3>
          <p class="text-xs text-white/60 mt-1">
            {{ (!isSwappedVideo && chatStore.isPeerCameraOff) ? (langStore.currentLang === 'zh' ? '对方已关闭摄像头' : 'Peer has turned off camera') : (isSwappedVideo && chatStore.isCallVideoStopped ? (langStore.currentLang === 'zh' ? '本地摄像头已关闭' : 'Local camera turned off') : `${chatStore.activeCall.callType.toUpperCase()} · WebRTC P2P 端到端加密连通中`) }}
          </p>
        </div>
        <span class="px-4 py-1.5 rounded-full bg-white/10 text-sm font-mono tabular-nums border border-white/10 mt-2">
          {{ formattedCallDuration }}
        </span>
      </div>

      <!-- Controls: Mute Mic, Toggle Video Camera, Hangup -->
      <div class="flex items-center justify-center gap-6 z-20 pb-4">
        <!-- Mic Toggle Button -->
        <button
          @click="chatStore.toggleCallAudio()"
          class="w-12 h-12 rounded-full flex items-center justify-center transition-colors cursor-pointer"
          :class="chatStore.isCallAudioMuted ? 'bg-amber-600 text-white' : 'bg-white/15 text-white hover:bg-white/25'"
          :title="chatStore.isCallAudioMuted ? '取消静音' : '静音麦克风'"
        >
          <MicOff v-if="chatStore.isCallAudioMuted" class="w-5 h-5 inline-icon" />
          <Mic v-else class="w-5 h-5 inline-icon" />
        </button>

        <!-- Video Camera Toggle Button -->
        <button
          @click="chatStore.toggleCallVideo()"
          class="w-12 h-12 rounded-full flex items-center justify-center transition-colors cursor-pointer"
          :class="chatStore.isCallVideoStopped ? 'bg-amber-600 text-white' : 'bg-white/15 text-white hover:bg-white/25'"
          :title="chatStore.isCallVideoStopped ? '开启摄像头' : '关闭摄像头'"
        >
          <VideoOff v-if="chatStore.isCallVideoStopped" class="w-5 h-5 inline-icon" />
          <Video v-else class="w-5 h-5 inline-icon" />
        </button>

        <!-- Mirror Mode Toggle Button -->
        <button
          v-if="chatStore.activeCall.callType === 'video' || chatStore.localMediaStream"
          @click="toggleMirror"
          class="w-12 h-12 rounded-full flex items-center justify-center transition-colors cursor-pointer"
          :class="isLocalMirrored ? 'bg-mim-accent text-white shadow-md shadow-mim-accent/30' : 'bg-white/15 text-white/70 hover:bg-white/25'"
          :title="isLocalMirrored ? (langStore.currentLang === 'zh' ? '镜像模式已开启 (左右像照镜子一样自然，点击关闭)' : 'Mirror mode active (Click to disable)') : (langStore.currentLang === 'zh' ? '镜像模式已关闭 (点击开启镜像)' : 'Mirror mode off (Click to enable)')"
        >
          <FlipHorizontal class="w-5 h-5 inline-icon" />
        </button>

        <!-- Hangup Button -->
        <button
          @click="chatStore.endActiveCall"
          class="w-14 h-14 rounded-full bg-red-600 hover:bg-red-500 flex items-center justify-center text-white transition-colors shadow-lg shadow-red-600/40 cursor-pointer"
          :title="langStore.t('chat.hangup')"
        >
          <PhoneOff class="w-6 h-6 inline-icon" />
        </button>
      </div>
    </div>

    <!-- Incoming Call Overlay -->
    <div
      v-if="chatStore.incomingCall"
      class="fixed inset-0 bg-black/70 backdrop-blur-md z-50 flex items-center justify-center p-4 select-none"
    >
      <div class="w-full max-w-sm rounded-2xl bg-mim-light-card dark:bg-mim-dark-card p-6 shadow-2xl border border-black/10 dark:border-white/10 flex flex-col items-center gap-4 text-center">
        <div class="w-16 h-16 rounded-2xl bg-mim-accent text-white flex items-center justify-center font-bold text-xl animate-pulse shadow-md">
          {{ chatStore.incomingCall.callerName.substring(0, 2).toUpperCase() }}
        </div>
        <div>
          <h3 class="text-base font-semibold text-mim-light-text dark:text-mim-dark-text">
            {{ chatStore.incomingCall.callerName }}
          </h3>
          <p class="text-xs text-mim-light-muted dark:text-mim-dark-muted mt-0.5">
            {{ langStore.t('chat.incomingCallDesc') }}
          </p>
        </div>
        <div class="flex items-center gap-6 mt-2">
          <button
            @click="chatStore.rejectIncomingCall"
            class="w-12 h-12 rounded-full bg-red-600 hover:bg-red-500 text-white flex items-center justify-center transition-colors shadow-lg shadow-red-600/30"
            :title="langStore.t('chat.decline')"
          >
            <PhoneOff class="w-5 h-5 inline-icon" />
          </button>
          <button
            @click="chatStore.acceptIncomingCall"
            class="w-12 h-12 rounded-full bg-emerald-600 hover:bg-emerald-500 text-white flex items-center justify-center transition-colors shadow-lg shadow-emerald-600/30 animate-bounce"
            :title="langStore.t('chat.accept')"
          >
            <Phone class="w-5 h-5 inline-icon" />
          </button>
        </div>
      </div>
    </div>

    <!-- Lightbox Modal for Images -->
    <div
      v-if="lightboxImage"
      @click="lightboxImage = null"
      class="fixed inset-0 bg-black/90 backdrop-blur-md z-50 flex items-center justify-center p-6 cursor-zoom-out"
    >
      <img :src="resolveApiUrl(lightboxImage)" class="max-w-full max-h-full rounded-xl object-contain shadow-2xl" />
    </div>

    <!-- Group Settings Modal (ui_group_set.png) -->
    <GroupSettingsModal
      v-if="showGroupSettingsModal && currentGroupId"
      :group-id="currentGroupId"
      @close="showGroupSettingsModal = false"
      @updated="refreshGroupData"
      @disbanded="handleGroupDisbanded"
    />

    <!-- Invite Modal (ui_invite.png) -->
    <InviteModal
      v-if="showInviteModal && currentGroupId"
      :group-id="currentGroupId"
      :existing-member-ids="groupMembers.map((m) => m.user_id)"
      @close="showInviteModal = false"
      @invited="refreshGroupData"
    />

    <!-- User Profile Modal -->
    <UserProfileModal
      v-if="selectedUserProfile"
      :user="selectedUserProfile"
      @close="selectedUserProfile = null"
    />

    <!-- Genshin Wish Gacha Modal -->
    <GenshinWishModal
      :visible="showWishModal"
      @close="showWishModal = false"
      @share="handleShareWishResult"
    />

    <!-- Custom Background Modal -->
    <CustomBgModal
      v-if="showBgModal"
      @close="showBgModal = false"
    />

    <!-- AI Summary Drawer (Structured Meeting/Chat Minutes) -->
    <AISummaryDrawer
      :visible="showSummaryDrawer"
      :loading="isSummarizing"
      :summary="summaryData.summary"
      :message-count="summaryData.messageCount"
      :participants="summaryData.participants"
      :conversation-title="chatStore.activeConversation?.title || ''"
      @close="showSummaryDrawer = false"
      @refresh="fetchAISummary(true)"
      @send="onSendSummaryToChat"
    />
  </main>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, watch, onMounted, onUnmounted, onBeforeUnmount, toRaw } from 'vue';
import { useRouter } from 'vue-router';
import {
  ArrowDown,
  Phone,
  Video,
  MoreHorizontal,
  ShieldCheck,
  ShieldAlert,
  Crown,
  Check,
  CheckCheck,
  Clock,
  Copy,
  Reply,
  Undo2,
  Play,
  Pause,
  X,
  Palette,
  Image,
  Paperclip,
  Mic,
  MicOff,
  Sparkles,
  Send,
  PhoneOff,
  VideoOff,
  Loader2,
  UploadCloud,
  FileText,
  Download,
  Users,
  Inbox,
  UserPlus,
  Search,
  MessageSquare,
  User,
  VolumeX,
  UserMinus,
  Minimize2,
  Maximize2,
  FlipHorizontal,
  ArrowLeftRight,
  Trash2,
  FolderPlus,
} from 'lucide-vue-next';
import { useChatStore, MessageItem } from '../stores/chat';
import { useLangStore } from '../stores/lang';
import { useContextMenuStore, ContextMenuItem } from '../stores/contextMenu';
import { confirmDialog, alertDialog } from '../stores/dialog';
import { getAvatarColor, getStatusRingClass } from '../utils/avatar';
import GroupSettingsModal from './GroupSettingsModal.vue';
import InviteModal from './InviteModal.vue';
import UserProfileModal from './UserProfileModal.vue';
import FileMessageCard from './FileMessageCard.vue';
import GenshinStickerPicker from './GenshinStickerPicker.vue';
import GenshinWishModal from './GenshinWishModal.vue';
import VodyanitsaStickerPicker from './VodyanitsaStickerPicker.vue';
import CustomBgModal from './CustomBgModal.vue';
import AISmartReplyBar, { type SmartReplySuggestion } from './AISmartReplyBar.vue';
import AISummaryDrawer from './AISummaryDrawer.vue';
import { useThemeStore } from '../stores/theme';
import type { GenshinSticker } from '../assets/genshin/stickers';
import { CHIBI_STICKERS, type ChibiSticker } from '../assets/genshin/chibiStickers';
import { getVodyanitsaStickerById } from '../assets/vodyanitsa/vodyanitsaStickers';
import vodyanitsaAvatar from '../assets/vodyanitsa/images/vodyanitsa_avatar.png';
import iconLily from '../assets/vodyanitsa/dock/nav_contacts_lily.png';
import iconConch from '../assets/vodyanitsa/dock/nav_calls_conch.png';
import iconCrest from '../assets/vodyanitsa/images/header_vodyanitsa_crest.png';
import iconPalette from '../assets/vodyanitsa/images/icon_stage_palette.png';
import iconBtnSticker from '../assets/vodyanitsa/images/icon_btn_sticker.png';
import iconBtnHorn from '../assets/vodyanitsa/images/icon_btn_horn.png';
import { useCustomStickersStore, type CustomSticker } from '../stores/customStickers';
import { playGenshinClick, playPrimogemSparkle } from '../utils/genshinAudio';
import { resolveApiUrl } from '../utils/api';

const router = useRouter();
const chatStore = useChatStore();
const langStore = useLangStore();
const themeStore = useThemeStore();
const contextMenuStore = useContextMenuStore();
const customStore = useCustomStickersStore();

const showGenshinStickers = ref(false);
const showVodyanitsaStickers = ref(false);
const showWishModal = ref(false);
const showBgModal = ref(false);

const handleSelectGenshinSticker = (sticker: GenshinSticker) => {
  chatStore.sendMessage(`【${sticker.character}】${sticker.quote}`, activeQuote.value || undefined);
  activeQuote.value = null;
  displayedQuote.value = null;
  showGenshinStickers.value = false;
  scrollToBottom(true);
};

const handleSelectChibi = (chibi: ChibiSticker) => {
  chatStore.sendMessage(`[PAIMON_CHIBI]:${chibi.id}`, activeQuote.value || undefined);
  activeQuote.value = null;
  displayedQuote.value = null;
  showGenshinStickers.value = false;
  scrollToBottom(true);
};

const handleSelectVodyanitsaSticker = (stickerId: string) => {
  chatStore.sendMessage(`[VODYANITSA_STICKER]:${stickerId}`, activeQuote.value || undefined);
  activeQuote.value = null;
  displayedQuote.value = null;
  showVodyanitsaStickers.value = false;
  scrollToBottom(true);
};

const handleSelectCustomSticker = (sticker: CustomSticker) => {
  const payload = JSON.stringify({ id: sticker.id, name: sticker.name, url: sticker.url });
  chatStore.sendMessage(`[CUSTOM_STICKER]:${encodeURIComponent(payload)}`, activeQuote.value || undefined);
  activeQuote.value = null;
  displayedQuote.value = null;
  showVodyanitsaStickers.value = false;
  scrollToBottom(true);
};

const isCustomSticker = (content: string) => {
  return typeof content === 'string' && content.includes('[CUSTOM_STICKER]:');
};

const parseCustomSticker = (content: string): { name: string; url: string } | null => {
  try {
    const raw = content.replace('[CUSTOM_STICKER]:', '').trim();
    const found = customStore.getStickerById(raw);
    if (found) return { name: found.name, url: found.url };
    const parsed = JSON.parse(decodeURIComponent(raw));
    return { name: parsed.name || '自定义表情', url: parsed.url };
  } catch {
    return null;
  }
};

const isVodyanitsaSticker = (content: string) => {
  return typeof content === 'string' && content.includes('[VODYANITSA_STICKER]:');
};

const parseVodyanitsaStickerId = (content: string) => {
  const match = content.match(/\[VODYANITSA_STICKER\]:\s*([a-zA-Z0-9_]+)/);
  return match ? match[1] : content.replace('[VODYANITSA_STICKER]:', '').trim();
};

const handleShareWishResult = (payload: { bannerName: string; pulls: any[]; summaryText: string }) => {
  chatStore.sendMessage(`[GENSHIN_WISH]:${JSON.stringify(payload)}`);
  scrollToBottom(true);
};

const isPaimonChibi = (content: string) => {
  return typeof content === 'string' && content.includes('[PAIMON_CHIBI]:');
};

const parseChibiId = (content: string) => {
  const match = content.match(/\[PAIMON_CHIBI\]:\s*([a-zA-Z0-9_]+)/);
  return match ? match[1] : content.replace('[PAIMON_CHIBI]:', '').trim();
};

const getChibiById = (id: string) => {
  return CHIBI_STICKERS.find((c) => c.id === id);
};

const isGenshinWish = (content: string) => {
  return typeof content === 'string' && content.startsWith('[GENSHIN_WISH]:');
};

const parseWishData = (content: string) => {
  try {
    const raw = content.replace('[GENSHIN_WISH]:', '').trim();
    return JSON.parse(raw);
  } catch {
    return { bannerName: '提瓦特祈愿', pulls: [], summaryText: '祈愿报告解析中' };
  }
};

const VISION_ELEMENTS = [
  { name: '岩', color: '#d97706' },
  { name: '水', color: '#0284c7' },
  { name: '雷', color: '#7c3aed' },
  { name: '草', color: '#16a34a' },
  { name: '风', color: '#0d9488' },
  { name: '火', color: '#dc2626' },
  { name: '冰', color: '#0ea5e9' },
];

const getVisionElement = (name: string): string => {
  if (!name) return '岩';
  if (name.includes('沃雅妮莎')) return '水';
  if (name.includes('钟离')) return '岩';
  if (name.includes('芙宁娜') || name.includes('水')) return '水';
  if (name.includes('雷') || name.includes('影')) return '雷';
  if (name.includes('纳西妲') || name.includes('草')) return '草';
  if (name.includes('温迪') || name.includes('风')) return '风';
  if (name.includes('胡桃') || name.includes('可莉') || name.includes('火')) return '火';
  let hash = 0;
  for (let i = 0; i < name.length; i++) hash += name.charCodeAt(i);
  return VISION_ELEMENTS[Math.abs(hash) % VISION_ELEMENTS.length].name;
};

const getVisionColor = (name: string): string => {
  const el = getVisionElement(name);
  return VISION_ELEMENTS.find((v) => v.name === el)?.color || '#caa159';
};

const getGenshinCharacterTitle = (name: string): string => {
  if (!name) return '✦ 提瓦特伙伴';
  if (name.includes('沃雅妮莎')) return '✦ 幽歌萦渊 · 首席女高音';
  if (name.includes('旅行者')) return '✦ 异乡的旅人';
  if (name.includes('钟离')) return '✦ 往生堂客卿';
  if (name.includes('胡桃')) return '✦ 往生堂堂主';
  if (name.includes('芙宁娜')) return '✦ 不休独舞';
  if (name.includes('雷电将军') || name.includes('影')) return '✦ 一心净土';
  if (name.includes('纳西妲')) return '✦ 白草净华';
  if (name.includes('温迪')) return '✦ 吟游诗人';
  if (name.includes('可莉')) return '✦ 逃跑的太阳';
  if (name.includes('刻晴')) return '✦ 霆霓快雨';
  if (name.includes('甘雨')) return '✦ 循循守月';
  if (name.includes('神里绫华') || name.includes('神里')) return '✦ 白鹭霜华';
  if (name.includes('派蒙')) return '✦ 最好的向导';
  return '✦ 提瓦特伙伴';
};

const getSpeakerName = (msg: { isMe?: boolean; senderName?: string }): string => {
  if (msg.isMe) return '旅行者';
  if (msg.senderName && msg.senderName !== 'Partner') return msg.senderName;
  if (chatStore.activeConversation?.title) return chatStore.activeConversation.title;
  return '提瓦特旅伴';
};

const inputContent = ref('');
const textareaRef = ref<HTMLTextAreaElement | null>(null);
const showSmartReply = ref(false);
const isReplying = ref(false);
const replySuggestions = ref<SmartReplySuggestion[]>([]);
let smartReplyAbortController: AbortController | null = null;
const activeQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const displayedQuote = ref<{ msgId: string; userName: string; content: string } | null>(null);
const copiedMsgId = ref<string | null>(null);
const scrollContainer = ref<HTMLDivElement | null>(null);

// 规范 1：中文输入法选字回车防误触 (IME Composition Enter Guard)
let isComposing = false;
let compositionCooldown = false;
let cooldownTimer: any = null;

const onCompositionStart = () => {
  isComposing = true;
  compositionCooldown = true;
  if (cooldownTimer) clearTimeout(cooldownTimer);
};

const onCompositionEnd = () => {
  isComposing = false;
  if (cooldownTimer) clearTimeout(cooldownTimer);
  cooldownTimer = setTimeout(() => {
    compositionCooldown = false;
  }, 100); // 100ms 阻断选字尾随回车，同时保障熟练打字快速发送
};

const onInputKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    if (isComposing || compositionCooldown || e.keyCode === 229) {
      e.preventDefault();
      return;
    }
    e.preventDefault();
    handleSend();
  }
};

// 规范 2：智能滚动吸底与「跳至最新」悬浮药丸 (Smart Scroll Bottom Following & Jump Pill)
const autoFollowBottom = ref(true);
const showJumpToBottom = ref(false);
let isProgrammaticScrolling = false;
let scrollTimer: ReturnType<typeof setTimeout> | null = null;

const onMessagesScroll = (e: Event) => {
  if (isProgrammaticScrolling) return;
  const el = e.target as HTMLElement;
  if (!el) return;
  const dist = el.scrollHeight - el.scrollTop - el.clientHeight;
  // 离底部 40px 以内才自动跟随新消息；用户一滚动就停止跟随
  autoFollowBottom.value = dist <= 40;
  // 离底部超过 120px 时显示「跳至最新」按钮
  showJumpToBottom.value = dist > 120;
};

const imageInputRef = ref<HTMLInputElement | null>(null);
const fileInputRef = ref<HTMLInputElement | null>(null);

const localVideoRef = ref<HTMLVideoElement | null>(null);
const remoteVideoRef = ref<HTMLVideoElement | null>(null);
const remoteAudioRef = ref<HTMLAudioElement | null>(null);

const isDragging = ref(false);
const isRecording = ref(false);
const recordingSeconds = ref(0);
const isSummarizing = ref(false);
const showSummaryDrawer = ref(false);
const summaryData = reactive({
  summary: '',
  messageCount: 0,
  participants: 0,
});
let summaryAbortController: AbortController | null = null;
const playingVoiceId = ref<string | null>(null);
const lightboxImage = ref<string | null>(null);
const isCallMinimized = ref(false);
const isSwappedVideo = ref(false);
const isLocalMirrored = ref(true); // Default to TRUE (mirror image like a real mirror, matching Zoom/FaceTime/WeChat standard)
const isRemoteMirrored = ref(false);

const toggleMirror = () => {
  isLocalMirrored.value = !isLocalMirrored.value;
};

const toggleVideoSwap = () => {
  isSwappedVideo.value = !isSwappedVideo.value;
  syncVideoStreams();
};

const returnToCallConversation = () => {
  if (chatStore.activeCall?.convId) {
    chatStore.selectConversation(chatStore.activeCall.convId);
    isCallMinimized.value = false;
  }
};

const bindRemoteVideo = (el: HTMLVideoElement | null) => {
  if (remoteVideoRef.value !== el) {
    remoteVideoRef.value = el;
    if (el) syncVideoStreams();
  }
};

const bindLocalVideo = (el: HTMLVideoElement | null) => {
  if (localVideoRef.value !== el) {
    localVideoRef.value = el;
    if (el) syncVideoStreams();
  }
};

const syncVideoStreams = () => {
  nextTick(() => {
    const rawLocal = chatStore.localMediaStream ? toRaw(chatStore.localMediaStream) : null;
    const rawRemote = chatStore.remoteMediaStream ? toRaw(chatStore.remoteMediaStream) : null;

    const mainStream = isSwappedVideo.value ? rawLocal : rawRemote;
    const pipStream = isSwappedVideo.value ? rawRemote : rawLocal;

    // Dedicated remote audio playback: keeps playing even when video is hidden or in audio call
    if (remoteAudioRef.value) {
      if (rawRemote && remoteAudioRef.value.srcObject !== rawRemote) {
        remoteAudioRef.value.srcObject = rawRemote;
      }
      if (rawRemote) {
        remoteAudioRef.value.play().catch(() => {});
      }
    }

    if (remoteVideoRef.value) {
      if (remoteVideoRef.value.srcObject !== mainStream) {
        remoteVideoRef.value.srcObject = mainStream;
      }
      // If remoteAudioRef is present, keep video muted to prevent audio doubling; otherwise mute local
      remoteVideoRef.value.muted = isSwappedVideo.value || (remoteAudioRef.value !== null);
      if (mainStream) {
        remoteVideoRef.value.play().catch(() => {});
      }
    }
    if (localVideoRef.value) {
      if (localVideoRef.value.srcObject !== pipStream) {
        localVideoRef.value.srcObject = pipStream;
      }
      localVideoRef.value.muted = true; // Local preview is always muted to prevent local feedback
      if (pipStream) {
        localVideoRef.value.play().catch(() => {});
      }
    }
  });
};

watch(isSwappedVideo, () => {
  syncVideoStreams();
});

watch(
  [
    () => isCallMinimized.value,
    () => chatStore.activeConvId,
    () => chatStore.localMediaStream,
    () => chatStore.remoteMediaStream,
    () => chatStore.isCallVideoStopped,
    () => chatStore.isPeerCameraOff,
    () => chatStore.activeCall?.callType,
  ],
  () => {
    syncVideoStreams();
  },
  { immediate: true, flush: 'post' }
);

watch(() => chatStore.activeCall, (val) => {
  if (!val) {
    isCallMinimized.value = false;
    isSwappedVideo.value = false;
  } else {
    syncVideoStreams();
  }
});

// Reset AI summary drawer and auto-minimize call when switching to another conversation
watch(() => chatStore.activeConvId, (newConvId) => {
  if (summaryAbortController) {
    summaryAbortController.abort();
    summaryAbortController = null;
  }
  showSummaryDrawer.value = false;
  isSummarizing.value = false;
  summaryData.summary = '';
  summaryData.messageCount = 0;
  summaryData.participants = 0;

  if (chatStore.activeCall) {
    if (newConvId !== chatStore.activeCall.convId) {
      isCallMinimized.value = true;
    } else {
      isCallMinimized.value = false;
    }
  }
});

// Group Feature States
const showGroupDrawer = ref(true);
const showGroupSettingsModal = ref(false);
const showInviteModal = ref(false);
const selectedUserProfile = ref<any | null>(null);

const groupDetails = ref<any | null>(null);
const groupMembers = ref<any[]>([]);
const isEditingNotice = ref(false);
const noticeEditContent = ref('');
const drawerMemberKeyword = ref('');

const activeMemberMenu = ref<{
  member: any;
  x: number;
  y: number;
} | null>(null);

const isGroup = computed(() => chatStore.activeConversation?.type === 'group');

const currentGroupId = computed(() => {
  if (!isGroup.value) return 0;
  const match = chatStore.activeConvId.match(/^group_(\d+)$/);
  if (match) return parseInt(match[1], 10);
  return chatStore.activeConversation?.targetId || 0;
});

const isGroupOwner = computed(() => {
  const currentM = groupMembers.value.find((m) => m.user_id === chatStore.currentUserId);
  return currentM?.role === 1 || groupDetails.value?.owner_id === chatStore.currentUserId;
});

const isGroupAdmin = computed(() => {
  const currentM = groupMembers.value.find((m) => m.user_id === chatStore.currentUserId);
  return currentM && (currentM.role === 1 || currentM.role === 2);
});

const myGroupRole = computed(() => {
  if (isGroupOwner.value) return 1;
  const currentM = groupMembers.value.find((m) => m.user_id === chatStore.currentUserId);
  return currentM?.role || 3;
});

const canManageTarget = (target: any) => {
  if (!target || target.user_id === chatStore.currentUserId) return false;
  if (isGroupOwner.value) return true;
  if (isGroupAdmin.value) return target.role === 3;
  return false;
};

const isAllMuted = computed(() => groupDetails.value?.is_all_muted === 1);

const isMutedInGroup = computed(() => {
  if (!isGroup.value) return false;
  if (isGroupAdmin.value) return false;
  if (isAllMuted.value) return true;
  const currentM = groupMembers.value.find((m) => m.user_id === chatStore.currentUserId);
  return currentM?.is_muted;
});

const filteredDrawerMembers = computed(() => {
  if (!drawerMemberKeyword.value.trim()) return groupMembers.value;
  const kw = drawerMemberKeyword.value.toLowerCase();
  return groupMembers.value.filter((m) => m.nickname.toLowerCase().includes(kw) || m.account.toLowerCase().includes(kw));
});

const formattedCallDuration = computed(() => {
  const d = chatStore.callDuration;
  const mins = Math.floor(d / 60).toString().padStart(2, '0');
  const secs = (d % 60).toString().padStart(2, '0');
  return `00:${mins}:${secs}`;
});

const clearQuote = () => {
  activeQuote.value = null;
  setTimeout(() => {
    displayedQuote.value = null;
  }, 250);
};

const handleSend = () => {
  if (!inputContent.value.trim() || isMutedInGroup.value) return;

  if (themeStore.currentPalette === 'genshin') {
    playGenshinClick();
  }

  chatStore.sendMessage(inputContent.value, activeQuote.value || undefined);
  inputContent.value = '';
  clearQuote();

  scrollToBottom(true);
};

const scrollToBottom = (force = false) => {
  if (!force && !autoFollowBottom.value) return;
  if (force) {
    isProgrammaticScrolling = true;
    showJumpToBottom.value = false;
    autoFollowBottom.value = true;
    if (scrollTimer) clearTimeout(scrollTimer);
    scrollTimer = setTimeout(() => {
      isProgrammaticScrolling = false;
    }, 350);
  }
  nextTick(() => {
    if (scrollContainer.value) {
      if (force) {
        scrollContainer.value.scrollTo({
          top: scrollContainer.value.scrollHeight,
          behavior: 'smooth',
        });
      } else {
        scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight;
      }
    }
  });
};

const copyMessage = (msgId: string, content: string) => {
  navigator.clipboard.writeText(content);
  copiedMsgId.value = msgId;
  setTimeout(() => {
    if (copiedMsgId.value === msgId) {
      copiedMsgId.value = null;
    }
  }, 2500);
};

const setQuote = (msg: MessageItem) => {
  const quoteData = {
    msgId: msg.id,
    userName: msg.senderName,
    content: msg.msgType === 'file' ? `[文件] ${msg.fileName || '文件'}` : msg.content,
  };
  displayedQuote.value = quoteData;
  activeQuote.value = quoteData;
};

const handleMessageContextMenu = (e: MouseEvent, msg: MessageItem) => {
  const items: ContextMenuItem[] = [
    {
      id: 'copy',
      label: '复制内容',
      icon: Copy,
      shortcut: 'Ctrl+C',
      action: () => copyMessage(msg.id, msg.content),
    },
    {
      id: 'quote',
      label: '引用回复',
      icon: Reply,
      action: () => setQuote(msg),
    },
  ];

  if (msg.msgType === 'file' || msg.msgType === 'image') {
    items.push({
      id: 'download',
      label: '另存为 / 下载',
      icon: Download,
      action: () => {
        const link = document.createElement('a');
        link.href = msg.content;
        link.download = msg.fileName || 'file';
        link.target = '_blank';
        link.click();
      },
    });
  }

  if (msg.msgType === 'image') {
    items.push({
      id: 'add-image-sticker',
      label: '添加到自定义表情',
      icon: FolderPlus,
      action: () => {
        const url = resolveApiUrl(msg.content);
        customStore.addSticker('收藏图片表情', url);
      },
    });
  } else if (isVodyanitsaSticker(msg.content)) {
    const stk = getVodyanitsaStickerById(parseVodyanitsaStickerId(msg.content));
    if (stk) {
      items.push({
        id: 'add-vodyanitsa-sticker',
        label: '收藏到自定义表情',
        icon: FolderPlus,
        action: () => {
          customStore.addSticker(stk.name, stk.imageSrc);
        },
      });
    }
  } else if (isCustomSticker(msg.content)) {
    const parsed = parseCustomSticker(msg.content);
    if (parsed) {
      items.push({
        id: 'add-custom-sticker',
        label: '收藏到自定义表情',
        icon: FolderPlus,
        action: () => {
          customStore.addSticker(parsed.name, parsed.url);
        },
      });
    }
  }

  if (msg.isMe && !msg.isRevoked) {
    items.push({
      id: 'revoke',
      label: '撤回此消息',
      icon: Undo2,
      divided: true,
      action: () => chatStore.revokeMessage(msg.id),
    });
  }

  items.push({
    id: 'delete',
    label: '删除消息',
    icon: Trash2,
    divided: !msg.isMe,
    danger: true,
    action: () => {
      chatStore.deleteMessage(chatStore.activeConvId, msg.id);
    },
  });

  contextMenuStore.show(e, items);
};

// Media & File Upload
const triggerImageInput = () => {
  imageInputRef.value?.click();
};

const triggerFileInput = () => {
  fileInputRef.value?.click();
};

const handleImageSelected = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  const files = input.files;
  if (!files || files.length === 0) return;
  const file = files[0];
  input.value = ''; // Always clear value to allow re-uploading the same image
  await uploadAndSendFile(file, 'image');
};

const handleFileSelected = async (e: Event) => {
  const input = e.target as HTMLInputElement;
  const files = input.files;
  if (!files || files.length === 0) return;
  const file = files[0];
  input.value = ''; // Always clear value to allow re-uploading the same file
  await uploadAndSendFile(file, 'file');
};

const handleFileDrop = async (e: DragEvent) => {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (!files || files.length === 0) return;

  const file = files[0];
  const isImg = file.type.startsWith('image/');
  await uploadAndSendFile(file, isImg ? 'image' : 'file');
};

const handlePaste = async (e: ClipboardEvent) => {
  const items = e.clipboardData?.items;
  if (!items) return;
  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    if (item.type.startsWith('image/')) {
      e.preventDefault();
      const file = item.getAsFile();
      if (file) {
        const ext = item.type.split('/')[1] || 'png';
        const namedFile = new File([file], `paste_${Date.now()}.${ext}`, { type: item.type });
        await uploadAndSendFile(namedFile, 'image');
      }
      break;
    }
  }
};

const uploadAndSendFile = async (file: File, type: 'image' | 'file') => {
  // Capture local absolute path if available in WebView2/Tauri environment
  const localPath = (file as any).path || undefined;
  const formData = new FormData();
  formData.append('file', file);
  formData.append('name', file.name);

  try {
    const res = await fetch(`/api/file/upload?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: formData,
    });
    const json = await res.json();
    if (json.code === 0 && json.data) {
      const formattedSize = (file.size / 1024 / 1024).toFixed(2) + ' MB';
      let fileUrl = json.data.file_url;
      if (!fileUrl.includes('name=')) {
        fileUrl += (fileUrl.includes('?') ? '&' : '?') + `name=${encodeURIComponent(file.name)}`;
      }
      if (!fileUrl.includes('size=')) {
        fileUrl += `&size=${encodeURIComponent(formattedSize)}`;
      }

      chatStore.sendMessage(fileUrl, undefined, type, {
        fileName: file.name,
        fileSize: formattedSize,
        filePath: localPath,
      });
      scrollToBottom(true);
    } else {
      console.error('File upload server error:', json);
      alertDialog({
        title: '发送失败',
        message: json.msg || '文件上传失败，请重试',
        type: 'danger',
      });
    }
  } catch (err: any) {
    console.error('File upload failed:', err);
    alertDialog({
      title: '发送失败',
      message: err.message || '网络连接异常，无法上传文件',
      type: 'danger',
    });
  }
};

// Real Voice Note Recording & Playback (MediaRecorder & Web Audio)
let mediaRecorder: MediaRecorder | null = null;
let audioChunks: Blob[] = [];
let recordingStartTime = 0;
let recordingTimer: any = null;
let activeAudio: HTMLAudioElement | null = null;

const toggleVoiceRecording = async () => {
  if (isRecording.value) {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop();
    }
    isRecording.value = false;
    clearInterval(recordingTimer);
  } else {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      audioChunks = [];
      recordingStartTime = Date.now();
      recordingSeconds.value = 0;

      let mimeType = 'audio/webm';
      let fileExt = 'webm';
      if (typeof MediaRecorder.isTypeSupported === 'function' && MediaRecorder.isTypeSupported('audio/mp4')) {
        mimeType = 'audio/mp4';
        fileExt = 'm4a';
      } else if (typeof MediaRecorder.isTypeSupported === 'function' && MediaRecorder.isTypeSupported('audio/webm;codecs=opus')) {
        mimeType = 'audio/webm;codecs=opus';
        fileExt = 'webm';
      } else if (typeof MediaRecorder.isTypeSupported === 'function' && MediaRecorder.isTypeSupported('audio/ogg;codecs=opus')) {
        mimeType = 'audio/ogg;codecs=opus';
        fileExt = 'ogg';
      }

      mediaRecorder = new MediaRecorder(stream, { mimeType });
      mediaRecorder.ondataavailable = (e) => {
        if (e.data && e.data.size > 0) {
          audioChunks.push(e.data);
        }
      };

      mediaRecorder.onstop = async () => {
        stream.getTracks().forEach((t) => t.stop());
        const durationSec = Math.max(1, Math.round((Date.now() - recordingStartTime) / 1000));
        const audioBlob = new Blob(audioChunks, { type: mimeType });
        const file = new File([audioBlob], `voice_${Date.now()}.${fileExt}`, { type: mimeType });

        const formData = new FormData();
        formData.append('file', file);
        formData.append('name', file.name);

        try {
          const res = await fetch(`/api/file/upload?user_id=${chatStore.currentUserId}`, {
            method: 'POST',
            headers: { Authorization: `Bearer ${chatStore.userToken}` },
            body: formData,
          });
          const json = await res.json();
          if (json.code === 0 && json.data) {
            const fileUrl = json.data.file_url;
            chatStore.sendMessage(fileUrl, undefined, 'voice', {
              duration: durationSec,
              fileName: file.name,
            });
            scrollToBottom(true);
          }
        } catch (err) {
          console.error('Upload voice failed:', err);
        }
      };

      mediaRecorder.start();
      isRecording.value = true;
      recordingTimer = setInterval(() => {
        recordingSeconds.value = Math.max(1, Math.round((Date.now() - recordingStartTime) / 1000));
      }, 500);
    } catch (err) {
      console.warn('Microphone access denied or unavailable:', err);
      alertDialog({
        title: '麦克风权限提示',
        message: '无法访问麦克风。请在浏览器或操作系统设置中允许麦克风权限后重试。',
        type: 'warning',
      });
    }
  }
};

const playVoice = (msgId: string) => {
  const msg = chatStore.activeMessages.find((m) => m.id === msgId);
  if (!msg) return;

  if (playingVoiceId.value === msgId) {
    if (activeAudio) {
      activeAudio.pause();
      activeAudio = null;
    }
    playingVoiceId.value = null;
    return;
  }

  if (activeAudio) {
    activeAudio.pause();
    activeAudio = null;
  }

  const audioUrl = msg.content?.startsWith('http') || msg.content?.startsWith('/')
    ? msg.content
    : '';

  if (audioUrl) {
    activeAudio = new Audio(resolveApiUrl(audioUrl));
    playingVoiceId.value = msgId;
    activeAudio.play().catch((err) => {
      console.warn('Audio playback error:', err);
      playingVoiceId.value = null;
    });
    activeAudio.onended = () => {
      playingVoiceId.value = null;
      activeAudio = null;
    };
  } else {
    playingVoiceId.value = msgId;
    setTimeout(() => {
      if (playingVoiceId.value === msgId) playingVoiceId.value = null;
    }, (msg.extra?.duration || 3) * 1000);
  }
};

// AI Structured Summary Drawer & Service
const fetchAISummary = async (forceRefresh = false) => {
  if (!chatStore.activeConvId) return;

  if (summaryAbortController) {
    summaryAbortController.abort();
  }
  summaryAbortController = new AbortController();
  const signal = summaryAbortController.signal;

  isSummarizing.value = true;
  if (forceRefresh) {
    summaryData.summary = '';
  }

  try {
    const res = await fetch(`/api/ai/summary?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        conversation_id: chatStore.activeConvId,
        lang: langStore.currentLang,
        limit: 50,
      }),
      signal,
    });

    if (signal.aborted) return;

    const json = await res.json();
    const data = json.data || json;
    summaryData.summary = data.summary || '';
    summaryData.messageCount = data.message_count || data.count || 0;
    summaryData.participants = data.participants || 0;
  } catch (err: any) {
    if (err.name === 'AbortError') return;
    console.error('Fetch AI summary error:', err);
  } finally {
    if (!signal.aborted) {
      isSummarizing.value = false;
    }
  }
};

const openSummaryDrawer = () => {
  showSummaryDrawer.value = !showSummaryDrawer.value;
  if (showSummaryDrawer.value) {
    if (!summaryData.summary && !isSummarizing.value) {
      fetchAISummary();
    }
  }
};

const triggerAISummary = () => {
  showSummaryDrawer.value = true;
  if (!summaryData.summary && !isSummarizing.value) {
    fetchAISummary();
  }
};

const onSendSummaryToChat = (summaryText: string) => {
  if (!summaryText.trim()) return;
  chatStore.sendMessage(summaryText);
  showSummaryDrawer.value = false;
  scrollToBottom(true);
};

// AI Smart Reply Methods
const fetchSmartReplies = async () => {
  if (!chatStore.activeConvId || isReplying.value) return;

  if (smartReplyAbortController) {
    smartReplyAbortController.abort();
  }
  smartReplyAbortController = new AbortController();
  const targetConvId = chatStore.activeConvId;

  isReplying.value = true;
  try {
    const res = await fetch(`/api/ai/reply?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        conversation_id: targetConvId,
      }),
      signal: smartReplyAbortController.signal,
    });
    const json = await res.json();
    if (chatStore.activeConvId !== targetConvId) return;

    if (json.code === 0 && json.data && json.data.suggestions) {
      replySuggestions.value = json.data.suggestions;
    }
  } catch (err: any) {
    if (err.name === 'AbortError') return;
    console.error('Failed to fetch smart replies:', err);
  } finally {
    if (chatStore.activeConvId === targetConvId) {
      isReplying.value = false;
    }
  }
};

const toggleSmartReply = async () => {
  if (showSmartReply.value) {
    showSmartReply.value = false;
    return;
  }
  showSmartReply.value = true;
  await fetchSmartReplies();
};

const onSmartReplySelect = (text: string) => {
  inputContent.value = text;
  nextTick(() => {
    textareaRef.value?.focus();
  });
};

const onSmartReplySend = (text: string) => {
  inputContent.value = text;
  handleSend();
  showSmartReply.value = false;
};

// WebRTC Calls
const startCall = (type: 'audio' | 'video') => {
  chatStore.initiateWebRTCCall(type);
};

const openLightbox = (url: string) => {
  lightboxImage.value = url;
};

// Group Data Fetching
const refreshGroupData = async () => {
  if (!currentGroupId.value) return;
  try {
    const [infoRes, membersRes] = await Promise.all([
      fetch(`/api/group/info?group_id=${currentGroupId.value}&user_id=${chatStore.currentUserId}`, {
        headers: { Authorization: `Bearer ${chatStore.userToken}` },
      }),
      fetch(`/api/group/members?group_id=${currentGroupId.value}&user_id=${chatStore.currentUserId}`, {
        headers: { Authorization: `Bearer ${chatStore.userToken}` },
      }),
    ]);
    const infoJson = await infoRes.json();
    const membersJson = await membersRes.json();
    if (infoJson.code === 0 && infoJson.data) {
      groupDetails.value = infoJson.data;
      noticeEditContent.value = infoJson.data.notice || '';
    }
    if (membersJson.code === 0 && Array.isArray(membersJson.data)) {
      groupMembers.value = membersJson.data;
    }
  } catch (_) {}
};

const saveNotice = async () => {
  if (!currentGroupId.value) return;
  try {
    const res = await fetch(`/api/group/update?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: currentGroupId.value,
        notice: noticeEditContent.value,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      if (groupDetails.value) {
        groupDetails.value.notice = noticeEditContent.value;
      }
      isEditingNotice.value = false;
    }
  } catch (_) {}
};

const openMemberMenu = (member: any, event: MouseEvent) => {
  activeMemberMenu.value = {
    member,
    x: event.clientX,
    y: Math.min(event.clientY, window.innerHeight - 180),
  };
};

const directMessageMember = (member: any) => {
  activeMemberMenu.value = null;
  const targetId = member.user_id;
  const u1 = Math.min(chatStore.currentUserId, targetId);
  const u2 = Math.max(chatStore.currentUserId, targetId);
  const convId = `${u1}_${u2}`;
  chatStore.selectConversation(convId);
  router.push(`/chats/${convId}`);
};

const openMemberProfile = (member: any) => {
  activeMemberMenu.value = null;
  selectedUserProfile.value = member;
};

const openPeerProfile = () => {
  const targetId = chatStore.activeConversation?.targetId;
  if (!targetId) return;
  selectedUserProfile.value = {
    user_id: targetId,
    nickname: chatStore.activeConversation?.title,
    avatar: chatStore.activeConversation?.avatarText,
    account: `user_${targetId}`,
    signature: 'Verified Mim Cluster Peer',
  };
};

const openUserProfileById = (senderId: number, senderName: string) => {
  selectedUserProfile.value = {
    user_id: senderId,
    nickname: senderName,
    account: `user_${senderId}`,
    signature: 'Verified Mim Cluster Member',
  };
};

const toggleMuteMember = async (member: any) => {
  activeMemberMenu.value = null;
  const duration = member.is_muted ? 0 : 600;
  try {
    const res = await fetch(`/api/group/mute-member?user_id=${chatStore.currentUserId}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${chatStore.userToken}`,
      },
      body: JSON.stringify({
        group_id: currentGroupId.value,
        target_id: member.user_id,
        duration,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      member.is_muted = !member.is_muted;
    }
  } catch (_) {}
};

const kickMemberFromGroup = async (member: any) => {
  activeMemberMenu.value = null;
  const ok = await confirmDialog({
    title: '移出群聊',
    message: `确定要将 ${member.nickname} 踢出群聊吗？`,
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
        group_id: currentGroupId.value,
        target_id: member.user_id,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      groupMembers.value = groupMembers.value.filter((m) => m.user_id !== member.user_id);
    }
  } catch (_) {}
};

const setGroupRole = async (member: any, newRole: number) => {
  activeMemberMenu.value = null;
  const roleName = newRole === 2 ? '管理员' : '普通成员';
  const ok = await confirmDialog({
    title: '调整群权限',
    message: `确定将 ${member.nickname} 设为 ${roleName} 吗？`,
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
        group_id: currentGroupId.value,
        target_id: member.user_id,
        role: newRole,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      member.role = newRole;
      await refreshGroupData();
    }
  } catch (_) {}
};

const transferGroupOwner = async (member: any) => {
  activeMemberMenu.value = null;
  const ok = await confirmDialog({
    title: '转让群主',
    message: `【警告】确定要将群主转让给 ${member.nickname} 吗？\n转让后您将降级为群管理员。`,
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
        group_id: currentGroupId.value,
        new_owner_id: member.user_id,
      }),
    });
    const json = await res.json();
    if (json.code === 0) {
      await refreshGroupData();
    }
  } catch (_) {}
};

const handleGroupDisbanded = () => {
  if (chatStore.conversations.length > 0) {
    chatStore.selectConversation(chatStore.conversations[0].id);
    router.push(`/chats/${chatStore.conversations[0].id}`);
  }
};

const closePopupOnGlobalClick = () => {
  activeMemberMenu.value = null;
};

watch(
  () => chatStore.activeConvId,
  () => {
    if (smartReplyAbortController) {
      smartReplyAbortController.abort();
      smartReplyAbortController = null;
    }
    isReplying.value = false;
    showSmartReply.value = false;
    replySuggestions.value = [];
    if (isGroup.value) {
      refreshGroupData();
    }
    autoFollowBottom.value = true;
    showJumpToBottom.value = false;
    scrollToBottom(true);
  },
  { immediate: true }
);

watch(
  () => chatStore.activeMessages.length,
  () => {
    scrollToBottom(false);
  }
);

const handleEscKey = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    if (showSummaryDrawer.value) {
      showSummaryDrawer.value = false;
      return;
    }
    if (showSmartReply.value) {
      showSmartReply.value = false;
    }
  }
};

onMounted(() => {
  window.addEventListener('click', closePopupOnGlobalClick);
  window.addEventListener('keydown', handleEscKey);
  if (isGroup.value) {
    refreshGroupData();
  }
});

onBeforeUnmount(() => {
  if (cooldownTimer) {
    clearTimeout(cooldownTimer);
    cooldownTimer = null;
  }
  if (scrollTimer) {
    clearTimeout(scrollTimer);
    scrollTimer = null;
  }
});

onUnmounted(() => {
  if (summaryAbortController) {
    summaryAbortController.abort();
    summaryAbortController = null;
  }
  if (cooldownTimer) {
    clearTimeout(cooldownTimer);
    cooldownTimer = null;
  }
  if (scrollTimer) {
    clearTimeout(scrollTimer);
    scrollTimer = null;
  }
  window.removeEventListener('click', closePopupOnGlobalClick);
  window.removeEventListener('keydown', handleEscKey);
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}
</style>
