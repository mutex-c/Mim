package group_test

import (
	"testing"
	"time"

	"mim/models"
	"mim/service/group"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGroupTest(t *testing.T) (*group.Service, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	svc := group.NewService(db)
	return svc, db
}

func TestGroupLifecycleAndPermissions(t *testing.T) {
	svc, db := setupGroupTest(t)

	// Users: 1: Owner, 2: Admin, 3: Member
	u1 := models.User{Account: "owner"}
	u2 := models.User{Account: "admin"}
	u3 := models.User{Account: "member"}
	db.Create(&u1)
	db.Create(&u2)
	db.Create(&u3)

	// 1. Create Group
	grp, err := svc.CreateGroup(&group.CreateGroupReq{
		Title:     "Mim Core Dev Team",
		OwnerID:   u1.ID,
		MemberIDs: []uint64{u2.ID, u3.ID},
	})
	if err != nil {
		t.Fatalf("create group failed: %v", err)
	}
	if grp.MemberCount != 3 {
		t.Errorf("expected 3 members, got: %d", grp.MemberCount)
	}

	// Promote u2 to Admin using SetMemberRole
	err = svc.SetMemberRole(u1.ID, grp.ID, u2.ID, models.GroupRoleAdmin)
	if err != nil {
		t.Fatalf("promote admin failed: %v", err)
	}

	// Non-owner trying to set role should fail
	err = svc.SetMemberRole(u3.ID, grp.ID, u2.ID, models.GroupRoleMember)
	if err != group.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got: %v", err)
	}

	// 2. Test Dynamic Mute
	// Admin mutes Member for 10 minutes
	err = svc.SetMemberMute(u2.ID, grp.ID, u3.ID, 10*time.Minute)
	if err != nil {
		t.Fatalf("set member mute failed: %v", err)
	}

	isMuted, err := svc.CheckMuted(grp.ID, u3.ID)
	if err != nil || !isMuted {
		t.Errorf("expected member 3 to be muted, got muted=%v, err=%v", isMuted, err)
	}

	// Owner should never be muted
	ownerMuted, _ := svc.CheckMuted(grp.ID, u1.ID)
	if ownerMuted {
		t.Errorf("owner should never be muted")
	}

	// 3. Test Permission on Kick
	// Member tries to kick Admin -> Should Fail
	err = svc.KickMember(u3.ID, grp.ID, u2.ID)
	if err != group.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got: %v", err)
	}

	// Admin tries to kick Owner -> Should Fail
	err = svc.KickMember(u2.ID, grp.ID, u1.ID)
	if err != group.ErrCannotKickHigher {
		t.Errorf("expected ErrCannotKickHigher, got: %v", err)
	}

	// Admin kicks Member -> Should Succeed
	err = svc.KickMember(u2.ID, grp.ID, u3.ID)
	if err != nil {
		t.Fatalf("admin kick member failed: %v", err)
	}

	var updatedGrp models.Group
	db.First(&updatedGrp, grp.ID)
	if updatedGrp.MemberCount != 2 {
		t.Errorf("expected 2 members after kick, got: %d", updatedGrp.MemberCount)
	}

	// 4. Test QuitGroup: Owner cannot quit directly
	err = svc.QuitGroup(u1.ID, grp.ID)
	if err != group.ErrOwnerCannotLeave {
		t.Errorf("expected ErrOwnerCannotLeave, got: %v", err)
	}

	// 5. Test TransferOwner
	err = svc.TransferOwner(u1.ID, grp.ID, u2.ID)
	if err != nil {
		t.Fatalf("transfer owner failed: %v", err)
	}
	db.First(&updatedGrp, grp.ID)
	if updatedGrp.OwnerID != u2.ID {
		t.Errorf("expected owner to be u2 (%d), got: %d", u2.ID, updatedGrp.OwnerID)
	}

	// Now u1 (demoted to admin) can quit group
	err = svc.QuitGroup(u1.ID, grp.ID)
	if err != nil {
		t.Fatalf("former owner quit failed: %v", err)
	}
	db.First(&updatedGrp, grp.ID)
	if updatedGrp.MemberCount != 1 {
		t.Errorf("expected 1 member after u1 quit, got: %d", updatedGrp.MemberCount)
	}
}
