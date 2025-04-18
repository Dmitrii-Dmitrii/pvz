package models

import (
	"github.com/jackc/pgx/pgtype"
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

type ReceptionKey struct {
	PvzID       pgtype.UUID
	ReceptionID pgtype.UUID
}
