package pvzs

import (
	"context"
	"github.com/jackc/pgx/pgtype"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers/pvzs"
	"pvz/internal/generated"
	"pvz/internal/models"
	"pvz/internal/models/custom_errors"
	"pvz/internal/services"
	"time"
)

type PvzService struct {
	driver pvzs.IPvzDriver
}

func NewPvzService(driver pvzs.IPvzDriver) *PvzService {
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

	pvz := &models.Pvz{Id: id, RegistrationDate: registrationDate, City: models.City(pvzDto.City)}

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
