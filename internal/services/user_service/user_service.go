package user_service

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/user_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/user_model"
	"pvz/internal/services"
	"time"
)

type UserService struct {
	driver user_driver.IUserDriver
}

func NewUserService(driver user_driver.IUserDriver) *UserService {
	return &UserService{driver: driver}
}

func (s *UserService) DummyLogin(ctx context.Context, roleDto generated.UserRole) (string, error) {
	role, err := mapRoleDtoToRole(roleDto)
	if err != nil {
		return "", err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return "", err
	}

	email := id.String() + "@example.com"
	passwordHash := make([]byte, 0)

	user := &user_model.User{Id: id, Email: email, Password: passwordHash, Role: role}
	err = s.driver.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}

	token, err := createToken(id.String(), email, string(role))
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGenerateJWTToken.Message)
		return "", custom_errors.ErrGenerateJWTToken
	}

	return token, nil
}

func (s *UserService) Register(ctx context.Context, email openapi_types.Email, password string, roleDto generated.UserRole) (*generated.User, string, error) {
	return nil, "", nil
}

func (s *UserService) Login(ctx context.Context, email openapi_types.Email, password string) (string, error) {
	return "", nil
}

func mapRoleDtoToRole(roleDto generated.UserRole) (user_model.UserRole, error) {
	switch roleDto {
	case generated.UserRoleEmployee:
		return user_model.Employee, nil
	case generated.UserRoleModerator:
		return user_model.Moderator, nil
	default:
		log.Error().Msg(custom_errors.ErrUserRole.Message)
		return "", custom_errors.ErrUserRole
	}
}

func createToken(userID, email, role string) (string, error) {
	claims := user_model.JWTClaims{
		UserId: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(user_model.JwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(user_model.GetJWTSecret())
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
