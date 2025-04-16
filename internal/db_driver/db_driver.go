package db_driver

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pvz/internal/models"
	"time"
)

type DBDriver struct {
	rwdb *pgxpool.Pool
}

func NewDBDriver(rwdb *pgxpool.Pool) *DBDriver {
	return &DBDriver{rwdb: rwdb}
}

func (d *DBDriver) CreatePvz(ctx context.Context, pvz *models.Pvz) (*models.Pvz, error) {
	err := d.rwdb.QueryRow(ctx, queryCreatePvz, pvz.Id, pvz.RegisterDate, pvz.City)
	if err != nil {
		return nil, fmt.Errorf("failed to create pvz: %w", err)
	}
	return pvz, nil
}

func (d *DBDriver) CreateReception(ctx context.Context, reception *models.Reception) error {
	err := d.rwdb.QueryRow(ctx, queryCreateReception, reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status)
	if err != nil {
		return fmt.Errorf("failed to create reception: %w", err)
	}
	return nil
}

func (d *DBDriver) GetReception(ctx context.Context, id pgtype.UUID) (*models.Reception, error) {
	var receptionTime time.Time
	var pvzId pgtype.UUID
	var status models.ReceptionStatus
	err := d.rwdb.QueryRow(ctx, queryGetReception, id).Scan(&receptionTime, &pvzId, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to get reception: %w", err)
	}
	reception := models.NewReception(id, pvzId, receptionTime, status)
	return reception, nil
}

func (d *DBDriver) CreateProducts(ctx context.Context, products []models.Product, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var receptionId pgtype.UUID
	err = tx.QueryRow(ctx, queryGetReceptionInProgressId, pvzId).Scan(&receptionId)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no open reception")
	}
	if err != nil {
		return fmt.Errorf("failed to get reception in progress id: %w", err)
	}

	batch := &pgx.Batch{}
	for _, product := range products {
		batch.Queue(queryCreateProduct, product.Id, product.AddingTime, product.ProductType, receptionId)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for range products {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failde to commit transaction: %w", err)
	}

	return nil
}
