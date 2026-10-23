package auth_test

import (
	"testing"
	"time"

	"mim/models"
	"mim/service/auth"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTest(t *testing.T) *auth.Service {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cfg := auth.Config{
		JwtSecret:     "auth_test_secret_2026",
		AccessExpire:  1 * time.Hour,
		RefreshExpire: 24 * time.Hour,
	}

	return auth.NewService(db, cfg)
}

func TestAuthFlow(t *testing.T) {
	svc := setupAuthTest(t)

	// 1. Test Register
	regReq := &auth.RegisterReq{
		Account:  "mim_user_1",
		Password: "password123",
		Nickname: "Mim User",
	}
	user, err := svc.Register(regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if user.ID == 0 || user.Account != "mim_user_1" {
		t.Errorf("unexpected registered user: %+v", user)
	}

	// Duplicate register should fail
	_, err = svc.Register(regReq)
	if err != auth.ErrAccountExists {
		t.Errorf("expected ErrAccountExists, got: %v", err)
	}

	// 2. Test Login
	loginReq := &auth.LoginReq{
		Account:  "mim_user_1",
		Password: "password123",
		Platform: "windows",
		DeviceID: "win-pc-001",
	}
	loginResp, err := svc.Login(loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == nil || loginResp.Token.AccessToken == "" {
		t.Errorf("missing access token in login response")
	}

	// Wrong password should fail
	badLogin := &auth.LoginReq{
		Account:  "mim_user_1",
		Password: "wrongpassword",
	}
	_, err = svc.Login(badLogin)
	if err != auth.ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}

	// 3. Test Refresh Token
	newPair, err := svc.RefreshToken(loginResp.Token.RefreshToken)
	if err != nil {
		t.Fatalf("refresh token failed: %v", err)
	}
	if newPair.AccessToken == "" {
		t.Errorf("empty new access token after refresh")
	}
}
