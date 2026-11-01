package models

import (
	"time"
)

// Message Type Constants (aligned with protobuf MessageType)
const (
	MsgTypeText               = 0 // 文本
	MsgTypeMarkdown           = 1 // Markdown
	MsgTypeImage              = 2 // 图片
	MsgTypeVoice              = 3 // 语音
	MsgTypeVideo              = 4 // 视频
	MsgTypeFile               = 5 // 文件
	MsgTypeSticker            = 6 // 表情贴纸
	MsgTypeRevoke             = 7 // 撤回通知
	MsgTypeWebRTC             = 8 // WebRTC 信令
	MsgTypeSystemNotification = 9 // 系统通知
)

// Message Lifecycle State Constants (ACK -> DELIVERED -> READ 状态机)
const (
	MsgStatusSending   = 0 // 客户端发送中
	MsgStatusSent      = 1 // 服务端已收录落盘 (SERVER_ACK)
	MsgStatusDelivered = 2 // 接收端已送达 (DELIVERED_ACK)
	MsgStatusRead      = 3 // 接收端已读 (READ_RECEIPT)
)

// Message 核心消息表 (mim_messages)
type Message struct {
	ID             uint64    `gorm:"primaryKey;column:id;comment:Snowflake 消息ID" json:"id"`
	ClientMsgID    string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_client_msg_id;column:client_msg_id;comment:客户端防重UUID" json:"client_msg_id"`
	ConversationID string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_conv_seq,priority:1;index:idx_conv_created,priority:1;column:conversation_id;comment:会话ID" json:"conversation_id"`
	ConvType       int8      `gorm:"type:tinyint;not null;default:0;column:conv_type;comment:0单聊 1群聊" json:"conv_type"`
	SenderID       uint64    `gorm:"not null;column:sender_id;comment:发送方ID" json:"sender_id"`
	ReceiverID     uint64    `gorm:"not null;default:0;column:receiver_id;comment:接收方ID(单聊有效)" json:"receiver_id"`
	SeqID          uint64    `gorm:"not null;uniqueIndex:uk_conv_seq,priority:2;column:seq_id;comment:会话内连续自增序号" json:"seq_id"`
	MsgType        int8      `gorm:"type:tinyint;not null;default:0;column:msg_type;comment:消息类型 0文本 1MD 2图片 3语音 4视频 5文件 6贴纸 7撤回 8WebRTC 9系统" json:"msg_type"`
	Content        string    `gorm:"type:text;column:content;comment:消息文本或媒体结构体JSON" json:"content"`
	QuotedMsgID    uint64    `gorm:"not null;default:0;column:quoted_msg_id;comment:引用的父消息ID" json:"quoted_msg_id"`
	QuotedContent  string    `gorm:"type:varchar(255);not null;default:'';column:quoted_content;comment:引用消息摘要" json:"quoted_content"`
	Status         int8      `gorm:"type:tinyint;not null;default:1;column:status;comment:消息生命周期状态: 0发送中 1已发送 2已送达 3已读" json:"status"`
	IsRevoked      int8      `gorm:"type:tinyint;not null;default:0;column:is_revoked;comment:是否撤回 0否 1是" json:"is_revoked"`
	IsEncrypted    int8      `gorm:"type:tinyint;not null;default:0;column:is_encrypted;comment:是否端到端加密 0否 1是" json:"is_encrypted"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP;index:idx_conv_created,priority:2;comment:创建时间" json:"created_at"`
}

func (Message) TableName() string {
	return "mim_messages"
}
