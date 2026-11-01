package user_test

import (
	"testing"

	"mim/models"
	"mim/service/user"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserTest(t *testing.T) (*user.Service, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	svc := user.NewService(db)
	return svc, db
}

func TestUserService(t *testing.T) {
	svc, db := setupUserTest(t)

	// Create test users
	u1 := models.User{Account: "alice", Nickname: "Alice", Status: models.UserStatusNormal}
	u2 := models.User{Account: "bob", Nickname: "Bob", Status: models.UserStatusNormal}
	db.Create(&u1)
	db.Create(&u2)

	// 1. Test Get & Update Profile
	profile, err := svc.GetProfile(u1.ID)
	if err != nil {
		t.Fatalf("failed to get profile: %v", err)
	}
	if profile.User.Account != "alice" {
		t.Errorf("account mismatch: %s", profile.User.Account)
	}

	err = svc.UpdateProfile(u1.ID, "Alice in Wonderland", "", "My bio")
	if err != nil {
		t.Fatalf("update profile failed: %v", err)
	}

	// 2. Test Friend Request Flow
	_, err = svc.SendFriendRequest(u1.ID, u2.ID, "Hi, I am Alice")
	if err != nil {
		t.Fatalf("send friend request failed: %v", err)
	}

	var req models.FriendVerification
	db.First(&req, "sender_id = ? AND receiver_id = ?", u1.ID, u2.ID)

	// Receiver agrees
	err = svc.HandleFriendRequest(req.ID, u2.ID, models.FriendVerificationAgreed)
	if err != nil {
		t.Fatalf("handle friend request failed: %v", err)
	}

	// Verify both have each other in friend list
	friendsOfAlice, err := svc.GetFriendList(u1.ID)
	if err != nil || len(friendsOfAlice) != 1 {
		t.Fatalf("expected 1 friend for Alice, got: %d, err: %v", len(friendsOfAlice), err)
	}
	if friendsOfAlice[0].FriendID != u2.ID {
		t.Errorf("expected friend Bob, got ID: %d", friendsOfAlice[0].FriendID)
	}

	friendsOfBob, err := svc.GetFriendList(u2.ID)
	if err != nil || len(friendsOfBob) != 1 {
		t.Fatalf("expected 1 friend for Bob, got: %d", len(friendsOfBob))
	}

	// 3. Test Delete Friend
	err = svc.DeleteFriend(u1.ID, u2.ID)
	if err != nil {
		t.Fatalf("delete friend failed: %v", err)
	}
	friendsAfterDelete, _ := svc.GetFriendList(u1.ID)
	if len(friendsAfterDelete) != 0 {
		t.Errorf("expected 0 friends after delete, got %d", len(friendsAfterDelete))
	}
}
