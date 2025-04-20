package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/pvz_model"
	"pvz/internal/services/pvz_service"
	"testing"
	"time"
)

type MockPvzDriver struct {
	mock.Mock
}

func (m *MockPvzDriver) GetPvzById(ctx context.Context, id pgtype.UUID) (*pvz_model.Pvz, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pvz_model.Pvz), args.Error(1)
}

func (m *MockPvzDriver) CreatePvz(ctx context.Context, pvz *pvz_model.Pvz) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *MockPvzDriver) GetPvz(ctx context.Context, limit uint32, offset uint32, startDate *time.Time, endDate *time.Time) ([]map[string]interface{}, error) {
	args := m.Called(ctx, limit, offset, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func TestCreatePvz(t *testing.T) {
	ctx := context.Background()

	t.Run("Create pvz with new id", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)
		city := generated.Москва
		pvzDto := generated.PVZ{
			City: city,
		}

		mockDriver.On("GetPvzById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrPvzNotFound)
		mockDriver.On("CreatePvz", ctx, mock.AnythingOfType("*pvz_model.Pvz")).Return(nil)

		result, err := service.CreatePvz(ctx, pvzDto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, city, result.City)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Create pvz with provided id", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)
		id := uuid.New()
		city := generated.СанктПетербург
		pvzDto := generated.PVZ{
			Id:   &id,
			City: city,
		}

		mockDriver.On("GetPvzById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrPvzNotFound)
		mockDriver.On("CreatePvz", ctx, mock.AnythingOfType("*pvz_model.Pvz")).Return(nil)

		result, err := service.CreatePvz(ctx, pvzDto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, id, *result.Id)
		assert.Equal(t, city, result.City)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Create already exists pvz", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)
		id := uuid.New()
		city := generated.Казань
		pvzDto := generated.PVZ{
			Id:   &id,
			City: city,
		}

		existingPvz := &pvz_model.Pvz{
			City: pvz_model.Kazan,
		}

		mockDriver.On("GetPvzById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(existingPvz, nil)

		result, err := service.CreatePvz(ctx, pvzDto)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrPvzExists, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreatePvz")
	})

	t.Run("Create pvz with invalid city", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		id := uuid.New()
		invalidCity := generated.PVZCity("Неизвестный_город")
		pvzDto := generated.PVZ{
			Id:   &id,
			City: invalidCity,
		}

		mockDriver.On("GetPvzById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrPvzNotFound)

		result, err := service.CreatePvz(ctx, pvzDto)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrPvzCity, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreatePvz")
	})

	t.Run("Create pvz with driver error", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		city := generated.Москва
		pvzDto := generated.PVZ{
			City: city,
		}

		expectedError := errors.New("database connection error")

		mockDriver.On("GetPvzById", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, expectedError)

		result, err := service.CreatePvz(ctx, pvzDto)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreatePvz")
	})
}

func TestGetPvz(t *testing.T) {
	ctx := context.Background()

	t.Run("Get pvz with default params", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		params := generated.GetPvzParams{}
		expectedPvzList := []map[string]interface{}{
			{"id": "123", "city": "Москва"},
			{"id": "456", "city": "Санкт-Петербург"},
		}

		mockDriver.On("GetPvz", ctx, uint32(10), uint32(0), (*time.Time)(nil), (*time.Time)(nil)).Return(expectedPvzList, nil)

		result, err := service.GetPvz(ctx, params)

		assert.NoError(t, err)
		assert.Equal(t, expectedPvzList, result)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Get pvz with custom params", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		limit := 20
		page := 2
		now := time.Now()
		startDate := now.AddDate(0, -1, 0)
		endDate := now

		params := generated.GetPvzParams{
			Limit:     &limit,
			Page:      &page,
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		expectedPvzList := []map[string]interface{}{
			{"id": "789", "city": "Казань"},
		}

		mockDriver.On("GetPvz", ctx, uint32(20), uint32(20), &startDate, &endDate).Return(expectedPvzList, nil)

		result, err := service.GetPvz(ctx, params)

		assert.NoError(t, err)
		assert.Equal(t, expectedPvzList, result)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Get pvz with invalid date range", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		endDate := time.Now().AddDate(0, -2, 0)
		startDate := time.Now().AddDate(0, -1, 0)

		params := generated.GetPvzParams{
			StartDate: &startDate,
			EndDate:   &endDate,
		}

		result, err := service.GetPvz(ctx, params)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrDateRange, err)
		assert.Nil(t, result)
		mockDriver.AssertNotCalled(t, "GetPvz")
	})

	t.Run("Get pvz with invalid limit", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		t.Run("Get pvz with too large limit", func(t *testing.T) {
			tooLargeLimit := 50
			params := generated.GetPvzParams{
				Limit: &tooLargeLimit,
			}

			result, err := service.GetPvz(ctx, params)

			assert.Error(t, err)
			assert.Equal(t, custom_errors.ErrLimitValue, err)
			assert.Nil(t, result)
			mockDriver.AssertNotCalled(t, "GetPvz")
		})

		t.Run("Get pvz with too small limit", func(t *testing.T) {
			mockDriver := new(MockPvzDriver)
			service := pvz_service.NewPvzService(mockDriver)

			tooSmallLimit := 0
			params := generated.GetPvzParams{
				Limit: &tooSmallLimit,
			}

			result, err := service.GetPvz(ctx, params)

			assert.Error(t, err)
			assert.Equal(t, custom_errors.ErrLimitValue, err)
			assert.Nil(t, result)
			mockDriver.AssertNotCalled(t, "GetPvz")
		})
	})

	t.Run("Get pvz with invalid page", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		invalidPage := 0
		params := generated.GetPvzParams{
			Page: &invalidPage,
		}

		result, err := service.GetPvz(ctx, params)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrPageValue, err)
		assert.Nil(t, result)
		mockDriver.AssertNotCalled(t, "GetPvz")
	})

	t.Run("Get pvz with driver error", func(t *testing.T) {
		mockDriver := new(MockPvzDriver)
		service := pvz_service.NewPvzService(mockDriver)

		params := generated.GetPvzParams{}
		expectedError := errors.New("database connection error")

		mockDriver.On("GetPvz", ctx, uint32(10), uint32(0), (*time.Time)(nil), (*time.Time)(nil)).Return(nil, expectedError)

		result, err := service.GetPvz(ctx, params)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
	})
}
