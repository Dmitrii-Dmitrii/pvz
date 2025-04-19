package pvz_driver

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"pvz/internal/models/pvz_model"
	"time"
)

type IPvzDriver interface {
	CreatePvz(ctx context.Context, pvz *pvz_model.Pvz) (*pvz_model.Pvz, error)
	GetPvzById(ctx context.Context, id pgtype.UUID) (*pvz_model.Pvz, error)
	GetPvz(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error)
}
