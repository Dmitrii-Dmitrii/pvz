package drivers

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"testing"
	"time"

	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/reception_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateReception(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := reception_driver.NewReceptionDriver(mockAdapter)

	t.Run("Create reception", func(t *testing.T) {
		reception := &reception_model.Reception{
			Id:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
			ReceptionTime: time.Now(),
			PvzId:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Status:        reception_model.InProgress,
		}

		params := []interface{}{reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status}
		mockAdapter.On("Exec", ctx, drivers.QueryCreateReception, params).Return(pgconn.CommandTag{}, nil).Once()

		err := driver.CreateReception(ctx, reception)

		require.NoError(t, err)
		mockAdapter.AssertExpectations(t)
	})

	t.Run("Create reception with error", func(t *testing.T) {
		reception := &reception_model.Reception{
			Id:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
			ReceptionTime: time.Now(),
			PvzId:         pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Status:        reception_model.InProgress,
		}

		params := []interface{}{reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status}
		mockAdapter.On("Exec", ctx, drivers.QueryCreateReception, params).Return(pgconn.CommandTag{}, errors.New("database error")).Once()

		err := driver.CreateReception(ctx, reception)

		assert.Equal(t, custom_errors.ErrCreateReception, err)
		mockAdapter.AssertExpectations(t)
	})
}

func TestGetLastReceptionStatus(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := reception_driver.NewReceptionDriver(mockAdapter)

	t.Run("Get last reception status", func(t *testing.T) {
		pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		expectedStatus := reception_model.InProgress

		mockRow := new(MockRow)
		params := []interface{}{pvzId}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetLastReceptionStatus, params).Return(mockRow).Once()

		mockRow.On("Scan", mock.AnythingOfType("*reception_model.ReceptionStatus")).
			Run(func(args mock.Arguments) {
				*(args.Get(0).(*reception_model.ReceptionStatus)) = expectedStatus
			}).Return(nil).Once()

		status, err := driver.GetLastReceptionStatus(ctx, pvzId)

		require.NoError(t, err)
		assert.Equal(t, expectedStatus, *status)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Get last reception status without rows", func(t *testing.T) {
		pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockRow := new(MockRow)
		params := []interface{}{pvzId}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetLastReceptionStatus, params).Return(mockRow).Once()

		mockRow.On("Scan", mock.AnythingOfType("*reception_model.ReceptionStatus")).
			Return(pgx.ErrNoRows).Once()

		status, err := driver.GetLastReceptionStatus(ctx, pvzId)

		assert.Nil(t, status)
		assert.Equal(t, custom_errors.ErrNoReception, err)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Get last reception status with error", func(t *testing.T) {
		pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		mockRow := new(MockRow)
		params := []interface{}{pvzId}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetLastReceptionStatus, params).Return(mockRow).Once()
		mockRow.On("Scan", mock.AnythingOfType("*reception_model.ReceptionStatus")).
			Return(errors.New("database error")).Once()

		status, err := driver.GetLastReceptionStatus(ctx, pvzId)

		assert.Nil(t, status)
		assert.Equal(t, custom_errors.ErrGetLastReceptionStatus, err)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestCloseReception(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := reception_driver.NewReceptionDriver(mockAdapter)

	pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	receptionId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	receptionTime := time.Now()
	status := reception_model.Close

	mockTx := new(MockTx)
	mockAdapter.On("Begin", ctx).Return(mockTx, nil)

	mockRowReceptionId := new(MockRow)
	params := []interface{}{pvzId}
	mockTx.On("QueryRow", ctx, drivers.QueryGetReceptionInProgressId, params).Return(mockRowReceptionId)
	mockRowReceptionId.On("Scan", mock.AnythingOfType("*pgtype.UUID")).
		Run(func(args mock.Arguments) {
			*(args.Get(0).(*pgtype.UUID)) = receptionId
		}).Return(nil)

	params = []interface{}{receptionId}
	mockTx.On("Exec", ctx, drivers.QueryCloseReception, params).Return(pgconn.CommandTag{}, nil)

	mockTx.On("Commit", ctx).Return(nil)
	mockTx.On("Rollback", ctx).Return(nil)

	mockRowReception := new(MockRow)
	mockAdapter.On("QueryRow", ctx, drivers.QueryGetReception, params).Return(mockRowReception)
	mockRowReception.On("Scan",
		mock.AnythingOfType("*time.Time"),
		mock.AnythingOfType("*pgtype.UUID"),
		mock.AnythingOfType("*reception_model.ReceptionStatus"),
	).Run(func(args mock.Arguments) {
		*(args.Get(0).(*time.Time)) = receptionTime
		*(args.Get(1).(*pgtype.UUID)) = pvzId
		*(args.Get(2).(*reception_model.ReceptionStatus)) = status
	}).Return(nil)

	reception, err := driver.CloseReception(ctx, pvzId)

	require.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, receptionId, reception.Id)
	assert.Equal(t, pvzId, reception.PvzId)
	assert.Equal(t, receptionTime, reception.ReceptionTime)
	assert.Equal(t, status, reception.Status)

	mockAdapter.AssertExpectations(t)
	mockTx.AssertExpectations(t)
	mockRowReceptionId.AssertExpectations(t)
	mockRowReception.AssertExpectations(t)
}
