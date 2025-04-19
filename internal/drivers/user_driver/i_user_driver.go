package user_driver

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"pvz/internal/models/user_model"
)

type IUserDriver interface {
	CreateUser(ctx context.Context, user *user_model.User) error
	GetUserByEmail(ctx context.Context, email string) (*user_model.User, error)
	GetUserById(ctx context.Context, id pgtype.UUID) (*user_model.User, error)
}
