package drivers

import (
	"context"
	"errors"
	"fmt"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/pvz_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/product_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/pvz_model"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreatePvz(t *testing.T) {
	ctx := context.Background()

	t.Run("Create pvz", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		pvz := &pvz_model.Pvz{
			Id:               id,
			RegistrationDate: time.Now(),
			City:             pvz_model.Moscow,
		}

		params := []interface{}{pvz.Id, pvz.RegistrationDate, pvz.City}
		mockAdapter.On("Exec", ctx, drivers.QueryCreatePvz, params).Return(pgconn.CommandTag{}, nil)

		err := driver.CreatePvz(ctx, pvz)

		assert.NoError(t, err)
		mockAdapter.AssertExpectations(t)
	})

	t.Run("Create pvz with error", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		pvz := &pvz_model.Pvz{
			Id:               id,
			RegistrationDate: time.Now(),
			City:             pvz_model.Moscow,
		}

		params := []interface{}{pvz.Id, pvz.RegistrationDate, pvz.City}
		mockAdapter.On("Exec", ctx, drivers.QueryCreatePvz, params).
			Return(pgconn.CommandTag{}, errors.New("db error"))

		err := driver.CreatePvz(ctx, pvz)

		assert.Equal(t, custom_errors.ErrCreatePvz, err)
		mockAdapter.AssertExpectations(t)
	})
}

func TestGetPvzById(t *testing.T) {
	ctx := context.Background()

	t.Run("Get pvz by id", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRow := new(MockRow)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		registrationDate := time.Now()
		city := pvz_model.Moscow

		params := []interface{}{id}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetPvzById, params).Return(mockRow)
		mockRow.On("Scan", mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Run(func(args mock.Arguments) {
				*(args.Get(0).(*time.Time)) = registrationDate
				*(args.Get(1).(*pvz_model.City)) = city
			}).
			Return(nil)

		pvz, err := driver.GetPvzById(ctx, id)

		require.NoError(t, err)
		assert.Equal(t, id, pvz.Id)
		assert.Equal(t, registrationDate, pvz.RegistrationDate)
		assert.Equal(t, city, pvz.City)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Get pvz by non-existing id", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRow := new(MockRow)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		params := []interface{}{id}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetPvzById, params).Return(mockRow)
		mockRow.On("Scan", mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Return(pgx.ErrNoRows)

		pvz, err := driver.GetPvzById(ctx, id)

		assert.Nil(t, pvz)
		assert.Equal(t, custom_errors.ErrPvzNotFound, err)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Get pvz by id with error", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRow := new(MockRow)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id := pgtype.UUID{Bytes: uuid.New(), Valid: true}

		params := []interface{}{id}
		mockAdapter.On("QueryRow", ctx, drivers.QueryGetPvzById, params).Return(mockRow)
		mockRow.On("Scan", mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Return(errors.New("db error"))

		pvz, err := driver.GetPvzById(ctx, id)

		assert.Nil(t, pvz)
		assert.Equal(t, custom_errors.ErrGetPvz, err)
		mockAdapter.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestGetAllPvz(t *testing.T) {
	ctx := context.Background()

	t.Run("Get all pvz", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRows := new(MockRows)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		id1 := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		id2 := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		date1 := time.Now()
		date2 := time.Now().Add(-24 * time.Hour)
		city1 := pvz_model.Moscow
		city2 := pvz_model.SPb

		mockAdapter.On("Query", ctx, drivers.QueryGetAllPvz).Return(mockRows, nil)
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Scan", mock.AnythingOfType("*pgtype.UUID"), mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Run(func(args mock.Arguments) {
				*(args.Get(0).(*pgtype.UUID)) = id1
				*(args.Get(1).(*time.Time)) = date1
				*(args.Get(2).(*pvz_model.City)) = city1
			}).
			Return(nil).Once()
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Scan", mock.AnythingOfType("*pgtype.UUID"), mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Run(func(args mock.Arguments) {
				*(args.Get(0).(*pgtype.UUID)) = id2
				*(args.Get(1).(*time.Time)) = date2
				*(args.Get(2).(*pvz_model.City)) = city2
			}).
			Return(nil).Once()
		mockRows.On("Next").Return(false)
		mockRows.On("Close").Return()

		pvzList, err := driver.GetAllPvz(ctx)

		require.NoError(t, err)
		assert.Len(t, pvzList, 2)
		assert.Equal(t, id1, pvzList[0].Id)
		assert.Equal(t, date1, pvzList[0].RegistrationDate)
		assert.Equal(t, city1, pvzList[0].City)
		assert.Equal(t, id2, pvzList[1].Id)
		assert.Equal(t, date2, pvzList[1].RegistrationDate)
		assert.Equal(t, city2, pvzList[1].City)
		mockAdapter.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})

	t.Run("Get all pvz with query error", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		mockAdapter.On("Query", ctx, drivers.QueryGetAllPvz).Return((*MockRows)(nil), errors.New("db error"))

		pvzList, err := driver.GetAllPvz(ctx)

		assert.Nil(t, pvzList)
		assert.Equal(t, custom_errors.ErrGetPvz, err)
		mockAdapter.AssertExpectations(t)
	})

	t.Run("Get all pvz with scan error", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRows := new(MockRows)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		mockAdapter.On("Query", ctx, drivers.QueryGetAllPvz).Return(mockRows, nil)
		mockRows.On("Next").Return(true)
		mockRows.On("Scan", mock.AnythingOfType("*pgtype.UUID"), mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*pvz_model.City")).
			Return(errors.New("scan error"))
		mockRows.On("Close").Return()

		pvzList, err := driver.GetAllPvz(ctx)

		assert.Nil(t, pvzList)
		assert.Equal(t, custom_errors.ErrScanRow, err)
		mockAdapter.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})
}

func TestGetPvzFullInfo(t *testing.T) {
	ctx := context.Background()

	t.Run("Get pvz full info", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		mockRows := new(MockRows)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		limit := uint32(10)
		offset := uint32(0)
		startTime := time.Now().Add(-7 * 24 * time.Hour)
		endTime := time.Now()

		query, params := getTestQueryGetPvz(limit, offset, &startTime, &endTime)

		mockAdapter.On("Query", ctx, query, params).Return(mockRows, nil)

		setupMockRowsForGetPvzFullInfo(mockRows)

		mockRows.On("Close").Return()

		result, err := driver.GetPvzFullInfo(ctx, limit, offset, &startTime, &endTime)

		require.NoError(t, err)
		assert.NotEmpty(t, result)
		mockAdapter.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})

	t.Run("Get pvz full info with query error", func(t *testing.T) {
		mockAdapter := new(MockAdapter)
		driver := pvz_driver.NewPvzDriver(mockAdapter)

		limit := uint32(10)
		offset := uint32(0)

		query, params := getTestQueryGetPvz(limit, offset, nil, nil)

		mockAdapter.On("Query", ctx, query, params).Return((*MockRows)(nil), errors.New("db error"))

		result, err := driver.GetPvzFullInfo(ctx, limit, offset, nil, nil)

		assert.Nil(t, result)
		assert.Equal(t, custom_errors.ErrGetPvz, err)
		mockAdapter.AssertExpectations(t)
	})
}

func getTestQueryGetPvz(limit, offset uint32, startInterval, endInterval *time.Time) (string, []interface{}) {
	query := drivers.QueryGetPvz
	var params []interface{}
	paramCnt := 0

	if startInterval != nil || endInterval != nil {
		query += " WHERE "

		if startInterval != nil {
			paramCnt++
			query += fmt.Sprintf("r.reception_time >= $%d", paramCnt)
			params = append(params, *startInterval)
		}

		if startInterval != nil && endInterval != nil {
			query += " AND "
		}

		if endInterval != nil {
			paramCnt++
			query += fmt.Sprintf("r.reception_time <= $%d", paramCnt)
			params = append(params, *endInterval)
		}
	}

	paramCnt++
	query += fmt.Sprintf(" ORDER BY p.id, r.reception_time DESC LIMIT $%d", paramCnt)
	params = append(params, limit)
	paramCnt++
	query += fmt.Sprintf(" OFFSET $%d", paramCnt)
	params = append(params, offset)

	return query, params
}

func setupMockRowsForGetPvzFullInfo(mockRows *MockRows) {
	pvzId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	receptionId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
	productId := pgtype.UUID{Bytes: uuid.New(), Valid: true}

	registrationDate := time.Now().Add(-30 * 24 * time.Hour)
	receptionTime := time.Now().Add(-7 * 24 * time.Hour)
	addingTime := time.Now().Add(-6 * 24 * time.Hour)

	pvzCity := pvz_model.Moscow
	receptionStatus := reception_model.Close
	productType := product_model.Electronics

	mockRows.On("Next").Return(true).Once()
	mockRows.On("Scan",
		mock.AnythingOfType("*pgtype.UUID"),
		mock.AnythingOfType("**time.Time"),
		mock.AnythingOfType("*pvz_model.City"),
		mock.AnythingOfType("*pgtype.UUID"),
		mock.AnythingOfType("**time.Time"),
		mock.AnythingOfType("**reception_model.ReceptionStatus"),
		mock.AnythingOfType("*pgtype.UUID"),
		mock.AnythingOfType("**time.Time"),
		mock.AnythingOfType("**product_model.ProductType"),
	).Run(func(args mock.Arguments) {
		*(args.Get(0).(*pgtype.UUID)) = pvzId
		*(args.Get(1).(**time.Time)) = &registrationDate
		*(args.Get(2).(*pvz_model.City)) = pvzCity
		*(args.Get(3).(*pgtype.UUID)) = receptionId
		*(args.Get(4).(**time.Time)) = &receptionTime
		*(args.Get(5).(**reception_model.ReceptionStatus)) = &receptionStatus
		*(args.Get(6).(*pgtype.UUID)) = productId
		*(args.Get(7).(**time.Time)) = &addingTime
		*(args.Get(8).(**product_model.ProductType)) = &productType
	}).Return(nil).Once()

	mockRows.On("Next").Return(false)
}
