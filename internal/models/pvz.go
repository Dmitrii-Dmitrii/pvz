package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Pvz struct {
	id           pgtype.UUID
	registerDate time.Time
	city         City
}

type City int

const (
	Moscow City = iota
	SPb
	Kazan
)
