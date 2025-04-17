package services

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func GenerateUuid() (pgtype.UUID, error) {
	setUuid := uuid.New()

	pgUUID := pgtype.UUID{}
	err := pgUUID.Set(setUuid)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to set uuid: %w", err)
	}

	return pgUUID, nil
}

func ConvertOpenAPIUuidToPgType(openapiUuid openapi_types.UUID) (pgtype.UUID, error) {
	stdUuid, err := uuid.Parse(openapiUuid.String())
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID format: %w", err)
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Set(stdUuid)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("failed to convert to pgtype: %w", err)
	}

	return pgUUID, nil
}
