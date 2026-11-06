package user

import (
	"errors"
	"strings"
	"fmt"
	"time"

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
	ID                   uint64 `json:"id"`
	Account              string `json:"account"`
	Nickname             string `json:"nickname"`
	Avatar               string `json:"avatar"`
	Signature            string `json:"signature"`
	IsFriend             bool   `json:"is_friend"`
	FriendVerifyMode     int8   `json:"friend_verify_mode"`
	FriendVerifyQuestion string `json:"friend_verify_question"`
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
	var existing models.UserConfig
	if err := s.db.Where("user_id = ?", userID).First(&existing).Error; err != nil {
		cfg.UserID = userID
		if cfg.RecallTimeLimit == 0 {
			cfg.RecallTimeLimit = 120
		}
		return s.db.Create(cfg).Error
	}
	updates := map[string]interface{}{
		"search_by_account":      cfg.SearchByAccount,
		"search_by_nickname":     cfg.SearchByNickname,
		"friend_verify_mode":     cfg.FriendVerifyMode,
		"friend_verify_question": cfg.FriendVerifyQuestion,
		"friend_verify_answer":   cfg.FriendVerifyAnswer,
	}
	if cfg.RecallTimeLimit > 0 {
		updates["recall_time_limit"] = cfg.RecallTimeLimit
	}
	return s.db.Model(&existing).Updates(updates).Error
}

// SearchUser finds users by account or nickname (respecting privacy configs)
func (s *Service) SearchUser(keyword string, currentUserID uint64) ([]*UserSearchResult, error) {
	if keyword == "" {
		return nil, nil
	}

	type UserWithConfig struct {
		models.User
		FriendVerifyMode     int8   `gorm:"column:friend_verify_mode"`
		FriendVerifyQuestion string `gorm:"column:friend_verify_question"`
	}
	var users []UserWithConfig
	// Join with user_configs to respect privacy
	err := s.db.Raw(`
		SELECT u.*,
		       COALESCE(c.friend_verify_mode, 2) AS friend_verify_mode,
		       COALESCE(c.friend_verify_question, '') AS friend_verify_question
		FROM mim_users u
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
			ID:                   u.ID,
			Account:              u.Account,
			Nickname:             u.Nickname,
			Avatar:               u.Avatar,
			Signature:            u.Signature,
			IsFriend:             friendCount > 0,
			FriendVerifyMode:     u.FriendVerifyMode,
			FriendVerifyQuestion: u.FriendVerifyQuestion,
		})
	}

	return results, nil
}

// establishFriendship binds two users as mutual friends, creates conversation entries, and posts a welcome message
func (s *Service) establishFriendship(tx *gorm.DB, u1, u2 uint64) error {
	// Clean up any existing friendship records (including soft-deleted) to avoid unique constraint violations
	if err := tx.Unscoped().Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", u1, u2, u2, u1).Delete(&models.Friend{}).Error; err != nil {
		return err
	}

	now := time.Now()
	f1 := models.Friend{UserID: u1, FriendID: u2, CreatedAt: now, UpdatedAt: now}
	f2 := models.Friend{UserID: u2, FriendID: u1, CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&f1).Error; err != nil {
		return err
	}
	if err := tx.Create(&f2).Error; err != nil {
		return err
	}

	// Update any pending friend verifications between u1 and u2 to agreed
	_ = tx.Model(&models.FriendVerification{}).
		Where("((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)) AND status = ?", u1, u2, u2, u1, models.FriendVerificationPending).
		Update("status", models.FriendVerificationAgreed).Error

	minID := u1
	maxID := u2
	if minID > maxID {
		minID, maxID = maxID, minID
	}
	convID := fmt.Sprintf("%d_%d", minID, maxID)

	var maxSeq uint64
	row := tx.Model(&models.Message{}).Where("conversation_id = ?", convID).Select("COALESCE(MAX(seq_id), 0)").Row()
	_ = row.Scan(&maxSeq)
	nextSeq := maxSeq + 1

	c1 := models.Conversation{
		ConversationID: convID,
		UserID:         u1,
		TargetID:       u2,
		ConvType:       models.ConvTypeSingle,
		LastMsgPreview: "你们已成为好友，现在可以开始聊天了",
		LastMsgTime:    now,
		LastMsgSeq:     nextSeq,
	}
	_ = tx.Where("conversation_id = ? AND user_id = ?", convID, u1).FirstOrCreate(&c1).Error
	_ = tx.Model(&models.Conversation{}).Where("conversation_id = ? AND user_id = ?", convID, u1).Updates(map[string]interface{}{
		"last_msg_preview": "你们已成为好友，现在可以开始聊天了",
		"last_msg_time":    now,
		"last_msg_seq":     nextSeq,
	}).Error

	c2 := models.Conversation{
		ConversationID: convID,
		UserID:         u2,
		TargetID:       u1,
		ConvType:       models.ConvTypeSingle,
		LastMsgPreview: "你们已成为好友，现在可以开始聊天了",
		LastMsgTime:    now,
		LastMsgSeq:     nextSeq,
	}
	_ = tx.Where("conversation_id = ? AND user_id = ?", convID, u2).FirstOrCreate(&c2).Error
	_ = tx.Model(&models.Conversation{}).Where("conversation_id = ? AND user_id = ?", convID, u2).Updates(map[string]interface{}{
		"last_msg_preview": "你们已成为好友，现在可以开始聊天了",
		"last_msg_time":    now,
		"last_msg_seq":     nextSeq,
	}).Error

	welcomeMsg := models.Message{
		ClientMsgID:    fmt.Sprintf("sys_friend_%d_%d_%d", minID, maxID, now.UnixNano()),
		ConversationID: convID,
		ConvType:       models.ConvTypeSingle,
		SenderID:       0,
		ReceiverID:     0,
		SeqID:          nextSeq,
		MsgType:        models.MsgTypeText,
		Content:        "你们已成为好友，现在可以开始聊天了。",
		CreatedAt:      now,
	}
	return tx.Create(&welcomeMsg).Error
}

// SendFriendRequest creates a friend verification record or directly establishes friendship
func (s *Service) SendFriendRequest(senderID, receiverID uint64, verifyMsg string) (bool, error) {
	if senderID == receiverID {
		return false, ErrCannotAddSelf
	}

	// Check if already friends
	var count int64
	s.db.Model(&models.Friend{}).Where("user_id = ? AND friend_id = ?", senderID, receiverID).Count(&count)
	if count > 0 {
		return false, ErrAlreadyFriends
	}

	// Check if receiver already sent a pending request to sender (cross-request -> instant mutual friendship)
	var reverseVerify models.FriendVerification
	if err := s.db.Where("sender_id = ? AND receiver_id = ? AND status = ?", receiverID, senderID, models.FriendVerificationPending).First(&reverseVerify).Error; err == nil {
		err = s.HandleFriendRequest(reverseVerify.ID, senderID, models.FriendVerificationAgreed)
		return true, err
	}

	// Check receiver's FriendVerifyMode
	var receiverConfig models.UserConfig
	if err := s.db.Where("user_id = ?", receiverID).First(&receiverConfig).Error; err == nil {
		if receiverConfig.FriendVerifyMode == models.VerifyModeRejectAll {
			return false, errors.New("对方已开启拒绝添加好友")
		}
		if receiverConfig.FriendVerifyMode == models.VerifyModeDirect { // 1 = Direct Add (No verification required)
			err := s.db.Transaction(func(tx *gorm.DB) error {
				return s.establishFriendship(tx, senderID, receiverID)
			})
			return true, err
		}
		if receiverConfig.FriendVerifyMode == models.VerifyModeQuestion {
			if receiverConfig.FriendVerifyAnswer != "" {
				if strings.TrimSpace(verifyMsg) != strings.TrimSpace(receiverConfig.FriendVerifyAnswer) {
					return false, errors.New("验证答案不正确")
				}
				// 正确回答问题直接通过
				err := s.db.Transaction(func(tx *gorm.DB) error {
					return s.establishFriendship(tx, senderID, receiverID)
				})
				return true, err
			}
		}
	}

	// Check existing pending request from sender to receiver
	var existing models.FriendVerification
	if err := s.db.Where("sender_id = ? AND receiver_id = ? AND status = ?", senderID, receiverID, models.FriendVerificationPending).First(&existing).Error; err == nil {
		existing.VerifyMsg = verifyMsg
		existing.UpdatedAt = time.Now()
		return false, s.db.Save(&existing).Error
	}

	req := models.FriendVerification{
		SenderID:   senderID,
		ReceiverID: receiverID,
		VerifyMsg:  verifyMsg,
		Status:     models.FriendVerificationPending,
	}
	return false, s.db.Create(&req).Error
}

// FriendRequestItem represents a pending or handled friend request item
type FriendRequestItem struct {
	ID        uint64 `json:"id"`
	SenderID  uint64 `json:"sender_id"`
	Account   string `json:"account"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	VerifyMsg string `json:"verify_msg"`
	Status    int8   `json:"status"` // 1: Pending, 2: Agreed, 3: Rejected
	CreatedAt string `json:"created_at"`
}

// GetFriendRequests returns all incoming verification requests for a user
func (s *Service) GetFriendRequests(userID uint64) ([]FriendRequestItem, error) {
	var records []models.FriendVerification
	err := s.db.Where("receiver_id = ?", userID).Order("status ASC, created_at DESC").Limit(50).Find(&records).Error
	if err != nil {
		return nil, err
	}

	items := make([]FriendRequestItem, 0, len(records))
	for _, r := range records {
		var u models.User
		s.db.Select("id, account, nickname, avatar").First(&u, r.SenderID)
		items = append(items, FriendRequestItem{
			ID:        r.ID,
			SenderID:  r.SenderID,
			Account:   u.Account,
			Nickname:  u.Nickname,
			Avatar:    u.Avatar,
			VerifyMsg: r.VerifyMsg,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	return items, nil
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

		// If agreed, establish mutual friendship and create conversations
		if action == models.FriendVerificationAgreed {
			if err := s.establishFriendship(tx, verify.SenderID, verify.ReceiverID); err != nil {
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
	return s.db.Unscoped().Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&models.Friend{}).Error
}
