package receptions

import (
	"context"
	"errors"
	"github.com/jackc/pgx/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/receptions"
	"pvz/internal/generated"
	"pvz/internal/models"
	"pvz/internal/models/custom_errors"
	"pvz/internal/services"
	"time"
)

type ReceptionService struct {
	driver receptions.IReceptionDriver
}

func NewReceptionService(driver receptions.IReceptionDriver) *ReceptionService {
	return &ReceptionService{driver: driver}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error) {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return nil, err
	}

	err = s.CheckLastReceptionStatus(ctx, pvzId)
	if err != nil && !errors.Is(err, custom_errors.ErrNoReception) {
		return nil, err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return nil, err
	}

	reception := &models.Reception{Id: id, ReceptionTime: time.Now(), PvzId: pvzId, Status: models.InProgress}
	err = s.driver.CreateReception(ctx, reception)
	if err != nil {
		return nil, err
	}

	idDto, err := services.ConvertPgUuidToOpenAPI(reception.Id)
	if err != nil {
		return nil, err
	}

	receptionDto := &generated.Reception{
		Id:       &idDto,
		DateTime: reception.ReceptionTime,
		PvzId:    pvzIdDto,
		Status:   generated.ReceptionStatus(reception.Status),
	}

	return receptionDto, nil
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error) {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return nil, err
	}

	err = s.CheckLastReceptionStatus(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	reception, err := s.driver.CloseReception(ctx, pvzId)
	if err != nil {
		return nil, err
	}

	idDto, err := services.ConvertPgUuidToOpenAPI(reception.Id)
	if err != nil {
		return nil, err
	}

	receptionDto := &generated.Reception{
		Id:       &idDto,
		DateTime: reception.ReceptionTime,
		PvzId:    pvzIdDto,
		Status:   generated.ReceptionStatus(reception.Status),
	}

	return receptionDto, nil
}

func (s *ReceptionService) CheckLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) error {
	status, err := s.driver.GetLastReceptionStatus(ctx, pvzId)
	if err != nil {
		return err
	}

	if *status == models.Close {
		log.Error().Msg(custom_errors.ErrNoOpenReception.Message)
		return custom_errors.ErrNoOpenReception
	}

	return nil
}
