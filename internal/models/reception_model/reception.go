package reception_model

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Reception struct {
	Id            pgtype.UUID
	ReceptionTime time.Time
	PvzId         pgtype.UUID
	ProductIds    []pgtype.UUID
	Status        ReceptionStatus
}

type ReceptionStatus string

const (
	InProgress ReceptionStatus = "in_progress"
	Close      ReceptionStatus = "close"
)
