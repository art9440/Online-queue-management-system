package main

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"Online-queue-management-system/services/auth/internal/infrastructure/app"
)

func main() {
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})

	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("failed to initialize auth app", "err", err)
		os.Exit(1)
	}
	defer a.Close()

	if err := a.Run(ctx); err != nil {
		slog.Error("auth app stopped with error", "err", err)
		os.Exit(1)
	}
}
