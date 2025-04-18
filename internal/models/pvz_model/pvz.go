package pvzs

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Pvz struct {
	Id               pgtype.UUID
	RegistrationDate time.Time
	City             City
}

type City string

const (
	Moscow City = "Москва"
	SPb    City = "Санкт-Петербург"
	Kazan  City = "Казань"
)
