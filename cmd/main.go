package main

import (
	"URLShortener/internal/api"
	"URLShortener/internal/config"
	"URLShortener/internal/repository"
	"URLShortener/internal/repository/postgres"
	"URLShortener/internal/repository/redis"
	"URLShortener/internal/service"
	tokenmanager "URLShortener/internal/tokenManager"
	"URLShortener/internal/worker"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := &slog.Logger{}
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
	repo := repository.NewRepository(db, redisCache, 6*time.Minute, logger) //ttl потом брать из конфига

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	clickWorker := worker.NewClickWorker(repo, 5*time.Second, logger)
	go func() {

		clickWorker.Run(ctx)
	}()
	service := service.NewService(repo, cfg.CodeLength, cfg.MaxRetry, logger)
	tokman := tokenmanager.NewTokenManager(repo, 5*time.Minute, 3*24*time.Hour, []byte("аняня")) //дополнить конфиг
	h := api.NewHandler(service, tokman)
	engine := h.InitRoutes()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv.Shutdown(shutdownCtx)

	fmt.Println("service closed correctly")
}
