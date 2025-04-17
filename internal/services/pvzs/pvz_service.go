package pvzs

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"pvz/internal/drivers/pvzs"
	"pvz/internal/generated"
	"pvz/internal/models"
	"pvz/internal/services"
	"time"
)

type PvzService struct {
	driver pvzs.IPvzDriver
}

func NewPvzService(driver pvzs.IPvzDriver) *PvzService {
	return &PvzService{driver: driver}
}

func (s *PvzService) CreatePvz(ctx context.Context, cityDto generated.PVZCity) (*generated.PVZ, error) {
	city, err := convertPvzCityToCity(cityDto)
	if err != nil {
		return nil, err
	}

	id, err := services.GenerateUuid()
	if err != nil {
		return nil, err
	}

	registrationDate := time.Now()
	pvz := &models.Pvz{Id: id, RegistrationDate: registrationDate, City: city}

	_, err = s.driver.CreatePvz(ctx, pvz)
	if err != nil {
		return nil, err
	}

	idDto, err := convertPgUuidToOpenAPI(id)
	if err != nil {
		return nil, err
	}

	return &generated.PVZ{Id: &idDto, City: cityDto, RegistrationDate: &registrationDate}, nil
}

func (s *PvzService) DeleteLastProduct(ctx context.Context, pvzId string) error {
	return nil
}

func convertPvzCityToCity(pvzCity generated.PVZCity) (models.City, error) {
	switch pvzCity {
	case generated.Москва:
		return models.Moscow, nil
	case generated.СанктПетербург:
		return models.SPb, nil
	case generated.Казань:
		return models.Kazan, nil
	default:
		return 0, fmt.Errorf("unknown city: %s", pvzCity)
	}
}

func convertCityToPvzCity(city models.City) (generated.PVZCity, error) {
	switch city {
	case models.Moscow:
		return generated.Москва, nil
	case models.SPb:
		return generated.СанктПетербург, nil
	case models.Kazan:
		return generated.Казань, nil
	default:
		return "", fmt.Errorf("unknown city: %d", city)
	}
}

func convertPgUuidToOpenAPI(pgUuid pgtype.UUID) (openapi_types.UUID, error) {
	if pgUuid.Status != pgtype.Present {
		return openapi_types.UUID{}, fmt.Errorf("invalid UUID: not present")
	}

	stdUuid, err := uuid.FromBytes(pgUuid.Bytes[:])
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("failed to convert UUID bytes: %w", err)
	}

	return stdUuid, nil
}
