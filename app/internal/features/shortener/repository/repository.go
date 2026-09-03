package repository

import (
	"context"
)

type DB interface {
	CreateURL(ctx context.Context, original_url, short_code string, userID int) error
	GetOriginalURL(ctx context.Context, shortCode string) (string, error)
}

type repository struct {
	db DB
}

func NewRepository(db DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateNewURL(ctx context.Context, original_url, short_code string, userID int) error {
	return r.db.CreateURL(ctx, original_url, short_code, userID)
}

func (r *repository) OriginalURL(ctx context.Context, shortCode string) (string, error) {
	return r.db.GetOriginalURL(ctx, shortCode)
}
