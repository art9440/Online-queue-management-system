package app

import (
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/services/booking/config"
	"context"
)

type App struct {
}

func NewApp(ctx context.Context, cfg config.BookingConfig, dbCfg libconfig.DBConfig) (*App, error) {
	return &App{}, nil
}

func (a *App) Run(ctx context.Context) error {
	return nil
}
