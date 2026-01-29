package jwt

import "context"

type contextKey string

const (
	// ClaimsKey is the key used to store JWT claims in the context
	SessionIdKey contextKey = "sessionIdOfClaims"
	UserIdKey    contextKey = "userIdOfClaims"
)

func setSessionIdToContext(ctx context.Context, sessionId string) context.Context {
	return context.WithValue(ctx, SessionIdKey, sessionId)
}

func SessionIdFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(SessionIdKey).(string); ok {
		return val
	}
	return ""
}

func setUserIdToContext(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, UserIdKey, userId)
}

func UserIdFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(UserIdKey).(string); ok {
		return val
	}
	return ""
}
