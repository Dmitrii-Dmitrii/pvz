package pvzs

import (
	"context"
	"pvz/internal/generated"
)

type IPvzService interface {
	CreatePvz(ctx context.Context, cityDto generated.PVZCity) (*generated.PVZ, error)
	// GetPvz methods
}
