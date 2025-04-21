package product_service

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type IProductService interface {
	CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeJson generated.PostProductsJSONBodyType) (*generated.Product, error)
	DeleteLastProduct(ctx context.Context, pvzIdDto openapi_types.UUID) error
}
