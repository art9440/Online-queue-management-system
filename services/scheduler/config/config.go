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
	RedisCfg       libconfig.RedisConfig
	EmailSenderCfg libconfig.EmailSenderConfig
	PollInterval   time.Duration
	BatchSize      int
	BotURL         string
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

	smtpHost, err := libconfig.MustGet(ctx, "SMTP_HOST")
	if err != nil {
		return nil, err
	}

	smtpPort, err := libconfig.GetInt(ctx, "SMTP_PORT")
	if err != nil {
		return nil, err
	}

	smtpUser, err := libconfig.MustGet(ctx, "SMTP_USER")
	if err != nil {
		return nil, err
	}

	smtpPass, err := libconfig.MustGet(ctx, "SMTP_PASS")
	if err != nil {
		return nil, err
	}

	return &Config{
		RedisCfg: libconfig.RedisConfig{
			RedisAddr:     redisAddr,
			RedisPassword: redisPassword,
			RedisDB:       redisDB,
		},
		EmailSenderCfg: libconfig.EmailSenderConfig{
			SMTPHost:    smtpHost,
			SMTPPort:    smtpPort,
			SMTPUser:    smtpUser,
			SMTPPass:    smtpPass,
			SendTimeOut: libconfig.GetDurationDefault(ctx, "EMAIL_TIMEOUT", 20*time.Second),
		},
		PollInterval: libconfig.GetDurationDefault(ctx, "SCHEDULER_POLL_INTERVAL", defaultPollInterval),
		BatchSize:    libconfig.GetIntDefault(ctx, "SCHEDULER_BATCH_SIZE", defaultBatchSize),
		BotURL:       libconfig.Get(ctx, "BOT_URL", defaultBotURL),
	}, nil
}
