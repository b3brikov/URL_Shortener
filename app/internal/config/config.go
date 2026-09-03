package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

const Version = "dev"

type Config struct {
	CodeLength       int           `env:"CODE_LEN" env-default:"8"`
	MaxRetry         int           `env:"MAX_RETRY" env-default:"5"`
	PostgresDriver   string        `env:"POSTGRES_DRIVER" env-required:"true"`
	PostgresUser     string        `env:"POSTGRES_USER" env-required:"true"`
	PostgresPassword string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	PostgresHost     string        `env:"POSTGRES_HOST" env-required:"true"`
	PostgresPort     string        `env:"POSTGRES_PORT" env-default:"5432"`
	PostgresDB       string        `env:"POSTGRES_DB" env-required:"true"`
	RedisAddr        string        `env:"REDIS_ADDR" env-required:"true"`
	HTTPPort         string        `env:"HTTP_PORT" env-default:"8080"`
	ShortCodeTTL     time.Duration `env:"SHORT_CODE_TTL"`
	AccessTokenTTL   time.Duration `env:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL  time.Duration `env:"REFRESH_TOKEN_TTL"`
	WorkerInterval   time.Duration `env:"WORKER_INTERVAL"`
	JWTSecret        string        `env:"JWT_SECRET" env-required:"true"`
	Version          string
}

func Load() (*Config, error) {
	cfg := &Config{}

	_ = godotenv.Load()

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalln("Load config error:", err.Error())
	}

	cfg.Version = Version

	log.Println("service version:", cfg.Version)

	return cfg, nil
}

func (c *Config) DSN() string {
	DSN := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
	return DSN
}
