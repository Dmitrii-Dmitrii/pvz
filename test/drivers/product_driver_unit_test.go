package drivers

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers/product_driver"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/product_model"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateProduct(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := product_driver.NewProductDriver(mockAdapter)

	pvzID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	product := &product_model.Product{
		Id:          pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
		AddingTime:  time.Now(),
		ProductType: "test",
	}
	receptionID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}

	mockTx := new(MockTx)
	mockRow := new(MockRow)

	mockAdapter.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("QueryRow", ctx, drivers.QueryGetReceptionInProgressId, []interface{}{pvzID}).
		Return(mockRow)
	mockRow.On("Scan", mock.AnythingOfType("*pgtype.UUID")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = receptionID
		}).
		Return(nil)
	mockTx.On("Exec", ctx, drivers.QueryCreateProduct, []interface{}{
		product.Id, product.AddingTime, product.ProductType, receptionID,
	}).Return(pgconn.CommandTag{}, nil)
	mockTx.On("Commit", ctx).Return(nil)

	result, err := driver.CreateProduct(ctx, product, pvzID)

	require.NoError(t, err)
	assert.Equal(t, receptionID, *result)
	mockAdapter.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestDeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	mockAdapter := new(MockAdapter)
	driver := product_driver.NewProductDriver(mockAdapter)

	pvzID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	receptionID := pgtype.UUID{Bytes: [16]byte{3}, Valid: true}

	mockTx := new(MockTx)
	mockRow := new(MockRow)

	mockAdapter.On("Begin", ctx).Return(mockTx, nil)
	mockTx.On("Rollback", ctx).Return(nil)
	mockTx.On("QueryRow", ctx, drivers.QueryGetReceptionInProgressId, []interface{}{pvzID}).
		Return(mockRow)
	mockRow.On("Scan", mock.AnythingOfType("*pgtype.UUID")).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*pgtype.UUID) = receptionID
		}).
		Return(nil)
	mockTx.On("Exec", ctx, drivers.QueryDeleteLastProduct, []interface{}{receptionID}).
		Return(pgconn.CommandTag{}, nil)
	mockTx.On("Commit", ctx).Return(nil)

	err := driver.DeleteLastProduct(ctx, pvzID)

	require.NoError(t, err)
	mockAdapter.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}
