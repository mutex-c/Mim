package models

import (
	"time"

	"gorm.io/gorm"
)

// Friend Verification Status Constants
const (
	FriendVerificationPending  = 1 // 待处理
	FriendVerificationAgreed   = 2 // 同意
	FriendVerificationRejected = 3 // 拒绝
	FriendVerificationIgnored  = 4 // 忽略
)

// Friend 好友关系链表 (mim_friends)
type Friend struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    uint64         `gorm:"not null;uniqueIndex:uk_user_friend,priority:1;column:user_id;comment:用户ID" json:"user_id"`
	FriendID  uint64         `gorm:"not null;uniqueIndex:uk_user_friend,priority:2;index:idx_friend_id;column:friend_id;comment:好友ID" json:"friend_id"`
	Remark    string         `gorm:"type:varchar(64);not null;default:'';column:remark;comment:好友备注" json:"remark"`
	IsPinned  int8           `gorm:"type:tinyint;not null;default:0;column:is_pinned;comment:是否置顶 0否 1是" json:"is_pinned"`
	IsBlack   int8           `gorm:"type:tinyint;not null;default:0;column:is_black;comment:是否拉黑 0否 1是" json:"is_black"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Friend) TableName() string {
	return "mim_friends"
}

// FriendVerification 好友申请验证流转表 (mim_friend_verifications)
type FriendVerification struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SenderID   uint64    `gorm:"not null;index:idx_sender_receiver,priority:1;column:sender_id;comment:申请人ID" json:"sender_id"`
	ReceiverID uint64    `gorm:"not null;index:idx_receiver_status,priority:1;index:idx_sender_receiver,priority:2;column:receiver_id;comment:被申请人ID" json:"receiver_id"`
	VerifyMsg  string    `gorm:"type:varchar(255);not null;default:'';column:verify_msg;comment:申请留言" json:"verify_msg"`
	Status     int8      `gorm:"type:tinyint;not null;default:1;index:idx_receiver_status,priority:2;column:status;comment:1待处理 2同意 3拒绝 4忽略" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (FriendVerification) TableName() string {
	return "mim_friend_verifications"
}
