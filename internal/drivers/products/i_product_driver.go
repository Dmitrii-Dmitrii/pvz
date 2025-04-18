package products

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	"pvz/internal/models"
)

type IProductDriver interface {
	CreateProduct(ctx context.Context, product *models.Product, pvzId pgtype.UUID) (*pgtype.UUID, error)
	DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error
}
