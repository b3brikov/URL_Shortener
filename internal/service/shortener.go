package service

import (
	"URLShortener/internal/models"
	"URLShortener/internal/repository"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Repository interface {
	CreateURL(ctx context.Context, original_url, short_code string, userID int) error
	GetOriginalURL(ctx context.Context, short_code string) (string, error)
}

type Service struct {
	repo              Repository
	codeLen, maxRetry int
	Logger            *slog.Logger
}

func NewService(repo Repository, codeLen, maxRetry int, logger *slog.Logger) *Service {
	return &Service{
		repo:     repo,
		codeLen:  codeLen,
		maxRetry: maxRetry,
		Logger:   logger,
	}
}

func (s *Service) GenerateCode() string {
	return generate(s.codeLen)
}

func (s *Service) CreateNewCode(ctx context.Context, url string, userID int) (models.URLModel, error) {
	var err error
	var shortCode string
	for i := 0; i < s.maxRetry; i++ {
		shortCode = s.GenerateCode()
		err = s.repo.CreateURL(ctx, url, shortCode, userID)
		if repository.IsUniqueViolation(err) {
			continue
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return models.URLModel{}, ErrTimeOut
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

func (s *Service) GetOriginalURL(ctx context.Context, short_code string) (string, error) {
	res, err := s.repo.GetOriginalURL(ctx, short_code)
	if err != nil {
		if !errors.Is(err, repository.ErrCodeNotFound) {
			s.Logger.Error("cannot get original url", slog.Any("error", err.Error()), slog.Any("short_code", short_code))
			return "", fmt.Errorf("get original url: %w", err)
		}
		if errors.Is(err, context.Canceled) {
			return "", ErrTimeOut
		}
		return "", err
	}
	return res, nil
}
