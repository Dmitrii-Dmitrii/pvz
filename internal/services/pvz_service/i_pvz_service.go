package pvz_service

import (
	"context"
	"pvz/internal/generated"
)

type IPvzService interface {
	CreatePvz(ctx context.Context, pvzDto generated.PVZ) (*generated.PVZ, error)
	GetPvz(ctx context.Context, pvzParams generated.GetPvzParams) ([]map[string]interface{}, error)
}
