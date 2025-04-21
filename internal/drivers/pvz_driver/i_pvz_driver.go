package pvz_driver

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/pvz_model"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type IPvzDriver interface {
	CreatePvz(ctx context.Context, pvz *pvz_model.Pvz) error
	GetPvzById(ctx context.Context, id pgtype.UUID) (*pvz_model.Pvz, error)
	GetPvzFullInfo(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error)
	GetAllPvz(ctx context.Context) ([]pvz_model.Pvz, error)
}
