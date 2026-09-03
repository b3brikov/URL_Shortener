package redis

import (
	"URLShortener/internal/core/models"
	"context"
	"fmt"
	"strings"
	"time"

	rds "github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	Redis *rds.Client
}

func NewRedisClient(Addr string) (*RedisStorage, error) {
	rdb := rds.NewClient(&rds.Options{
		Addr:     Addr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStorage{
		Redis: rdb,
	}, nil
}

func (r *RedisStorage) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.Redis.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStorage) GetValue(ctx context.Context, key string) (string, error) {
	res := r.Redis.Get(ctx, key)
	if res.Err() != nil {
		return "", res.Err()
	}
	return res.Val(), nil
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	return r.Redis.Del(ctx, key).Err()
}

func (r *RedisStorage) AddClick(ctx context.Context, key string) error {
	return r.Redis.Incr(ctx, models.ClickNameSpace+key).Err()
}

func (r *RedisStorage) FlushClicks(ctx context.Context) (map[string]int, error) {
	keys, err := r.Redis.Keys(ctx, models.ClickNameSpace+"*").Result()
	if err != nil {
		return map[string]int{}, fmt.Errorf("failed to get keys: %w", err)
	}
	if len(keys) == 0 {
		return map[string]int{}, nil
	}
	pipe := r.Redis.Pipeline()
	values := make(map[string]*rds.StringCmd)
	result := make(map[string]int)
	for _, key := range keys {
		values[key] = pipe.GetDel(ctx, key)
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != rds.Nil {
		return map[string]int{}, fmt.Errorf("pipeline exec: %w", err)
	}

	for k, v := range values {
		val, err := v.Int()
		if err == rds.Nil {
			continue
		}
		if err != nil {
			return map[string]int{}, err
		}
		result[strings.TrimPrefix(k, models.ClickNameSpace)] = val
	}

	return result, nil
}
