import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { CmdType, encodeFrame, decodeFrame, FLAG_JSON } from '../protocol/frame';

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
  isRevoked?: boolean;
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
  const currentUserId = ref<number>(1);
  const currentUserName = ref<string>('Alex Chen');
  const currentUserAvatar = ref<string>('AC');
  const userToken = ref<string>('');
  const isWsConnected = ref<boolean>(false);
  const demoUsers = ref<DemoUser[]>([]);

  const activeTab = ref<string>('chats'); // 'chats', 'contacts', 'calls'
  const activeConvId = ref<string>('1_2');
  const conversations = ref<ConversationItem[]>([]);
  const messagesByConv = ref<Record<string, MessageItem[]>>({});

  // WebRTC Call State
  const incomingCall = ref<{ callerId: number; callerName: string; callType: string; callId: string } | null>(null);
  const activeCall = ref<{ peerId: number; peerName: string; callType: string; callId: string } | null>(null);
  const callDuration = ref<number>(0);
  let callTimerInterval: any = null;

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
        const initial = demoUsers.value.find((u) => u.user.id === currentUserId.value) || demoUsers.value[0];
        if (initial) {
          currentUserId.value = initial.user.id;
          currentUserName.value = initial.user.nickname;
          currentUserAvatar.value = initial.user.avatar || initial.user.nickname.substring(0, 2).toUpperCase();
          userToken.value = initial.token.access_token;
        }
      }
    } catch (e) {
      console.warn('Could not fetch demo users from backend:', e);
    }
  };

  // Switch Account
  const switchAccount = (u: DemoUser) => {
    currentUserId.value = u.user.id;
    currentUserName.value = u.user.nickname;
    currentUserAvatar.value = u.user.avatar || u.user.nickname.substring(0, 2).toUpperCase();
    userToken.value = u.token.access_token;

    // Disconnect and reconnect WebSocket
    connectWs();
    fetchConversations();
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
        conversations.value = json.data.map((c: any) => ({
          id: c.conversation_id,
          targetId: c.target_id,
          title: c.title || (c.conv_type === 1 ? 'Group Chat' : 'Direct Message'),
          type: c.conv_type === 1 ? 'group' : 'single',
          avatarText: c.avatar_text || 'MI',
          lastMessage: c.last_msg_preview || 'No messages yet',
          lastTime: c.last_msg_time ? new Date(c.last_msg_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false }) : '',
          unreadCount: c.unread_count || 0,
          isPinned: c.is_pinned === 1,
          isMuted: false,
          isOnline: c.is_online,
        }));

        if (conversations.value.length > 0) {
          const currentExists = conversations.value.find((c) => c.id === activeConvId.value);
          if (!currentExists) {
            selectConversation(conversations.value[0].id);
          } else {
            fetchMessages(activeConvId.value);
          }
        }
      }
    } catch (e) {
      console.warn('Fetch conversations error:', e);
    }
  };

  // Fetch messages from Go backend for conversation
  const fetchMessages = async (convId: string) => {
    try {
      const res = await fetch(`/api/chat/messages?conv_id=${convId}&user_id=${currentUserId.value}`, {
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
          if (m.msg_type === 7) parsedType = 'revoke';
          else if (m.msg_type === 2) parsedType = 'image';
          else if (m.msg_type === 3) parsedType = 'voice';
          else if (m.msg_type === 5) parsedType = 'file';

          return {
            id: String(m.id),
            clientMsgId: m.client_msg_id,
            seqId: m.seq_id,
            senderId: m.sender_id,
            senderName: isMe ? currentUserName.value : (m.sender_name || 'Partner'),
            content: m.content,
            msgType: parsedType,
            time: timeStr,
            isMe,
            status: isMe ? 'read' : 'delivered',
            isRevoked: m.is_revoked === 1,
            quote: m.quoted_msg_id > 0 ? {
              msgId: String(m.quoted_msg_id),
              userName: 'Quoted',
              content: m.quoted_content,
            } : undefined,
          };
        });
        messagesByConv.value[convId] = list;
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
    if (socket) {
      try {
        socket.close();
      } catch (_) {}
      socket = null;
    }
    clearInterval(heartbeatTimer);
    clearTimeout(reconnectTimer);

    const wsProto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProto}//${location.host}/ws?user_id=${currentUserId.value}&token=${userToken.value}&platform=web`;

    try {
      socket = new WebSocket(wsUrl);
      socket.binaryType = 'arraybuffer';

      socket.onopen = () => {
        isWsConnected.value = true;
        console.log(`[Mim WS] Connected as User ${currentUserId.value} (${currentUserName.value})`);

        heartbeatTimer = setInterval(() => {
          if (socket && socket.readyState === WebSocket.OPEN) {
            const hb = encodeFrame(CmdType.HeartbeatPing, 0n, 0, new Uint8Array(0));
            socket.send(hb);
          }
        }, 25000);
      };

      socket.onclose = () => {
        isWsConnected.value = false;
        clearInterval(heartbeatTimer);
        reconnectTimer = setTimeout(connectWs, 2500);
      };

      socket.onerror = (err) => {
        console.warn('[Mim WS] Error:', err);
      };

      socket.onmessage = (event: MessageEvent) => {
        if (event.data instanceof ArrayBuffer) {
          handleIncomingBinaryFrame(event.data);
        }
      };
    } catch (err) {
      console.error('[Mim WS] Failed to create WebSocket:', err);
      reconnectTimer = setTimeout(connectWs, 3000);
    }
  };

  // Handle incoming 24-byte protocol frame
  const handleIncomingBinaryFrame = (buffer: ArrayBuffer) => {
    try {
      const frame = decodeFrame(buffer);
      const textDecoder = new TextDecoder();

      switch (frame.cmdType) {
        case CmdType.HeartbeatPong:
          break;

        case CmdType.MessageAck: {
          const payloadStr = textDecoder.decode(frame.payload);
          const ack = JSON.parse(payloadStr);
          const list = messagesByConv.value[activeConvId.value] || [];
          const target = list.find((m) => m.clientMsgId === ack.client_msg_id);
          if (target) {
            target.status = 'sent';
            target.id = String(ack.msg_id);
            target.seqId = Number(ack.seq_id);
            setTimeout(() => {
              target.status = 'delivered';
            }, 300);
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
          if (pushMsg.msg_type === 7) parsedType = 'revoke';
          else if (pushMsg.msg_type === 2) parsedType = 'image';
          else if (pushMsg.msg_type === 3) parsedType = 'voice';
          else if (pushMsg.msg_type === 5) parsedType = 'file';

          const newMsg: MessageItem = {
            id: String(pushMsg.msg_id),
            clientMsgId: pushMsg.client_msg_id,
            seqId: Number(pushMsg.seq_id),
            senderId: pushMsg.sender_id,
            senderName: isMe ? currentUserName.value : (pushMsg.sender_name || 'Sarah Jenkins'),
            content: pushMsg.content,
            msgType: parsedType,
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
          }

          const conv = conversations.value.find((c) => c.id === convId);
          if (conv) {
            conv.lastMessage = pushMsg.content;
            conv.lastTime = timeStr;
            if (activeConvId.value !== convId) {
              conv.unreadCount++;
            }
          }
          break;
        }

        case CmdType.RevokeNotify: {
          const payloadStr = textDecoder.decode(frame.payload);
          const revoke = JSON.parse(payloadStr);
          const list = messagesByConv.value[revoke.conversation_id];
          if (list) {
            const target = list.find((m) => m.id === String(revoke.msg_id) || m.seqId === Number(revoke.seq_id));
            if (target) {
              target.isRevoked = true;
              target.content = '消息已撤回';
            }
          }
          break;
        }

        case CmdType.WebRTCSignal: {
          const payloadStr = textDecoder.decode(frame.payload);
          const signal = JSON.parse(payloadStr);

          if (signal.signal_type === 'OFFER') {
            // Incoming Call
            const caller = demoUsers.value.find((u) => u.user.id === signal.from_user_id)?.user.nickname || `User ${signal.from_user_id}`;
            incomingCall.value = {
              callerId: signal.from_user_id,
              callerName: caller,
              callType: signal.data?.includes('video') ? 'Video' : 'Audio',
              callId: signal.call_id,
            };
          } else if (signal.signal_type === 'ANSWER') {
            // Call accepted by peer
            startCallTimer();
          } else if (signal.signal_type === 'HANGUP' || signal.signal_type === 'REJECT') {
            // Call ended or declined
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
  const sendMessage = (content: string, quote?: MessageItem['quote'], msgType: MessageItem['msgType'] = 'text', extra?: { fileName?: string; fileSize?: string; duration?: number }) => {
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
      duration: extra?.duration,
    };

    list.push(newMsg);
    messagesByConv.value[activeConvId.value] = list;

    const conv = conversations.value.find((c) => c.id === activeConvId.value);
    if (conv) {
      conv.lastMessage = msgType === 'image' ? '[Image]' : (msgType === 'file' ? `[File] ${extra?.fileName || ''}` : content);
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
      target.isRevoked = true;
      target.content = '你撤回了一条消息';

      try {
        await fetch('/api/chat/revoke', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${userToken.value}`,
          },
          body: JSON.stringify({ msg_id: Number(msgId) || 0 }),
        });
      } catch (_) {}
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

  const initiateWebRTCCall = (type: 'audio' | 'video') => {
    const activeC = activeConversation.value;
    if (!activeC || activeC.type === 'group') return;

    const callId = 'call_' + Date.now();
    activeCall.value = {
      peerId: activeC.targetId,
      peerName: activeC.title,
      callType: type,
      callId,
    };

    sendWebRTCSignal('OFFER', activeC.targetId, callId, type);
    startCallTimer();
  };

  const acceptIncomingCall = () => {
    if (!incomingCall.value) return;
    const inc = incomingCall.value;
    activeCall.value = {
      peerId: inc.callerId,
      peerName: inc.callerName,
      callType: inc.callType,
      callId: inc.callId,
    };
    sendWebRTCSignal('ANSWER', inc.callerId, inc.callId, inc.callType);
    incomingCall.value = null;
    startCallTimer();
  };

  const rejectIncomingCall = () => {
    if (!incomingCall.value) return;
    sendWebRTCSignal('REJECT', incomingCall.value.callerId, incomingCall.value.callId);
    incomingCall.value = null;
  };

  const endActiveCall = () => {
    if (!activeCall.value) return;
    sendWebRTCSignal('HANGUP', activeCall.value.peerId, activeCall.value.callId);
    activeCall.value = null;
    stopCallTimer();
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

  // Auto initialize on load
  loadDemoUsers().then(() => {
    connectWs();
    fetchConversations();
  });

  return {
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
    switchAccount,
    selectConversation,
    sendMessage,
    revokeMessage,
    fetchConversations,
    initiateWebRTCCall,
    acceptIncomingCall,
    rejectIncomingCall,
    endActiveCall,
  };
});
