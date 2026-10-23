package auth

import (
	"errors"
	"time"

	"mim/models"
	"mim/pkg/encrypt"
	"mim/pkg/jwtx"

	"gorm.io/gorm"
)

var (
	ErrAccountExists   = errors.New("account already exists")
	ErrInvalidPassword = errors.New("invalid account or password")
	ErrUserBanned      = errors.New("user account is banned")
	ErrUserNotFound    = errors.New("user not found")
)

type Config struct {
	JwtSecret     string
	AccessExpire  time.Duration
	RefreshExpire time.Duration
}

type Service struct {
	db  *gorm.DB
	cfg Config
}

func NewService(db *gorm.DB, cfg Config) *Service {
	return &Service{db: db, cfg: cfg}
}

type RegisterReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}

type LoginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	Platform string `json:"platform"` // web, windows, android
	DeviceID string `json:"device_id"`
}

type LoginResp struct {
	User  models.User     `json:"user"`
	Token *jwtx.TokenPair `json:"token"`
}

func (s *Service) Register(req *RegisterReq) (*models.User, error) {
	if len(req.Account) < 3 || len(req.Password) < 6 {
		return nil, errors.New("account min 3 chars, password min 6 chars")
	}

	var count int64
	s.db.Model(&models.User{}).Where("account = ?", req.Account).Count(&count)
	if count > 0 {
		return nil, ErrAccountExists
	}

	hash, err := encrypt.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Account
	}

	user := models.User{
		Account:      req.Account,
		Nickname:     nickname,
		PasswordHash: hash,
		Avatar:       req.Avatar,
		Status:       models.UserStatusNormal,
		Role:         models.UserRoleNormal,
	}

	// Transaction to create User and UserConfig
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		userConfig := models.UserConfig{
			UserID:           user.ID,
			SearchByAccount:  1,
			SearchByNickname: 1,
			FriendVerifyMode: models.VerifyModeNeedConfirm,
			RecallTimeLimit:  120,
		}
		return tx.Create(&userConfig).Error
	})

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) Login(req *LoginReq) (*LoginResp, error) {
	var user models.User
	if err := s.db.Where("account = ?", req.Account).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidPassword
		}
		return nil, err
	}

	if user.Status == models.UserStatusBanned {
		return nil, ErrUserBanned
	}

	if !encrypt.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, ErrInvalidPassword
	}

	platform := req.Platform
	if platform == "" {
		platform = "web"
	}

	tokenPair, err := jwtx.GenerateTokenPair(s.cfg.JwtSecret, s.cfg.AccessExpire, s.cfg.RefreshExpire, user.ID, user.Account, platform, req.DeviceID)
	if err != nil {
		return nil, err
	}

	return &LoginResp{
		User:  user,
		Token: tokenPair,
	}, nil
}

func (s *Service) RefreshToken(refreshToken string) (*jwtx.TokenPair, error) {
	claims, err := jwtx.ParseToken(s.cfg.JwtSecret, refreshToken)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid refresh token type")
	}

	return jwtx.GenerateTokenPair(s.cfg.JwtSecret, s.cfg.AccessExpire, s.cfg.RefreshExpire, claims.UserID, claims.Account, claims.Platform, claims.DeviceID)
}
