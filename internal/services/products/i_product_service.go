package products

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/generated"
)

type IProductService interface {
	CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeDto generated.ProductType) (*generated.Product, error)
	DeleteLastProduct(ctx context.Context, pvzIdDto openapi_types.UUID) error
}
