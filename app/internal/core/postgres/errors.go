package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFoundUser   = errors.New("user not found")
	ErrCodeNotFound   = errors.New("short code not found")
	ErrNoAffectedRows = errors.New("no affected rows")
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
