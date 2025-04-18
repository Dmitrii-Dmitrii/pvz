package pvzs

import (
	"context"
	"pvz/internal/models"
	"time"
)

type IPvzDriver interface {
	CreatePvz(ctx context.Context, pvz *models.Pvz) (*models.Pvz, error)
	GetPvz(ctx context.Context, limit, offset uint32, startInterval, endInterval *time.Time) ([]map[string]interface{}, error)
}
