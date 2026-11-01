package main

import (
	"log"
	"time"

	"mim/models"
	"mim/pkg/encrypt"

	"gorm.io/gorm"
)

// Seed initial test data
func seedInitialData(db *gorm.DB) {
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	hash, _ := encrypt.HashPassword("123456")

	// Users: Alex Chen (1), Sarah Jenkins (2), Michael Zhang (3), David Miller (4)
	users := []models.User{
		{
			ID:           1,
			Account:      "alex",
			Nickname:     "Alex Chen",
			PasswordHash: hash,
			Avatar:       "AC",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleAdmin,
		},
		{
			ID:           2,
			Account:      "sarah",
			Nickname:     "Sarah Jenkins",
			PasswordHash: hash,
			Avatar:       "SJ",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
		{
			ID:           3,
			Account:      "michael",
			Nickname:     "Michael Zhang",
			PasswordHash: hash,
			Avatar:       "MZ",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
		{
			ID:           4,
			Account:      "david",
			Nickname:     "David Miller",
			PasswordHash: hash,
			Avatar:       "DM",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
	}

	for _, u := range users {
		db.Create(&u)
		db.Create(&models.UserConfig{
			UserID:           u.ID,
			SearchByAccount:  1,
			SearchByNickname: 1,
			FriendVerifyMode: models.VerifyModeDirect,
			RecallTimeLimit:  120,
		})
	}

	// Establish Friends
	friends := []models.Friend{
		{UserID: 1, FriendID: 2},
		{UserID: 2, FriendID: 1},
		{UserID: 1, FriendID: 3},
		{UserID: 3, FriendID: 1},
		{UserID: 1, FriendID: 4},
		{UserID: 4, FriendID: 1},
		{UserID: 2, FriendID: 3},
		{UserID: 3, FriendID: 2},
	}
	db.Create(&friends)

	// Seed Group 101: Core Infrastructure Team
	grp := models.Group{
		ID:           101,
		Title:        "Core Infrastructure Team",
		Avatar:       "CI",
		Notice:       "v1.4.2 canary deployed. All cluster nodes nominal.",
		OwnerID:      1,
		MemberCount:  4,
		MaxMembers:   500,
		IsSearchable: 1,
	}
	db.Create(&grp)

	grpMembers := []models.GroupMember{
		{GroupID: 101, UserID: 1, MemberNickname: "Alex Chen", Role: models.GroupRoleOwner},
		{GroupID: 101, UserID: 2, MemberNickname: "Sarah Jenkins", Role: models.GroupRoleAdmin},
		{GroupID: 101, UserID: 3, MemberNickname: "Michael Zhang", Role: models.GroupRoleMember},
		{GroupID: 101, UserID: 4, MemberNickname: "David Miller", Role: models.GroupRoleMember},
	}
	db.Create(&grpMembers)

	// Establish Initial Conversations
	now := time.Now()
	conv12 := models.Conversation{
		ConversationID: "1_2",
		UserID:         1,
		TargetID:       2,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      2,
		LastMsgSeq:     2,
		LastMsgPreview: "The Q3 architecture review is ready for sign-off.",
		LastMsgTime:    now.Add(-5 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	conv21 := models.Conversation{
		ConversationID: "1_2",
		UserID:         2,
		TargetID:       1,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      2,
		LastMsgSeq:     2,
		LastMsgPreview: "The Q3 architecture review is ready for sign-off.",
		LastMsgTime:    now.Add(-5 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	conv13 := models.Conversation{
		ConversationID: "1_3",
		UserID:         1,
		TargetID:       3,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      3,
		LastMsgSeq:     1,
		LastMsgPreview: "All integration test suites passing.",
		LastMsgTime:    now.Add(-1 * time.Hour),
		UnreadCount:    0,
		IsPinned:       0,
	}
	convGrp1 := models.Conversation{
		ConversationID: "group_101",
		UserID:         1,
		TargetID:       101,
		ConvType:       models.ConvTypeGroup,
		LastMsgID:      4,
		LastMsgSeq:     1,
		LastMsgPreview: "David: Deployed v1.4.2 to canary cluster.",
		LastMsgTime:    now.Add(-20 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	convGrp2 := models.Conversation{
		ConversationID: "group_101",
		UserID:         2,
		TargetID:       101,
		ConvType:       models.ConvTypeGroup,
		LastMsgID:      4,
		LastMsgSeq:     1,
		LastMsgPreview: "David: Deployed v1.4.2 to canary cluster.",
		LastMsgTime:    now.Add(-20 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	db.Create(&conv12)
	db.Create(&conv21)
	db.Create(&conv13)
	db.Create(&convGrp1)
	db.Create(&convGrp2)

	// Seed Messages
	msgs := []models.Message{
		{
			ID:             1,
			ClientMsgID:    "init_c1",
			ConversationID: "1_2",
			ConvType:       models.ConvTypeSingle,
			SenderID:       2,
			ReceiverID:     1,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "Hi Alex, did you review the distributed SeqId monotonic generator proposal?",
			CreatedAt:      now.Add(-10 * time.Minute),
		},
		{
			ID:             2,
			ClientMsgID:    "init_c2",
			ConversationID: "1_2",
			ConvType:       models.ConvTypeSingle,
			SenderID:       2,
			ReceiverID:     1,
			SeqID:          2,
			MsgType:        models.MsgTypeText,
			Content:        "The Q3 architecture review is ready for sign-off.",
			CreatedAt:      now.Add(-5 * time.Minute),
		},
		{
			ID:             3,
			ClientMsgID:    "init_c3",
			ConversationID: "1_3",
			ConvType:       models.ConvTypeSingle,
			SenderID:       3,
			ReceiverID:     1,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "All integration test suites passing.",
			CreatedAt:      now.Add(-1 * time.Hour),
		},
		{
			ID:             4,
			ClientMsgID:    "init_c4",
			ConversationID: "group_101",
			ConvType:       models.ConvTypeGroup,
			SenderID:       4,
			ReceiverID:     0,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "Deployed v1.4.2 to canary cluster. Observability metrics are nominal.",
			CreatedAt:      now.Add(-20 * time.Minute),
		},
	}
	db.Create(&msgs)

	// Seed Initial Call Records (Specifically for Alex Chen and Sarah Jenkins)
	calls := []models.CallRecord{
		{
			CallID:     "call_init_1",
			CallerID:   1,
			CallerName: "Alex Chen",
			CalleeID:   2,
			CalleeName: "Sarah Jenkins",
			CallType:   "audio",
			Status:     "Completed",
			Duration:   184,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			CallID:     "call_init_2",
			CallerID:   1,
			CallerName: "Alex Chen",
			CalleeID:   3,
			CalleeName: "Michael Zhang",
			CallType:   "video",
			Status:     "Completed",
			Duration:   340,
			CreatedAt:  now.Add(-24 * time.Hour),
		},
	}
	db.Create(&calls)

	log.Printf("Seeded 4 demo users, 1 group, initial conversations and calls.")
}
