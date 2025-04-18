package pvz_service

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/pvz_driver"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	pvzs2 "pvz/internal/models/pvz_model"
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
		id, err = services.GenerateUuid()
		if err != nil {
			return nil, err
		}
	} else {
		id, err = services.ConvertOpenAPIUuidToPgType(*pvzDto.Id)
		if err != nil {
			return nil, err
		}
	}

	registrationDate := time.Now()
	if pvzDto.RegistrationDate != nil {
		registrationDate = *pvzDto.RegistrationDate
	}

	city, err := mapCityDtoToCity(pvzDto.City)
	if err != nil {
		return nil, err
	}

	pvz := &pvzs2.Pvz{Id: id, RegistrationDate: registrationDate, City: city}

	_, err = s.driver.CreatePvz(ctx, pvz)
	if err != nil {
		return nil, err
	}

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

func mapCityDtoToCity(cityDto generated.PVZCity) (pvzs2.City, error) {
	switch cityDto {
	case generated.Москва:
		return pvzs2.Moscow, nil
	case generated.СанктПетербург:
		return pvzs2.SPb, nil
	case generated.Казань:
		return pvzs2.Kazan, nil
	default:
		log.Error().Msg(custom_errors.ErrPvzCity.Message)
		return "", custom_errors.ErrPvzCity
	}
}
