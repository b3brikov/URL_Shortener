package postgres

import (
	"URLShortener/internal/core/models"
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(db *sql.DB) *PostgresDB {
	return &PostgresDB{
		db: db,
	}
}

func (r *PostgresDB) CreateURL(ctx context.Context, original_url, short_code string, userID *int) error {
	query := `INSERT INTO urls (original_url,short_code,user_id) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, original_url, short_code, userID)

	return err
}

func (r *PostgresDB) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code=$1`
	var url string

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&url)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrCodeNotFound
		}
		return "", err
	}
	return url, nil
}

func (r *PostgresDB) Batch(ctx context.Context, flushed map[string]int) error {
	if len(flushed) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := `UPDATE urls SET clicks=clicks+$1 WHERE short_code=$2`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for code, data := range flushed {
		_, err = stmt.ExecContext(ctx, data, code)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *PostgresDB) CreateUser(ctx context.Context, username, email, hash string) error {
	query := `INSERT INTO users (username,email,password_hash) VALUES ($1,$2,$3)`

	_, err := r.db.ExecContext(ctx, query, username, email, hash)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresDB) GetUser(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash FROM users WHERE email=$1`
	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.HashPass)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFoundUser
		}
		return nil, err
	}
	return &user, nil
}
