package drivers

import (
	"context"
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/product_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/reception_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/product_model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateProductIntegration(t *testing.T) {
	pool, cleanup := SetupPostgresContainer(t)
	defer cleanup()

	driver := product_driver.NewProductDriver(pool)
	ctx := context.Background()

	pvzIds, receptionIds, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)
	require.NotEmpty(t, receptionIds)

	idBytes := uuid.New()
	id := pgtype.UUID{Bytes: idBytes, Valid: true}
	addingTime := time.Now().UTC()
	productType := product_model.Electronics

	product := &product_model.Product{
		Id:          id,
		AddingTime:  addingTime,
		ProductType: productType,
	}

	result, err := driver.CreateProduct(ctx, product, pvzIds[0])
	require.NoError(t, err)
	assert.Equal(t, receptionIds[0], *result)

	var dbTime time.Time
	var dbReceptionId pgtype.UUID
	var dbProductType string
	err = pool.QueryRow(ctx, queryGetProduct, id).Scan(&dbTime, &dbProductType, &dbReceptionId)

	require.NoError(t, err)
	assert.Equal(t, addingTime, dbTime)
	assert.Equal(t, receptionIds[0], dbReceptionId)
	assert.Equal(t, string(productType), dbProductType)
}

func TestDeleteLastProductIntegration(t *testing.T) {
	pool, cleanup := SetupPostgresContainer(t)
	defer cleanup()

	productDriver := product_driver.NewProductDriver(pool)
	receptionDriver := reception_driver.NewReceptionDriver(pool)
	ctx := context.Background()

	pvzIds, receptionIds, _, err := createTestData(ctx, pool)
	require.NoError(t, err)
	require.NotEmpty(t, pvzIds)
	require.NotEmpty(t, receptionIds)

	t.Run("Product with existing pvzId in reception in progress", func(t *testing.T) {
		idBytes := uuid.New()
		id := pgtype.UUID{Bytes: idBytes, Valid: true}
		addingTime := time.Now().UTC()
		productType := product_model.Electronics

		product := &product_model.Product{
			Id:          id,
			AddingTime:  addingTime,
			ProductType: productType,
		}

		_, err := productDriver.CreateProduct(ctx, product, pvzIds[0])

		err = productDriver.DeleteLastProduct(ctx, pvzIds[0])
		require.NoError(t, err)

		var dbTime time.Time
		var dbReceptionId pgtype.UUID
		var dbProductType string
		err = pool.QueryRow(ctx, queryGetProduct, id).Scan(&dbTime, &dbProductType, &dbReceptionId)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})

	t.Run("Product with existing pvzId in reception closed", func(t *testing.T) {
		_, err := receptionDriver.CloseReception(ctx, pvzIds[1])
		require.NoError(t, err)

		err = productDriver.DeleteLastProduct(ctx, pvzIds[1])
		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
	})

	t.Run("Product with non-existing pvzId", func(t *testing.T) {
		nonExistentId := pgtype.UUID{Bytes: uuid.New(), Valid: true}
		err = productDriver.DeleteLastProduct(ctx, nonExistentId)

		assert.Error(t, err)
		assert.Equal(t, custom_errors.ErrNoOpenReception, err)
	})
}
