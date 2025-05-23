package user_driver

import (
	"context"
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/user_model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type UserDriver struct {
	adapter drivers.Adapter
}

func NewUserDriver(adapter drivers.Adapter) *UserDriver {
	return &UserDriver{adapter: adapter}
}

func (d *UserDriver) CreateUser(ctx context.Context, user *user_model.User) error {
	_, err := d.adapter.Exec(ctx, drivers.QueryCreateUser, user.Id, user.Email, user.PasswordHash, user.Role)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreateUser.Message)
		return custom_errors.ErrCreateUser
	}

	return nil
}

func (d *UserDriver) GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) {
	var id pgtype.UUID
	var passwordHash []byte
	var userRole user_model.UserRole

	err := d.adapter.QueryRow(ctx, drivers.QueryGetUserByEmail, email).Scan(&id, &passwordHash, &userRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, custom_errors.ErrUserNotFound
		}

		log.Error().Err(err).Msg(custom_errors.ErrGetUserByEmail.Message)
		return nil, custom_errors.ErrGetUserByEmail
	}

	user := &user_model.User{Id: id, Email: email, PasswordHash: passwordHash, Role: userRole}
	return user, nil
}

func (d *UserDriver) GetUserById(ctx context.Context, id pgtype.UUID) (*user_model.User, error) {
	var email string
	var passwordHash []byte
	var userRole user_model.UserRole

	err := d.adapter.QueryRow(ctx, drivers.QueryGetUserById, id).Scan(&email, &passwordHash, &userRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, custom_errors.ErrUserNotFound
		}

		log.Error().Err(err).Msg(custom_errors.ErrGetUserById.Message)
		return nil, custom_errors.ErrGetUserById
	}

	user := &user_model.User{Id: id, Email: email, PasswordHash: passwordHash, Role: userRole}
	return user, nil
}
