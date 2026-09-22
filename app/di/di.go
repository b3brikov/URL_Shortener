package di

import (
	"URLShortener/internal/config"
	"URLShortener/internal/core/cache/redis"
	"URLShortener/internal/core/db"
	"URLShortener/internal/core/middleware"
	"URLShortener/internal/core/postgres"
	"URLShortener/internal/core/transport/gRPC/interceptor"
	"URLShortener/internal/core/transport/gRPC/proto"
	server "URLShortener/internal/core/transport/http"
	authservice "URLShortener/internal/features/auth/service"
	authtransport "URLShortener/internal/features/auth/transport"
	shortenerrepository "URLShortener/internal/features/shortener/repository"
	shortenerservice "URLShortener/internal/features/shortener/service"
	shortenertransport "URLShortener/internal/features/shortener/transport/api"
	grpcapi "URLShortener/internal/features/shortener/transport/grpc_api"
	"URLShortener/internal/features/shortener/worker"
	"database/sql"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

type DI struct {
	logger             *slog.Logger
	config             *config.Config         //
	middleware         *middleware.Middleware //
	auth               *authservice.Service   //
	authtransport      *authtransport.Handlers
	postgres           *postgres.PostgresDB
	db                 *sql.DB
	cache              *redis.RedisStorage
	shortenerRepo      *shortenerrepository.Repository
	clickWorker        *worker.ClickWorker
	shortenerService   *shortenerservice.Service
	shortenerTransport *shortenertransport.Handlers
	server             *server.Server
	grpcShortener      *grpcapi.Server
	grpcServer         *grpc.Server
}

func (d *DI) RunGRPC() error {
	lis, err := net.Listen("tcp", ":"+d.Config().GRPCPort)
	if err != nil {
		return err
	}

	return d.GrpcServer().Serve(lis)
}

func (d *DI) GrpcServer() *grpc.Server {
	if d.grpcServer != nil {
		return d.grpcServer
	}

	server := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptor.RecoverInterceptor(d.Logger())))
	proto.RegisterShortenerServer(server, d.GrpcShortener())
	d.grpcServer = server
	return d.grpcServer
}

func (d *DI) GrpcShortener() *grpcapi.Server {
	if d.grpcShortener != nil {
		return d.grpcShortener
	}

	server := grpcapi.NewServer(d.shortenerService)
	d.grpcShortener = server
	return d.grpcShortener
}

func (d *DI) Server() *server.Server {
	if d.server != nil {
		return d.server
	}

	server := server.NewServer(d.Config().HTTPPort, d.AuthTransport(), d.ShortenerTransport())

	d.server = server

	return d.server
}

func (d *DI) ShortenerService() *shortenerservice.Service {
	if d.shortenerService != nil {
		return d.shortenerService
	}
	cfg := d.Config()
	config := shortenerservice.Config{
		CodeLen:  cfg.CodeLength,
		MaxRetry: cfg.MaxRetry,
		CodeTtl:  cfg.ShortCodeTTL,
	}
	service := shortenerservice.NewService(d.ShortenerRepo(), config, d.Cache(), d.Logger())
	d.shortenerService = service

	return d.shortenerService
}

func (d *DI) ShortenerTransport() *shortenertransport.Handlers {
	if d.shortenerTransport != nil {
		return d.shortenerTransport
	}

	transport := shortenertransport.ShortenerHandlers(d.ShortenerService(), d.Middleware(), d.Logger())

	d.shortenerTransport = transport

	return d.shortenerTransport
}

func (d *DI) Middleware() *middleware.Middleware {
	if d.middleware != nil {
		return d.middleware
	}

	m := middleware.NewMiddleware(d.AuthService())

	d.middleware = m
	return d.middleware
}

func (d *DI) AuthTransport() *authtransport.Handlers {
	if d.authtransport != nil {
		return d.authtransport
	}

	transport := authtransport.NewHandlers(d.AuthService(), d.Middleware())

	d.authtransport = transport

	return d.authtransport
}

func (d *DI) Logger() *slog.Logger {
	if d.logger != nil {
		return d.logger
	}

	logger := slog.Default()

	d.logger = logger

	return d.logger
}

func (d *DI) ClickWorker() *worker.ClickWorker {
	if d.clickWorker != nil {
		return d.clickWorker
	}

	worker := worker.NewClickWorker(d.PostgresDB(), d.Cache(), d.Config().WorkerInterval, d.Logger())

	d.clickWorker = worker
	return d.clickWorker
}

func (d *DI) ShortenerRepo() *shortenerrepository.Repository {
	if d.shortenerRepo != nil {
		return d.shortenerRepo
	}

	repo := shortenerrepository.NewRepository(d.PostgresDB())
	d.shortenerRepo = repo
	return d.shortenerRepo
}

func (d *DI) PostgresDB() *postgres.PostgresDB {
	if d.postgres != nil {
		return d.postgres
	}

	postgres := postgres.NewPostgresDB(d.DB())

	d.postgres = postgres
	return d.postgres
}

func (d *DI) AuthService() *authservice.Service {
	if d.auth != nil {
		return d.auth
	}

	auth := authservice.NewService(d.PostgresDB(), d.Cache(), d.Config().AccessTokenTTL, d.Config().RefreshTokenTTL, []byte(d.Config().JWTSecret))
	d.auth = auth
	return d.auth
}

func (d *DI) Cache() *redis.RedisStorage {
	if d.cache != nil {
		return d.cache
	}

	cache, err := redis.NewRedisClient(d.Config().RedisAddr)
	if err != nil {
		panic(err)
	}

	d.cache = cache
	return d.cache
}

func (d *DI) DB() *sql.DB {
	if d.db != nil {
		return d.db
	}

	db, err := db.GetDB(d.Config().DSN(), d.Config().PostgresDriver)
	if err != nil {
		panic(err)
	}

	d.db = db
	return db
}

func (d *DI) Config() *config.Config {
	if d.config != nil {
		return d.config
	}

	cfg, err := config.Load()

	if err != nil {
		panic(err)
	}
	d.config = cfg
	return d.config
}
