package receptions

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type IReceptionService interface {
	CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) error
	CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) error
}
