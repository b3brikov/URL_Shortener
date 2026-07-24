package service

import (
	"URLShortener/internal/repository"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RepositoryMock struct {
	CreateCalled bool
}

func (r *RepositoryMock) CreateURL(ctx context.Context, original_url, short_code string, userID int) error {
	r.CreateCalled = true
	if original_url == "created" {
		return repository.ErrNoAffectedRows
	}
	return nil
}

func (r *RepositoryMock) GetOriginalURL(ctx context.Context, short_code string) (string, error) {
	if short_code == "existed" {
		return "exists_code", nil
	}

	return "", repository.ErrCodeNotFound
}

func TestShortenerCreateURL(t *testing.T) {
	repo := &RepositoryMock{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s := NewService(repo, 5, 3, slog.Default())

	cases := []struct {
		Name        string
		Url         string
		UserID      int
		ExpectedErr error
	}{
		{"дефолт 1", "google.com", 1, nil},
		{"дефолт 2", "анняфываыфва", 1488, nil},
		{"необходима ошибка", "created", 4, repository.ErrNoAffectedRows},
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
