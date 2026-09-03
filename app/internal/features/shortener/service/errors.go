package service

import "errors"

var (
	CannotCreateNewUnique = errors.New("cannot create new unique code")
	ErrCodeNotFound       = errors.New("short code not found")
	ErrUnexpectedError    = errors.New("unexpected error")
	ErrTimeOut            = errors.New("timeout by context")
)
