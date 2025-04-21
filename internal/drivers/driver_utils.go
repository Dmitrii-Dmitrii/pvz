package drivers

import (
	"context"
	"errors"
	"github.com/Dmitrii-Dmitrii/pvz/internal/models/custom_errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

func GetReceptionInProgressId(ctx context.Context, tx pgx.Tx, pvzId pgtype.UUID) (pgtype.UUID, error) {
	var receptionId pgtype.UUID
	err := tx.QueryRow(ctx, QueryGetReceptionInProgressId, pvzId).Scan(&receptionId)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Msg(custom_errors.ErrNoOpenReception.Message)
		return pgtype.UUID{}, custom_errors.ErrNoOpenReception
	}
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrGetReceptionInProgress.Message)
		return pgtype.UUID{}, custom_errors.ErrGetReceptionInProgress
	}

	return receptionId, nil
}
