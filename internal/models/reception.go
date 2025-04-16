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

func NewReception(id, pvzId pgtype.UUID, receptionTime time.Time, status ReceptionStatus) *Reception {
	return &Reception{Id: id, PvzId: pvzId, ReceptionTime: receptionTime, Status: status}
}
