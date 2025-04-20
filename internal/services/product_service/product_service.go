package product_service

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/product_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/product_model"
	"pvz/internal/models/reception_model"
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

	status, err := s.receptionService.GetLastReceptionStatus(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	if *status == reception_model.Close {
		log.Warn().Msg(custom_errors.ErrNoOpenReception.Message)
		return nil, custom_errors.ErrNoOpenReception
	}

	id := services.GenerateUuid()

	product := &product_model.Product{Id: id, AddingTime: time.Now(), ProductType: productType}
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

	status, err := s.receptionService.GetLastReceptionStatus(ctx, pvzId)
	if err != nil {
		return err
	}

	if *status == reception_model.Close {
		log.Warn().Msg(custom_errors.ErrNoOpenReception.Message)
		return custom_errors.ErrNoOpenReception
	}

	return s.driver.DeleteLastProduct(ctx, pvzId)
}

func mapJsonToProductType(productType generated.PostProductsJSONBodyType) (product_model.ProductType, error) {
	switch productType {
	case "электроника":
		return product_model.Electronics, nil
	case "обувь":
		return product_model.Shoes, nil
	case "одежда":
		return product_model.Clothes, nil
	default:
		log.Error().Msg(custom_errors.ErrProductType.Message)
		return "", custom_errors.ErrProductType
	}
}
