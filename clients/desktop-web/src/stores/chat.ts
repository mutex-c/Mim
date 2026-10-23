import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { CmdType, encodeFrame, decodeFrame } from '../protocol/frame';

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
  duration?: number; // for voice
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

export const useChatStore = defineStore('chat', () => {
  const currentUserId = ref<number>(1);
  const currentUserName = ref<string>('Alex Chen');
  const activeConvId = ref<string>('conv_single_2');

  const conversations = ref<ConversationItem[]>([
    {
      id: 'conv_single_2',
      targetId: 2,
      title: 'Sarah Jenkins',
      type: 'single',
      avatarText: 'SJ',
      lastMessage: 'The Q3 architecture review is ready for sign-off.',
      lastTime: '12:45',
      unreadCount: 0,
      isPinned: true,
      isMuted: false,
      isOnline: true,
    },
    {
      id: 'conv_group_101',
      targetId: 101,
      title: 'Core Infrastructure Team',
      type: 'group',
      avatarText: 'CI',
      lastMessage: 'David: Deployed v1.4.2 to canary cluster.',
      lastTime: '11:20',
      unreadCount: 3,
      isPinned: true,
      isMuted: false,
    },
    {
      id: 'conv_single_3',
      targetId: 3,
      title: 'Michael Zhang',
      type: 'single',
      avatarText: 'MZ',
      lastMessage: 'Voice message (12s)',
      lastTime: 'Yesterday',
      unreadCount: 0,
      isPinned: false,
      isMuted: false,
      isOnline: false,
    },
    {
      id: 'conv_single_4',
      targetId: 4,
      title: 'AI Summary Assistant',
      type: 'single',
      avatarText: 'AI',
      lastMessage: '3 unread discussions summarized into key action items.',
      lastTime: 'Yesterday',
      unreadCount: 0,
      isPinned: false,
      isMuted: false,
      isOnline: true,
    },
  ]);

  const messagesByConv = ref<Record<string, MessageItem[]>>({
    'conv_single_2': [
      {
        id: 'msg_1',
        clientMsgId: 'c_uuid_1',
        seqId: 1,
        senderId: 2,
        senderName: 'Sarah Jenkins',
        content: 'Hi Alex, did you review the distributed SeqId monotonic generator proposal?',
        msgType: 'text',
        time: '12:30',
        isMe: false,
        status: 'read',
      },
      {
        id: 'msg_2',
        clientMsgId: 'c_uuid_2',
        seqId: 2,
        senderId: 1,
        senderName: 'Alex Chen',
        content: 'Yes, I checked the Redis atomic INCR counter coupled with the client ACK protocol. It guarantees zero message loss even during network partition recovery.',
        msgType: 'text',
        time: '12:32',
        isMe: true,
        status: 'read',
        quote: {
          msgId: 'msg_1',
          userName: 'Sarah Jenkins',
          content: 'Hi Alex, did you review the distributed SeqId monotonic generator proposal?',
        },
      },
      {
        id: 'msg_3',
        clientMsgId: 'c_uuid_3',
        seqId: 3,
        senderId: 2,
        senderName: 'Sarah Jenkins',
        content: 'Audio Note',
        msgType: 'voice',
        duration: 8,
        time: '12:40',
        isMe: false,
        status: 'read',
      },
      {
        id: 'msg_4',
        clientMsgId: 'c_uuid_4',
        seqId: 4,
        senderId: 2,
        senderName: 'Sarah Jenkins',
        content: 'The Q3 architecture review is ready for sign-off.',
        msgType: 'text',
        time: '12:45',
        isMe: false,
        status: 'read',
      },
    ],
    'conv_group_101': [
      {
        id: 'msg_101',
        clientMsgId: 'c_uuid_101',
        seqId: 1,
        senderId: 5,
        senderName: 'David Miller',
        content: 'Deployed v1.4.2 to canary cluster. Observability metrics are nominal.',
        msgType: 'text',
        time: '11:20',
        isMe: false,
        status: 'sent',
      },
    ],
  });

  const activeConversation = computed(() => {
    return conversations.value.find((c) => c.id === activeConvId.value);
  });

  const activeMessages = computed(() => {
    return messagesByConv.value[activeConvId.value] || [];
  });

  const selectConversation = (convId: string) => {
    activeConvId.value = convId;
    const conv = conversations.value.find((c) => c.id === convId);
    if (conv) {
      conv.unreadCount = 0;
    }
  };

  const sendMessage = (content: string, quote?: MessageItem['quote']) => {
    if (!content.trim() || !activeConvId.value) return;

    const list = messagesByConv.value[activeConvId.value] || [];
    const nextSeq = list.length > 0 ? list[list.length - 1].seqId + 1 : 1;
    const clientMsgId = 'client_' + Date.now() + '_' + Math.random().toString(36).substring(2, 8);

    const now = new Date();
    const timeStr = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });

    const newMsg: MessageItem = {
      id: 'msg_' + Date.now(),
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

    // Update conversation list preview
    const conv = conversations.value.find((c) => c.id === activeConvId.value);
    if (conv) {
      conv.lastMessage = content;
      conv.lastTime = timeStr;
    }

    // Simulate Server ACK
    setTimeout(() => {
      newMsg.status = 'sent';
      setTimeout(() => {
        newMsg.status = 'delivered';
      }, 300);
    }, 150);
  };

  const revokeMessage = (msgId: string) => {
    const list = messagesByConv.value[activeConvId.value];
    if (!list) return;
    const target = list.find((m) => m.id === msgId);
    if (target && target.isMe) {
      target.isRevoked = true;
      target.msgType = 'revoke';
      target.content = 'You recalled a message';

      const conv = conversations.value.find((c) => c.id === activeConvId.value);
      if (conv) {
        conv.lastMessage = 'You recalled a message';
      }
    }
  };

  return {
    currentUserId,
    currentUserName,
    activeConvId,
    conversations,
    activeConversation,
    activeMessages,
    selectConversation,
    sendMessage,
    revokeMessage,
  };
});
