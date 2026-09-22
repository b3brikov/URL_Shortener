package main

import (
	"URLShortener/di"
	"context"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	app := &di.DI{}

	server := app.Server()
	worker := app.ClickWorker()
	errCh := make(chan error, 3)

	go worker.Run()
	go func() { errCh <- server.Run() }()
	go func() { errCh <- app.RunGRPC() }()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	select {
	case <-ctx.Done():

	case err := <-errCh:
		app.Logger().Error("server failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Shutdown(shutdownCtx); err != nil {
			app.Logger().Error("http shutdown", "err", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		done := make(chan struct{})
		go func() {
			app.GrpcServer().GracefulStop()
			close(done)
		}()
		select {
		case <-done:
		case <-shutdownCtx.Done():
			app.GrpcServer().Stop()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := worker.Shutdown(shutdownCtx); err != nil {
			app.Logger().Error("worker shutdown", "err", err)
		}
	}()

	wg.Wait()
}
