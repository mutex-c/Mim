package ctxdata

import (
	"context"
)

type ctxKey string

const (
	CtxKeyUserID   ctxKey = "user_id"
	CtxKeyAccount  ctxKey = "account"
	CtxKeyPlatform ctxKey = "platform"
	CtxKeyDeviceID ctxKey = "device_id"
	CtxKeyTraceID  ctxKey = "trace_id"
)

func SetUserInfo(ctx context.Context, userID uint64, account, platform, deviceID, traceID string) context.Context {
	ctx = context.WithValue(ctx, CtxKeyUserID, userID)
	ctx = context.WithValue(ctx, CtxKeyAccount, account)
	ctx = context.WithValue(ctx, CtxKeyPlatform, platform)
	ctx = context.WithValue(ctx, CtxKeyDeviceID, deviceID)
	ctx = context.WithValue(ctx, CtxKeyTraceID, traceID)
	return ctx
}

func GetUserID(ctx context.Context) uint64 {
	if val, ok := ctx.Value(CtxKeyUserID).(uint64); ok {
		return val
	}
	return 0
}

func GetAccount(ctx context.Context) string {
	if val, ok := ctx.Value(CtxKeyAccount).(string); ok {
		return val
	}
	return ""
}

func GetPlatform(ctx context.Context) string {
	if val, ok := ctx.Value(CtxKeyPlatform).(string); ok {
		return val
	}
	return ""
}

func GetTraceID(ctx context.Context) string {
	if val, ok := ctx.Value(CtxKeyTraceID).(string); ok {
		return val
	}
	return ""
}
