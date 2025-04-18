package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"pvz/internal/drivers/pvzs"
)

func main() {
	connString := "postgres://pvz_user:pvz_password@localhost:5432/pvz_database"
	ctx := context.Background()
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	fmt.Println("Connected to database")
	defer dbpool.Close()

	driver := pvzs.NewPvzDriver(dbpool)
	result, err := driver.GetPvz(ctx, 10, 0, nil, nil)
	fmt.Println(result, err)
}
