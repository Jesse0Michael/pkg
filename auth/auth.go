package auth

import (
	"context"
)

type contextKey string

const (
	AuthorizationContextKey = contextKey("authorization")
	SubjectContextKey       = contextKey("subject")
	AdminContextKey         = contextKey("admin")
	ReadOnlyContextKey      = contextKey("readOnly")
	JTIContextKey           = contextKey("jti")
	RequestContextKey       = contextKey("request")
)

func Authorization(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(AuthorizationContextKey).(string)
	return val, ok
}

func Subject(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(SubjectContextKey).(string)
	return val, ok
}

func Admin(ctx context.Context) (bool, bool) {
	val, ok := ctx.Value(AdminContextKey).(bool)
	return val, ok
}

func ReadOnly(ctx context.Context) (bool, bool) {
	val, ok := ctx.Value(ReadOnlyContextKey).(bool)
	return val, ok
}

func JTI(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(JTIContextKey).(string)
	return val, ok
}

func Check(ctx context.Context, subject string) bool {
	if admin, ok := Admin(ctx); admin && ok {
		return true
	}
	if sub, ok := Subject(ctx); ok && subject != "" && subject == sub {
		return true
	}
	return false
}
