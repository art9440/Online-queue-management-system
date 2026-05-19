package config

import (
	libconfig "Online-queue-management-system/libs/config"
	"context"
	"time"
)

const (
	defaultBotPort         = "8085"
	defaultPollingTimeout  = 30
	defaultPollingInterval = time.Second
)

type Config struct {
	RedisCfg        libconfig.RedisConfig
	TelegramToken   string
	BotPort         string
	PollingTimeout  int
	PollingInterval time.Duration
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
		TelegramToken:   libconfig.Get(ctx, "TELEGRAM_BOT_TOKEN", ""),
		BotPort:         libconfig.Get(ctx, "BOT_PORT", defaultBotPort),
		PollingTimeout:  libconfig.GetIntDefault(ctx, "TELEGRAM_POLLING_TIMEOUT", defaultPollingTimeout),
		PollingInterval: libconfig.GetDurationDefault(ctx, "TELEGRAM_POLLING_INTERVAL", defaultPollingInterval),
	}, nil
}
