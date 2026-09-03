package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type ClickRepository interface {
	Batch(ctx context.Context, flushed map[string]int) error
}

type Cache interface {
	FlushClicks(ctx context.Context) (map[string]int, error)
}

type ClickWorker struct {
	interval time.Duration
	cache    Cache
	repo     ClickRepository
	Logger   *slog.Logger
}

func NewClickWorker(repo ClickRepository, cache Cache, interval time.Duration, logger *slog.Logger) *ClickWorker {
	return &ClickWorker{
		interval: interval,
		repo:     repo,
		Logger:   logger,
	}
}

func (c *ClickWorker) Flush(ctx context.Context) error {
	data, err := c.cache.FlushClicks(ctx)
	if err != nil {
		return err
	}
	if err := c.repo.Batch(ctx, data); err != nil {
		return err
	}
	fmt.Println("service flushed", data)
	return nil
}

func (c *ClickWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := c.Flush(ctx); err != nil {
				c.Logger.Debug("flush error", slog.Any("error", err.Error()), slog.Any("time", time.Now()), slog.Any("from", "ClickWorker"))
			}
		}
	}
}
