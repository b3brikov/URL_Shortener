package postgres

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresDB struct {
	db *sql.DB
}

func NewPostgresDB(db *sql.DB) *postgresDB {
	return &postgresDB{
		db: db,
	}
}

func (r *postgresDB) CreateURL(ctx context.Context, original_url, short_code string, userID int) error {
	query := `INSERT INTO urls (original_url,short_code,user_id) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, query, original_url, short_code, userID)

	return err
}

func (r *postgresDB) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_code=$1`
	var url string

	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(&url)

	if err != nil {
		return "", err
	}
	return url, nil
}

func (r *postgresDB) Batch(ctx context.Context, flushed map[string]int) error {
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

func (r *postgresDB) CreateUser(ctx context.Context, username, email, hash string) error {
	query := `INSERT INTO users (username,email,password_hash) VALUES ($1,$2,$3)`

	_, err := r.db.ExecContext(ctx, query, username, email, hash)
	if err != nil {
		return err
	}

	return nil
}
