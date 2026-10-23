package models

import (
	"time"

	"gorm.io/gorm"
)

// User Status Constants
const (
	UserStatusNormal = 1 // 正常
	UserStatusMuted  = 2 // 禁言
	UserStatusBanned = 3 // 封禁
)

// User Role Constants
const (
	UserRoleNormal = 1 // 普通用户
	UserRoleAdmin  = 2 // 管理员
)

// Friend Verify Mode Constants
const (
	VerifyModeDirect      = 1 // 直接添加
	VerifyModeNeedConfirm = 2 // 需要验证
	VerifyModeQuestion    = 3 // 问答验证
	VerifyModeRejectAll   = 4 // 拒绝添加
)

// User 用户基础模型 (mim_users)
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;column:id;comment:用户ID" json:"id"`
	Account      string         `gorm:"type:varchar(32);not null;uniqueIndex:uk_account;column:account;comment:用户唯一账号" json:"account"`
	Nickname     string         `gorm:"type:varchar(64);not null;column:nickname;comment:用户昵称" json:"nickname"`
	PasswordHash string         `gorm:"type:varchar(128);not null;column:password_hash;comment:密码bcrypt哈希" json:"-"`
	Avatar       string         `gorm:"type:varchar(255);not null;default:'';column:avatar;comment:头像URL" json:"avatar"`
	Signature    string         `gorm:"type:varchar(255);not null;default:'';column:signature;comment:个性签名" json:"signature"`
	Email        string         `gorm:"type:varchar(128);not null;default:'';index:idx_email;column:email;comment:绑定邮箱" json:"email"`
	Phone        string         `gorm:"type:varchar(32);not null;default:'';index:idx_phone;column:phone;comment:绑定手机" json:"phone"`
	Status       int8           `gorm:"type:tinyint;not null;default:1;column:status;comment:1正常 2禁言 3封禁" json:"status"`
	Role         int8           `gorm:"type:tinyint;not null;default:1;column:role;comment:1普通用户 2管理员" json:"role"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (User) TableName() string {
	return "mim_users"
}

// UserConfig 用户个性化配置表 (mim_user_configs)
type UserConfig struct {
	UserID               uint64    `gorm:"primaryKey;column:user_id;comment:关联用户ID" json:"user_id"`
	SearchByAccount      int8      `gorm:"type:tinyint;not null;default:1;column:search_by_account;comment:允许账号搜索: 0禁止 1允许" json:"search_by_account"`
	SearchByNickname     int8      `gorm:"type:tinyint;not null;default:1;column:search_by_nickname;comment:允许昵称搜索: 0禁止 1允许" json:"search_by_nickname"`
	FriendVerifyMode     int8      `gorm:"type:tinyint;not null;default:2;column:friend_verify_mode;comment:1直接添加 2需要验证 3问答验证 4拒绝添加" json:"friend_verify_mode"`
	FriendVerifyQuestion string    `gorm:"type:varchar(255);not null;default:'';column:friend_verify_question;comment:验证问题" json:"friend_verify_question"`
	FriendVerifyAnswer   string    `gorm:"type:varchar(255);not null;default:'';column:friend_verify_answer;comment:验证答案" json:"friend_verify_answer"`
	RecallTimeLimit      int32     `gorm:"type:int;not null;default:120;column:recall_time_limit;comment:撤回时效限制(秒)" json:"recall_time_limit"`
	UpdatedAt            time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (UserConfig) TableName() string {
	return "mim_user_configs"
}
