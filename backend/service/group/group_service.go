package group

import (
	"errors"
	"fmt"
	"time"

	"mim/models"

	"gorm.io/gorm"
)

var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrPermissionDenied    = errors.New("group permission denied")
	ErrCannotKickHigher    = errors.New("cannot kick member with equal or higher role")
	ErrOwnerCannotLeave    = errors.New("group owner cannot leave, please disband or transfer group")
	ErrAlreadyInGroup      = errors.New("user already in group")
	ErrMemberMuted         = errors.New("user is currently muted in this group")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateGroupReq struct {
	Title     string   `json:"title"`
	Avatar    string   `json:"avatar"`
	Notice    string   `json:"notice"`
	OwnerID   uint64   `json:"owner_id"`
	MemberIDs []uint64 `json:"member_ids"`
}

// CreateGroup creates a new group and initializes owner & members in a single transaction
func (s *Service) CreateGroup(req *CreateGroupReq) (*models.Group, error) {
	if req.Title == "" {
		return nil, errors.New("group title cannot be empty")
	}

	grp := models.Group{
		Title:        req.Title,
		Avatar:       req.Avatar,
		Notice:       req.Notice,
		OwnerID:      req.OwnerID,
		MemberCount:  int32(len(req.MemberIDs) + 1),
		MaxMembers:   500,
		IsSearchable: 1,
		JoinMode:     models.GroupJoinModeVerify,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&grp).Error; err != nil {
			return err
		}

		// Add Owner
		ownerMember := models.GroupMember{
			GroupID: grp.ID,
			UserID:  req.OwnerID,
			Role:    models.GroupRoleOwner,
		}
		if err := tx.Create(&ownerMember).Error; err != nil {
			return err
		}

		// Add Members
		for _, uid := range req.MemberIDs {
			if uid == req.OwnerID {
				continue
			}
			m := models.GroupMember{
				GroupID: grp.ID,
				UserID:  uid,
				Role:    models.GroupRoleMember,
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		// Create conversation records for owner and members
		convID := fmt.Sprintf("group_%d", grp.ID)
		allUserIDs := append([]uint64{req.OwnerID}, req.MemberIDs...)
		for _, uid := range allUserIDs {
			c := models.Conversation{
				ConversationID: convID,
				UserID:         uid,
				TargetID:       grp.ID,
				ConvType:       models.ConvTypeGroup,
				LastMsgPreview: "群聊已创建",
				LastMsgTime:    time.Now(),
			}
			_ = tx.Create(&c).Error
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &grp, nil
}

// CheckMuted checks if user is allowed to send message in the group
func (s *Service) CheckMuted(groupID, userID uint64) (bool, error) {
	var grp models.Group
	if err := s.db.First(&grp, groupID).Error; err != nil {
		return true, ErrGroupNotFound
	}

	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return true, errors.New("user is not in this group")
	}

	// Owner and Admin are immune to muting
	if member.Role == models.GroupRoleOwner || member.Role == models.GroupRoleAdmin {
		return false, nil
	}

	// Check All Muted
	if grp.IsAllMuted == 1 {
		return true, nil
	}

	// Check Personal Mute Timer
	if member.MutedUntil != nil && time.Now().Before(*member.MutedUntil) {
		return true, nil
	}

	return false, nil
}

// SetMemberMute sets dynamic mute countdown for a member
func (s *Service) SetMemberMute(operatorID, groupID, targetUserID uint64, duration time.Duration) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}
	if opMember.Role != models.GroupRoleOwner && opMember.Role != models.GroupRoleAdmin {
		return ErrPermissionDenied
	}

	var targetMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, targetUserID).First(&targetMember).Error; err != nil {
		return errors.New("target member not found")
	}

	// Admin cannot mute another Admin or Owner
	if opMember.Role == models.GroupRoleAdmin && targetMember.Role <= models.GroupRoleAdmin {
		return ErrPermissionDenied
	}

	var muteTime *time.Time
	if duration > 0 {
		t := time.Now().Add(duration)
		muteTime = &t
	}

	return s.db.Model(&targetMember).Update("muted_until", muteTime).Error
}

// SetAllMuted toggles group-wide mute
func (s *Service) SetAllMuted(operatorID, groupID uint64, allMuted bool) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}
	if opMember.Role != models.GroupRoleOwner && opMember.Role != models.GroupRoleAdmin {
		return ErrPermissionDenied
	}

	val := int8(0)
	if allMuted {
		val = 1
	}

	return s.db.Model(&models.Group{}).Where("id = ?", groupID).Update("is_all_muted", val).Error
}

// KickMember removes a member from the group
func (s *Service) KickMember(operatorID, groupID, targetUserID uint64) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}
	if opMember.Role != models.GroupRoleOwner && opMember.Role != models.GroupRoleAdmin {
		return ErrPermissionDenied
	}

	var targetMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, targetUserID).First(&targetMember).Error; err != nil {
		return errors.New("target member not found")
	}

	// Prevent kicking equal or higher role
	if opMember.Role >= targetMember.Role {
		return ErrCannotKickHigher
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&targetMember).Error; err != nil {
			return err
		}
		// Also delete conversation for kicked user
		convID := fmt.Sprintf("group_%d", groupID)
		_ = tx.Where("user_id = ? AND conversation_id = ?", targetUserID, convID).Delete(&models.Conversation{}).Error
		return tx.Model(&models.Group{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("member_count - 1")).Error
	})
}

// InviteMembers adds new members to the group
func (s *Service) InviteMembers(operatorID, groupID uint64, memberIDs []uint64) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}

	var grp models.Group
	if err := s.db.First(&grp, groupID).Error; err != nil {
		return ErrGroupNotFound
	}

	// Permission checks based on JoinMode
	if grp.JoinMode == models.GroupJoinModeForbid {
		return errors.New("该群已开启禁止加入")
	}
	if grp.JoinMode == models.GroupJoinModeVerify {
		if opMember.Role != models.GroupRoleOwner && opMember.Role != models.GroupRoleAdmin {
			return errors.New("该群开启了进群审批，仅管理员和群主可邀请新成员")
		}
	}

	// Capacity check
	var currentCount int64
	s.db.Model(&models.GroupMember{}).Where("group_id = ?", groupID).Count(&currentCount)
	if int32(currentCount)+int32(len(memberIDs)) > grp.MaxMembers {
		return errors.New("群成员已达上限")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		addedCount := 0
		for _, uid := range memberIDs {
			var count int64
			tx.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, uid).Count(&count)
			if count > 0 {
				continue
			}
			newM := models.GroupMember{
				GroupID: groupID,
				UserID:  uid,
				Role:    models.GroupRoleMember,
			}
			if err := tx.Create(&newM).Error; err != nil {
				return err
			}
			// Ensure conversation exists for invited user
			convID := fmt.Sprintf("group_%d", groupID)
			var userConv models.Conversation
			tx.Where("user_id = ? AND conversation_id = ?", uid, convID).FirstOrCreate(&userConv, models.Conversation{
				ConversationID: convID,
				UserID:         uid,
				TargetID:       groupID,
				ConvType:       models.ConvTypeGroup,
				LastMsgPreview: "欢迎加入群聊",
				LastMsgTime:    time.Now(),
			})
			addedCount++
		}
		if addedCount > 0 {
			return tx.Model(&models.Group{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("member_count + ?", addedCount)).Error
		}
		return nil
	})
}

// UpdateGroupSettings updates group info and preferences
func (s *Service) UpdateGroupSettings(operatorID, groupID uint64, title, notice, avatar string, isSearchable, joinMode, isAllMuted int8) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}
	if opMember.Role != models.GroupRoleOwner && opMember.Role != models.GroupRoleAdmin {
		return ErrPermissionDenied
	}

	updates := map[string]interface{}{}
	if title != "" {
		updates["title"] = title
	}
	if notice != "" {
		updates["notice"] = notice
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if isSearchable >= 0 {
		updates["is_searchable"] = isSearchable
	}
	if joinMode > 0 {
		updates["join_mode"] = joinMode
	}
	if isAllMuted >= 0 {
		updates["is_all_muted"] = isAllMuted
	}

	return s.db.Model(&models.Group{}).Where("id = ?", groupID).Updates(updates).Error
}

// DisbandGroup dissolves a group
func (s *Service) DisbandGroup(operatorID, groupID uint64) error {
	var grp models.Group
	if err := s.db.First(&grp, groupID).Error; err != nil {
		return ErrGroupNotFound
	}
	if grp.OwnerID != operatorID {
		return ErrPermissionDenied
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		convID := fmt.Sprintf("group_%d", groupID)
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("conversation_id = ?", convID).Delete(&models.Conversation{}).Error; err != nil {
			return err
		}
		return tx.Delete(&grp).Error
	})
}

// SetMemberRole sets or revokes admin role for a member (only Owner can perform)
func (s *Service) SetMemberRole(operatorID, groupID, targetUserID uint64, newRole int8) error {
	var opMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, operatorID).First(&opMember).Error; err != nil {
		return ErrPermissionDenied
	}
	if opMember.Role != models.GroupRoleOwner {
		return ErrPermissionDenied
	}

	if targetUserID == operatorID {
		return errors.New("cannot change owner's role directly, please transfer ownership instead")
	}

	if newRole != models.GroupRoleAdmin && newRole != models.GroupRoleMember {
		return errors.New("invalid role")
	}

	var targetMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, targetUserID).First(&targetMember).Error; err != nil {
		return errors.New("target member not found")
	}

	return s.db.Model(&targetMember).Update("role", newRole).Error
}

// TransferOwner transfers group ownership to another member
func (s *Service) TransferOwner(operatorID, groupID, newOwnerID uint64) error {
	var grp models.Group
	if err := s.db.First(&grp, groupID).Error; err != nil {
		return ErrGroupNotFound
	}
	if grp.OwnerID != operatorID {
		return ErrPermissionDenied
	}

	if newOwnerID == operatorID {
		return errors.New("already the owner")
	}

	var newOwnerMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, newOwnerID).First(&newOwnerMember).Error; err != nil {
		return errors.New("target member is not in this group")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&grp).Update("owner_id", newOwnerID).Error; err != nil {
			return err
		}
		if err := tx.Model(&newOwnerMember).Update("role", models.GroupRoleOwner).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, operatorID).Update("role", models.GroupRoleAdmin).Error; err != nil {
			return err
		}
		return nil
	})
}

// QuitGroup removes user from the group (Owner cannot quit directly)
func (s *Service) QuitGroup(userID, groupID uint64) error {
	var member models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return errors.New("user is not in this group")
	}

	if member.Role == models.GroupRoleOwner {
		return ErrOwnerCannotLeave
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&member).Error; err != nil {
			return err
		}
		convID := fmt.Sprintf("group_%d", groupID)
		_ = tx.Where("user_id = ? AND conversation_id = ?", userID, convID).Delete(&models.Conversation{}).Error
		return tx.Model(&models.Group{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("member_count - 1")).Error
	})
}

// GetGroupDetails returns full group model
func (s *Service) GetGroupDetails(groupID uint64) (*models.Group, error) {
	var grp models.Group
	if err := s.db.First(&grp, groupID).Error; err != nil {
		return nil, ErrGroupNotFound
	}
	return &grp, nil
}

