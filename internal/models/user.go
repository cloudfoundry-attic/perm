package models

import "context"

type ctxKey struct{}

type User struct {
	ID string
}

func NewUserContext(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, ctxKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(ctxKey{}).(User)
	return user, ok
}
