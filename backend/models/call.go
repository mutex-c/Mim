package models

import (
	"time"
)

// CallRecord 通话记录表 (mim_calls)
type CallRecord struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CallID     string    `gorm:"type:varchar(64);not null;index:idx_call_id;column:call_id" json:"call_id"`
	CallerID   uint64    `gorm:"not null;index:idx_caller;column:caller_id" json:"caller_id"`
	CallerName string    `gorm:"type:varchar(64);column:caller_name" json:"caller_name"`
	CalleeID   uint64    `gorm:"not null;index:idx_callee;column:callee_id" json:"callee_id"`
	CalleeName string    `gorm:"type:varchar(64);column:callee_name" json:"callee_name"`
	CallType   string    `gorm:"type:varchar(16);column:call_type" json:"call_type"` // audio | video
	Status     string    `gorm:"type:varchar(32);column:status" json:"status"`       // Completed, Missed, Rejected
	Duration   int       `gorm:"column:duration;default:0" json:"duration"`          // seconds
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP;index:idx_call_created" json:"created_at"`
}

func (CallRecord) TableName() string {
	return "mim_calls"
}
