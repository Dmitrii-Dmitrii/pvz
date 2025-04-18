package services

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"
	"pvz/internal/models/custom_errors"
)

func GenerateUuid() (pgtype.UUID, error) {
	setUuid := uuid.New()

	pgUUID := pgtype.UUID{}
	err := pgUUID.Set(setUuid)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrSetUuid.Message)
		return pgtype.UUID{}, custom_errors.ErrSetUuid
	}

	return pgUUID, nil
}

func ConvertOpenAPIUuidToPgType(openapiUuid openapi_types.UUID) (pgtype.UUID, error) {
	stdUuid, err := uuid.Parse(openapiUuid.String())
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrUuidFormat.Message)
		return pgtype.UUID{}, custom_errors.ErrUuidFormat
	}

	var pgUUID pgtype.UUID
	err = pgUUID.Set(stdUuid)
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrConvertUuidToPgtype.Message)
		return pgtype.UUID{}, custom_errors.ErrConvertUuidToPgtype
	}

	return pgUUID, nil
}

func ConvertPgUuidToOpenAPI(pgUuid pgtype.UUID) (openapi_types.UUID, error) {
	if pgUuid.Status != pgtype.Present {
		log.Error().Msg(custom_errors.ErrUuidNotPresent.Message)
		return openapi_types.UUID{}, custom_errors.ErrUuidNotPresent
	}

	stdUuid, err := uuid.FromBytes(pgUuid.Bytes[:])
	if err != nil {
		log.Error().Err(err).Msg(custom_errors.ErrConvertUuidToOpenapi.Message)
		return openapi_types.UUID{}, custom_errors.ErrConvertUuidToOpenapi
	}

	return stdUuid, nil
}
