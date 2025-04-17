package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Reception struct {
	Id            pgtype.UUID
	ReceptionTime time.Time
	PvzId         pgtype.UUID
	Products      []Product
	Status        ReceptionStatus
}

type ReceptionStatus int

const (
	InProgress ReceptionStatus = iota
	Close
)
