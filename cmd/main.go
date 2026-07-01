package main

import (
	"URLShortener/internal/api"
	"URLShortener/internal/config"
	"URLShortener/internal/repository"
	"URLShortener/internal/repository/postgres"
	"URLShortener/internal/repository/redis"
	"URLShortener/internal/service"
	tokenmanager "URLShortener/internal/tokenManager"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	fmt.Println(sql.Drivers())
	cfg, err := config.Load()
	if err != nil {
		log.Println("cfg")
		panic(err)
	}
	db, err := postgres.GetDB(cfg.DSN())
	if err != nil {
		log.Println("DB")
		panic(err)
	}
	redisCache, err := redis.NewRedisClient(cfg.RedisAddr)
	if err != nil {
		log.Println("redis")
		panic(err)
	}
	repo := repository.NewRepository(db, redisCache, 6*time.Minute) //ttl потом брать из конфига
	service := service.NewService(repo, cfg.CodeLength, cfg.MaxRetry)
	tokman := tokenmanager.NewTokenManager(repo, 5*time.Minute, 3*24*time.Hour, []byte("аняня")) //дополнить конфиг
	h := api.NewHandler(service, tokman)
	engine := h.InitRoutes()
	engine.Run(":8080")
}
