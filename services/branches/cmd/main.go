package main

import (
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/tracing"
	branchesConfig "Online-queue-management-system/services/branches/config"
	"Online-queue-management-system/services/branches/internal/infrastructure/app"
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
	shutdownTracing := tracing.InitFromEnv(ctx, "branches", log)
	defer shutdownTracing()

	if err := run(ctx); err != nil {
		slog.Error("something went wrong while running branches service", "err", err)
		return 1
	}

	return 0
}

func run(ctx context.Context) error {
	log := logger.From(ctx)
	cfg, err := branchesConfig.LoadConfig(ctx)
	if err != nil {
		log.Error("error loading config", "err", err)
		return err
	}

	dbCfg, err := libconfig.LoadDBConfig(ctx)
	if err != nil {
		log.Error("error loading db config", "err", err)
		return err
	}

	branchesApp, err := app.NewApp(ctx, *cfg, dbCfg)
	if err != nil {
		log.Error("error creating branches app", "err", err)
		return err
	}

	if err := branchesApp.Run(ctx); err != nil {
		log.Error("error starting branches service", "err", err)
		return err
	}

	return nil
}
