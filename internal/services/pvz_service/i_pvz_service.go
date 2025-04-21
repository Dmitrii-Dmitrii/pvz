package pvz_service

import (
	"context"
	"github.com/Dmitrii-Dmitrii/pvz/internal/generated"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/pvz_model"
)

type IPvzService interface {
	CreatePvz(ctx context.Context, pvzDto generated.PVZ) (*generated.PVZ, error)
	GetPvzFullInfo(ctx context.Context, pvzParams generated.GetPvzParams) ([]map[string]interface{}, error)
	GetAllPvz(ctx context.Context) ([]pvz_model.Pvz, error)
}
