package service

import (
	"URLShortener/internal/core/models"
	"URLShortener/internal/core/postgres"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	AddClick(ctx context.Context, key string) error
	GetValue(ctx context.Context, key string) (string, error)
}

type URLRepository interface {
	CreateURL(ctx context.Context, original_url, short_code string, userID int) error
	GetOriginalURL(ctx context.Context, short_code string) (string, error)
}

type service struct {
	config Config
	cache  Cache
	repo   URLRepository
	Logger *slog.Logger
}

func NewService(repo URLRepository, cfg Config, cache Cache, logger *slog.Logger) *service {
	return &service{
		repo:   repo,
		Logger: logger,
		config: cfg,
		cache:  cache,
	}
}

func (s *service) generateCode() string {
	return generate(s.config.CodeLen)
}

func (s *service) CreateNewCode(ctx context.Context, url string, userID int) (models.URLModel, error) {
	var err error
	var shortCode string
	for i := 0; i < s.config.MaxRetry; i++ {
		shortCode = s.generateCode()
		err = s.repo.CreateURL(ctx, url, shortCode, userID)
		if postgres.IsUniqueViolation(err) {
			continue
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return models.URLModel{}, ErrTimeOut
			}
			if errors.Is(err, postgres.ErrNoAffectedRows) {
				return models.URLModel{}, err
			}
			s.Logger.Error("cannot create new code", slog.Any("error", err.Error()))
			return models.URLModel{}, ErrUnexpectedError
		}
		return models.URLModel{Original_url: url, Short_code: shortCode}, nil
	}
	if errors.Is(err, context.Canceled) {
		return models.URLModel{}, ErrTimeOut
	}

	s.Logger.Error("cannot create new code", slog.Any("error", err.Error()))

	return models.URLModel{}, ErrUnexpectedError
}

func (s *service) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	if res, err := s.getCode(ctx, shortCode); err != nil {
		s.Logger.Debug("value not in cache", slog.Any("key", shortCode))
	} else {
		go func() {
			s.addValue(shortCode, res)
		}()
		defer s.addClick(shortCode)
		return res, nil
	}

	res, err := s.repo.GetOriginalURL(ctx, shortCode)

	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.Logger.Debug("get original url context cancelled", slog.Any("error", err.Error()), slog.Any("short_code", shortCode))
			return "", ErrTimeOut
		}
		if !errors.Is(err, postgres.ErrCodeNotFound) {
			s.Logger.Error("cannot get original url", slog.Any("error", err.Error()), slog.Any("short_code", shortCode))
			return "", fmt.Errorf("get original url: %w", err)
		}
		return "", err
	}
	go func() {
		s.addValue(shortCode, res)
	}()
	defer s.addClick(shortCode)
	return res, nil
}

func (s *service) addClick(shortCode string) {
	cacheContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.cache.AddClick(cacheContext, shortCode); err != nil {
		s.Logger.Error(
			"cannot add click",
			slog.String("short_code", shortCode),
			slog.Any("error", err),
		)
	}
}

func (s *service) getCode(ctx context.Context, key string) (string, error) {
	return s.cache.GetValue(ctx, key)
}

func (s *service) addValue(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.cache.Set(ctx, key, value, s.config.CodeTtl)
}
