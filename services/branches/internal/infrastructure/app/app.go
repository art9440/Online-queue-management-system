package app

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/redisclient"
	branchesConfig "Online-queue-management-system/services/branches/config"
	"Online-queue-management-system/services/branches/internal/infrastructure/repos"
	"Online-queue-management-system/services/registration/config"
	"context"
	"fmt"
	"time"
)

type BranchesApp struct {
}

func NewApp(ctx context.Context, cfg branchesConfig.Config, dbCfg config.DBConfig) (*BranchesApp, error) {
	log := logger.From(ctx)
	redisClient, err := redisclient.New(ctx, cfg.RedisCfg, 5*time.Second)

	if err := redisclient.WaitForRedis(ctx, redisClient); err != nil {
		log.Error("redis not ready", "err", err)
		return nil, fmt.Errorf("redis not ready: %w", err)
	}

	repoRedis := repos.NewRegistrationRepoRedis(redisClient)
	repoPostgres, err := repos.NewRegistrationRepoPostgres(dbCfg.DSN)
	if err != nil {
		log.Error("error creating registration repo", "err", err)
		return nil, err
	}

	return &BranchesApp{}, nil
}

func (a *BranchesApp) Run(ctx context.Context) error {

}
