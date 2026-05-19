package main

import (
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/booking/config"
	"Online-queue-management-system/services/booking/internal/infrastructure/app"
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
		slog.Error("something went wrong while running booking service", "err", err)
		stop()
		os.Exit(1)
	}

	stop()
}

func run(ctx context.Context) error {
	log := logger.From(ctx)
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Error("error loading config", "err", err)
		return err
	}

	dbCfg, err := libconfig.LoadDBConfig(ctx)
	if err != nil {
		log.Error("error loading db config", "err", err)
		return err
	}

	bookingApp, err := app.NewApp(ctx, cfg, dbCfg)
	if err != nil {
		log.Error("error creating booking app", "err", err)
		return err
	}
	defer bookingApp.Close()

	if err := bookingApp.Run(ctx); err != nil {
		log.Error("error starting booking service", "err", err)
		return err
	}

	return nil
}
