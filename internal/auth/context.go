package auth

import (
	"context"

	"github.com/dmksnnk/octo/internal"
)

type ctxKey string

const ctxKeyUser ctxKey = "user"

func ContextWithUser(ctx context.Context, user internal.User) context.Context {
	return context.WithValue(ctx, ctxKeyUser, user)
}

func ContextUser(ctx context.Context) (internal.User, bool) {
	user, ok := ctx.Value(ctxKeyUser).(internal.User)
	return user, ok
}
