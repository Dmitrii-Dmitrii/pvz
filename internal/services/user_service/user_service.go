package user_service

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
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

	email := "dummy." + string(role) + "@example.com"

	user, err := s.driver.GetUserByEmail(ctx, email)
	var userErr *custom_errors.UserError
	if !errors.As(err, &userErr) && err != nil {
		return "", err
	}

	var id pgtype.UUID
	if user == nil {
		id = services.GenerateUuid()
		passwordHash := make([]byte, 0)

		user = &user_model.User{Id: id, Email: email, PasswordHash: passwordHash, Role: role}
		err = s.driver.CreateUser(ctx, user)
		if err != nil {
			return "", err
		}
	} else {
		id = user.Id
	}

	token, err := createToken(id.String(), email, string(role))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) Register(ctx context.Context, emailDto openapi_types.Email, password string, roleDto generated.UserRole) (*generated.User, string, error) {
	role, err := mapRoleDtoToRole(roleDto)
	if err != nil {
		return nil, "", err
	}

	err = validateEmail(emailDto)
	if err != nil {
		return nil, "", err
	}

	id := services.GenerateUuid()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrHashPassword.Message)
		return nil, "", custom_errors.ErrHashPassword
	}

	email := string(emailDto)
	user := &user_model.User{
		Id:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}

	err = s.driver.CreateUser(ctx, user)
	if err != nil {
		return nil, "", err
	}

	token, err := createToken(id.String(), email, string(role))
	if err != nil {
		return nil, "", err
	}

	idDto, err := services.ConvertPgUuidToOpenAPI(id)
	if err != nil {
		return nil, "", err
	}

	userDto := generated.User{
		Id:    &idDto,
		Email: emailDto,
		Role:  roleDto,
	}

	return &userDto, token, nil
}

func (s *UserService) Login(ctx context.Context, emailDto openapi_types.Email, password string) (string, error) {
	err := validateEmail(emailDto)
	if err != nil {
		return "", err
	}

	email := string(emailDto)

	user, err := s.driver.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrHashPassword.Message)
		return "", custom_errors.ErrLoginPassword
	}

	token, err := createToken(user.Id.String(), email, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *UserService) ValidateToken(ctx context.Context, token string) (*user_model.User, error) {
	claims, err := user_model.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	pgUuid, err := mapStringUuidToPgUuid(claims.UserId)
	if err != nil {
		return nil, err
	}

	user, err := s.driver.GetUserById(ctx, pgUuid)
	if err != nil {
		return nil, custom_errors.ErrUserNotFound
	}

	return user, nil
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
	claims := user_model.JwtClaims{
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
	signedToken, err := token.SignedString(user_model.GetJwtSecret())
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGenerateJWTToken.Message)
		return "", custom_errors.ErrGenerateJWTToken
	}

	return signedToken, nil
}

func validateEmail(email openapi_types.Email) error {
	if _, err := email.MarshalJSON(); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrEmailFormat.Message)
		return custom_errors.ErrEmailFormat
	}

	return nil
}

func mapStringUuidToPgUuid(s string) (pgtype.UUID, error) {
	parsedUuid, err := uuid.Parse(s)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrInvalidUuid.Message)
		return pgtype.UUID{}, custom_errors.ErrInvalidUuid
	}

	pgUuid := pgtype.UUID{
		Bytes: parsedUuid,
		Valid: true,
	}

	return pgUuid, nil
}
