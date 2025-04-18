package receptions

import (
	"context"
	"errors"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"pvz/internal/drivers"
	"pvz/internal/models"
	"pvz/internal/models/custom_errors"
	"time"
)

type ReceptionDriver struct {
	rwdb *pgxpool.Pool
}

func NewReceptionDriver(rwdb *pgxpool.Pool) *ReceptionDriver {
	return &ReceptionDriver{rwdb: rwdb}
}

func (d *ReceptionDriver) CreateReception(ctx context.Context, reception *models.Reception) error {
	_, err := d.rwdb.Exec(ctx, drivers.QueryCreateReception, reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreateReception.Message)
		return custom_errors.ErrCreateReception
	}

	return nil
}

func (d *ReceptionDriver) GetReception(ctx context.Context, id pgtype.UUID) (*models.Reception, error) {
	var receptionTime time.Time
	var pvzId pgtype.UUID
	var status models.ReceptionStatus
	err := d.rwdb.QueryRow(ctx, drivers.QueryGetReception, id).Scan(&receptionTime, &pvzId, &status)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetReception.Message)
		return nil, custom_errors.ErrGetReception
	}

	reception := &models.Reception{Id: id, PvzId: pvzId, ReceptionTime: receptionTime, Status: status}
	return reception, nil
}

func (d *ReceptionDriver) CloseReception(ctx context.Context, pvzId pgtype.UUID) (*models.Reception, error) {
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

	_, err = tx.Exec(ctx, drivers.QueryCloseReception, receptionId)

	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCloseReception.Message)
		return nil, custom_errors.ErrCloseReception
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCommitTransaction.Message)
		return nil, custom_errors.ErrCommitTransaction
	}

	reception, err := d.GetReception(ctx, receptionId)
	if err != nil {
		return nil, err
	}

	return reception, nil
}

func (d *ReceptionDriver) GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*models.ReceptionStatus, error) {
	var status models.ReceptionStatus
	err := d.rwdb.QueryRow(ctx, drivers.QueryGetLastReceptionStatus, pvzId).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Msg(custom_errors.ErrNoReception.Message)
		return nil, custom_errors.ErrNoReception
	}

	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetLastReceptionStatus.Message)
		return nil, custom_errors.ErrGetLastReceptionStatus
	}

	return &status, nil
}
