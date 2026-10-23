import { defineStore } from 'pinia';
import { ref } from 'vue';

export type Language = 'zh' | 'en';

export const translations = {
  zh: {
    // Navigation Dock
    nav: {
      chats: '消息会话',
      contacts: '通讯录与群组',
      calls: '音视频通话',
      settings: '系统设置',
      theme: '主题模式',
      language: '语言切换',
      switchAccount: '切换演示账号',
      backendWs: '长连接网关',
      online: '在线',
      offline: '离线',
    },
    // Conversation List
    conv: {
      messages: '消息',
      searchPlaceholder: '搜索会话、联系人、消息...',
      newChat: '新建会话',
      singleChat: '私聊会话',
      groupChat: '团队群组',
      noMessages: '暂无消息记录',
      recalledPreview: '撤回了一条消息',
      imagePreview: '[图片]',
      filePreview: '[文件]',
      voicePreview: '[语音消息]',
      aiPreview: '[AI 智能摘要]',
    },
    // Chat Stage
    chat: {
      onlineE2EE: '在线 (端到端加密会话)',
      clusterEncrypted: '集群加密会话',
      audioCall: '语音通话',
      videoCall: '视频通话',
      convInfo: '会话详情',
      copy: '复制',
      reply: '回复 / 引用',
      recall: '撤回消息',
      recalledSelf: '你撤回了一条消息',
      recalledPeer: '对方撤回了一条消息',
      replyingTo: '正在回复',
      inputPlaceholder: '输入消息... (Enter 发送，Shift+Enter 换行)',
      send: '发送',
      attachImage: '发送图片',
      attachFile: '发送文件 (SHA-256 秒传)',
      voiceRecording: '语音录制',
      aiSummary: 'AI 会话摘要',
      dropOverlay: '释放文件即可触发即时秒传上传',
      hangup: '挂断',
      accept: '接听',
      decline: '拒绝',
      incomingCall: '实时来电',
      incomingCallDesc: '收到来自好友的音视频呼叫...',
      voiceMessage: '语音留言',
      aiCopilot: 'Mim 智能助理',
      aiSynthesisTitle: 'Mim 智能会话提炼总结',
    },
    // Contacts & Teams
    contacts: {
      title: '通讯录与团队',
      subtitle: '查看并管理你的组织架构、团队群组与好友关系',
      addFriend: '添加好友',
      newGroup: '新建群聊',
      teamGroups: '团队群组',
      directContacts: '我的好友',
      membersCount: '名成员',
      chatBtn: '发消息',
      addFriendModalTitle: '添加好友',
      searchPlaceholder: '输入账号或昵称精确查找 (如 sarah, michael, david)...',
      searchBtn: '查找',
      addBtn: '添加',
      noUserFound: '未找到匹配的用户，可尝试搜索 "sarah" 或 "david"',
      createGroupModalTitle: '创建新团队群组',
      groupTitleLabel: '群聊名称',
      groupTitlePlaceholder: '例如：核心基础设施架构组',
      groupNoticeLabel: '群公告',
      groupNoticePlaceholder: '填写面向全员的公告信息...',
      cancel: '取消',
      create: '立即创建',
    },
    // Calls
    calls: {
      title: '通话记录与 WebRTC 归档',
      subtitle: '实时点对点音视频会话归档与历史',
      newAudioCall: '发起语音通话',
      newVideoCall: '发起视频通话',
      empty: '暂无通话记录，点击上方按钮即可体验真实点对点 WebRTC 通话！',
      audio: '语音通话',
      video: '视频通话',
      duration: '通话时长',
      completed: '已完成',
      missed: '未接听',
      redial: '再次呼叫',
    },
    // Settings Modal
    settings: {
      title: 'Mim 企业级系统与协议架构设置',
      transportTitle: '传输层底层架构',
      transportDesc: '严格对齐 24 字节内存私有二进制帧，全字段网络大端序 (Big-Endian)，单调递增 SeqId，CRC32 全包校验，CAS 会话预览防乱序守卫。',
      dfaTitle: 'DFA 敏感词安全过滤',
      dfaDesc: '基于确定性有限状态自动机 (DFA) 的毫秒级违规关键词检测，入库前自动脱敏替换为星号，守护企业合规与安全边界。',
      dedupTitle: 'SHA-256 存储去重秒传',
      dedupDesc: '客户端预先计算文件的 SHA-256 哈希值，命中服务已有存储文件时无需再次落盘，秒级直接返回可用 URL。',
      webrtcTitle: 'WebRTC 纯内存信令通道',
      webrtcDesc: 'Offer、Answer、Candidate 信令纯走 WebSocket 内存转发，绝不入库污染消息表；仅在通话结束归档通话时长记录。',
      done: '完成',
    },
  },
  en: {
    // Navigation Dock
    nav: {
      chats: 'Conversations',
      contacts: 'Contacts & Teams',
      calls: 'Audio/Video Calls',
      settings: 'Settings',
      theme: 'Theme Mode',
      language: 'Language',
      switchAccount: 'Switch Demo Account',
      backendWs: 'Cluster WebSocket',
      online: 'Online',
      offline: 'Offline',
    },
    // Conversation List
    conv: {
      messages: 'Messages',
      searchPlaceholder: 'Search conversations, files...',
      newChat: 'New Chat',
      singleChat: 'Direct Message',
      groupChat: 'Group Chat',
      noMessages: 'No messages yet',
      recalledPreview: 'recalled a message',
      imagePreview: '[Image]',
      filePreview: '[File]',
      voicePreview: '[Voice Note]',
      aiPreview: '[AI Summary]',
    },
    // Chat Stage
    chat: {
      onlineE2EE: 'Online (E2EE Active)',
      clusterEncrypted: 'Cluster Encrypted',
      audioCall: 'Audio Call',
      videoCall: 'Video Call',
      convInfo: 'Conversation Info',
      copy: 'Copy',
      reply: 'Reply / Quote',
      recall: 'Recall Message',
      recalledSelf: 'You recalled a message',
      recalledPeer: 'Recalled a message',
      replyingTo: 'Replying to',
      inputPlaceholder: 'Write a message... (Press Enter to send, Shift+Enter for newline)',
      send: 'Send',
      attachImage: 'Attach Image',
      attachFile: 'Attach Document (Instant Upload)',
      voiceRecording: 'Voice Recording',
      aiSummary: 'AI Summary',
      dropOverlay: 'Drop files here to upload (SHA-256 Instant Upload)',
      hangup: 'Hang Up',
      accept: 'Accept',
      decline: 'Decline',
      incomingCall: 'Incoming Call',
      incomingCallDesc: 'Incoming real-time WebRTC audio/video call...',
      voiceMessage: 'Voice Message',
      aiCopilot: 'Mim AI Copilot',
      aiSynthesisTitle: 'Mim AI Intelligence Summary',
    },
    // Contacts & Teams
    contacts: {
      title: 'Contacts & Teams',
      subtitle: 'Manage your directory, team groups and relationships',
      addFriend: 'Add Friend',
      newGroup: 'New Group',
      teamGroups: 'Team Groups',
      directContacts: 'Direct Contacts',
      membersCount: 'members',
      chatBtn: 'Chat',
      addFriendModalTitle: 'Add Friend',
      searchPlaceholder: 'Search by account or nickname (e.g. sarah, michael)...',
      searchBtn: 'Search',
      addBtn: 'Add',
      noUserFound: 'No users found. Try searching "sarah" or "david".',
      createGroupModalTitle: 'Create New Team Group',
      groupTitleLabel: 'Group Title',
      groupTitlePlaceholder: 'e.g. Distributed Core Team',
      groupNoticeLabel: 'Group Announcement',
      groupNoticePlaceholder: 'Notice for team members...',
      cancel: 'Cancel',
      create: 'Create',
    },
    // Calls
    calls: {
      title: 'Call History & WebRTC Sessions',
      subtitle: 'Real-time peer-to-peer audio and video records',
      newAudioCall: 'New Audio Call',
      newVideoCall: 'New Video Call',
      empty: 'No call records yet. Click above to start a peer-to-peer WebRTC session!',
      audio: 'Audio Call',
      video: 'Video Call',
      duration: 'Duration',
      completed: 'Completed',
      missed: 'Missed',
      redial: 'Redial',
    },
    // Settings Modal
    settings: {
      title: 'Mim Enterprise System Settings',
      transportTitle: 'Transport Layer Architecture',
      transportDesc: 'Network Byte Order (Big-Endian) across all multi-byte fields. Monotonic gap-tolerant SeqId, CRC32 packet integrity validation, CAS concurrency guard for conversation previews.',
      dfaTitle: 'DFA Sensitive Word Filter',
      dfaDesc: 'Real-time keyword filtering for unauthorized leaks, phishing, and abuse. Automatically masks malicious payloads with asterisks prior to database persistence.',
      dedupTitle: 'SHA-256 Storage Deduplication',
      dedupDesc: 'Files are hashed with SHA-256 upon selection. Matching existing file hashes bypass disk writing and return instantaneous zero-latency upload responses.',
      webrtcTitle: 'WebRTC Ephemeral Signaling',
      webrtcDesc: 'Offer, Answer, and ICE Candidate packets (0x0500) route purely in WebSocket memory without polluting message tables. Only finalized call logs are archived.',
      done: 'Done',
    },
  },
};

export const useLangStore = defineStore('lang', () => {
  const savedLang = (localStorage.getItem('mim_lang') as Language) || 'zh';
  const currentLang = ref<Language>(savedLang);

  const setLang = (lang: Language) => {
    currentLang.value = lang;
    localStorage.setItem('mim_lang', lang);
  };

  const toggleLang = () => {
    setLang(currentLang.value === 'zh' ? 'en' : 'zh');
  };

  // Helper to fetch translated text
  const t = (path: string): string => {
    const keys = path.split('.');
    let curr: any = translations[currentLang.value];
    for (const k of keys) {
      if (curr && typeof curr === 'object' && k in curr) {
        curr = curr[k];
      } else {
        return path;
      }
    }
    return typeof curr === 'string' ? curr : path;
  };

  return {
    currentLang,
    setLang,
    toggleLang,
    t,
  };
});
