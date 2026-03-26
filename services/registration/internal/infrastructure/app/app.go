package app

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/config"
	"Online-queue-management-system/services/registration/internal/infrastructure/repos"
	"context"
)

type App struct {
}

func NewApp(ctx context.Context, cfg config.Config, dbCfg config.DBConfig) (*App, error) {
	log := logger.From(ctx)
	repoRedis, err := repos.NewRegistrationRepoRedis(log, cfg.RedisURL)
	repoPostgres, err := repos.NewRegistrationRepo(log, dbCfg.DSN)
	if err != nil {
		log.Error("error creating registration repo", "err", err)
		return nil, err
	}

	return &App{}, nil
}

func (a *App) Run(ctx context.Context) error {
	return nil
}
