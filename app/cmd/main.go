package main

import (
	"URLShortener/di"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	di := di.DI{}
	di.Run()
}
