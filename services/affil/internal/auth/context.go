package auth

import "context"

type contextKey string
const userIDKey contextKey = "user_id"

func UserID(ctx context.Context) int64 {
	val, _ := ctx.Value(userIDKey).(int64)
	return val
}

func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}
