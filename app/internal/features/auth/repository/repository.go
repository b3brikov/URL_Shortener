package repository

import (
	"URLShortener/internal/core/models"
	"context"
	"database/sql"
)

type Repository struct {
}

func (r *Repository) GetUser(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash FROM users WHERE email=$1`
	var user models.User
	row := r.db.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.HashPass)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFoundUser
		}
		return nil, err
	}
	return &user, nil
}
