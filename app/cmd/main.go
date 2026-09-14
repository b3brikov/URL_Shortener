package main

import (
	"URLShortener/di"
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	app := &di.DI{}

	server := app.Server()
	worker := app.ClickWorker()

	go worker.Run()
	go server.Run()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	go server.Shutdown(shutdownCtx)
	if err := worker.Shutdown(shutdownCtx); err != nil {
		log.Panicln("воркер не доработал все записи")
	}
}
