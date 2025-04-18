package product_service

import (
	"context"
	"errors"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/product_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	products2 "pvz/internal/models/product_model"
	"pvz/internal/services"
	"pvz/internal/services/reception_service"
	"time"
)

type ProductService struct {
	driver           product_driver.IProductDriver
	receptionService reception_service.IReceptionService
}

func NewProductService(driver product_driver.IProductDriver, receptionService reception_service.IReceptionService) *ProductService {
	return &ProductService{driver: driver, receptionService: receptionService}
}

func (s *ProductService) CreateProduct(ctx context.Context, pvzIdDto openapi_types.UUID, productTypeJson generated.PostProductsJSONBodyType) (*generated.Product, error) {
	productType, err := mapJsonToProductType(productTypeJson)
	if err != nil {
		return nil, err
	}

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

	product := &products2.Product{Id: id, AddingTime: time.Now(), ProductType: productType}
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
		Type:        generated.ProductType(productType),
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

func mapJsonToProductType(productType generated.PostProductsJSONBodyType) (products2.ProductType, error) {
	switch productType {
	case "электроника":
		return products2.Electronics, nil
	case "обувь":
		return products2.Shoes, nil
	case "одежда":
		return products2.Clothes, nil
	default:
		log.Error().Msg(custom_errors.ErrProductType.Message)
		return "", custom_errors.ErrProductType
	}
}
