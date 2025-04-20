package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"pvz/internal/generated"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/reception_model"
	"pvz/internal/services/reception_service"
	"testing"
	"time"
)

type MockReceptionDriver struct {
	mock.Mock
}

func (m *MockReceptionDriver) CreateReception(ctx context.Context, reception *reception_model.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *MockReceptionDriver) CloseReception(ctx context.Context, pvzId pgtype.UUID) (*reception_model.Reception, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception_model.Reception), args.Error(1)
}

func (m *MockReceptionDriver) GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error) {
	args := m.Called(ctx, pvzId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*reception_model.ReceptionStatus), args.Error(1)
}

func TestCreateReception(t *testing.T) {
	ctx := context.Background()

	t.Run("Create reception with previous receptions close", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		status := reception_model.Close
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("CreateReception", ctx, mock.AnythingOfType("*reception_model.Reception")).Return(nil)

		result, err := service.CreateReception(ctx, pvzIdDto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, pvzIdDto, result.PvzId)
		assert.Equal(t, generated.InProgress, result.Status)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Create reception without previous receptions", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrNoReception)
		mockDriver.On("CreateReception", ctx, mock.AnythingOfType("*reception_model.Reception")).Return(nil)

		result, err := service.CreateReception(ctx, pvzIdDto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, pvzIdDto, result.PvzId)
		assert.Equal(t, generated.InProgress, result.Status)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Create reception with previous receptions in progress", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		status := reception_model.InProgress
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		result, err := service.CreateReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrInProgressReception, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreateReception")
	})

	t.Run("Create reception with GetLastReceptionStatus error", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		expectedError := errors.New("database error")
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, expectedError)

		result, err := service.CreateReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CreateReception")
	})

	t.Run("Create reception with error", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		status := reception_model.Close
		expectedError := errors.New("database error")

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("CreateReception", ctx, mock.AnythingOfType("*reception_model.Reception")).Return(expectedError)

		result, err := service.CreateReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
	})
}

func TestCloseReception(t *testing.T) {
	ctx := context.Background()

	t.Run("Close reception with previous receptions close", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()
		pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		receptionId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		status := reception_model.InProgress
		closedStatus := reception_model.Close
		receptionTime := time.Now()

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		closedReception := &reception_model.Reception{
			Id:            receptionId,
			ReceptionTime: receptionTime,
			PvzId:         pvzId,
			Status:        closedStatus,
		}
		mockDriver.On("CloseReception", ctx, mock.AnythingOfType("pgtype.UUID")).Return(closedReception, nil)

		result, err := service.CloseReception(ctx, pvzIdDto)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, pvzIdDto, result.PvzId)
		assert.Equal(t, generated.Close, result.Status)
		assert.Equal(t, receptionTime, result.DateTime)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Close reception without previous receptions", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrNoReception)

		result, err := service.CloseReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoReception, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CloseReception")
	})

	t.Run("Close reception with previous receptions in progress", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		status := reception_model.Close
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		result, err := service.CloseReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CloseReception")
	})

	t.Run("Close reception with GetLastReceptionStatus error", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		expectedError := errors.New("database error")
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, expectedError)

		result, err := service.CloseReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
		mockDriver.AssertNotCalled(t, "CloseReception")
	})

	t.Run("Close reception with error", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzIdDto := uuid.New()

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(pvzIdDto)

		status := reception_model.InProgress
		expectedError := errors.New("database error")

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)
		mockDriver.On("CloseReception", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, expectedError)

		result, err := service.CloseReception(ctx, pvzIdDto)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
	})
}

func TestGetLastReceptionStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("Get last reception status close", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(uuid.New())

		status := reception_model.Close
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		result, err := service.GetLastReceptionStatus(ctx, pvzId)

		assert.NoError(t, err)
		assert.Equal(t, status, *result)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Get last reception status in progress", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(uuid.New())

		status := reception_model.InProgress
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(&status, nil)

		result, err := service.GetLastReceptionStatus(ctx, pvzId)

		assert.NoError(t, err)
		assert.Equal(t, status, *result)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Get last reception status without existing Receptions", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(uuid.New())

		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, custom_errors.ErrNoReception)

		result, err := service.GetLastReceptionStatus(ctx, pvzId)

		require.Error(t, err)
		assert.True(t, errors.Is(err, custom_errors.ErrNoReception))
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
	})

	t.Run("Get last reception status with error", func(t *testing.T) {
		mockDriver := new(MockReceptionDriver)
		service := reception_service.NewReceptionService(mockDriver)

		pvzId := pgtype.UUID{}
		_ = pvzId.Scan(uuid.New())

		expectedError := errors.New("database error")
		mockDriver.On("GetLastReceptionStatus", ctx, mock.AnythingOfType("pgtype.UUID")).Return(nil, expectedError)

		result, err := service.GetLastReceptionStatus(ctx, pvzId)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Nil(t, result)
		mockDriver.AssertExpectations(t)
	})
}
