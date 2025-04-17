package db_driver

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	"pvz/internal/models"
)

type IDBDriver interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) (*models.Pvz, error)
	CreateReception(ctx context.Context, reception *models.Reception) error
	GetReception(ctx context.Context, id pgtype.UUID) (*models.Reception, error)
	CreateProducts(ctx context.Context, products []models.Product, pvzId pgtype.UUID) error
	DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) (pgtype.UUID, error)
	CloseReception(ctx context.Context, pvzId pgtype.UUID) error
}
