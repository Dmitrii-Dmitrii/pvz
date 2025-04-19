package drivers

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz/internal/drivers/reception_driver"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/reception_model"
	"testing"
	"time"
)

func TestCreateReception(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := reception_driver.NewReceptionDriver(pool)
	ctx := context.Background()

	pvzIds, _, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}
	receptionTime := time.Now().UTC()
	status := reception_model.InProgress

	reception := &reception_model.Reception{
		Id:            id,
		ReceptionTime: receptionTime,
		PvzId:         pvzIds[0],
		Status:        status,
	}

	err = driver.CreateReception(ctx, reception)
	require.NoError(t, err)

	var dbTime time.Time
	var dbPvzId pgtype.UUID
	var dbStatus string
	err = pool.QueryRow(ctx, queryGetReception, id).Scan(&dbTime, &dbPvzId, &dbStatus)

	require.NoError(t, err)
	assert.Equal(t, receptionTime, dbTime)
	assert.Equal(t, pvzIds[0], dbPvzId)
	assert.Equal(t, string(status), dbStatus)
}

func TestCloseReception(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := reception_driver.NewReceptionDriver(pool)
	ctx := context.Background()

	pvzIds, receptionIds, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)
	require.NotEmpty(t, receptionIds)

	t.Run("Existing reception in progress", func(t *testing.T) {
		result, err := driver.CloseReception(ctx, pvzIds[0])

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, receptionIds[0], result.Id)
		assert.NotZero(t, result.ReceptionTime)
		assert.Equal(t, pvzIds[0], result.PvzId)
		assert.Equal(t, reception_model.Close, result.Status)
	})

	t.Run("Existing reception closed", func(t *testing.T) {
		result, err := driver.CloseReception(ctx, pvzIds[1])
		require.NoError(t, err)

		result, err = driver.CloseReception(ctx, pvzIds[1])
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
		assert.Nil(t, result)
	})

	t.Run("Not existing reception", func(t *testing.T) {
		nonExistentId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		result, err := driver.CloseReception(ctx, nonExistentId)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
		assert.Nil(t, result)
	})
}

func TestGetLastReceptionStatus(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := reception_driver.NewReceptionDriver(pool)
	ctx := context.Background()

	pvzIds, receptionIds, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)
	require.NotEmpty(t, receptionIds)

	t.Run("Reception with existing pvzId", func(t *testing.T) {
		result, err := driver.GetLastReceptionStatus(ctx, pvzIds[0])

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, *result == reception_model.InProgress || *result == reception_model.Close)
	})

	t.Run("Reception with non-existing pvzId", func(t *testing.T) {
		nonExistentId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		result, err := driver.GetLastReceptionStatus(ctx, nonExistentId)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoReception, err)
		assert.Nil(t, result)
	})
}
