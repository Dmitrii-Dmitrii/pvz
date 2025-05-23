package reception_driver

import (
	"context"
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/drivers"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/reception_model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"time"
)

type ReceptionDriver struct {
	adapter drivers.Adapter
}

func NewReceptionDriver(adapter drivers.Adapter) *ReceptionDriver {
	return &ReceptionDriver{adapter: adapter}
}

func (d *ReceptionDriver) CreateReception(ctx context.Context, reception *reception_model.Reception) error {
	_, err := d.adapter.Exec(ctx, drivers.QueryCreateReception, reception.Id, reception.ReceptionTime, reception.PvzId, reception.Status)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrCreateReception.Message)
		return custom_errors.ErrCreateReception
	}

	return nil
}

func (d *ReceptionDriver) CloseReception(ctx context.Context, pvzId pgtype.UUID) (*reception_model.Reception, error) {
	tx, err := d.adapter.Begin(ctx)
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

	reception, err := d.getReception(ctx, receptionId)
	if err != nil {
		return nil, err
	}

	return reception, nil
}

func (d *ReceptionDriver) GetLastReceptionStatus(ctx context.Context, pvzId pgtype.UUID) (*reception_model.ReceptionStatus, error) {
	var status reception_model.ReceptionStatus
	err := d.adapter.QueryRow(ctx, drivers.QueryGetLastReceptionStatus, pvzId).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, custom_errors.ErrNoReception
	}

	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetLastReceptionStatus.Message)
		return nil, custom_errors.ErrGetLastReceptionStatus
	}

	return &status, nil
}

func (d *ReceptionDriver) getReception(ctx context.Context, id pgtype.UUID) (*reception_model.Reception, error) {
	var receptionTime time.Time
	var pvzId pgtype.UUID
	var status reception_model.ReceptionStatus
	err := d.adapter.QueryRow(ctx, drivers.QueryGetReception, id).Scan(&receptionTime, &pvzId, &status)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetReception.Message)
		return nil, custom_errors.ErrGetReception
	}

	reception := &reception_model.Reception{Id: id, PvzId: pvzId, ReceptionTime: receptionTime, Status: status}
	return reception, nil
}
