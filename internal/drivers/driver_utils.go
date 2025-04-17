package drivers

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/pgtype"
	"github.com/jackc/pgx/v5"
)

func GetReceptionInProgressId(ctx context.Context, tx pgx.Tx, pvzId pgtype.UUID) (pgtype.UUID, error) {
	var receptionId pgtype.UUID
	err := tx.QueryRow(ctx, QueryGetReceptionInProgressId, pvzId).Scan(&receptionId)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, fmt.Errorf("no open receptions")
	}
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to get receptions in progress id: %w", err)
	}

	return receptionId, nil
}
