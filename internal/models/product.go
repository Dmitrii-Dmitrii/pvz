package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Product struct {
	Id          pgtype.UUID
	AddingTime  time.Time
	ProductType ProductType
}

type ProductType int

const (
	Electronics ProductType = iota
	Clothes
	Shoes
)
