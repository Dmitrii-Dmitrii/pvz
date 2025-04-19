package reception_service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/reception_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/reception_model"
	"pvz/internal/services"
	"time"
)

type ReceptionService struct {
	driver reception_driver.IReceptionDriver
}

func NewReceptionService(driver reception_driver.IReceptionDriver) *ReceptionService {
	return &ReceptionService{driver: driver}
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzIdDto openapi_types.UUID) (*generated.Reception, error) {
	pvzId, err := services.ConvertOpenAPIUuidToPgType(pvzIdDto)
	if err != nil {
		return nil, err
	}

	isClosed, err := s.IsLastReceptionStatusClose(ctx, pvzId)
	if err != nil && !errors.Is(err, custom_errors.ErrNoReception) {
		return nil, err
	}

	if !isClosed {
		log.Warn().Msg(custom_errors.ErrInProgressReception.Message)
		return nil, custom_errors.ErrInProgressReception
	}

	id := services.GenerateUuid()

	reception := &reception_model.Reception{Id: id, ReceptionTime: time.Now(), PvzId: pvzId, Status: reception_model.InProgress}
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

	isClosed, err := s.IsLastReceptionStatusClose(ctx, pvzId)
	if err != nil && errors.Is(err, custom_errors.ErrNoReception) {
		return nil, err
	}

	if isClosed {
		log.Warn().Msg(custom_errors.ErrNoOpenReception.Message)
		return nil, custom_errors.ErrNoOpenReception
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

func (s *ReceptionService) IsLastReceptionStatusClose(ctx context.Context, pvzId pgtype.UUID) (bool, error) {
	status, err := s.driver.GetLastReceptionStatus(ctx, pvzId)
	if errors.Is(err, custom_errors.ErrNoReception) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	if *status != reception_model.Close {
		return false, nil
	}

	return true, nil
}
