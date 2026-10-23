package jwtx_test

import (
	"testing"
	"time"

	"mim/pkg/jwtx"
)

func TestTokenPairGenerationAndParse(t *testing.T) {
	secret := "mim_super_secret_key_2026"
	pair, err := jwtx.GenerateTokenPair(secret, 2*time.Hour, 14*24*time.Hour, 1001, "testuser", "windows", "dev-pc-01")
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	claims, err := jwtx.ParseToken(secret, pair.AccessToken)
	if err != nil {
		t.Fatalf("failed to parse access token: %v", err)
	}

	if claims.UserID != 1001 || claims.Account != "testuser" || claims.Platform != "windows" {
		t.Errorf("claims mismatch: %+v", claims)
	}
	if claims.TokenType != "access" {
		t.Errorf("expected access token type, got %s", claims.TokenType)
	}

	refClaims, err := jwtx.ParseToken(secret, pair.RefreshToken)
	if err != nil {
		t.Fatalf("failed to parse refresh token: %v", err)
	}
	if refClaims.TokenType != "refresh" {
		t.Errorf("expected refresh token type, got %s", refClaims.TokenType)
	}
}
