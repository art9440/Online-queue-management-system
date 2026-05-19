package config

import (
	"Online-queue-management-system/libs/config"
	"context"
)

type BranchesConfig struct {
	BranchesPort    string
	JWTAccessSecret string
}

type Config struct {
	BranchesCfg BranchesConfig
}

func LoadConfig(ctx context.Context) (*Config, error) {
	// branches config
	branchesPort, err := config.MustGet(ctx, "BRANCHES_PORT")
	if err != nil {
		return nil, err
	}
	jwtAccessSecret, err := config.MustGet(ctx, "JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	branchesCfg := BranchesConfig{
		BranchesPort:    branchesPort,
		JWTAccessSecret: jwtAccessSecret,
	}

	return &Config{
		BranchesCfg: branchesCfg,
	}, nil
}
