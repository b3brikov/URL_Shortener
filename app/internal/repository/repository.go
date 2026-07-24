package repository

import (
	"URLShortener/internal/models"
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"time"
)

var ErrNoAffectedRows = errors.New("no affected rows")

type Cache interface {
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	AddClick(ctx context.Context, key string) error
	FlushClicks(ctx context.Context) (map[string]int, error)
	GetValue(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type Repository struct {
	db     *sql.DB
	cache  Cache
	ttl    time.Duration
	logger *slog.Logger
}

func NewRepository(db *sql.DB, cache Cache, ttl time.Duration, logger *slog.Logger) *Repository {
	return &Repository{
		db:     db,
		cache:  cache,
		ttl:    ttl,
		logger: logger,
	}
}

// Создание нового с ретраями
func (r *Repository) CreateURL(ctx context.Context, original_url, short_code string, userID int) error {
	query := `INSERT INTO urls (original_url,short_code,user_id) VALUES ($1, $2, $3)`

	res, err := r.db.ExecContext(ctx, query, original_url, short_code, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNoAffectedRows
	}
	return nil
}

// Получение по короткому коду
func (r *Repository) GetOriginalURL(ctx context.Context, shortCode string) (string, error) {
	resCache, err := r.cache.GetValue(ctx, shortCode)
	if err == nil {
		log.Println("Value from cache: ", resCache)
		go func() {
			ctxInc, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := r.cache.AddClick(ctxInc, shortCode); err != nil {
				r.logger.Error("cannot increment click to %s: %s", shortCode, err.Error())
			} else {
				r.logger.Info("Added new click!")
			}
		}()
		return resCache, nil
	}

	r.logger.Info("cache miss for %s: %v", shortCode, err)

	query := `SELECT original_url FROM urls WHERE short_code=$1`

	row := r.db.QueryRowContext(ctx, query, shortCode)

	var res string

	err = row.Scan(&res)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrCodeNotFound
		}
		return "", err
	}
	go func() {
		ctxInc, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := r.cache.AddClick(ctxInc, shortCode); err != nil {
			r.logger.Error("cannot increment click to %s: %s", shortCode, err.Error())
		} else {
			r.logger.Info("Added new click!")
		}
	}()

	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := r.cache.Set(cacheCtx, shortCode, res, r.ttl); err != nil {
			r.logger.Error("cannot add value to cache", slog.Any("error", err.Error()))
		}

	}()
	return res, nil
}

func (r *Repository) CreateUser(ctx context.Context, username, email, hash string) error {
	query := `INSERT INTO users (username,email,password_hash) VALUES ($1,$2,$3)`

	res, err := r.db.ExecContext(ctx, query, username, email, hash)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNoAffectedRows
	}
	return nil
}

func (r *Repository) SetToken(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.cache.Set(ctx, key, value, ttl)
}

func (r *Repository) GetToken(ctx context.Context, key string) (string, error) {
	return r.cache.GetValue(ctx, key)
}

func (r *Repository) DeleteToken(ctx context.Context, key string) error {
	return r.cache.Delete(ctx, key)
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

func (r *Repository) Batch(ctx context.Context, flushed map[string]int) error {
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

	return tx.Commit()
}

func (r *Repository) FlushClicks(ctx context.Context) (map[string]int, error) {
	return r.cache.FlushClicks(ctx)
}
