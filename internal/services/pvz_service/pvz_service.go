package pvz_service

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"pvz/internal"
	"pvz/internal/drivers/pvz_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/pvz_model"
	"pvz/internal/services"
	"time"
)

type PvzService struct {
	driver pvz_driver.IPvzDriver
}

func NewPvzService(driver pvz_driver.IPvzDriver) *PvzService {
	return &PvzService{driver: driver}
}

func (s *PvzService) CreatePvz(ctx context.Context, pvzDto generated.PVZ) (*generated.PVZ, error) {
	var id pgtype.UUID
	var err error
	if pvzDto.Id == nil {
		id = services.GenerateUuid()
	} else {
		id, err = services.ConvertOpenAPIUuidToPgType(*pvzDto.Id)
		if err != nil {
			return nil, err
		}
	}

	pvz, err := s.driver.GetPvzById(ctx, id)
	if !errors.Is(err, custom_errors.ErrPvzNotFound) && err != nil {
		return nil, err
	}

	if pvz != nil {
		log.Warn().Msg(custom_errors.ErrPvzExists.Message)
		return nil, custom_errors.ErrPvzExists
	}

	registrationDate := time.Now()
	if pvzDto.RegistrationDate != nil {
		registrationDate = *pvzDto.RegistrationDate
	}

	city, err := mapCityDtoToCity(pvzDto.City)
	if err != nil {
		return nil, err
	}

	pvz = &pvz_model.Pvz{Id: id, RegistrationDate: registrationDate, City: city}

	err = s.driver.CreatePvz(ctx, pvz)
	if err != nil {
		return nil, err
	}

	idDto, err := services.ConvertPgUuidToOpenAPI(id)
	if err != nil {
		return nil, err
	}

	pvzDto.Id = &idDto
	pvzDto.RegistrationDate = &registrationDate

	internal.PvzCreatedTotal.Inc()

	return &pvzDto, nil
}

func (s *PvzService) GetPvz(ctx context.Context, pvzParams generated.GetPvzParams) ([]map[string]interface{}, error) {
	if pvzParams.StartDate != nil && pvzParams.EndDate != nil {
		if pvzParams.EndDate.Before(*pvzParams.StartDate) {
			log.Error().Msg(custom_errors.ErrDateRange.Message)
			return nil, custom_errors.ErrDateRange
		}
	}

	limit := 10
	if pvzParams.Limit != nil {
		if *pvzParams.Limit < 1 || *pvzParams.Limit > 30 {
			log.Error().Msg(custom_errors.ErrLimitValue.Message)
			return nil, custom_errors.ErrLimitValue
		}

		limit = *pvzParams.Limit
	}

	page := 1
	if pvzParams.Page != nil {
		if *pvzParams.Page < 1 {
			log.Error().Msg(custom_errors.ErrPageValue.Message)
			return nil, custom_errors.ErrPageValue
		}

		page = *pvzParams.Page
	}

	offset := (page - 1) * limit

	return s.driver.GetPvz(ctx, uint32(limit), uint32(offset), pvzParams.StartDate, pvzParams.EndDate)
}

func mapCityDtoToCity(cityDto generated.PVZCity) (pvz_model.City, error) {
	switch cityDto {
	case generated.Москва:
		return pvz_model.Moscow, nil
	case generated.СанктПетербург:
		return pvz_model.SPb, nil
	case generated.Казань:
		return pvz_model.Kazan, nil
	default:
		log.Error().Msg(custom_errors.ErrPvzCity.Message)
		return "", custom_errors.ErrPvzCity
	}
}
