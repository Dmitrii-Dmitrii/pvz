package receptions

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/generated"
)

type IReceptionService interface {
	CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	CheckLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) error
}
