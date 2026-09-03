package service

import (
	"URLShortener/internal/core/postgres"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CacheMock struct{}

func (c *CacheMock) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return nil
}

func (c *CacheMock) GetValue(ctx context.Context, key string) (string, error) {

	return "res.Val()", nil
}

func (c *CacheMock) AddClick(ctx context.Context, key string) error {
	return nil
}

type RepositoryMock struct {
	CreateCalled bool
}

func (r *RepositoryMock) CreateURL(ctx context.Context, original_url, short_code string, userID int) error {
	r.CreateCalled = true
	if original_url == "created" {
		return postgres.ErrNoAffectedRows
	}
	return nil
}

func (r *RepositoryMock) GetOriginalURL(ctx context.Context, short_code string) (string, error) {
	if short_code == "existed" {
		return "exists_code", nil
	}

	return "", postgres.ErrCodeNotFound
}

func TestShortenerCreateURL(t *testing.T) {
	repo := &RepositoryMock{}
	cache := &CacheMock{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cfg := Config{
		CodeLen:  5,
		MaxRetry: 3,
		CodeTtl:  3 * time.Second,
	}
	s := NewService(repo, cfg, cache, slog.Default())

	cases := []struct {
		Name        string
		Url         string
		UserID      int
		ExpectedErr error
	}{
		{"дефолт 1", "google.com", 1, nil},
		{"дефолт 2", "анняфываыфва", 1488, nil},
		{"необходима ошибка", "created", 4, postgres.ErrNoAffectedRows},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			model, err := s.CreateNewCode(ctx, c.Url, c.UserID)

			if c.ExpectedErr != nil {
				assert.ErrorIs(t, err, c.ExpectedErr)
				assert.Equal(t, model.Original_url, "")
			} else {
				require.NoError(t, err)
				assert.Equal(t, model.Original_url, c.Url)
				require.NotZero(t, model.Short_code)
			}
			assert.True(t, repo.CreateCalled)

			repo.CreateCalled = false
		})
	}
}
