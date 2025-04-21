package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"os"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
	"pvz/internal/services/user_service"
	"testing"
	"time"
)

type MockUserDriver struct {
	mock.Mock
}

func (m *MockUserDriver) GetUserByEmail(ctx context.Context, email string) (*user_model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_model.User), args.Error(1)
}

func (m *MockUserDriver) GetUserById(ctx context.Context, id pgtype.UUID) (*user_model.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_model.User), args.Error(1)
}

func (m *MockUserDriver) CreateUser(ctx context.Context, user *user_model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func setupJwtSecret(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")

	err := os.Setenv("JWT_SECRET", "test-secret-key")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.Setenv("JWT_SECRET", originalSecret)
	})
}

func TestDummyLogin(t *testing.T) {
	ctx := context.Background()
	setupJwtSecret(t)

	t.Run("Dummy login with non-existing user", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		role := generated.UserRoleEmployee
		email := "dummy.employee@example.com"

		mockDriver.On("GetUserByEmail", ctx, email).Return(nil, nil)
		mockDriver.On("CreateUser", ctx, mock.AnythingOfType("*user_model.User")).Return(nil)

		token, err := service.DummyLogin(ctx, role)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Dummy login with existing user", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		role := generated.UserRoleModerator
		email := "dummy.moderator@example.com"

		existingUser := &user_model.User{
			Id:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Email:        email,
			PasswordHash: []byte{},
			Role:         user_model.Moderator,
		}

		mockDriver.On("GetUserByEmail", ctx, email).Return(existingUser, nil)

		token, err := service.DummyLogin(ctx, role)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Dummy login with invalid role", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		invalidRole := generated.UserRole("invalid")

		token, err := service.DummyLogin(ctx, invalidRole)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, custom_errors.ErrUserRole, err)
		mockDriver.AssertNotCalled(t, "GetUserByEmail")
	})

	t.Run("Dummy login with driver error", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		role := generated.UserRoleEmployee
		email := "dummy.employee@example.com"
		expectedErr := errors.New("database error")

		mockDriver.On("GetUserByEmail", ctx, email).Return(nil, expectedErr)

		token, err := service.DummyLogin(ctx, role)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, expectedErr, err)
		mockDriver.AssertExpectations(t)
	})
}

func TestRegister(t *testing.T) {
	ctx := context.Background()
	setupJwtSecret(t)

	t.Run("Register user", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		password := "password123"
		role := generated.UserRoleEmployee

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(nil, custom_errors.ErrUserNotFound)
		mockDriver.On("CreateUser", ctx, mock.AnythingOfType("*user_model.User")).Return(nil)

		userDto, token, err := service.Register(ctx, email, password, role)

		assert.NoError(t, err)
		assert.NotNil(t, userDto)
		assert.Equal(t, email, userDto.Email)
		assert.Equal(t, role, userDto.Role)
		assert.NotEmpty(t, token)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Register user with invalid email", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		invalidEmail := openapi_types.Email("invalid-email")
		password := "password123"
		role := generated.UserRoleEmployee

		userDto, token, err := service.Register(ctx, invalidEmail, password, role)

		assert.Error(t, err)
		assert.Nil(t, userDto)
		assert.Empty(t, token)
		mockDriver.AssertNotCalled(t, "CreateUser")
	})

	t.Run("Register user with invalid role", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		password := "password123"
		invalidRole := generated.UserRole("invalid")

		userDto, token, err := service.Register(ctx, email, password, invalidRole)

		assert.Error(t, err)
		assert.Nil(t, userDto)
		assert.Empty(t, token)
		assert.Equal(t, custom_errors.ErrUserRole, err)
		mockDriver.AssertNotCalled(t, "CreateUser")
	})

	t.Run("Register user with existing email in db", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		password := "password123"
		invalidRole := generated.UserRoleEmployee

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(nil, custom_errors.ErrExistingUser)

		userDto, token, err := service.Register(ctx, email, password, invalidRole)

		assert.Error(t, err)
		assert.Nil(t, userDto)
		assert.Empty(t, token)
		assert.Equal(t, custom_errors.ErrExistingUser, err)
		mockDriver.AssertNotCalled(t, "CreateUser")
	})

	t.Run("Register user with driver error", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		password := "password123"
		role := generated.UserRoleEmployee
		expectedErr := errors.New("database error")

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(nil, custom_errors.ErrUserNotFound)
		mockDriver.On("CreateUser", ctx, mock.AnythingOfType("*user_model.User")).Return(expectedErr)

		userDto, token, err := service.Register(ctx, email, password, role)

		assert.Error(t, err)
		assert.Nil(t, userDto)
		assert.Empty(t, token)
		assert.Equal(t, expectedErr, err)
		mockDriver.AssertExpectations(t)
	})
}

func TestLogin(t *testing.T) {
	ctx := context.Background()
	setupJwtSecret(t)

	t.Run("Login existing user", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		password := "password123"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		require.NoError(t, err)

		existingUser := &user_model.User{
			Id:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Email:        string(email),
			PasswordHash: passwordHash,
			Role:         user_model.Employee,
		}

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(existingUser, nil)

		token, err := service.Login(ctx, email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Login non-existing user", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("nonexistent@example.com")
		password := "password123"
		expectedErr := custom_errors.ErrUserNotFound

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(nil, expectedErr)

		token, err := service.Login(ctx, email, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, expectedErr, err)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Login existing user with invalid password", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		email := openapi_types.Email("test@example.com")
		correctPassword := "password123"
		wrongPassword := "wrongpassword"
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
		require.NoError(t, err)

		existingUser := &user_model.User{
			Id:           pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Email:        string(email),
			PasswordHash: passwordHash,
			Role:         user_model.Employee,
		}

		mockDriver.On("GetUserByEmail", ctx, string(email)).Return(existingUser, nil)

		token, err := service.Login(ctx, email, wrongPassword)

		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Equal(t, custom_errors.ErrLoginPassword, err)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Login existing user with invalid email format", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		invalidEmail := openapi_types.Email("invalid-email")
		password := "password123"

		token, err := service.Login(ctx, invalidEmail, password)

		assert.Error(t, err)
		assert.Empty(t, token)
		mockDriver.AssertNotCalled(t, "GetUserByEmail")
	})
}

func TestValidateToken(t *testing.T) {
	ctx := context.Background()
	setupJwtSecret(t)

	t.Run("Validate valid token", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		userID := uuid.New().String()
		email := "test@example.com"
		role := "employee"

		claims := user_model.JwtClaims{
			UserId: userID,
			Email:  email,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("test-secret-key"))
		require.NoError(t, err)

		pgUuid := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		existingUser := &user_model.User{
			Id:           pgUuid,
			Email:        email,
			PasswordHash: []byte{},
			Role:         user_model.Employee,
		}

		mockDriver.On("GetUserById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(existingUser, nil)

		user, err := service.ValidateToken(ctx, signedToken)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, existingUser, user)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Validate invalid token", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		invalidToken := "invalid.token.string"

		user, err := service.ValidateToken(ctx, invalidToken)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockDriver.AssertNotCalled(t, "GetUserById")
	})

	t.Run("Validate expired token", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		userID := uuid.New().String()
		email := "test@example.com"
		role := "employee"

		claims := user_model.JwtClaims{
			UserId: userID,
			Email:  email,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("test-secret-key"))
		require.NoError(t, err)

		user, err := service.ValidateToken(ctx, signedToken)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockDriver.AssertNotCalled(t, "GetUserById")
	})

	t.Run("Validate toke: User not found", func(t *testing.T) {
		mockDriver := new(MockUserDriver)
		service := user_service.NewUserService(mockDriver)

		userID := uuid.New().String()
		email := "test@example.com"
		role := "employee"

		claims := user_model.JwtClaims{
			UserId: userID,
			Email:  email,
			Role:   role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, err := token.SignedString([]byte("test-secret-key"))
		require.NoError(t, err)

		mockDriver.On("GetUserById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrUserNotFound)

		user, err := service.ValidateToken(ctx, signedToken)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, custom_errors.ErrUserNotFound, err)
		mockDriver.AssertExpectations(t)
	})
}
