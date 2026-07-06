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

func (c *ClickWorker) Flush(ctx context.Context) error {
	data, err := c.repo.FlushClicks(ctx)
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
				fmt.Println("click worker error:", err)
			}
		}
	}
}
