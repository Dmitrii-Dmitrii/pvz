package products

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"pvz/internal/drivers"
	"pvz/internal/models"
)

type ProductDriver struct {
	rwdb *pgxpool.Pool
}

func NewProductDriver(rwdb *pgxpool.Pool) *ProductDriver {
	return &ProductDriver{rwdb: rwdb}
}

func (d *ProductDriver) CreateProduct(ctx context.Context, product *models.Product, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := drivers.GetReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, drivers.QueryCreateProduct, product.Id, product.ProductType, product.AddingTime, receptionId)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *ProductDriver) DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := drivers.GetReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, drivers.QueryDeleteLastProduct, receptionId)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
