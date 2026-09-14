package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type ClickRepository interface {
	Batch(ctx context.Context, flushed map[string]int) error
}

type Cache interface {
	FlushClicks(ctx context.Context) (map[string]int, error)
}

type ClickWorker struct {
	wg       sync.WaitGroup
	once     sync.Once
	stop     chan struct{}
	done     chan struct{}
	interval time.Duration
	cache    Cache
	repo     ClickRepository
	Logger   *slog.Logger
}

func (c *ClickWorker) Run() {
	c.wg.Add(1)
	go c.worker()

	c.Logger.Debug("worker started")

}

func (c *ClickWorker) Shutdown(ctx context.Context) error {
	called := false
	c.once.Do(func() {
		called = true
		close(c.stop)
	})
	if !called {
		return errors.New("already shutdown")
	}

	exit := make(chan struct{})

	go func() {
		c.wg.Wait()
		close(exit)
	}()

	select {
	case <-ctx.Done():
		close(c.done)
		<-exit
		return errors.New("context expired")
	case <-exit:
		return nil
	}
}

func NewClickWorker(repo ClickRepository, cache Cache, interval time.Duration, logger *slog.Logger) *ClickWorker {
	return &ClickWorker{
		interval: interval,
		repo:     repo,
		Logger:   logger,
		wg:       sync.WaitGroup{},
		once:     sync.Once{},
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		cache:    cache,
	}
}

func (c *ClickWorker) flush(ctx context.Context) error {
	defer func() {
		if p := recover(); p != nil {
			slog.Error("worker panic")
		}
	}()
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

func (c *ClickWorker) worker() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	defer c.wg.Done()
	for {
		select {
		case <-c.done:
			return

		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := c.flush(ctx); err != nil {
				c.Logger.Debug("flush error", slog.Any("error", err.Error()), slog.Any("time", time.Now()), slog.Any("from", "ClickWorker"))
			}
			cancel()
		case <-c.stop:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := c.flush(ctx); err != nil {
				c.Logger.Debug("flush error", slog.Any("error", err.Error()), slog.Any("time", time.Now()), slog.Any("from", "ClickWorker"))
			}
			cancel()
			return
		}
	}
}
