package user_driver

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
)

type UserDriver struct {
	rwdb *pgxpool.Pool
}

func NewUserDriver(rwdb *pgxpool.Pool) *UserDriver {
	return &UserDriver{rwdb: rwdb}
}

func (d *UserDriver) CreateUser(ctx context.Context, user *user_model.User) error {
	_, err := d.rwdb.Exec(ctx, drivers.QueryCreateUser, user.Id, user.Email, user.Password, user.Role)
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
	err := d.rwdb.QueryRow(ctx, drivers.QueryGetUserByEmail, email).Scan(&id, &passwordHash, &userRole)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetUserByEmail.Message)
		return nil, custom_errors.ErrGetUserByEmail
	}

	user := &user_model.User{Id: id, Email: email, Password: passwordHash, Role: userRole}
	return user, nil
}
