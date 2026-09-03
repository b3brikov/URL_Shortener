package service

import "time"

type Config struct {
	CodeLen  int
	MaxRetry int
	CodeTtl  time.Duration
}
