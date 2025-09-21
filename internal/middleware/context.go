package middleware

import (
	"backend/internal/db"
	"context"
)

type contextKey string

const ContextUserKey = contextKey("user")

// Attach user to context
func WithUser(ctx context.Context, user *db.User) context.Context {
	return context.WithValue(ctx, ContextUserKey, user)
}

// Retrieve user from context
func GetUser(ctx context.Context) *db.User {
	if v := ctx.Value(ContextUserKey); v != nil {
		if u, ok := v.(*db.User); ok {
			return u
		}
	}
	return nil
}
