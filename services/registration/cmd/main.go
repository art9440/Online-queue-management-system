package main

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	log := logger.New(logger.Config{
		Level:  slog.LevelInfo,
		JSON:   false,
		Source: true,
	})

	slog.SetDefault(log)

	if err := godotenv.Load(".env"); err != nil {
		slog.Warn(".env not found, using OS env", "err", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	ctx = logger.With(ctx, log)

	if err := run(ctx); err != nil {
		slog.Error("something went wrong while running scrapper", "err", err)
		stop()
		os.Exit(1)
	}

	stop()
}

func run(ctx context.Context) error {
	for{

	}
}
