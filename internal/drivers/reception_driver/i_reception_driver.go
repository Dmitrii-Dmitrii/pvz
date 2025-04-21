package reception_driver

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/jackc/pgx/v5/pgtype"
)

type IReceptionDriver interface {
	CreateReception(ctx context.Context, reception *reception_model.Reception) error
	CloseReception(ctx context.Context, pvzId pgtype.UUID) (*reception_model.Reception, error)
	GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error)
}
