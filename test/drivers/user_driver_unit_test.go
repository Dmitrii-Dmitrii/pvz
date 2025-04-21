package drivers

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/user_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/user_model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := user_driver.NewUserDriver(mockAdapter)

	user := &user_model.User{
		Id:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Email:        "test@example.com",
		PasswordHash: []byte("hash"),
		Role:         user_model.Moderator,
	}

	mockAdapter.On("Exec", ctx, drivers.QueryCreateUser, []interface{}{
		user.Id, user.Email, user.PasswordHash, user.Role,
	}).Return(pgconn.CommandTag{}, nil)

	err := driver.CreateUser(ctx, user)

	require.NoError(t, err)
	mockAdapter.AssertExpectations(t)
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := user_driver.NewUserDriver(mockAdapter)

	email := "test@example.com"
	expectedUser := &user_model.User{
		Id:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Email:        email,
		PasswordHash: []byte("hash"),
		Role:         user_model.Employee,
	}

	mockRow := new(MockRow)
	mockAdapter.On("QueryRow", ctx, drivers.QueryGetUserByEmail, []interface{}{email}).
		Return(mockRow)

	mockRow.On("Scan",
		mock.AnythingOfType("*pgtype.UUID"),
		mock.AnythingOfType("*[]uint8"),
		mock.AnythingOfType("*user_model.UserRole"),
	).Run(func(args mock.Arguments) {
		*args.Get(0).(*pgtype.UUID) = expectedUser.Id
		*args.Get(1).(*[]byte) = expectedUser.PasswordHash
		*args.Get(2).(*user_model.UserRole) = expectedUser.Role
	}).Return(nil)

	user, err := driver.GetUserByEmail(ctx, email)

	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockAdapter.AssertExpectations(t)
}

func TestGetUserById(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := user_driver.NewUserDriver(mockAdapter)

	userID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	expectedUser := &user_model.User{
		Id:           userID,
		Email:        "test@example.com",
		PasswordHash: []byte("hash"),
		Role:         user_model.Employee,
	}

	mockRow := new(MockRow)
	mockAdapter.On("QueryRow", ctx, drivers.QueryGetUserById, []interface{}{userID}).
		Return(mockRow)

	mockRow.On("Scan",
		mock.AnythingOfType("*string"),
		mock.AnythingOfType("*[]uint8"),
		mock.AnythingOfType("*user_model.UserRole"),
	).Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = expectedUser.Email
		*args.Get(1).(*[]byte) = expectedUser.PasswordHash
		*args.Get(2).(*user_model.UserRole) = expectedUser.Role
	}).Return(nil)

	user, err := driver.GetUserById(ctx, userID)

	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
	mockAdapter.AssertExpectations(t)
}
