package service

import (
	"context"
	"time"
)

type service struct {
	cache Cache
}

func (r *service) setAccessToken(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.cache.Set(ctx, key, value, ttl)
}

func (r *service) setRefreshToken(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.cache.Set(ctx, key, value, ttl)
}

func (r *service) getAccessToken(ctx context.Context, key string) (string, error) {
	return r.cache.GetValue(ctx, key)
}

func (r *service) getRefreshToken(ctx context.Context, key string) (string, error) {
	return r.cache.GetValue(ctx, key)
}

func (r *service) deleteAccessToken(ctx context.Context, key string) error {
	return r.cache.Delete(ctx, key)
}

func (r *service) deleteRefreshToken(ctx context.Context, key string) error {
	return r.cache.Delete(ctx, key)
}
