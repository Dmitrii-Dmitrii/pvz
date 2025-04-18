package pvz_driver

import (
	"context"
	"pvz/internal/models/pvz_model"
	"time"
)

type IPvzDriver interface {
	CreatePvz(ctx context.Context, pvz *pvzs.Pvz) (*pvzs.Pvz, error)
	GetPvz(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error)
}
