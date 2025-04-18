package user_service

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/generated"
)

type IUserService interface {
	DummyLogin(ctx context.Context, roleDto generated.UserRole) (string, error)
	Register(ctx context.Context, email openapi_types.Email, password string, roleDto generated.UserRole) (*generated.User, string, error)
	Login(ctx context.Context, email openapi_types.Email, password string) (string, error)
}
