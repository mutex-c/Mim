package models

import (
	"time"

	"gorm.io/gorm"
)

// Group Role Constants
const (
	GroupRoleOwner  = 1 // 群主
	GroupRoleAdmin  = 2 // 管理员
	GroupRoleMember = 3 // 普通成员
)

// Group Join Mode Constants
const (
	GroupJoinModeFree   = 1 // 自由加入
	GroupJoinModeVerify = 2 // 需审批
	GroupJoinModeForbid = 3 // 禁止加入
)

// Group 群组基础表 (mim_groups)
type Group struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;column:id;comment:群ID" json:"id"`
	Title        string         `gorm:"type:varchar(64);not null;column:title;comment:群名称" json:"title"`
	Avatar       string         `gorm:"type:varchar(255);not null;default:'';column:avatar;comment:群头像" json:"avatar"`
	Notice       string         `gorm:"type:text;column:notice;comment:群公告" json:"notice"`
	OwnerID      uint64         `gorm:"not null;index:idx_owner;column:owner_id;comment:群主ID" json:"owner_id"`
	MemberCount  int32          `gorm:"type:int;not null;default:1;column:member_count;comment:当前成员总数" json:"member_count"`
	MaxMembers   int32          `gorm:"type:int;not null;default:500;column:max_members;comment:最大成员上限" json:"max_members"`
	IsSearchable int8           `gorm:"type:tinyint;not null;default:1;column:is_searchable;comment:是否允许被搜索" json:"is_searchable"`
	JoinMode     int8           `gorm:"type:tinyint;not null;default:2;column:join_mode;comment:1自由加入 2需审批 3禁止加入" json:"join_mode"`
	IsAllMuted   int8           `gorm:"type:tinyint;not null;default:0;column:is_all_muted;comment:是否全员禁言 0否 1是" json:"is_all_muted"`
	CreatedAt    time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Group) TableName() string {
	return "mim_groups"
}

// GroupMember 群成员与权限关系表 (mim_group_members)
type GroupMember struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupID        uint64     `gorm:"not null;uniqueIndex:uk_group_user,priority:1;index:idx_user_group,priority:2;column:group_id;comment:群ID" json:"group_id"`
	UserID         uint64     `gorm:"not null;uniqueIndex:uk_group_user,priority:2;index:idx_user_group,priority:1;column:user_id;comment:用户ID" json:"user_id"`
	MemberNickname string     `gorm:"type:varchar(64);not null;default:'';column:member_nickname;comment:群内昵称" json:"member_nickname"`
	Role           int8       `gorm:"type:tinyint;not null;default:3;column:role;comment:1群主 2管理员 3普通成员" json:"role"`
	MutedUntil     *time.Time `gorm:"column:muted_until;comment:禁言到期时间(NULL表示未禁言)" json:"muted_until"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (GroupMember) TableName() string {
	return "mim_group_members"
}
