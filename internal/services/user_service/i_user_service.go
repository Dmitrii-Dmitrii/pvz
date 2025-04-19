package user_service

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/generated"
	"pvz/internal/models/user_model"
)

type IUserService interface {
	DummyLogin(ctx context.Context, roleDto generated.UserRole) (string, error)
	Register(ctx context.Context, emailDto openapi_types.Email, password string, roleDto generated.UserRole) (*generated.User, string, error)
	Login(ctx context.Context, emailDto openapi_types.Email, password string) (string, error)
	ValidateToken(ctx context.Context, token string) (*user_model.User, error)
}
