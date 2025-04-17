package pvzs

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pvz/internal/drivers"
	"pvz/internal/models"
	"time"
)

type PvzDriver struct {
	rwdb *pgxpool.Pool
}

func NewPvzDriver(rwdb *pgxpool.Pool) *PvzDriver {
	return &PvzDriver{rwdb: rwdb}
}

func (d *PvzDriver) CreatePvz(ctx context.Context, pvz *models.Pvz) (*models.Pvz, error) {
	err := d.rwdb.QueryRow(ctx, drivers.QueryCreatePvz, pvz.Id, pvz.RegistrationDate, pvz.City)
	if err != nil {
		return nil, fmt.Errorf("failed to create pvzs: %w", err)
	}
	return pvz, nil
}

func (d *PvzDriver) GetPvz(ctx context.Context, limit, offset uint32) ([]models.Pvz, error) {
	rows, err := d.rwdb.Query(ctx, drivers.QueryGetPvz, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pvzs: %w", err)
	}
	defer rows.Close()

	return makePvzList(rows)
}

func (d *PvzDriver) GetPvzWithReceptionInterval(ctx context.Context, limit, offset uint32, startInterval, endInterval time.Time) ([]models.Pvz, error) {
	rows, err := d.rwdb.Query(ctx, drivers.QueryGetPvzWithReceptionInterval, startInterval, endInterval, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pvzs with receptions interval: %w", err)
	}
	defer rows.Close()

	return makePvzList(rows)
}

func makePvzList(rows pgx.Rows) ([]models.Pvz, error) {
	var pvzList []models.Pvz
	for rows.Next() {
		var id pgtype.UUID
		var registrationDate time.Time
		var city models.City

		err := rows.Scan(&id, &registrationDate, &city)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		pvzList = append(pvzList, models.Pvz{
			Id:               id,
			RegistrationDate: registrationDate,
			City:             city,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pvzList, nil
}
