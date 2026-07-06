package worker

import (
	"context"
	"fmt"
	"time"
)

type ClickRepository interface {
	Batch(ctx context.Context, flushed map[string]int) error
	FlushClicks(ctx context.Context) (map[string]int, error)
}

type ClickWorker struct {
	interval time.Duration
	repo     ClickRepository
}

func NewClickWorker(repo ClickRepository, interval time.Duration) *ClickWorker {
	return &ClickWorker{
		interval: interval,
		repo:     repo,
	}
}

func (c *ClickWorker) Fluch(ctx context.Context) error {
	data, err := c.repo.FlushClicks(ctx)
	if err != nil {
		return err
	}
	return c.repo.Batch(ctx, data)
}

func (c *ClickWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := c.Fluch(ctx); err != nil {
				fmt.Println("click worker error:", err)
			}
		}
	}
}
