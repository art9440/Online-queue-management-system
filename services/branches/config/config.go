package config

import (
	"Online-queue-management-system/libs/config"
	"context"
)

type BranchesConfig struct {
	BranchesPort string
}

type Config struct {
	RedisCfg    config.RedisConfig
	BranchesCfg BranchesConfig
}

func LoadConfig(ctx context.Context) (*Config, error) {
	//redis config
	redisAddr, err := config.MustGet(ctx, "REDIS_ADDR")
	if err != nil {
		return nil, err
	}

	redisPassword, err := config.MustGet(ctx, "REDIS_PASSWORD")
	if err != nil {
		return nil, err
	}

	redisDB, err := config.GetInt(ctx, "REDIS_DB")
	if err != nil {
		return nil, err
	}

	redisCfg := config.RedisConfig{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
	}
	//branches config
	branchesPort, err := config.MustGet(ctx, "BRANCHES_PORT")
	if err != nil {
		return nil, err
	}
	branchesCfg := BranchesConfig{
		BranchesPort: branchesPort,
	}
	return &Config{
		RedisCfg:    redisCfg,
		BranchesCfg: branchesCfg,
	}, nil
}
