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
	reception := &models.Reception{Id: id, PvzId: pvzId, ReceptionTime: receptionTime, Status: status}
	return reception, nil
}

func (d *DBDriver) CreateProducts(ctx context.Context, products []models.Product, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := d.getReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	batch := &pgx.Batch{}
	for _, product := range products {
		batch.Queue(queryCreateProduct, product.Id, product.AddingTime, product.ProductType, receptionId)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for range products {
		_, err = results.Exec()
		if err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *DBDriver) DeleteLastProduct(ctx context.Context, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := d.getReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, queryDeleteLastProduct, receptionId)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *DBDriver) CloseReception(ctx context.Context, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := d.getReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, queryCloseReception, receptionId)

	if err != nil {
		return fmt.Errorf("failed to close reception: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *DBDriver) GetPvz(ctx context.Context, limit, offset uint32) ([]models.Pvz, error) {
	rows, err := d.rwdb.Query(ctx, queryGetPvz, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pvz: %w", err)
	}
	defer rows.Close()

	return makePvzList(rows)
}

func (d *DBDriver) GetPvzWithReceptionInterval(ctx context.Context, limit, offset uint32, startInterval, endInterval time.Time) ([]models.Pvz, error) {
	rows, err := d.rwdb.Query(ctx, queryGetPvzWithReceptionInterval, startInterval, endInterval, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pvz with reception interval: %w", err)
	}
	defer rows.Close()

	return makePvzList(rows)
}

func (d *DBDriver) getReceptionInProgressId(ctx context.Context, tx pgx.Tx, pvzId pgtype.UUID) (pgtype.UUID, error) {
	var receptionId pgtype.UUID
	err := tx.QueryRow(ctx, queryGetReceptionInProgressId, pvzId).Scan(&receptionId)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, fmt.Errorf("no open reception")
	}
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to get reception in progress id: %w", err)
	}

	return receptionId, nil
}

func makePvzList(rows pgx.Rows) ([]models.Pvz, error) {
	var pvzList []models.Pvz
	for rows.Next() {
		var id pgtype.UUID
		var registerDate time.Time
		var city models.City

		err := rows.Scan(&id, &registerDate, &city)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		pvzList = append(pvzList, models.Pvz{
			Id:           id,
			RegisterDate: registerDate,
			City:         city,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pvzList, nil
}
