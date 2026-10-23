package models

import (
	"gorm.io/gorm"
)

// AllModels returns all GORM entity models for migration
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&UserConfig{},
		&Friend{},
		&FriendVerification{},
		&Conversation{},
		&Message{},
		&Group{},
		&GroupMember{},
		&FileRecord{},
	}
}

// AutoMigrate migrates all database tables
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
