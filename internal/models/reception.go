package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Reception struct {
	id            pgtype.UUID
	receptionTime time.Time
	pvz           Pvz
	products      []Product
	state         ReceptionState
}

type ReceptionState int

const (
	InProgress ReceptionState = iota
	Close
)
