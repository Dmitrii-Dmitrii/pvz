package receptions

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	"pvz/internal/models"
)

type IReceptionDriver interface {
	CreateReception(ctx context.Context, reception *models.Reception) error
	GetReception(ctx context.Context, id pgtype.UUID) (*models.Reception, error)
	CloseReception(ctx context.Context, pvzId pgtype.UUID) (*models.Reception, error)
	GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*models.ReceptionStatus, error)
}
