package products

import (
	"context"
	"errors"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/drivers/products"
	"pvz/internal/generated"
	"pvz/internal/models"
	"pvz/internal/models/custom_errors"
	"pvz/internal/services"
	"pvz/internal/services/receptions"
	"time"
)

type ProductService struct {
	driver           products.IProductDriver
	receptionService receptions.IReceptionService
}

func NewProductService(driver products.IProductDriver, receptionService receptions.IReceptionService) *ProductService {
	return &ProductService{driver: driver, receptionService: receptionService}
}

func (s *ProductService) CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeDto generated.ProductType) (*generated.Product, error) {
	productType := models.ProductType(productTypeDto)

	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return nil, err
	}

	err = s.receptionService.CheckLastReceptionStatus(ctx, pvzId)
	if err != nil && !errors.Is(err, custom_errors.ErrNoReception) {
		return nil, err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return nil, err
	}

	product := &models.Product{Id: id, AddingTime: time.Now(), ProductType: productType}
	receptionId, err := s.driver.CreateProduct(ctx, product, pvzId)
	if err != nil {
		return nil, err
	}

	idDto, err := services.ConvertPgUuidToOpenAPI(product.Id)
	if err != nil {
		return nil, err
	}

	receptionIdDto, err := services.ConvertPgUuidToOpenAPI(*receptionId)
	if err != nil {
		return nil, err
	}

	productDto := &generated.Product{
		Id:          &idDto,
		DateTime:    &product.AddingTime,
		ReceptionId: receptionIdDto,
		Type:        productTypeDto,
	}

	return productDto, nil
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzIdDto openapi_types.UUID) error {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return err
	}

	err = s.receptionService.CheckLastReceptionStatus(ctx, pvzId)
	if err != nil && !errors.Is(err, custom_errors.ErrNoReception) {
		return err
	}

	return s.driver.DeleteLastProduct(ctx, pvzId)
}
