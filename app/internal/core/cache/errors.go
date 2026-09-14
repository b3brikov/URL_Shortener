package cache

import "errors"

var (
	ErrMissingValue error = errors.New("missing cache value")
)
