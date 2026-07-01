package service

import (
	"URLShortener/internal/models"
	"URLShortener/internal/repository"
	"context"
	"fmt"
)

type Repository interface {
	CreateURL(ctx context.Context, original_url, short_code string, userID int) error
	GetOriginalURL(ctx context.Context, short_code string) (string, error)
}

type Service struct {
	repo              Repository
	codeLen, maxRetry int
}

func NewService(repo Repository, codeLen, maxRetry int) *Service {
	return &Service{
		repo:     repo,
		codeLen:  codeLen,
		maxRetry: maxRetry,
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
			return models.URLModel{}, fmt.Errorf("cannot create new code: %w", err)
		}
		return models.URLModel{Original_url: url, Short_code: shortCode}, nil
	}
	return models.URLModel{}, fmt.Errorf("cannot create new code: %w", err)
}

func (s *Service) GetOriginalURL(ctx context.Context, url string) (string, error) {
	return s.repo.GetOriginalURL(ctx, url)
}
