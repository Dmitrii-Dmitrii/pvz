package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Pvz struct {
	Id           pgtype.UUID
	RegisterDate time.Time
	City         City
}

type City int

const (
	Moscow City = iota
	SPb
	Kazan
)

func NewPvz(id pgtype.UUID, registerDate time.Time, city City) *Pvz {
	return &Pvz{Id: id, RegisterDate: registerDate, City: city}
}
