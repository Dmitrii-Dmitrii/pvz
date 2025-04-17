package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Pvz struct {
	Id               pgtype.UUID
	RegistrationDate time.Time
	City             City
}

type City int

const (
	Moscow City = iota
	SPb
	Kazan
)
