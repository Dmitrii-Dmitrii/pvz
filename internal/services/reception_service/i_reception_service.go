package reception_service

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type IReceptionService interface {
	CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error)
	GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error)
}
