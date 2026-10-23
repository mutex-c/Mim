package user

import (
	"errors"

	"mim/models"

	"gorm.io/gorm"
)

var (
	ErrCannotAddSelf       = errors.New("cannot add yourself as friend")
	ErrAlreadyFriends      = errors.New("already friends")
	ErrVerificationExists  = errors.New("friend request already pending")
	ErrVerificationNotFound = errors.New("friend request not found")
	ErrForbidden           = errors.New("permission denied")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type UserProfileResp struct {
	User   models.User       `json:"user"`
	Config models.UserConfig `json:"config"`
}

type UserSearchResult struct {
	ID        uint64 `json:"id"`
	Account   string `json:"account"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	IsFriend  bool   `json:"is_friend"`
}

type FriendItem struct {
	FriendID  uint64 `json:"friend_id"`
	Account   string `json:"account"`
	Nickname  string `json:"nickname"`
	Remark    string `json:"remark"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	IsPinned  int8   `json:"is_pinned"`
}

// GetProfile retrieves user info and config
func (s *Service) GetProfile(userID uint64) (*UserProfileResp, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var cfg models.UserConfig
	s.db.First(&cfg, "user_id = ?", userID)

	return &UserProfileResp{User: user, Config: cfg}, nil
}

// UpdateProfile updates nickname, avatar, and signature
func (s *Service) UpdateProfile(userID uint64, nickname, avatar, signature string) error {
	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if signature != "" {
		updates["signature"] = signature
	}

	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// UpdateConfig updates privacy & anti-harassment configs
func (s *Service) UpdateConfig(userID uint64, cfg *models.UserConfig) error {
	cfg.UserID = userID
	return s.db.Save(cfg).Error
}

// SearchUser finds users by account or nickname (respecting privacy configs)
func (s *Service) SearchUser(keyword string, currentUserID uint64) ([]*UserSearchResult, error) {
	if keyword == "" {
		return nil, nil
	}

	var users []models.User
	// Join with user_configs to respect privacy
	err := s.db.Raw(`
		SELECT u.* FROM mim_users u
		LEFT JOIN mim_user_configs c ON u.id = c.user_id
		WHERE u.id != ? AND u.status = 1 AND (
			(u.account = ? AND (c.search_by_account IS NULL OR c.search_by_account = 1)) OR
			(u.nickname LIKE ? AND (c.search_by_nickname IS NULL OR c.search_by_nickname = 1))
		)
		LIMIT 20
	`, currentUserID, keyword, "%"+keyword+"%").Scan(&users).Error

	if err != nil {
		return nil, err
	}

	results := make([]*UserSearchResult, 0, len(users))
	for _, u := range users {
		var friendCount int64
		s.db.Model(&models.Friend{}).Where("user_id = ? AND friend_id = ?", currentUserID, u.ID).Count(&friendCount)

		results = append(results, &UserSearchResult{
			ID:        u.ID,
			Account:   u.Account,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			Signature: u.Signature,
			IsFriend:  friendCount > 0,
		})
	}

	return results, nil
}

// SendFriendRequest creates a friend verification record
func (s *Service) SendFriendRequest(senderID, receiverID uint64, verifyMsg string) error {
	if senderID == receiverID {
		return ErrCannotAddSelf
	}

	// Check if already friends
	var count int64
	s.db.Model(&models.Friend{}).Where("user_id = ? AND friend_id = ?", senderID, receiverID).Count(&count)
	if count > 0 {
		return ErrAlreadyFriends
	}

	// Check existing pending request
	var pendingCount int64
	s.db.Model(&models.FriendVerification{}).
		Where("sender_id = ? AND receiver_id = ? AND status = ?", senderID, receiverID, models.FriendVerificationPending).
		Count(&pendingCount)
	if pendingCount > 0 {
		return ErrVerificationExists
	}

	req := models.FriendVerification{
		SenderID:   senderID,
		ReceiverID: receiverID,
		VerifyMsg:  verifyMsg,
		Status:     models.FriendVerificationPending,
	}
	return s.db.Create(&req).Error
}

// HandleFriendRequest processes a verification (agreed, rejected, ignored)
func (s *Service) HandleFriendRequest(verificationID, currentUserID uint64, action int8) error {
	var verify models.FriendVerification
	if err := s.db.First(&verify, verificationID).Error; err != nil {
		return ErrVerificationNotFound
	}

	if verify.ReceiverID != currentUserID {
		return ErrForbidden
	}

	if verify.Status != models.FriendVerificationPending {
		return errors.New("friend request already processed")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		verify.Status = action
		if err := tx.Save(&verify).Error; err != nil {
			return err
		}

		// If agreed, establish mutual friendship
		if action == models.FriendVerificationAgreed {
			f1 := models.Friend{UserID: verify.SenderID, FriendID: verify.ReceiverID}
			f2 := models.Friend{UserID: verify.ReceiverID, FriendID: verify.SenderID}
			if err := tx.FirstOrCreate(&f1, f1).Error; err != nil {
				return err
			}
			if err := tx.FirstOrCreate(&f2, f2).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetFriendList returns friends with user info
func (s *Service) GetFriendList(userID uint64) ([]*FriendItem, error) {
	type Result struct {
		FriendID  uint64
		Account   string
		Nickname  string
		Remark    string
		Avatar    string
		Signature string
		IsPinned  int8
	}

	var rawResults []Result
	err := s.db.Raw(`
		SELECT f.friend_id, u.account, u.nickname, f.remark, u.avatar, u.signature, f.is_pinned
		FROM mim_friends f
		INNER JOIN mim_users u ON f.friend_id = u.id
		WHERE f.user_id = ? AND f.deleted_at IS NULL AND u.deleted_at IS NULL
		ORDER BY f.is_pinned DESC, u.nickname ASC
	`, userID).Scan(&rawResults).Error

	if err != nil {
		return nil, err
	}

	items := make([]*FriendItem, len(rawResults))
	for i, r := range rawResults {
		items[i] = &FriendItem{
			FriendID:  r.FriendID,
			Account:   r.Account,
			Nickname:  r.Nickname,
			Remark:    r.Remark,
			Avatar:    r.Avatar,
			Signature: r.Signature,
			IsPinned:  r.IsPinned,
		}
	}
	return items, nil
}

// DeleteFriend removes friendship
func (s *Service) DeleteFriend(userID, friendID uint64) error {
	return s.db.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&models.Friend{}).Error
}
