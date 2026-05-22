package main

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/tracing"
	"Online-queue-management-system/services/bot/config"
	"Online-queue-management-system/services/bot/internal/infrastructure/app"
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
		JSON:   false,
		Source: true,
	})

	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx = logger.With(ctx, log)
	shutdownTracing := tracing.InitFromEnv(ctx, "bot", log)
	defer shutdownTracing()

	if err := run(ctx); err != nil {
		slog.Error("telegram bot stopped with error", "err", err)
		return 1
	}

	return 0
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
