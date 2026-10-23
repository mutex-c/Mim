package models

import (
	"time"
)

// Conversation Type Constants
const (
	ConvTypeSingle = 0 // 单聊
	ConvTypeGroup  = 1 // 群聊
)

// Conversation 会话列表状态表 (mim_conversations)
type Conversation struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ConversationID string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_user_conv,priority:2;column:conversation_id;comment:会话唯一标识" json:"conversation_id"`
	UserID         uint64    `gorm:"not null;uniqueIndex:uk_user_conv,priority:1;index:idx_user_pinned_time,priority:1;column:user_id;comment:所属用户" json:"user_id"`
	TargetID       uint64    `gorm:"not null;column:target_id;comment:单聊目标用户ID或群聊ID" json:"target_id"`
	ConvType       int8      `gorm:"type:tinyint;not null;default:0;column:conv_type;comment:0单聊 1群聊" json:"conv_type"`
	LastMsgID      uint64    `gorm:"not null;default:0;column:last_msg_id;comment:最新一条消息ID" json:"last_msg_id"`
	LastMsgSeq     uint64    `gorm:"not null;default:0;column:last_msg_seq;comment:最新一条消息SeqId(并发乱序守卫)" json:"last_msg_seq"`
	LastMsgPreview string    `gorm:"type:varchar(255);not null;default:'';column:last_msg_preview;comment:最新消息预览" json:"last_msg_preview"`
	LastMsgTime    time.Time `gorm:"column:last_msg_time;not null;default:CURRENT_TIMESTAMP;index:idx_user_pinned_time,priority:3;comment:最新消息时间" json:"last_msg_time"`
	LastReadSeq    uint64    `gorm:"not null;default:0;column:last_read_seq;comment:已读最大SeqId" json:"last_read_seq"`
	UnreadCount    int32     `gorm:"type:int;not null;default:0;column:unread_count;comment:未读数" json:"unread_count"`
	IsPinned       int8      `gorm:"type:tinyint;not null;default:0;index:idx_user_pinned_time,priority:2;column:is_pinned;comment:是否置顶" json:"is_pinned"`
	IsMuted        int8      `gorm:"type:tinyint;not null;default:0;column:is_muted;comment:免打扰 0否 1是" json:"is_muted"`
	CreatedAt      time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (Conversation) TableName() string {
	return "mim_conversations"
}
