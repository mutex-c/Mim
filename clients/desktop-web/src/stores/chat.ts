import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { router } from '../router';
import { CmdType, encodeFrame, decodeFrame, FLAG_JSON } from '../protocol/frame';
import { alertDialog } from './dialog';
import { getWsUrl, resolveApiUrl } from '../utils/api';

export interface MessageItem {
  id: string;
  clientMsgId: string;
  seqId: number;
  senderId: number;
  senderName: string;
  content: string;
  msgType: 'text' | 'image' | 'voice' | 'file' | 'video' | 'revoke' | 'ai_summary';
  time: string;
  isMe: boolean;
  status: 'sending' | 'sent' | 'delivered' | 'read' | 'failed';
  quote?: {
    msgId: string;
    userName: string;
    content: string;
  };
  duration?: number;
  fileName?: string;
  fileSize?: string;
  filePath?: string;
  isRevoked?: boolean;
  extra?: {
    fileName?: string;
    fileSize?: string;
    filePath?: string;
    duration?: number;
    audioUrl?: string;
  };
}

export interface ConversationItem {
  id: string;
  targetId: number;
  title: string;
  type: 'single' | 'group';
  avatarText: string;
  lastMessage: string;
  lastTime: string;
  unreadCount: number;
  isPinned: boolean;
  isMuted: boolean;
  isOnline?: boolean;
}

export interface UserProfile {
  id: number;
  account: string;
  nickname: string;
  avatar: string;
  signature: string;
  email: string;
  phone: string;
  role: number;
  created_at?: string;
}

export interface UserConfig {
  user_id: number;
  search_by_account: number;
  search_by_nickname: number;
  friend_verify_mode: number;
  friend_verify_question: string;
  friend_verify_answer: string;
  recall_time_limit: number;
}

export interface AIConfigData {
  provider: string; // 'deepseek' | 'kimi' | 'openai' | 'ollama' | 'custom'
  base_url: string;
  api_key: string;
  model: string;
  temperature: number;
}

export interface DemoUser {
  user: {
    id: number;
    account: string;
    nickname: string;
    avatar: string;
    role: number;
  };
  token: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
}

export const useChatStore = defineStore('chat', () => {
  const savedUid = typeof localStorage !== 'undefined' ? Number(localStorage.getItem('mim_active_uid')) : 0;
  const savedToken = typeof localStorage !== 'undefined' ? localStorage.getItem('mim_token') || '' : '';
  const isExplicitlyLoggedOut = typeof localStorage !== 'undefined' && localStorage.getItem('mim_logged_out') === 'true';

  const savedSidebarWidth = typeof localStorage !== 'undefined' ? Number(localStorage.getItem('mim_sidebar_width')) : 0;
  const sidebarWidth = ref<number>(savedSidebarWidth || 320);

  const setSidebarWidth = (w: number) => {
    const clamped = Math.max(220, Math.min(520, w));
    sidebarWidth.value = clamped;
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('mim_sidebar_width', String(clamped));
    }
  };

  const currentUserId = ref<number>(savedUid || 0);
  const currentUserName = ref<string>('');
  const currentUserAvatar = ref<string>('');
  const userToken = ref<string>(savedToken);
  const isLoggedIn = ref<boolean>(!isExplicitlyLoggedOut && !!savedToken && savedUid > 0);
  const showAuthModal = ref<boolean>(false);
  const isWsConnected = ref<boolean>(false);
  const demoUsers = ref<DemoUser[]>([]);

  const userProfile = ref<UserProfile | null>(null);
  const userConfig = ref<UserConfig | null>(null);

  const fetchUserProfile = async () => {
    if (!currentUserId.value || !userToken.value) return;
    try {
      const res = await fetch(`/api/user/profile?user_id=${currentUserId.value}`, {
        headers: { Authorization: `Bearer ${userToken.value}` },
      });
      const json = await res.json();
      if (json.code === 0 && json.data) {
        userProfile.value = json.data.user;
        userConfig.value = json.data.config;
      }
    } catch (_) {}
  };

  const aiConfig = ref<AIConfigData>({
    provider: 'deepseek',
    base_url: 'https://api.deepseek.com/v1',
    api_key: '',
    model: 'deepseek-chat',
    temperature: 0.7,
  });

  const fetchAIConfig = async (): Promise<void> => {
    if (!currentUserId.value || !userToken.value) return;
    try {
      const url = resolveApiUrl(`/api/ai/config?user_id=${currentUserId.value}`);
      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${userToken.value}` },
      });
      const json = await res.json();
      if (json.code === 0 && json.data) {
        aiConfig.value = {
          provider: json.data.provider || 'deepseek',
          base_url: json.data.base_url || 'https://api.deepseek.com/v1',
          api_key: json.data.api_key || '',
          model: json.data.model || 'deepseek-chat',
          temperature: typeof json.data.temperature === 'number' ? json.data.temperature : 0.7,
        };
      }
    } catch (err) {
      console.warn('[AI] Failed to fetch AI config:', err);
    }
  };

  const saveAIConfig = async (config: AIConfigData): Promise<{ success: boolean; error?: string }> => {
    if (!currentUserId.value || !userToken.value) {
      return { success: false, error: 'User not logged in' };
    }
    try {
      const url = resolveApiUrl(`/api/ai/config?user_id=${currentUserId.value}`);
      const res = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify(config),
      });
      const json = await res.json();
      if (json.code === 0) {
        aiConfig.value = { ...config };
        return { success: true };
      }
      return { success: false, error: json.msg || 'Failed to save AI config' };
    } catch (err: any) {
      return { success: false, error: err.message || 'Network error' };
    }
  };

  const testAIConnection = async (configToTest?: AIConfigData): Promise<{ success: boolean; latencyMs?: number; error?: string }> => {
    const targetConfig = configToTest || aiConfig.value;
    if (!targetConfig.base_url) {
      return { success: false, error: 'Base URL 不能为空' };
    }
    if (!targetConfig.api_key && targetConfig.provider !== 'ollama') {
      return { success: false, error: 'API Key 不能为空' };
    }

    const startTime = performance.now();

    // 1. Direct endpoint test (works seamlessly in Tauri Desktop and local Ollama)
    try {
      let endpoint = targetConfig.base_url.replace(/\/+$/, '');
      if (!endpoint.endsWith('/chat/completions')) {
        endpoint = `${endpoint}/chat/completions`;
      }

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 8000);

      const res = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${targetConfig.api_key}`,
        },
        body: JSON.stringify({
          model: targetConfig.model || 'deepseek-chat',
          messages: [{ role: 'user', content: 'Ping' }],
          max_tokens: 5,
        }),
        signal: controller.signal,
      });
      clearTimeout(timeoutId);

      const latencyMs = Math.round(performance.now() - startTime);

      if (res.ok) {
        return { success: true, latencyMs };
      }
      if (res.status === 401) {
        return { success: false, latencyMs, error: '认证失败: API Key 无效 (401)' };
      }
      if (res.status === 404) {
        return { success: false, latencyMs, error: '端点不存在或模型名称错误 (404)' };
      }
      const errText = await res.text().catch(() => '');
      if (res.status === 400 && errText.toLowerCase().includes('model')) {
        return { success: true, latencyMs };
      }
      return { success: false, latencyMs, error: `HTTP ${res.status}: ${errText.slice(0, 100)}` };
    } catch (directErr: any) {
      // 2. If direct fetch fails (e.g. browser CORS for public APIs), fallback to backend proxy test
      if (currentUserId.value && userToken.value) {
        try {
          await saveAIConfig(targetConfig);
          const convId = activeConvId.value || (conversations.value[0]?.id || '');
          if (convId) {
            const replyUrl = resolveApiUrl(`/api/ai/reply?user_id=${currentUserId.value}`);
            const replyRes = await fetch(replyUrl, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${userToken.value}`,
              },
              body: JSON.stringify({
                conversation_id: convId,
                count: 1,
              }),
            });
            const latencyMs = Math.round(performance.now() - startTime);
            if (replyRes.ok) {
              const data = await replyRes.json();
              if (data.code === 0) {
                return { success: true, latencyMs };
              }
            }
          }
        } catch (_) {}
      }

      const latencyMs = Math.round(performance.now() - startTime);
      const msg = directErr.name === 'AbortError' ? '请求超时 (超过 8 秒)' : (directErr.message || '网络连接失败');
      return { success: false, latencyMs, error: msg };
    }
  };

  const activeTab = ref<string>('chats'); // 'chats', 'contacts', 'calls'
  const activeConvId = ref<string>('');
  const conversations = ref<ConversationItem[]>([]);
  const messagesByConv = ref<Record<string, MessageItem[]>>({});

  // WebRTC Call State
  const incomingCall = ref<{ callerId: number; callerName: string; callType: string; callId: string } | null>(null);
  const activeCall = ref<{ peerId: number; peerName: string; callType: string; callId: string; convId: string } | null>(null);
  const callDuration = ref<number>(0);
  let callTimerInterval: any = null;

  const localMediaStream = ref<MediaStream | null>(null);
  const remoteMediaStream = ref<MediaStream | null>(null);
  const isCallAudioMuted = ref<boolean>(false);
  const isCallVideoStopped = ref<boolean>(false);
  const isPeerCameraOff = ref<boolean>(false);
  let peerConnection: RTCPeerConnection | null = null;
  let pendingOfferSdp: any = null;
  let pendingIceCandidates: RTCIceCandidateInit[] = [];

  const ICE_SERVERS: RTCIceServer[] = [
    { urls: 'stun:stun.qq.com:3478' },
    { urls: 'stun:stun.miwifi.com:3478' },
    { urls: 'stun:stun.cloudflare.com:3478' },
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
  ];

  const playNotificationSound = () => {
    try {
      const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
      if (!AudioCtx) return;
      const ctx = new AudioCtx();
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(587.33, ctx.currentTime);
      osc.frequency.exponentialRampToValueAtTime(880, ctx.currentTime + 0.12);
      gain.gain.setValueAtTime(0.08, ctx.currentTime);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.35);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start();
      osc.stop(ctx.currentTime + 0.35);
    } catch (_) {}
  };

  const sendDesktopNotification = (title: string, body: string, convId?: string) => {
    try {
      if (typeof Notification === 'undefined') return;
      if (Notification.permission === 'granted') {
        const n = new Notification(title, {
          body,
          icon: '/favicon.ico',
          tag: convId ? `mim_${convId}` : undefined,
        });
        n.onclick = () => {
          if (typeof window !== 'undefined') {
            window.focus();
          }
          if (convId) {
            selectConversation(convId);
          }
        };
      } else if (Notification.permission !== 'denied') {
        Notification.requestPermission();
      }
    } catch (_) {}
  };

  let socket: WebSocket | null = null;
  let heartbeatTimer: any = null;
  let reconnectTimer: any = null;

  // Active getters
  const activeConversation = computed(() => {
    return conversations.value.find((c) => c.id === activeConvId.value);
  });

  const activeMessages = computed(() => {
    return messagesByConv.value[activeConvId.value] || [];
  });

  // Load demo users from Go backend
  const loadDemoUsers = async () => {
    try {
      const res = await fetch('/api/auth/demo-users');
      const json = await res.json();
      if (json.code === 0 && Array.isArray(json.data)) {
        demoUsers.value = json.data;
        const storedUid = typeof localStorage !== 'undefined' ? Number(localStorage.getItem('mim_active_uid')) : 0;
        const storedToken = typeof localStorage !== 'undefined' ? localStorage.getItem('mim_token') || '' : '';

        if (!isExplicitlyLoggedOut && (storedUid || currentUserId.value)) {
          const uid = storedUid || currentUserId.value;
          const initial = demoUsers.value.find((u) => u.user.id === uid) || demoUsers.value[0];
          if (initial) {
            currentUserId.value = initial.user.id;
            currentUserName.value = initial.user.nickname;
            currentUserAvatar.value = initial.user.avatar || initial.user.nickname.substring(0, 2).toUpperCase();
            userToken.value = storedToken || initial.token.access_token;
            isLoggedIn.value = true;
            if (typeof localStorage !== 'undefined') {
              localStorage.setItem('mim_active_uid', String(initial.user.id));
              localStorage.setItem('mim_token', userToken.value);
            }
          }
        }
      }
    } catch (e) {
      console.warn('Could not fetch demo users from backend:', e);
    }
  };

  // Helper to purge all in-memory user data when switching or logging out
  const clearUserData = () => {
    if (activeCall.value) {
      endActiveCall();
    }
    cleanUpWebRTC();
    incomingCall.value = null;
    activeConvId.value = '';
    conversations.value = [];
    messagesByConv.value = {};
    userProfile.value = null;
    userConfig.value = null;
    aiConfig.value = {
      provider: 'deepseek',
      base_url: 'https://api.deepseek.com/v1',
      api_key: '',
      model: 'deepseek-chat',
      temperature: 0.7,
    };
    if (socket) {
      try {
        socket.onclose = null;
        socket.onerror = null;
        socket.onmessage = null;
        socket.close();
      } catch (_) {}
      socket = null;
    }
    isWsConnected.value = false;
    clearInterval(heartbeatTimer);
    clearTimeout(reconnectTimer);
  };

  // Switch Account
  const switchAccount = (u: DemoUser) => {
    clearUserData();
    currentUserId.value = u.user.id;
    currentUserName.value = u.user.nickname;
    currentUserAvatar.value = u.user.avatar || u.user.nickname.substring(0, 2).toUpperCase();
    userToken.value = u.token.access_token;
    isLoggedIn.value = true;
    showAuthModal.value = false;
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('mim_active_uid', String(u.user.id));
      localStorage.setItem('mim_token', u.token.access_token);
      localStorage.removeItem('mim_logged_out');
    }

    // Connect and fetch data for this account
    connectWs();
    fetchConversations();
    fetchUserProfile();
    fetchAIConfig();
  };

  // Login
  const login = async (account: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          account,
          password,
          platform: 'desktop',
          device_id: 'mim-client-' + (typeof navigator !== 'undefined' ? navigator.userAgent.slice(0, 15) : 'web'),
        }),
      });
      const json = await res.json();
      if (json.code === 0 && json.data) {
        clearUserData();
        const u = json.data.user;
        const t = json.data.token;
        currentUserId.value = u.id;
        currentUserName.value = u.nickname;
        currentUserAvatar.value = u.avatar || u.nickname.substring(0, 2).toUpperCase();
        userToken.value = t.access_token;
        isLoggedIn.value = true;
        showAuthModal.value = false;
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('mim_active_uid', String(u.id));
          localStorage.setItem('mim_token', t.access_token);
          localStorage.removeItem('mim_logged_out');
        }
        await fetchUserProfile();
        await fetchAIConfig();
        await fetchConversations();
        connectWs();
        return { success: true };
      } else {
        return { success: false, error: json.msg || 'Invalid account or password' };
      }
    } catch (err: any) {
      return { success: false, error: err.message || 'Network error' };
    }
  };

  // Register
  const register = async (account: string, password: string, nickname?: string, avatar?: string): Promise<{ success: boolean; error?: string }> => {
    try {
      const res = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          account,
          password,
          nickname: nickname || account,
          avatar: avatar || (nickname || account).substring(0, 2).toUpperCase(),
        }),
      });
      const json = await res.json();
      if (json.code === 0 && json.data) {
        clearUserData();
        const u = json.data.user;
        const t = json.data.token;
        currentUserId.value = u.id;
        currentUserName.value = u.nickname;
        currentUserAvatar.value = u.avatar || u.nickname.substring(0, 2).toUpperCase();
        userToken.value = t.access_token;
        isLoggedIn.value = true;
        showAuthModal.value = false;
        if (typeof localStorage !== 'undefined') {
          localStorage.setItem('mim_active_uid', String(u.id));
          localStorage.setItem('mim_token', t.access_token);
          localStorage.removeItem('mim_logged_out');
        }
        await fetchUserProfile();
        await fetchConversations();
        connectWs();
        return { success: true };
      } else {
        return { success: false, error: json.msg || 'Registration failed' };
      }
    } catch (err: any) {
      return { success: false, error: err.message || 'Network error' };
    }
  };

  // Logout
  const logout = () => {
    clearUserData();
    currentUserId.value = 0;
    currentUserName.value = '';
    currentUserAvatar.value = '';
    userToken.value = '';
    isLoggedIn.value = false;
    showAuthModal.value = false;
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem('mim_active_uid');
      localStorage.removeItem('mim_token');
      localStorage.setItem('mim_logged_out', 'true');
    }
    router.push('/login');
  };

  // Fetch conversations from Go backend
  const fetchConversations = async () => {
    try {
      const res = await fetch(`/api/chat/conversations?user_id=${currentUserId.value}`, {
        headers: {
          Authorization: `Bearer ${userToken.value}`,
        },
      });
      const json = await res.json();
      if (json.code === 0 && Array.isArray(json.data)) {
        conversations.value = json.data.map((c: any) => {
          let preview = c.last_msg_preview || 'No messages yet';
          if (typeof preview === 'string' && (preview.startsWith('/api/file/download') || preview.includes('file_id'))) {
            preview = '[文件]';
          }
          return {
            id: c.conversation_id,
            targetId: c.target_id,
            title: c.title || (c.conv_type === 1 ? 'Group Chat' : 'Direct Message'),
            type: c.conv_type === 1 ? 'group' : 'single',
            avatarText: c.avatar_text || 'MI',
            lastMessage: preview,
            lastTime: c.last_msg_time ? new Date(c.last_msg_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false }) : '',
            unreadCount: c.unread_count || 0,
            isPinned: c.is_pinned === 1,
            isMuted: c.is_muted === 1,
            isOnline: c.is_online,
          };
        });

        if (conversations.value.length > 0) {
          const currentExists = conversations.value.find((c) => c.id === activeConvId.value);
          if (!currentExists) {
            selectConversation(conversations.value[0].id);
          } else {
            fetchMessages(activeConvId.value);
          }
        } else {
          activeConvId.value = '';
        }
      }
    } catch (e) {
      console.warn('Fetch conversations error:', e);
    }
  };

  // Fetch messages from Go backend for conversation
  const fetchMessages = async (convId: string, lastSeqId?: number) => {
    try {
      let url = `/api/chat/messages?conv_id=${convId}&user_id=${currentUserId.value}`;
      if (lastSeqId && lastSeqId > 0) {
        url += `&last_seq_id=${lastSeqId}`;
      }
      const res = await fetch(url, {
        headers: {
          Authorization: `Bearer ${userToken.value}`,
        },
      });
      const json = await res.json();
      if (json.code === 0 && Array.isArray(json.data)) {
        const list: MessageItem[] = json.data.map((m: any) => {
          const isMe = m.sender_id === currentUserId.value;
          const timeStr = m.created_at
            ? new Date(m.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false })
            : '';
          
          let parsedType: MessageItem['msgType'] = 'text';
          const contentStr = typeof m.content === 'string' ? m.content : '';
          const isFileUrl = contentStr.startsWith('/api/file/download') || contentStr.includes('file_id') || m.msg_type === 5;

          if (m.is_revoked === 1 || m.msg_type === 7) {
            parsedType = 'revoke';
          } else if (m.msg_type === 2) {
            parsedType = 'image';
          } else if (m.msg_type === 3) {
            parsedType = 'voice';
          } else if (m.msg_type === 5 || isFileUrl) {
            parsedType = 'file';
          }

          let fName = m.file_name;
          let fSize = m.file_size;

          if (contentStr.startsWith('{')) {
            try {
              const meta = JSON.parse(contentStr);
              if (meta.name || meta.fileName) fName = meta.name || meta.fileName;
              if (meta.size || meta.fileSize) fSize = meta.size || meta.fileSize;
            } catch (_) {}
          }

          if (contentStr && contentStr.includes('?')) {
            try {
              const u = new URL(contentStr, 'http://localhost');
              const qName = u.searchParams.get('name');
              if (qName && !fName) {
                try {
                  fName = qName.includes('%') ? decodeURIComponent(qName) : qName;
                } catch (_) {
                  fName = qName;
                }
              }
              const qSize = u.searchParams.get('size');
              if (qSize && !fSize) {
                try {
                  fSize = qSize.includes('%') ? decodeURIComponent(qSize) : qSize;
                } catch (_) {
                  fSize = qSize;
                }
              }
            } catch (_) {}
          }

          if (!fName && isFileUrl) {
            fName = '文件附件';
          }

          return {
            id: String(m.id),
            clientMsgId: m.client_msg_id,
            seqId: m.seq_id,
            senderId: m.sender_id,
            senderName: isMe
              ? currentUserName.value
              : (m.sender_name || (activeConversation.value?.type === 'single' ? activeConversation.value?.title : '') || 'Partner'),
            content: (m.is_revoked === 1 || parsedType === 'revoke')
              ? (isMe ? '你撤回了一条消息' : `${m.sender_name || '对方'} 撤回了一条消息`)
              : m.content,
            msgType: parsedType,
            fileName: fName,
            fileSize: fSize,
            time: timeStr,
            isMe,
            status: isMe ? 'read' : 'delivered',
            isRevoked: m.is_revoked === 1 || parsedType === 'revoke',
            quote: m.quoted_msg_id > 0 ? {
              msgId: String(m.quoted_msg_id),
              userName: 'Quoted',
              content: m.quoted_content,
            } : undefined,
          };
        });

        if (lastSeqId && lastSeqId > 0) {
          if (!messagesByConv.value[convId]) {
            messagesByConv.value[convId] = [];
          }
          const existingIds = new Set(messagesByConv.value[convId].map((m) => m.id));
          const existingClientIds = new Set(messagesByConv.value[convId].map((m) => m.clientMsgId).filter(Boolean));
          for (const newMsg of list) {
            if (!existingIds.has(newMsg.id) && (!newMsg.clientMsgId || !existingClientIds.has(newMsg.clientMsgId))) {
              messagesByConv.value[convId].push(newMsg);
              existingIds.add(newMsg.id);
            }
          }
        } else {
          messagesByConv.value[convId] = list;
        }
      }
    } catch (e) {
      console.warn('Fetch messages error:', e);
    }
  };

  // Select a conversation
  const selectConversation = (convId: string) => {
    activeConvId.value = convId;
    const conv = conversations.value.find((c) => c.id === convId);
    if (conv) {
      conv.unreadCount = 0;
    }
    fetchMessages(convId);
  };

  // WebSocket Connection
  const connectWs = () => {
    if (!currentUserId.value || !userToken.value) {
      isWsConnected.value = false;
      return;
    }
    if (socket) {
      try {
        socket.onclose = null;
        socket.onerror = null;
        socket.onmessage = null;
        socket.close();
      } catch (_) {}
      socket = null;
    }
    clearInterval(heartbeatTimer);
    clearTimeout(reconnectTimer);

    let tabId = sessionStorage.getItem('mim_tab_id');
    if (!tabId) {
      tabId = 'tab_' + Math.random().toString(36).substring(2, 9);
      sessionStorage.setItem('mim_tab_id', tabId);
    }
    const wsUrl = getWsUrl(currentUserId.value, userToken.value, `web_${tabId}`);

    try {
      const s = new WebSocket(wsUrl);
      socket = s;
      s.binaryType = 'arraybuffer';

      s.onopen = () => {
        if (socket !== s) return;
        isWsConnected.value = true;
        console.log(`[Mim WS] Connected as User ${currentUserId.value} (${currentUserName.value})`);

        // Sync conversations & active conversation messages on connect/reconnect
        fetchConversations();
        if (activeConvId.value) {
          const existing = messagesByConv.value[activeConvId.value] || [];
          const maxSeq = existing.reduce((max, m) => Math.max(max, m.seqId || 0), 0);
          fetchMessages(activeConvId.value, maxSeq);
        }

        heartbeatTimer = setInterval(() => {
          if (socket === s && s.readyState === WebSocket.OPEN) {
            const hb = encodeFrame(CmdType.HeartbeatPing, 0n, 0, new Uint8Array(0));
            s.send(hb);
          }
        }, 25000);
      };

      s.onclose = () => {
        if (socket !== s) return;
        isWsConnected.value = false;
        clearInterval(heartbeatTimer);
        if (isLoggedIn.value && currentUserId.value && userToken.value) {
          reconnectTimer = setTimeout(connectWs, 2500);
        }
      };

      s.onerror = (err) => {
        if (socket !== s) return;
        console.warn('[Mim WS] Error:', err);
      };

      s.onmessage = (event: MessageEvent) => {
        if (socket !== s) return;
        if (event.data instanceof ArrayBuffer) {
          handleIncomingBinaryFrame(event.data);
        }
      };
    } catch (err) {
      console.error('[Mim WS] Failed to create WebSocket:', err);
      if (isLoggedIn.value && currentUserId.value && userToken.value) {
        reconnectTimer = setTimeout(connectWs, 3000);
      }
    }
  };

  // Handle incoming 24-byte protocol frame
  const handleIncomingBinaryFrame = async (buffer: ArrayBuffer) => {
    try {
      const frame = decodeFrame(buffer);
      const textDecoder = new TextDecoder();

      switch (frame.cmdType) {
        case CmdType.HeartbeatPong:
          break;

        case CmdType.MessageAck: {
          const payloadStr = textDecoder.decode(frame.payload);
          const ack = JSON.parse(payloadStr);
          const clientMsgId = ack.client_msg_id || ack.ClientMsgId;
          const msgId = ack.msg_id || ack.MsgId;
          const seqId = ack.seq_id || ack.SeqId;

          for (const convMsgs of Object.values(messagesByConv.value)) {
            const target = convMsgs.find((m) => m.clientMsgId === clientMsgId);
            if (target) {
              target.status = 'sent';
              if (msgId) target.id = String(msgId);
              if (seqId) target.seqId = Number(seqId);
              setTimeout(() => {
                target.status = 'delivered';
              }, 300);
              break;
            }
          }
          break;
        }

        case CmdType.ChatMessagePush: {
          const payloadStr = textDecoder.decode(frame.payload);
          const pushMsg = JSON.parse(payloadStr);

          const convId = pushMsg.conversation_id;
          const isMe = pushMsg.sender_id === currentUserId.value;

          const now = new Date(pushMsg.send_time || Date.now());
          const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

          let parsedType: MessageItem['msgType'] = 'text';
          const contentStr = typeof pushMsg.content === 'string' ? pushMsg.content : '';
          const isFileUrl = contentStr.startsWith('/api/file/download') || contentStr.includes('file_id') || pushMsg.msg_type === 5;

          if (pushMsg.msg_type === 7) parsedType = 'revoke';
          else if (pushMsg.msg_type === 2) parsedType = 'image';
          else if (pushMsg.msg_type === 3) parsedType = 'voice';
          else if (pushMsg.msg_type === 5 || isFileUrl) parsedType = 'file';

          let pushFName = pushMsg.file_name;
          let pushFSize = pushMsg.file_size;

          if (contentStr.startsWith('{')) {
            try {
              const meta = JSON.parse(contentStr);
              if (meta.name || meta.fileName) pushFName = meta.name || meta.fileName;
              if (meta.size || meta.fileSize) pushFSize = meta.size || meta.fileSize;
            } catch (_) {}
          }

          if (contentStr && contentStr.includes('?')) {
            try {
              const u = new URL(contentStr, 'http://localhost');
              const qName = u.searchParams.get('name');
              if (qName && !pushFName) {
                try {
                  pushFName = qName.includes('%') ? decodeURIComponent(qName) : qName;
                } catch (_) {
                  pushFName = qName;
                }
              }
              const qSize = u.searchParams.get('size');
              if (qSize && !pushFSize) {
                try {
                  pushFSize = qSize.includes('%') ? decodeURIComponent(qSize) : qSize;
                } catch (_) {
                  pushFSize = qSize;
                }
              }
            } catch (_) {}
          }

          if (!pushFName && isFileUrl) {
            pushFName = '文件附件';
          }

          const newMsg: MessageItem = {
            id: String(pushMsg.msg_id),
            clientMsgId: pushMsg.client_msg_id,
            seqId: Number(pushMsg.seq_id),
            senderId: pushMsg.sender_id,
            senderName: isMe
              ? currentUserName.value
              : (pushMsg.sender_name || (activeConversation.value?.type === 'single' ? activeConversation.value?.title : '') || 'Partner'),
            content: pushMsg.content,
            msgType: parsedType,
            fileName: isFileUrl ? pushFName : undefined,
            fileSize: isFileUrl ? pushFSize : undefined,
            time: timeStr,
            isMe,
            status: 'read',
            quote: pushMsg.quote ? {
              msgId: String(pushMsg.quote.quoted_msg_id),
              userName: pushMsg.quote.quoted_user_name || 'Quoted',
              content: pushMsg.quote.preview_content,
            } : undefined,
          };

          if (!messagesByConv.value[convId]) {
            messagesByConv.value[convId] = [];
          }

          const existing = messagesByConv.value[convId].find((m) => m.clientMsgId === pushMsg.client_msg_id || m.id === String(pushMsg.msg_id));
          if (!existing) {
            messagesByConv.value[convId].push(newMsg);
          } else {
            existing.id = String(pushMsg.msg_id);
            existing.status = 'sent';
            if (!existing.fileName && pushFName) existing.fileName = pushFName;
            if (!existing.fileSize && pushFSize) existing.fileSize = pushFSize;
          }

          const conv = conversations.value.find((c) => c.id === convId);
          if (conv) {
            let preview = pushMsg.content;
            if (parsedType === 'file') preview = '[文件]';
            else if (parsedType === 'image') preview = '[图片]';
            else if (parsedType === 'voice') preview = '[语音]';
            else if (parsedType === 'revoke') preview = '一条消息已被撤回';

            conv.lastMessage = preview;
            conv.lastTime = timeStr;
            if (activeConvId.value !== convId) {
              conv.unreadCount++;
            }
          }

          // Trigger Desktop Notification and Sound
          if (!isMe) {
            playNotificationSound();
            if (typeof document !== 'undefined' && (document.hidden || activeConvId.value !== convId)) {
              sendDesktopNotification(pushMsg.sender_name || 'Mim 新消息', pushMsg.content, convId);
            }
          }
          break;
        }

        case CmdType.RevokeNotify: {
          const payloadStr = textDecoder.decode(frame.payload);
          const revoke = JSON.parse(payloadStr);
          const convId = revoke.conversation_id || revoke.ConversationId;
          const msgId = String(revoke.msg_id || revoke.MsgId || '');
          const clientMsgId = revoke.client_msg_id || revoke.ClientMsgId || '';
          const seqId = Number(revoke.seq_id || revoke.SeqId || 0);

          const updateInList = (list: MessageItem[]) => {
            const target = list.find((m) =>
              (clientMsgId && m.clientMsgId === clientMsgId) ||
              (msgId && m.id === msgId) ||
              (seqId > 0 && m.seqId === seqId)
            );
            if (target) {
              target.isRevoked = true;
              target.msgType = 'revoke';
              target.content = target.isMe ? '你撤回了一条消息' : `${target.senderName || '对方'} 撤回了一条消息`;
              return true;
            }
            return false;
          };

          if (convId && messagesByConv.value[convId]) {
            updateInList(messagesByConv.value[convId]);
          } else {
            for (const list of Object.values(messagesByConv.value)) {
              if (updateInList(list)) break;
            }
          }

          const conv = conversations.value.find((c) => c.id === convId);
          if (conv) {
            conv.lastMessage = '一条消息已被撤回';
          }
          break;
        }

        case CmdType.WebRTCSignal: {
          const payloadStr = textDecoder.decode(frame.payload);
          const signal = JSON.parse(payloadStr);

          if (signal.signal_type === 'OFFER') {
            const caller = demoUsers.value.find((u) => u.user.id === signal.from_user_id)?.user.nickname || `User ${signal.from_user_id}`;
          let callType = signal.data?.includes('video') ? 'Video' : 'Audio';
          pendingOfferSdp = null;

          try {
            const parsed = JSON.parse(signal.data);
            if (parsed.sdp) {
              pendingOfferSdp = parsed.sdp;
              callType = parsed.callType === 'video' ? 'Video' : 'Audio';
            }
          } catch (_) {
            if (signal.data?.toLowerCase().includes('video')) callType = 'Video';
          }

          incomingCall.value = {
            callerId: signal.from_user_id,
            callerName: caller,
            callType,
            callId: signal.call_id,
          };
          playNotificationSound();
        } else if (signal.signal_type === 'ANSWER') {
          startCallTimer();
          if (peerConnection && signal.data) {
            try {
              const parsed = JSON.parse(signal.data);
              const sdp = parsed.sdp || parsed;
              await peerConnection.setRemoteDescription(new RTCSessionDescription(sdp));
              await flushPendingIceCandidates();
            } catch (err) {
              console.warn('[WebRTC] Remote description error on answer:', err);
            }
          }
        } else if (signal.signal_type === 'CANDIDATE') {
          if (signal.data) {
            try {
              const candidateInit = JSON.parse(signal.data);
              if (peerConnection && peerConnection.remoteDescription && peerConnection.remoteDescription.type) {
                await peerConnection.addIceCandidate(new RTCIceCandidate(candidateInit));
              } else {
                pendingIceCandidates.push(candidateInit);
              }
            } catch (err) {
              console.warn('[WebRTC] Ice candidate error:', err);
            }
          }
        } else if (signal.signal_type === 'CAMERA_STATE') {
          if (signal.data) {
            try {
              const parsed = JSON.parse(signal.data);
              isPeerCameraOff.value = !!parsed.isVideoMuted;
            } catch (_) {}
          }
        } else if (signal.signal_type === 'HANGUP' || signal.signal_type === 'REJECT') {
          cleanUpWebRTC();
          incomingCall.value = null;
          activeCall.value = null;
          stopCallTimer();
        }
        break;
        }

        default:
          console.log(`[WS] Frame received: 0x${frame.cmdType.toString(16)}`);
      }
    } catch (e) {
      console.error('[WS] Frame handling error:', e);
    }
  };

  // Send message over WebSocket using 24-byte binary Frame
  const sendMessage = (content: string, quote?: MessageItem['quote'], msgType: MessageItem['msgType'] = 'text', extra?: { fileName?: string; fileSize?: string; filePath?: string; duration?: number }) => {
    if (!content.trim() || !activeConvId.value) return;

    const list = messagesByConv.value[activeConvId.value] || [];
    const nextSeq = list.length > 0 ? list[list.length - 1].seqId + 1 : 1;
    const clientMsgId = 'c_' + Date.now() + '_' + Math.random().toString(36).substring(2, 8);

    const now = new Date();
    const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

    // Numeric message type enum mapping
    let numType = 0;
    if (msgType === 'image') numType = 2;
    else if (msgType === 'voice') numType = 3;
    else if (msgType === 'file') numType = 5;

    const newMsg: MessageItem = {
      id: 'local_' + Date.now(),
      clientMsgId,
      seqId: nextSeq,
      senderId: currentUserId.value,
      senderName: currentUserName.value,
      content,
      msgType,
      time: timeStr,
      isMe: true,
      status: 'sending',
      quote,
      fileName: extra?.fileName,
      fileSize: extra?.fileSize,
      filePath: extra?.filePath,
      duration: extra?.duration,
    };

    list.push(newMsg);
    messagesByConv.value[activeConvId.value] = list;

    const conv = conversations.value.find((c) => c.id === activeConvId.value);
    if (conv) {
      conv.lastMessage = msgType === 'image' ? '[图片]' : (msgType === 'file' ? '[文件]' : (msgType === 'voice' ? '[语音]' : content));
      conv.lastTime = timeStr;
    }

    if (socket && socket.readyState === WebSocket.OPEN) {
      const activeC = activeConversation.value;
      const payloadObj = {
        client_msg_id: clientMsgId,
        conversation_id: activeConvId.value,
        conv_type: activeC?.type === 'group' ? 1 : 0,
        sender_id: currentUserId.value,
        receiver_id: activeC?.targetId || 0,
        msg_type: numType,
        content,
        quote: quote ? {
          quoted_msg_id: Number(quote.msgId) || 0,
          preview_content: quote.content,
        } : undefined,
      };

      const payloadBytes = new TextEncoder().encode(JSON.stringify(payloadObj));
      const frameBuffer = encodeFrame(CmdType.ChatMessageSend, BigInt(nextSeq), FLAG_JSON, payloadBytes);
      socket.send(frameBuffer);
    } else {
      setTimeout(() => {
        newMsg.status = 'sent';
        setTimeout(() => {
          newMsg.status = 'delivered';
        }, 300);
      }, 150);
    }
  };

  // Revoke message
  const revokeMessage = async (msgId: string) => {
    const list = messagesByConv.value[activeConvId.value];
    if (!list) return;
    const target = list.find((m) => m.id === msgId);
    if (target && target.isMe) {
      try {
        const res = await fetch(`/api/chat/revoke?user_id=${currentUserId.value}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${userToken.value}`,
          },
          body: JSON.stringify({
            msg_id: Number(msgId) || 0,
            client_msg_id: target.clientMsgId || '',
          }),
        });
        const json = await res.json();
        if (res.ok && json.code === 0) {
          target.isRevoked = true;
          target.msgType = 'revoke';
          target.content = '你撤回了一条消息';
          const conv = conversations.value.find((c) => c.id === activeConvId.value);
          if (conv) {
            conv.lastMessage = '你撤回了一条消息';
          }
        } else {
          console.warn('[Revoke] Server rejected revoke:', json.msg);
          alertDialog({
            title: '撤回提示',
            message: json.msg || '撤回失败',
            type: 'warning',
          });
        }
      } catch (err) {
        console.error('[Revoke] Error:', err);
      }
    }
  };

  // WebRTC Signaling Functions
  const sendWebRTCSignal = (signalType: string, toUserId: number, callId: string, data: string = '') => {
    if (!socket || socket.readyState !== WebSocket.OPEN) return;

    const signalObj = {
      call_id: callId,
      from_user_id: currentUserId.value,
      to_user_id: toUserId,
      signal_type: signalType,
      data: data,
    };
    const payloadBytes = new TextEncoder().encode(JSON.stringify(signalObj));
    const frameBuffer = encodeFrame(CmdType.WebRTCSignal, 0n, FLAG_JSON, payloadBytes);
    socket.send(frameBuffer);
  };

  const flushPendingIceCandidates = async () => {
    if (!peerConnection || !peerConnection.remoteDescription) return;
    const candidates = [...pendingIceCandidates];
    pendingIceCandidates = [];
    for (const cand of candidates) {
      try {
        await peerConnection.addIceCandidate(new RTCIceCandidate(cand));
      } catch (err) {
        console.warn('[WebRTC] Error adding buffered ICE candidate:', err);
      }
    }
  };

  const createFallbackVideoStream = (name: string): MediaStream => {
    try {
      const canvas = document.createElement('canvas');
      canvas.width = 1280;
      canvas.height = 720;
      const ctx = canvas.getContext('2d');
      let frame = 0;
      const draw = () => {
        if (!ctx) return;
        frame++;
        const grad = ctx.createLinearGradient(0, 0, 1280, 720);
        grad.addColorStop(0, '#1e293b');
        grad.addColorStop(1, '#0f172a');
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, 1280, 720);

        // Avatar Circle
        ctx.beginPath();
        ctx.arc(640, 290, 100, 0, Math.PI * 2);
        ctx.fillStyle = '#3b82f6';
        ctx.fill();

        ctx.fillStyle = '#ffffff';
        ctx.font = 'bold 56px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText((name || 'U').substring(0, 2).toUpperCase(), 640, 290);

        ctx.font = 'bold 28px sans-serif';
        ctx.fillStyle = '#f8fafc';
        ctx.fillText(name || 'Virtual Camera', 640, 440);

        ctx.font = '18px monospace';
        ctx.fillStyle = '#10b981';
        const dots = '.'.repeat(frame % 4);
        ctx.fillText(`MIM HD 720P Camera${dots}`, 640, 485);
      };
      draw();
      const timer = setInterval(draw, 33); // 30 FPS for crisp fluid HD streaming
      const stream = (canvas as any).captureStream ? (canvas as any).captureStream(30) : new MediaStream();
      stream.getTracks().forEach((track: any) => {
        const origStop = track.stop.bind(track);
        track.stop = () => {
          clearInterval(timer);
          origStop();
        };
      });
      return stream;
    } catch (_) {
      return new MediaStream();
    }
  };

  const boostSenderQuality = async (pc: RTCPeerConnection) => {
    try {
      const senders = pc.getSenders();
      for (const sender of senders) {
        if (sender.track && sender.track.kind === 'video') {
          const params = sender.getParameters();
          if (!params.encodings || params.encodings.length === 0) {
            params.encodings = [{}];
          }
          for (const enc of params.encodings) {
            enc.maxBitrate = 4_000_000; // 4.0 Mbps for crisp 1080p/720p HD
            enc.maxFramerate = 30;
            enc.scaleResolutionDownBy = 1.0; // Strictly prevent browser downscaling
          }
          await sender.setParameters(params);
          console.log('[WebRTC] Desktop video sender boosted to 4Mbps / 30fps / scale 1.0');
        }
      }
    } catch (e) {
      console.warn('[WebRTC] boostSenderQuality failed:', e);
    }
  };

  const enhanceSdpBitrate = (sdp: string): string => {
    const lines = sdp.split('\r\n');
    const result: string[] = [];
    let inVideo = false;
    let addedBandwidth = false;

    for (const line of lines) {
      if (line.startsWith('m=video')) {
        inVideo = true;
      } else if (line.startsWith('m=audio') || line.startsWith('m=application')) {
        inVideo = false;
      }

      result.push(line);

      if (inVideo && !addedBandwidth && (line.startsWith('c=IN') || line.startsWith('a=rtcp-mux'))) {
        result.push('b=AS:4000');
        result.push('b=TIAS:4000000');
        addedBandwidth = true;
      }
    }
    return result.join('\r\n');
  };

  const acquireLocalMedia = async (isVideo: boolean): Promise<MediaStream> => {
    if (typeof navigator !== 'undefined' && navigator.mediaDevices) {
      // 1. Try full requested HD constraints (1080p / 720p)
      try {
        const stream = await navigator.mediaDevices.getUserMedia({
          audio: {
            echoCancellation: true,
            noiseSuppression: true,
            autoGainControl: true,
          },
          video: isVideo ? {
            width: { ideal: 1920, min: 1280 },
            height: { ideal: 1080, min: 720 },
            frameRate: { ideal: 30, min: 24 },
          } : false,
        });
        return stream;
      } catch (err: any) {
        console.warn('[WebRTC] Full 1080p/720p media capture failed, falling back to basic 720p:', err);
      }

      // 1b. Try more permissive 720p constraints
      if (isVideo) {
        try {
          const stream = await navigator.mediaDevices.getUserMedia({
            audio: true,
            video: { width: { ideal: 1280 }, height: { ideal: 720 }, frameRate: { ideal: 30 } },
          });
          return stream;
        } catch (_) {}
      }

      // 2. If video was requested but failed (e.g. webcam in use by another tab/app on Windows)
      if (isVideo) {
        try {
          const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
          const virtualVideoStream = createFallbackVideoStream(currentUserName.value || 'User');
          virtualVideoStream.getVideoTracks().forEach((vt) => audioStream.addTrack(vt));
          return audioStream;
        } catch (_) {}
      }

      // 3. Try audio only
      try {
        const audioStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
        if (isVideo) {
          const virtualVideoStream = createFallbackVideoStream(currentUserName.value || 'User');
          virtualVideoStream.getVideoTracks().forEach((vt) => audioStream.addTrack(vt));
        }
        return audioStream;
      } catch (_) {}
    }

    // 4. Pure fallback virtual stream
    if (isVideo) {
      return createFallbackVideoStream(currentUserName.value || 'User');
    }
    return new MediaStream();
  };

  const setupPeerConnection = (pc: RTCPeerConnection) => {
    pc.ontrack = (event) => {
      console.log('[WebRTC] ontrack received track:', event.track.kind, 'stream count:', event.streams.length);
      const incomingTrack = event.track;
      const currentTracks = remoteMediaStream.value ? remoteMediaStream.value.getTracks() : [];
      const updatedTracks = currentTracks.filter((t) => t.id !== incomingTrack.id);
      updatedTracks.push(incomingTrack);

      // Re-instantiate MediaStream with all tracks to guarantee Vue 3 reactivity update!
      remoteMediaStream.value = new MediaStream(updatedTracks);
    };

    pc.oniceconnectionstatechange = () => {
      console.log('[WebRTC] ICE Connection State:', pc.iceConnectionState);
      if (pc.iceConnectionState === 'connected' || pc.iceConnectionState === 'completed') {
        boostSenderQuality(pc);
      }
    };

    pc.onconnectionstatechange = () => {
      console.log('[WebRTC] Connection State:', pc.connectionState);
      if (pc.connectionState === 'connected') {
        boostSenderQuality(pc);
      }
    };
  };

  const initiateWebRTCCall = async (type: 'audio' | 'video') => {
    const activeC = activeConversation.value;
    if (!activeC || activeC.type === 'group') return;

    const callId = 'call_' + Date.now();
    activeCall.value = {
      peerId: activeC.targetId,
      peerName: activeC.title,
      callType: type,
      callId,
      convId: activeC.id,
    };
    isCallAudioMuted.value = false;
    isCallVideoStopped.value = false;
    isPeerCameraOff.value = false;
    pendingIceCandidates = [];

    try {
      const stream = await acquireLocalMedia(type === 'video');
      localMediaStream.value = stream;

      peerConnection = new RTCPeerConnection({ iceServers: ICE_SERVERS });
      setupPeerConnection(peerConnection);

      stream.getTracks().forEach((track) => {
        peerConnection?.addTrack(track, stream);
      });

      peerConnection.onicecandidate = (event) => {
        if (event.candidate && activeCall.value) {
          sendWebRTCSignal('CANDIDATE', activeCall.value.peerId, activeCall.value.callId, JSON.stringify(event.candidate));
        }
      };

      const offer = await peerConnection.createOffer({
        offerToReceiveAudio: true,
        offerToReceiveVideo: type === 'video',
      });
      const hdOfferSdp = enhanceSdpBitrate(offer.sdp || '');
      const hdOffer = { type: offer.type, sdp: hdOfferSdp };
      await peerConnection.setLocalDescription(hdOffer);
      boostSenderQuality(peerConnection);

      sendWebRTCSignal('OFFER', activeC.targetId, callId, JSON.stringify({ sdp: hdOffer, callType: type }));
      startCallTimer();
    } catch (err) {
      console.error('[WebRTC] initiateWebRTCCall failed:', err);
      sendWebRTCSignal('OFFER', activeC.targetId, callId, JSON.stringify({ callType: type }));
      startCallTimer();
    }
  };

  const acceptIncomingCall = async () => {
    if (!incomingCall.value) return;
    const inc = incomingCall.value;
    const matchedConv = conversations.value.find((c) => c.targetId === inc.callerId);
    const convId = matchedConv ? matchedConv.id : `${Math.min(currentUserId.value, inc.callerId)}_${Math.max(currentUserId.value, inc.callerId)}`;
    const isVideo = inc.callType.toLowerCase().includes('video');

    activeCall.value = {
      peerId: inc.callerId,
      peerName: inc.callerName,
      callType: isVideo ? 'video' : 'audio',
      callId: inc.callId,
      convId,
    };
    isCallAudioMuted.value = false;
    isCallVideoStopped.value = false;
    isPeerCameraOff.value = false;
    incomingCall.value = null;
    startCallTimer();

    if (activeConvId.value !== convId) {
      selectConversation(convId);
    }

    try {
      const stream = await acquireLocalMedia(isVideo);
      localMediaStream.value = stream;

      peerConnection = new RTCPeerConnection({ iceServers: ICE_SERVERS });
      setupPeerConnection(peerConnection);

      stream.getTracks().forEach((track) => {
        peerConnection?.addTrack(track, stream);
      });

      peerConnection.onicecandidate = (event) => {
        if (event.candidate && activeCall.value) {
          sendWebRTCSignal('CANDIDATE', inc.callerId, inc.callId, JSON.stringify(event.candidate));
        }
      };

      if (pendingOfferSdp) {
        await peerConnection.setRemoteDescription(new RTCSessionDescription(pendingOfferSdp));
        await flushPendingIceCandidates();

        const answer = await peerConnection.createAnswer();
        const hdAnswerSdp = enhanceSdpBitrate(answer.sdp || '');
        const hdAnswer = { type: answer.type, sdp: hdAnswerSdp };
        await peerConnection.setLocalDescription(hdAnswer);
        boostSenderQuality(peerConnection);
        sendWebRTCSignal('ANSWER', inc.callerId, inc.callId, JSON.stringify({ sdp: hdAnswer }));
        return;
      }
    } catch (err) {
      console.error('[WebRTC] acceptIncomingCall error:', err);
    }

    sendWebRTCSignal('ANSWER', inc.callerId, inc.callId, inc.callType);
  };

  const rejectIncomingCall = () => {
    if (!incomingCall.value) return;
    sendWebRTCSignal('REJECT', incomingCall.value.callerId, incomingCall.value.callId);
    cleanUpWebRTC();
    incomingCall.value = null;
  };

  const endActiveCall = () => {
    if (!activeCall.value) return;
    const data = JSON.stringify({
      callType: activeCall.value.callType,
      duration: callDuration.value,
    });
    sendWebRTCSignal('HANGUP', activeCall.value.peerId, activeCall.value.callId, data);
    cleanUpWebRTC();
    activeCall.value = null;
    stopCallTimer();
  };

  const cleanUpWebRTC = () => {
    if (localMediaStream.value) {
      localMediaStream.value.getTracks().forEach((t) => t.stop());
      localMediaStream.value = null;
    }
    if (remoteMediaStream.value) {
      remoteMediaStream.value.getTracks().forEach((t) => t.stop());
      remoteMediaStream.value = null;
    }
    if (peerConnection) {
      try {
        peerConnection.close();
      } catch (_) {}
      peerConnection = null;
    }
    pendingOfferSdp = null;
    pendingIceCandidates = [];
    isCallAudioMuted.value = false;
    isCallVideoStopped.value = false;
    isPeerCameraOff.value = false;
  };

  const toggleCallAudio = () => {
    if (localMediaStream.value) {
      localMediaStream.value.getAudioTracks().forEach((t) => {
        t.enabled = !t.enabled;
        isCallAudioMuted.value = !t.enabled;
      });
    } else {
      isCallAudioMuted.value = !isCallAudioMuted.value;
    }
  };

  const toggleCallVideo = async () => {
    if (!localMediaStream.value) {
      isCallVideoStopped.value = !isCallVideoStopped.value;
      if (activeCall.value) {
        sendWebRTCSignal('CAMERA_STATE', activeCall.value.peerId, activeCall.value.callId, JSON.stringify({ isVideoMuted: isCallVideoStopped.value }));
      }
      return;
    }

    const videoTracks = localMediaStream.value.getVideoTracks();
    if (videoTracks.length === 0) {
      try {
        if (typeof navigator !== 'undefined' && navigator.mediaDevices) {
          const videoStream = await navigator.mediaDevices.getUserMedia({
            video: { width: { ideal: 1920, min: 1280 }, height: { ideal: 1080, min: 720 }, frameRate: { ideal: 30 } }
          });
          const newTrack = videoStream.getVideoTracks()[0];
          if (newTrack) {
            localMediaStream.value.addTrack(newTrack);
            if (peerConnection) {
              peerConnection.addTrack(newTrack, localMediaStream.value);
              boostSenderQuality(peerConnection);
            }
            if (activeCall.value) {
              activeCall.value.callType = 'video';
            }
            isCallVideoStopped.value = false;
            if (activeCall.value) {
              sendWebRTCSignal('CAMERA_STATE', activeCall.value.peerId, activeCall.value.callId, JSON.stringify({ isVideoMuted: false }));
            }
          }
        }
      } catch (err) {
        console.warn('Failed to acquire camera track:', err);
      }
    } else {
      videoTracks.forEach((t) => {
        t.enabled = !t.enabled;
        isCallVideoStopped.value = !t.enabled;
      });
      if (activeCall.value) {
        if (!isCallVideoStopped.value) {
          activeCall.value.callType = 'video';
        }
        sendWebRTCSignal('CAMERA_STATE', activeCall.value.peerId, activeCall.value.callId, JSON.stringify({ isVideoMuted: isCallVideoStopped.value }));
      }
    }
  };

  const startCallTimer = () => {
    stopCallTimer();
    callDuration.value = 0;
    callTimerInterval = setInterval(() => {
      callDuration.value++;
    }, 1000);
  };

  const stopCallTimer = () => {
    clearInterval(callTimerInterval);
    callTimerInterval = null;
  };

  const togglePinConversation = async (convId: string) => {
    const c = conversations.value.find((item) => item.id === convId);
    if (!c) return;
    c.isPinned = !c.isPinned;
    try {
      await fetch(`/api/chat/pin-conversation?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
          is_pinned: c.isPinned ? 1 : 0,
        }),
      });
    } catch (_) {}
  };

  const toggleMuteConversation = async (convId: string) => {
    const c = conversations.value.find((item) => item.id === convId);
    if (!c) return;
    c.isMuted = !c.isMuted;
    try {
      await fetch(`/api/chat/mute-conversation?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
          is_muted: c.isMuted ? 1 : 0,
        }),
      });
    } catch (_) {}
  };

  const markConversationAsRead = async (convId: string) => {
    const c = conversations.value.find((item) => item.id === convId);
    if (!c) return;
    c.unreadCount = 0;
    try {
      await fetch(`/api/chat/read-conversation?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
          unread_count: 0,
        }),
      });
    } catch (_) {}
  };

  const markConversationAsUnread = async (convId: string) => {
    const c = conversations.value.find((item) => item.id === convId);
    if (!c) return;
    c.unreadCount = (c.unreadCount || 0) + 1;
    try {
      await fetch(`/api/chat/read-conversation?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
          unread_count: c.unreadCount,
        }),
      });
    } catch (_) {}
  };

  const deleteConversation = async (convId: string) => {
    conversations.value = conversations.value.filter((item) => item.id !== convId);
    delete messagesByConv.value[convId];
    if (activeConvId.value === convId && conversations.value.length > 0) {
      selectConversation(conversations.value[0].id);
    }
    try {
      await fetch(`/api/chat/delete-conversation?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
        }),
      });
    } catch (_) {}
  };

  const clearConversationMessages = async (convId: string) => {
    messagesByConv.value[convId] = [];
    const c = conversations.value.find((item) => item.id === convId);
    if (c) {
      c.lastMessage = '';
    }
    try {
      await fetch(`/api/chat/clear-messages?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
        }),
      });
    } catch (_) {}
  };

  const deleteMessage = async (convId: string, msgId: string) => {
    const list = messagesByConv.value[convId];
    const target = list?.find((m) => m.id === msgId);
    if (list) {
      messagesByConv.value[convId] = list.filter((m) => m.id !== msgId);
    }
    try {
      await fetch(`/api/chat/delete-message?user_id=${currentUserId.value}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${userToken.value}`,
        },
        body: JSON.stringify({
          conversation_id: convId,
          msg_id: Number(msgId) || 0,
          client_msg_id: target?.clientMsgId || '',
        }),
      });
    } catch (_) {}
  };

  // Auto initialize on load
  loadDemoUsers().then(() => {
    if (isLoggedIn.value) {
      connectWs();
      fetchConversations();
      fetchUserProfile();
      fetchAIConfig();
    }
  });

  return {
    isLoggedIn,
    showAuthModal,
    sidebarWidth,
    setSidebarWidth,
    currentUserId,
    currentUserName,
    currentUserAvatar,
    userToken,
    isWsConnected,
    demoUsers,
    activeTab,
    activeConvId,
    conversations,
    activeConversation,
    activeMessages,
    incomingCall,
    activeCall,
    callDuration,
    userProfile,
    userConfig,
    aiConfig,
    fetchAIConfig,
    saveAIConfig,
    testAIConnection,
    login,
    register,
    logout,
    clearUserData,
    fetchUserProfile,
    switchAccount,
    selectConversation,
    sendMessage,
    revokeMessage,
    fetchConversations,
    initiateWebRTCCall,
    acceptIncomingCall,
    rejectIncomingCall,
    endActiveCall,
    localMediaStream,
    remoteMediaStream,
    isCallAudioMuted,
    isCallVideoStopped,
    isPeerCameraOff,
    toggleCallAudio,
    toggleCallVideo,
    togglePinConversation,
    toggleMuteConversation,
    markConversationAsRead,
    markConversationAsUnread,
    deleteConversation,
    clearConversationMessages,
    deleteMessage,
    playNotificationSound,
    sendDesktopNotification,
  };
});
