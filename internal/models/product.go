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

type ProductType string

const (
	Electronics ProductType = "electronics"
	Clothes     ProductType = "closes"
	Shoes       ProductType = "shoes"
)
