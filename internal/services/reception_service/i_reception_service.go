package reception_service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/generated"
)

type IReceptionService interface {
	CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	IsLastReceptionStatusClose(ctx context.Context, pvzId pgtype.UUID) (bool, error)
}
