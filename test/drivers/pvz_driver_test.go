package drivers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"pvz/internal/drivers/pvz_driver"
	"pvz/internal/models/custom_errors"
	"pvz/internal/models/pvz_model"
)

func TestCreatePvz(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := pvz_driver.NewPvzDriver(pool)
	ctx := context.Background()

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}
	registrationDate := time.Now().UTC()
	city := pvz_model.Moscow

	pvz := &pvz_model.Pvz{
		Id:               id,
		RegistrationDate: registrationDate,
		City:             city,
	}

	err := driver.CreatePvz(ctx, pvz)
	require.NoError(t, err)

	var dbCity string
	var dbDate time.Time
	err = pool.QueryRow(ctx, queryGetPvz, id).Scan(&dbDate, &dbCity)

	require.NoError(t, err)
	assert.Equal(t, registrationDate, dbDate)
	assert.Equal(t, string(city), dbCity)
}

func TestGetPvzById(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := pvz_driver.NewPvzDriver(pool)
	ctx := context.Background()

	pvzIds, _, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)

	t.Run("Existing pvz", func(t *testing.T) {
		result, err := driver.GetPvzById(ctx, pvzIds[0])

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, pvzIds[0], result.Id)
		assert.NotZero(t, result.RegistrationDate)
		assert.True(t, result.City == pvz_model.Moscow || result.City == pvz_model.SPb)
	})

	t.Run("Non-existent pvz", func(t *testing.T) {
		nonExistentId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		result, err := driver.GetPvzById(ctx, nonExistentId)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrPvzNotFound, err)
		assert.Nil(t, result)
	})
}

func TestGetPvz(t *testing.T) {
	pool, cleanup := setupPostgresContainer(t)
	defer cleanup()

	driver := pvz_driver.NewPvzDriver(pool)
	ctx := context.Background()

	_, _, _, err := createTestData(ctx, pool)
	require.NoError(t, err)

	t.Run("All pvz", func(t *testing.T) {
		limit := uint32(10)
		offset := uint32(0)

		results, err := driver.GetPvz(ctx, limit, offset, nil, nil)

		require.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.LessOrEqual(t, len(results), 2)

		if len(results) > 0 {
			pvzData, ok := results[0]["pvz"]
			assert.True(t, ok, "Result should contain 'pvz' key")
			assert.NotNil(t, pvzData)

			receptions, ok := results[0]["receptions"]
			assert.True(t, ok, "Result should contain 'receptions' key")
			assert.NotNil(t, receptions)
		}
	})

	t.Run("Pvz with time filter", func(t *testing.T) {
		limit := uint32(10)
		offset := uint32(0)
		startTime := time.Now().Add(-36 * time.Hour)
		endTime := time.Now()

		results, err := driver.GetPvz(ctx, limit, offset, &startTime, &endTime)

		require.NoError(t, err)
		if len(results) > 0 {
			pvzData, ok := results[0]["pvz"]
			assert.True(t, ok, "Result should contain 'pvz' key")
			assert.NotNil(t, pvzData)
		}
	})

	t.Run("Pagination", func(t *testing.T) {
		limit := uint32(1)
		firstOffset := uint32(0)
		secondOffset := uint32(1)

		firstPage, err := driver.GetPvz(ctx, limit, firstOffset, nil, nil)
		require.NoError(t, err)

		if len(firstPage) == 0 {
			t.Skip("No data returned for pagination test")
		}

		secondPage, err := driver.GetPvz(ctx, limit, secondOffset, nil, nil)
		require.NoError(t, err)

		if len(firstPage) > 0 && len(secondPage) > 0 {
			firstPvzMap, ok1 := firstPage[0]["pvz"].(map[string]interface{})
			secondPvzMap, ok2 := secondPage[0]["pvz"].(map[string]interface{})

			if ok1 && ok2 && firstPvzMap["id"] != nil && secondPvzMap["id"] != nil {
				assert.NotEqual(t, firstPvzMap["id"], secondPvzMap["id"],
					"Different pages should return different pvz")
			}
		}
	})
}
