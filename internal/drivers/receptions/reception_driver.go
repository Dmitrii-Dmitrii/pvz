package receptions

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"pvz/internal/drivers"
	"pvz/internal/models"
	"time"
)

type ReceptionDriver struct {
	rwdb *pgxpool.Pool
}

func NewReceptionDriver(rwdb *pgxpool.Pool) *ReceptionDriver {
	return &ReceptionDriver{rwdb: rwdb}
}

func (d *ReceptionDriver) CreateReception(ctx context.Context, reception *models.Reception) error {
	err := d.rwdb.QueryRow(ctx, drivers.QueryCreateReception, reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status)
	if err != nil {
		return fmt.Errorf("failed to create receptions: %w", err)
	}
	return nil
}

func (d *ReceptionDriver) GetReception(ctx context.Context, id pgtype.UUID) (*models.Reception, error) {
	var receptionTime time.Time
	var pvzId pgtype.UUID
	var status models.ReceptionStatus
	err := d.rwdb.QueryRow(ctx, drivers.QueryGetReception, id).Scan(&receptionTime, &pvzId, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to get receptions: %w", err)
	}
	reception := &models.Reception{Id: id, PvzId: pvzId, ReceptionTime: receptionTime, Status: status}
	return reception, nil
}

func (d *ReceptionDriver) CloseReception(ctx context.Context, pvzId pgtype.UUID) error {
	tx, err := d.rwdb.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	receptionId, err := drivers.GetReceptionInProgressId(ctx, tx, pvzId)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, drivers.QueryCloseReception, receptionId)

	if err != nil {
		return fmt.Errorf("failed to close receptions: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
