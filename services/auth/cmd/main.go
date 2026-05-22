package main

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/tracing"
	"context"
	"log/slog"
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
	ctx = logger.With(ctx, log)
	shutdownTracing := tracing.InitFromEnv(ctx, "auth", log)
	defer shutdownTracing()

	a, err := app.New(ctx)
	if err != nil {
		slog.Error("failed to initialize auth app", "err", err)
		stop()
		return
	}
	defer a.Close()

	if err := a.Run(ctx); err != nil {
		slog.Error("auth app stopped with error", "err", err)
		stop()
		return
	}
}
