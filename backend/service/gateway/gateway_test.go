package gateway_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"mim/pkg/ctxdata"
	"mim/pkg/jwtx"
	"mim/service/gateway"
)

func TestGatewayMiddlewares(t *testing.T) {
	jwtSecret := "gateway_test_secret_key"
	whitelist := []*regexp.Regexp{
		regexp.MustCompile(`^/api/v1/auth/login`),
		regexp.MustCompile(`^/api/v1/auth/register`),
	}

	authMw := gateway.NewAuthMiddleware(gateway.MiddlewareConfig{
		JwtSecret: jwtSecret,
		Whitelist: whitelist,
	})

	// 1. Whitelisted route should pass without token
	var passedWhitelist bool
	handlerWhitelist := authMw.Handle(func(w http.ResponseWriter, r *http.Request) {
		passedWhitelist = true
		w.WriteHeader(http.StatusOK)
	})

	reqWhite := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	recWhite := httptest.NewRecorder()
	handlerWhitelist(recWhite, reqWhite)

	if !passedWhitelist || recWhite.Code != http.StatusOK {
		t.Fatalf("expected whitelisted path to pass")
	}

	// 2. Protected route without token should return 401
	handlerProtected := authMw.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	reqProtected := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	recProtected := httptest.NewRecorder()
	handlerProtected(recProtected, reqProtected)

	if recProtected.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got: %d", recProtected.Code)
	}

	// 3. Protected route with valid token should inject user info into context
	tokenPair, _ := jwtx.GenerateTokenPair(jwtSecret, time.Hour, 24*time.Hour, 888, "alice", "web", "browser-id")

	var extractedUserID uint64
	var extractedTraceID string
	handlerWithToken := gateway.TraceMiddleware(authMw.Handle(func(w http.ResponseWriter, r *http.Request) {
		extractedUserID = ctxdata.GetUserID(r.Context())
		extractedTraceID = ctxdata.GetTraceID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	reqWithToken := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	reqWithToken.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	recWithToken := httptest.NewRecorder()

	handlerWithToken(recWithToken, reqWithToken)

	if recWithToken.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got: %d", recWithToken.Code)
	}
	if extractedUserID != 888 {
		t.Errorf("expected user ID 888, got: %d", extractedUserID)
	}
	if extractedTraceID == "" {
		t.Errorf("expected injected TraceId in context")
	}
	if recWithToken.Header().Get("X-Trace-Id") == "" {
		t.Errorf("expected X-Trace-Id in response header")
	}
}
