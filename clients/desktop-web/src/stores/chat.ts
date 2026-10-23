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

  const activeConvId = ref<string>('1_2');
  const conversations = ref<ConversationItem[]>([]);
  const messagesByConv = ref<Record<string, MessageItem[]>>({});

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
        // If current user is not logged in, select user 1
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
          return {
            id: String(m.id),
            clientMsgId: m.client_msg_id,
            seqId: m.seq_id,
            senderId: m.sender_id,
            senderName: isMe ? currentUserName.value : (m.sender_name || 'Partner'),
            content: m.content,
            msgType: m.msg_type === 7 ? 'revoke' : (m.msg_type === 3 ? 'voice' : 'text'),
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

        // Start Heartbeat every 25 seconds
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
        // Auto reconnect after 2.5s
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
          // Heartbeat acknowledged
          break;

        case CmdType.MessageAck: {
          const payloadStr = textDecoder.decode(frame.payload);
          const ack = JSON.parse(payloadStr);
          // Update status of matching message in active conversation
          const list = messagesByConv.value[activeConvId.value] || [];
          const target = list.find((m) => m.clientMsgId === ack.client_msg_id);
          if (target) {
            target.status = 'sent';
            target.id = String(ack.msg_id);
            target.seqId = Number(ack.seq_id);

            // Simulate delivered tick
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

          const newMsg: MessageItem = {
            id: String(pushMsg.msg_id),
            clientMsgId: pushMsg.client_msg_id,
            seqId: Number(pushMsg.seq_id),
            senderId: pushMsg.sender_id,
            senderName: isMe ? currentUserName.value : (pushMsg.sender_name || 'Sarah Jenkins'),
            content: pushMsg.content,
            msgType: pushMsg.msg_type === 7 ? 'revoke' : (pushMsg.msg_type === 3 ? 'voice' : 'text'),
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

          // Check duplicate
          const existing = messagesByConv.value[convId].find((m) => m.clientMsgId === pushMsg.client_msg_id || m.id === String(pushMsg.msg_id));
          if (!existing) {
            messagesByConv.value[convId].push(newMsg);
          }

          // Update conversation list preview
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

        default:
          console.log(`[WS] Frame received: 0x${frame.cmdType.toString(16)}`);
      }
    } catch (e) {
      console.error('[WS] Frame handling error:', e);
    }
  };

  // Send message over WebSocket using 24-byte binary Frame
  const sendMessage = (content: string, quote?: MessageItem['quote']) => {
    if (!content.trim() || !activeConvId.value) return;

    const list = messagesByConv.value[activeConvId.value] || [];
    const nextSeq = list.length > 0 ? list[list.length - 1].seqId + 1 : 1;
    const clientMsgId = 'c_' + Date.now() + '_' + Math.random().toString(36).substring(2, 8);

    const now = new Date();
    const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

    // Local optimistic message
    const newMsg: MessageItem = {
      id: 'local_' + Date.now(),
      clientMsgId,
      seqId: nextSeq,
      senderId: currentUserId.value,
      senderName: currentUserName.value,
      content,
      msgType: 'text',
      time: timeStr,
      isMe: true,
      status: 'sending',
      quote,
    };

    list.push(newMsg);
    messagesByConv.value[activeConvId.value] = list;

    // Update conversation item preview
    const conv = conversations.value.find((c) => c.id === activeConvId.value);
    if (conv) {
      conv.lastMessage = content;
      conv.lastTime = timeStr;
    }

    // Send binary frame over WebSocket if connected
    if (socket && socket.readyState === WebSocket.OPEN) {
      const activeC = activeConversation.value;
      const payloadObj = {
        client_msg_id: clientMsgId,
        conversation_id: activeConvId.value,
        conv_type: activeC?.type === 'group' ? 1 : 0,
        sender_id: currentUserId.value,
        receiver_id: activeC?.targetId || 0,
        msg_type: 0,
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
      // Mock fallback if disconnected
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

      // Call API
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
    activeConvId,
    conversations,
    activeConversation,
    activeMessages,
    switchAccount,
    selectConversation,
    sendMessage,
    revokeMessage,
    fetchConversations,
  };
});
