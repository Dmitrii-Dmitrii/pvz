package product_driver

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
)

type IProductDriver interface {
	CreateProduct(ctx context.Context, product *product_model.Product, pvzId pgtype.UUID) (*pgtype.UUID, error)
	DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error
}
