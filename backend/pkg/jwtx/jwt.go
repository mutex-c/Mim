package jwtx

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

type UserClaims struct {
	UserID   uint64 `json:"user_id"`
	Account  string `json:"account"`
	Platform string `json:"platform"`
	DeviceID string `json:"device_id"`
	TokenType string `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // AccessToken seconds
}

// GenerateTokenPair creates both an AccessToken and a RefreshToken
func GenerateTokenPair(secret string, accessExpire, refreshExpire time.Duration, userID uint64, account, platform, deviceID string) (*TokenPair, error) {
	now := time.Now()

	// 1. Access Token
	accessClaims := UserClaims{
		UserID:    userID,
		Account:   account,
		Platform:  platform,
		DeviceID:  deviceID,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "mim-auth",
			Subject:   account,
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	// 2. Refresh Token
	refreshClaims := UserClaims{
		UserID:    userID,
		Account:   account,
		Platform:  platform,
		DeviceID:  deviceID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpire)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "mim-auth",
			Subject:   account,
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessExpire.Seconds()),
	}, nil
}

// ParseToken validates and extracts claims from a JWT token
func ParseToken(secret, tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrTokenInvalid
}
