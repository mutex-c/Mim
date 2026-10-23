package group

import (
	"errors"
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
		return tx.Model(&models.Group{}).Where("id = ?", groupID).Update("member_count", gorm.Expr("member_count - 1")).Error
	})
}
