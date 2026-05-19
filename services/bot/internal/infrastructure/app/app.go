package app

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/middleware"
	"Online-queue-management-system/libs/redisclient"
	"Online-queue-management-system/services/bot/config"
	"Online-queue-management-system/services/bot/internal/application/service"
	httpserver "Online-queue-management-system/services/bot/internal/infrastructure/httpserver"
	"Online-queue-management-system/services/bot/internal/infrastructure/polling"
	redisrepo "Online-queue-management-system/services/bot/internal/infrastructure/redis"
	"Online-queue-management-system/services/bot/internal/infrastructure/telegram"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	httpServer *http.Server
	poller     *polling.Poller
	redis      *goredis.Client
	enabled    bool
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	log := logger.From(ctx)

	rdb, err := redisclient.New(ctx, cfg.RedisCfg, 5*time.Second)
	if err != nil {
		log.Error("failed to connect redis", "err", err)
		return nil, err
	}

	repo := redisrepo.NewChatRepository(rdb)

	var bot *service.Bot
	var poller *polling.Poller
	enabled := cfg.TelegramToken != ""
	if enabled {
		telegramClient := telegram.NewClient(cfg.TelegramToken)
		bot = service.New(repo, telegramClient)
		poller = polling.New(telegramClient, bot, cfg.PollingTimeout, cfg.PollingInterval)
	} else {
		bot = service.New(repo, nil)
		log.Warn("telegram bot token is not set, bot polling and sending are disabled")
	}

	serverImpl := httpserver.New(bot, enabled)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", serverImpl.Health)
	mux.HandleFunc("POST /telegram/notifications", serverImpl.SendNotification)

	httpServer := &http.Server{
		Addr:              ":" + cfg.BotPort,
		Handler:           middleware.RequestLogger(middleware.CORSMiddleware(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		httpServer: httpServer,
		poller:     poller,
		redis:      rdb,
		enabled:    enabled,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info("starting telegram bot service", "addr", a.httpServer.Addr, "enabled", a.enabled)

	errCh := make(chan error, 1)

	if a.poller != nil {
		go a.poller.Run(ctx)
	}

	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("telegram bot shutdown signal received")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown telegram bot http server: %w", err)
	}

	log.Info("telegram bot service stopped")
	return nil
}

func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
}
