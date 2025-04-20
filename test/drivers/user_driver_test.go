package drivers

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz/internal/drivers/user_driver"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
	"testing"
)

func TestCreateUser(t *testing.T) {
	pool, cleanup := SetupPostgresContainer(t)
	defer cleanup()

	driver := user_driver.NewUserDriver(pool)
	ctx := context.Background()

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}
	email := "test@example.com"
	passwordHash := "passwordHash"
	role := user_model.Employee

	user := &user_model.User{
		Id:           id,
		Email:        email,
		PasswordHash: []byte(passwordHash),
		Role:         role,
	}

	err := driver.CreateUser(ctx, user)
	require.NoError(t, err)

	var dbEmail string
	var dbPasswordHash string
	var dbRole string
	err = pool.QueryRow(ctx, queryGetUser, id).Scan(&dbEmail, &dbPasswordHash, &dbRole)

	require.NoError(t, err)
	assert.Equal(t, email, dbEmail)
	assert.NotEmpty(t, dbPasswordHash)
}

func TestGetUserByEmail(t *testing.T) {
	pool, cleanup := SetupPostgresContainer(t)
	defer cleanup()

	driver := user_driver.NewUserDriver(pool)
	ctx := context.Background()

	_, _, userIds, err := createTestData(ctx, pool)
	require.NoError(t, err)

	t.Run("Get user by existing email", func(t *testing.T) {
		result, err := driver.GetUserByEmail(ctx, "dummy.employee@example.com")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userIds[0], result.Id)
		assert.Equal(t, "dummy.employee@example.com", result.Email)
		assert.Equal(t, []byte{}, result.PasswordHash)
		assert.Equal(t, user_model.Employee, result.Role)
	})

	t.Run("Get user by non-existing email", func(t *testing.T) {
		result, err := driver.GetUserByEmail(ctx, "dummy@example.com")

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)
		assert.Nil(t, result)
	})
}

func TestGetUserById(t *testing.T) {
	pool, cleanup := SetupPostgresContainer(t)
	defer cleanup()

	driver := user_driver.NewUserDriver(pool)
	ctx := context.Background()

	_, _, userIds, err := createTestData(ctx, pool)
	require.NoError(t, err)

	t.Run("Get user by existing id", func(t *testing.T) {
		result, err := driver.GetUserById(ctx, userIds[0])

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userIds[0], result.Id)
		assert.Equal(t, "dummy.employee@example.com", result.Email)
		assert.Equal(t, []byte{}, result.PasswordHash)
		assert.Equal(t, user_model.Employee, result.Role)
	})

	t.Run("Get user by non-existing id", func(t *testing.T) {
		nonExistentId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		result, err := driver.GetUserById(ctx, nonExistentId)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)
		assert.Nil(t, result)
	})
}
