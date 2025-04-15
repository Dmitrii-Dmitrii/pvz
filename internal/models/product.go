package models

import (
	"github.com/jackc/pgx/pgtype"
	"time"
)

type Product struct {
	id          pgtype.UUID
	addingTime  time.Time
	productType ProductType
}

type ProductType int

const (
	Electronics ProductType = iota
	Clothes
	Shoes
)
