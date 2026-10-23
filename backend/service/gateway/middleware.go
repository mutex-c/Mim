package gateway

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"mim/pkg/ctxdata"
	"mim/pkg/jwtx"
	"mim/pkg/response"

	"github.com/google/uuid"
)

type MiddlewareConfig struct {
	JwtSecret string
	Whitelist []*regexp.Regexp
}

type AuthMiddleware struct {
	cfg MiddlewareConfig
}

func NewAuthMiddleware(cfg MiddlewareConfig) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

// TraceMiddleware generates or propagates X-Trace-Id
func TraceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		w.Header().Set("X-Trace-Id", traceID)

		ctx := context.WithValue(r.Context(), ctxdata.CtxKeyTraceID, traceID)
		next(w, r.WithContext(ctx))
	}
}

// Handle authenticates incoming HTTP requests
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Check Whitelist
		for _, re := range m.cfg.Whitelist {
			if re.MatchString(path) {
				next(w, r)
				return
			}
		}

		// Check Bearer Token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(w, "missing or malformed authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtx.ParseToken(m.cfg.JwtSecret, tokenStr)
		if err != nil {
			response.Unauthorized(w, "invalid or expired token")
			return
		}

		if claims.TokenType != "access" {
			response.Unauthorized(w, "token is not an access token")
			return
		}

		// Inject into context
		traceID, _ := r.Context().Value(ctxdata.CtxKeyTraceID).(string)
		ctx := ctxdata.SetUserInfo(r.Context(), claims.UserID, claims.Account, claims.Platform, claims.DeviceID, traceID)

		next(w, r.WithContext(ctx))
	}
}
