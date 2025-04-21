package product_model

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type Product struct {
	Id          pgtype.UUID
	AddingTime  time.Time
	ProductType ProductType
}

type ProductType string

const (
	Electronics ProductType = "электроника"
	Clothes     ProductType = "одежда"
	Shoes       ProductType = "обувь"
)
