package handler

import "context"

type userKeyType string

var userKey userKeyType = "authUser"

func NewUserContext(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUserFromContext(ctx context.Context) (AuthUser, bool) {
	val := ctx.Value(userKey)
	user, ok := val.(AuthUser)
	return user, ok
}

