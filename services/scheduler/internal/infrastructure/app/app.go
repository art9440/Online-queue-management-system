package app

import (
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/redisclient"
	"Online-queue-management-system/services/scheduler/config"
	"Online-queue-management-system/services/scheduler/internal/application/service"
	"Online-queue-management-system/services/scheduler/internal/infrastructure/bot"
	emaildispatcher "Online-queue-management-system/services/scheduler/internal/infrastructure/email"
	redisrepo "Online-queue-management-system/services/scheduler/internal/infrastructure/redis"
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	scheduler *service.Scheduler
	redis     *goredis.Client
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log := logger.From(ctx)

	rdb, err := redisclient.New(ctx, cfg.RedisCfg, 5*time.Second)
	if err != nil {
		log.Error("failed to connect redis", "err", err)
		return nil, err
	}

	if err := redisclient.WaitForRedis(ctx, rdb); err != nil {
		_ = rdb.Close()
		log.Error("redis not ready", "err", err)
		return nil, fmt.Errorf("redis not ready: %w", err)
	}

	repo := redisrepo.NewNotificationRepository(rdb)
	telegramDispatcher := bot.NewDispatcher(cfg.BotURL)
	emailSender := email.NewEmailSender(cfg.EmailSenderCfg)
	emailDispatcher := emaildispatcher.NewDispatcher(emailSender)
	scheduler := service.New(repo, telegramDispatcher, emailDispatcher, cfg.PollInterval, cfg.BatchSize)

	return &App{
		scheduler: scheduler,
		redis:     rdb,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.scheduler.Run(ctx)
}

func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
}
