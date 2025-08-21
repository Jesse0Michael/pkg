package auth

import "context"

type contextKey string

const AuthorizationContextKey = contextKey("authorization")
const SubjectContextKey = contextKey("subject")
const AdminContextKey = contextKey("admin")
const ReadOnlyContextKey = contextKey("readOnly")
const RequestContextKey = contextKey("request")

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
