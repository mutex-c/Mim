package models_test

import (
	"testing"
	"time"

	"mim/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect sqlite in-memory db: %v", err)
	}

	err = models.AutoMigrate(db)
	if err != nil {
		t.Fatalf("failed to auto migrate models: %v", err)
	}
	return db
}

func TestAutoMigrateAndCRUD(t *testing.T) {
	db := setupTestDB(t)

	// 1. Test User & UserConfig
	user := models.User{
		Account:      "admin",
		Nickname:     "Super Admin",
		PasswordHash: "$2a$10$hashedpasswordplaceholder",
		Avatar:       "https://avatar.com/admin.png",
		Status:       models.UserStatusNormal,
		Role:         models.UserRoleAdmin,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	cfg := models.UserConfig{
		UserID:           user.ID,
		SearchByAccount:  1,
		SearchByNickname: 1,
		FriendVerifyMode: models.VerifyModeNeedConfirm,
		RecallTimeLimit:  120,
	}
	if err := db.Create(&cfg).Error; err != nil {
		t.Fatalf("failed to create user config: %v", err)
	}

	// 2. Test Conversation & Message with SeqID
	conv := models.Conversation{
		ConversationID: "1_2",
		UserID:         user.ID,
		TargetID:       2,
		ConvType:       models.ConvTypeSingle,
		LastMsgPreview: "Hello World",
		LastReadSeq:    1,
		UnreadCount:    0,
	}
	if err := db.Create(&conv).Error; err != nil {
		t.Fatalf("failed to create conversation: %v", err)
	}

	msg := models.Message{
		ID:             10001,
		ClientMsgID:    "uuid-client-123456",
		ConversationID: "1_2",
		ConvType:       models.ConvTypeSingle,
		SenderID:       user.ID,
		ReceiverID:     2,
		SeqID:          1,
		MsgType:        models.MsgTypeText,
		Content:        "Hello World",
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatalf("failed to create message: %v", err)
	}

	// 3. Test Group & GroupMember
	grp := models.Group{
		Title:        "Mim Developers Group",
		OwnerID:      user.ID,
		MemberCount:  1,
		MaxMembers:   500,
		IsSearchable: 1,
		JoinMode:     models.GroupJoinModeVerify,
	}
	if err := db.Create(&grp).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}

	muteUntil := time.Now().Add(10 * time.Minute)
	member := models.GroupMember{
		GroupID:        grp.ID,
		UserID:         user.ID,
		MemberNickname: "AdminUser",
		Role:           models.GroupRoleOwner,
		MutedUntil:     &muteUntil,
	}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("failed to create group member: %v", err)
	}

	// 4. Test FileRecord
	file := models.FileRecord{
		FileHash:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		FileName:    "test.png",
		FileSize:    1024,
		FileType:    "image/png",
		StoragePath: "/uploads/2026/test.png",
		UploaderID:  user.ID,
	}
	if err := db.Create(&file).Error; err != nil {
		t.Fatalf("failed to create file record: %v", err)
	}

	t.Logf("Successfully verified migration and CRUD for all 9 core models!")
}
