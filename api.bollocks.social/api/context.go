package api

import "context"

type contextKey int

var userIdKey contextKey

func ContextWithUserId(ctx context.Context, userId string) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func ContextGetUserId(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIdKey).(string)
	return v, ok
}
