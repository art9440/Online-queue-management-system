package main

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/bot/config"
	"Online-queue-management-system/services/bot/internal/infrastructure/app"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})

	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	ctx = logger.With(ctx, log)

	if err := run(ctx); err != nil {
		slog.Error("telegram bot stopped with error", "err", err)
		stop()
		os.Exit(1)
	}

	stop()
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		return err
	}

	botApp, err := app.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer botApp.Close()

	return botApp.Run(ctx)
}
