package config

import (
	libconfig "Online-queue-management-system/libs/config"
	"context"
	"time"
)

const (
	defaultPollInterval = 30 * time.Second
	defaultBatchSize    = 100
	defaultBotURL       = "http://bot:8085"
)

type Config struct {
	RedisCfg     libconfig.RedisConfig
	PollInterval time.Duration
	BatchSize    int
	BotURL       string
}

func LoadConfig(ctx context.Context) (*Config, error) {
	redisAddr, err := libconfig.MustGet(ctx, "REDIS_ADDR")
	if err != nil {
		return nil, err
	}

	redisPassword, err := libconfig.MustGet(ctx, "REDIS_PASSWORD")
	if err != nil {
		return nil, err
	}

	redisDB, err := libconfig.GetInt(ctx, "REDIS_DB")
	if err != nil {
		return nil, err
	}

	return &Config{
		RedisCfg: libconfig.RedisConfig{
			RedisAddr:     redisAddr,
			RedisPassword: redisPassword,
			RedisDB:       redisDB,
		},
		PollInterval: libconfig.GetDurationDefault(ctx, "SCHEDULER_POLL_INTERVAL", defaultPollInterval),
		BatchSize:    libconfig.GetIntDefault(ctx, "SCHEDULER_BATCH_SIZE", defaultBatchSize),
		BotURL:       libconfig.Get(ctx, "BOT_URL", defaultBotURL),
	}, nil
}
