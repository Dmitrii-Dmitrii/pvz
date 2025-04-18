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

type City string

const (
	Moscow City = "moscow"
	SPb    City = "spb"
	Kazan  City = "kazan"
)
