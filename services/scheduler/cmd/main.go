package main

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/scheduler/config"
	"Online-queue-management-system/services/scheduler/internal/infrastructure/app"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   true,
		Source: true,
	})

	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx = logger.With(ctx, log)

	if err := run(ctx); err != nil {
		slog.Error("scheduler stopped with error", "err", err)
		return 1
	}

	return 0
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return err
	}

	schedulerApp, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer schedulerApp.Close()

	return schedulerApp.Run(ctx)
}
