package di

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
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type DI struct {
	Logger       *slog.Logger
	Config       *config.Config
	Database     *sql.DB
	Cache        *redis.RedisStorage
	Repository   *repository.Repository
	TokenManager *tokenmanager.TokenManager
	Shortener    *service.Service
	Server       *api.Handler
	Worker       *worker.ClickWorker
}

func (d *DI) GetLogger() *slog.Logger {
	if d.Logger != nil {
		return d.Logger
	}

	d.Logger = slog.Default()
	return d.Logger
}

func (d *DI) GetConfig() *config.Config {
	if d.Config != nil {
		return d.Config
	}
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	d.Config = cfg
	return d.Config
}

func (d *DI) GetDatabase() *sql.DB {
	if d.Database != nil {
		return d.Database
	}
	db, err := postgres.GetDB(d.GetConfig().DSN())
	if err != nil {
		panic(err)
	}
	d.Database = db
	return d.Database
}

func (d *DI) GetCache() *redis.RedisStorage {
	if d.Cache != nil {
		return d.Cache
	}

	rd, err := redis.NewRedisClient(d.Config.RedisAddr)
	if err != nil {
		panic(err)
	}

	d.Cache = rd
	return d.Cache
}

func (d *DI) GetRepository() *repository.Repository {
	if d.Repository != nil {
		return d.Repository
	}

	repo := repository.NewRepository(d.GetDatabase(), d.GetCache(), d.GetConfig().ShortCodeTTL, d.GetLogger())

	d.Repository = repo
	return d.Repository
}

func (d *DI) GetShortener() *service.Service {
	if d.Shortener != nil {
		return d.Shortener
	}

	service := service.NewService(d.GetRepository(), d.GetConfig().CodeLength, d.GetConfig().MaxRetry, d.GetLogger())

	d.Shortener = service
	return d.Shortener
}

func (d *DI) GetWorker() *worker.ClickWorker {
	if d.Worker != nil {
		return d.Worker
	}

	worker := worker.NewClickWorker(d.GetRepository(), d.GetConfig().WorkerInterval, d.GetLogger())

	d.Worker = worker
	return d.Worker
}

func (d *DI) GetTokenManager() *tokenmanager.TokenManager {
	if d.TokenManager != nil {
		return d.TokenManager
	}

	man := tokenmanager.NewTokenManager(d.GetRepository(), d.GetConfig().AccessTokenTTL, d.GetConfig().RefreshTokenTTL, []byte(d.GetConfig().JWTSecret))

	d.TokenManager = man
	return d.TokenManager
}

func (d *DI) GetServer() *api.Handler {
	if d.Server != nil {
		return d.Server
	}

	server := api.NewHandler(d.GetShortener(), d.GetTokenManager())

	d.Server = server

	return d.Server
}

func (d *DI) Run() error {
	server := d.GetServer()
	worker := d.GetWorker()

	routes := server.InitRoutes()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	go func() {

		worker.Run(ctx)
	}()
	srv := &http.Server{
		Addr:    ":" + d.GetConfig().HTTPPort,
		Handler: routes,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			d.GetLogger().Info(err.Error())
		}
	}()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
