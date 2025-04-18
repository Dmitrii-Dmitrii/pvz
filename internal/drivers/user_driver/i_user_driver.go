package user_driver

import (
	"context"
	"pvz/internal/models/user_model"
)

type IUserDriver interface {
	CreateUser(ctx context.Context, user *user_model.User) error
	GetUserByEmail(ctx context.Context, email string) (*user_model.User, error)
}
