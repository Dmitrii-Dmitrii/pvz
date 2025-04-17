package receptions

import (
	"context"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/drivers/receptions"
	"pvz/internal/models"
	"pvz/internal/services"
	"time"
)

type ReceptionService struct {
	driver receptions.IReceptionDriver
}

func NewReceptionService(driver receptions.IReceptionDriver) *ReceptionService {
	return &ReceptionService{driver: driver}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) error {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return err
	}

	reception := &models.Reception{Id: id, ReceptionTime: time.Now(), PvzId: pvzId, Status: models.InProgress}

	return s.driver.CreateReception(ctx, reception)
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) error {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return err
	}

	return s.driver.CloseReception(ctx, pvzId)
}
