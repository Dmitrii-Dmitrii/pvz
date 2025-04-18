package services

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/models/custom_errors"
)

func GenerateUuid() pgtype.UUID {
	newUuid := uuid.New()

	pgUuid := pgtype.UUID{
		Bytes: newUuid,
		Valid: true,
	}

	return pgUuid
}

func ConvertOpenAPIUuidToPgType(openapiUuid openapi_types.UUID) (pgtype.UUID, error) {
	stdUuid, err := uuid.Parse(openapiUuid.String())
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrUuidFormat.Message)
		return pgtype.UUID{}, custom_errors.ErrUuidFormat
	}

	pgUuid := pgtype.UUID{
		Bytes: stdUuid,
		Valid: true,
	}

	return pgUuid, nil
}

func ConvertPgUuidToOpenAPI(pgUuid pgtype.UUID) (openapi_types.UUID, error) {
	if !pgUuid.Valid {
		log.Error().Msg(custom_errors.ErrUuidNotValid.Message)
		return openapi_types.UUID{}, custom_errors.ErrUuidNotValid
	}

	stdUuid, err := uuid.FromBytes(pgUuid.Bytes[:])
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrConvertUuidToOpenapi.Message)
		return openapi_types.UUID{}, custom_errors.ErrConvertUuidToOpenapi
	}

	return stdUuid, nil
}
