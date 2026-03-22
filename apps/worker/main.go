package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnouarMohamed/StateSight/internal/config"
	"github.com/AnouarMohamed/StateSight/internal/jobs"
	"github.com/AnouarMohamed/StateSight/internal/observability"
	"github.com/AnouarMohamed/StateSight/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker service failed", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadWorker()
	if err != nil {
		return fmt.Errorf("load worker config: %w", err)
	}
	logger := observability.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer pool.Close()

	queue, err := jobs.NewRedisQueue(cfg.RedisURL, cfg.PollTimeout)
	if err != nil {
		return fmt.Errorf("create redis queue: %w", err)
	}
	defer queue.Close()

	processor := jobs.NewProcessor(storage.NewRepository(pool), logger)

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("worker consuming queue")
		if err := queue.Consume(workerCtx, processor.HandleMessage); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("queue consumer failed: %w", err)
	}

	cancel()
	time.Sleep(250 * time.Millisecond)
	logger.Info("worker shutdown complete")
	return nil
}
