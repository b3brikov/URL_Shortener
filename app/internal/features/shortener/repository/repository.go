package repository

import (
	"context"
)

type DB interface {
	CreateURL(ctx context.Context, original_url, short_code string, userID *int) error
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type Repository struct {
	db DB
}

func NewRepository(db DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateNewURL(ctx context.Context, original_url, short_code string, userID *int) error {
	return r.db.CreateURL(ctx, original_url, short_code, userID)
}

func (r *Repository) OriginalURL(ctx context.Context, shortCode string) (string, error) {
	return r.db.GetOriginalURL(ctx, shortCode)
}
