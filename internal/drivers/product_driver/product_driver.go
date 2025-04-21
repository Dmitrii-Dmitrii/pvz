package product_driver

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/product_model"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type ProductDriver struct {
	rwdb *pgxpool.Pool
}

func NewProductDriver(rwdb *pgxpool.Pool) *ProductDriver {
	return &ProductDriver{rwdb: rwdb}
}

func (d *ProductDriver) CreateProduct(ctx context.Context, product *product_model.Product, pvzId pgtype.UUID) (*pgtype.UUID, error) {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrBeginTransaction.Message)
		return nil, custom_errors.ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	receptionId, err := drivers.GetReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, drivers.QueryCreateProduct, product.Id, product.AddingTime, product.ProductType, receptionId)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreateProduct.Message)
		return nil, custom_errors.ErrCreateProduct
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCommitTransaction.Message)
		return nil, custom_errors.ErrCommitTransaction
	}

	return &receptionId, nil
}

func (d *ProductDriver) DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrBeginTransaction.Message)
		return custom_errors.ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	receptionId, err := drivers.GetReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, drivers.QueryDeleteLastProduct, receptionId)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrDeleteProduct.Message)
		return custom_errors.ErrDeleteProduct
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCommitTransaction.Message)
		return custom_errors.ErrCommitTransaction
	}
	return nil
}
