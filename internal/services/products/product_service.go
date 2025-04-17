package products

import (
	"context"
	"fmt"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/drivers/products"
	"pvz/internal/generated"
	"pvz/internal/models"
	"pvz/internal/services"
	"time"
)

type ProductService struct {
	driver products.IProductDriver
}

func (s *ProductService) CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeDto generated.ProductType) error {
	productType, err := convertProductTypeDtoToProductType(productTypeDto)
	if err != nil {
		return err
	}

	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return err
	}

	product := &models.Product{Id: id, AddingTime: time.Now(), ProductType: productType}

	return s.driver.CreateProduct(ctx, product, pvzId)
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzIdDto openapi_types.UUID) error {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return err
	}

	return s.driver.DeleteLastProduct(ctx, pvzId)
}

func convertProductTypeDtoToProductType(productTypeDto generated.ProductType) (models.ProductType, error) {
	switch productTypeDto {
	case generated.ProductTypeЭлектроника:
		return models.Electronics, nil
	case generated.ProductTypeОдежда:
		return models.Clothes, nil
	case generated.ProductTypeОбувь:
		return models.Shoes, nil
	default:
		return 0, fmt.Errorf("unknown city: %s", productTypeDto)
	}
}
