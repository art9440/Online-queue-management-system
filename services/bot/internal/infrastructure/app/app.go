package app

import (
	libconfig "Online-queue-management-system/libs/config"
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
	"net"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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
		lock := redisrepo.NewLeaderLock(rdb, instanceID(), pollerLockTTL)
		offsets := redisrepo.NewOffsetStore(rdb)
		poller = polling.New(telegramClient, bot, lock, offsets, cfg.PollingTimeout, cfg.PollingInterval, pollerLockTTL)
	} else {
		bot = service.New(repo, nil)
		log.Warn("telegram bot token is not set, bot polling and sending are disabled")
	}

	serverImpl := httpserver.New(bot, enabled)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", serverImpl.Health)
	mux.HandleFunc("POST /telegram/notifications", serverImpl.SendNotification)

	httpServer := &http.Server{
		Addr:    ":" + cfg.BotPort,
		Handler: middleware.RequestLogger(middleware.CORSMiddleware(libconfig.LoadCORSAllowedOrigins(ctx), mux)),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
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

	pollerDone := make(chan struct{})
	if a.poller != nil {
		go func() {
			defer close(pollerDone)
			a.poller.Run(ctx)
		}()
	} else {
		close(pollerDone)
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

	// Wait for the poller to release the leader lock, so another replica takes over at once
	// instead of waiting for the lock TTL (12-factor IX).
	select {
	case <-pollerDone:
	case <-shutdownCtx.Done():
		log.Warn("telegram poller did not stop in time")
	}

	log.Info("telegram bot service stopped")
	return nil
}

func (a *App) Close() {
	if a.redis != nil {
		_ = a.redis.Close()
	}
}

// pollerLockTTL: if the polling replica dies, another one takes over within this time.
const pollerLockTTL = 30 * time.Second

// instanceID identifies this replica as the owner of the poller lock.
func instanceID() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "bot"
	}
	return host + "-" + uuid.NewString()[:8]
}
