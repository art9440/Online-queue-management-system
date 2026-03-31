package app

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"Online-queue-management-system/services/auth/internal/application/service"
	"Online-queue-management-system/services/auth/internal/infrastructure/config"
	httpapi "Online-queue-management-system/services/auth/internal/infrastructure/http"
	"Online-queue-management-system/services/auth/internal/infrastructure/jwt"
	"Online-queue-management-system/services/auth/internal/infrastructure/postgres"
	redisrepo "Online-queue-management-system/services/auth/internal/infrastructure/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
)

type App struct {
	server *http.Server
	db     *pgxpool.Pool
	redis  *goredis.Client
}

func New(ctx context.Context) (*App, error) {
	log := logger.From(ctx)

	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load auth config", "err", err)
		return nil, err
	}

	db, err := newPostgres(ctx, cfg)
	if err != nil {
		log.Error("failed to connect postgres", "err", err)
		return nil, err
	}

	rdb, err := newRedis(ctx, cfg)
	if err != nil {
		log.Error("failed to connect redis", "err", err)
		db.Close()
		return nil, err
	}

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := redisrepo.NewSessionRepository(rdb, cfg.RefreshTTL)
	tokenManager := jwt.New(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.AccessTTL,
		cfg.RefreshTTL,
	)

	authService := service.New(userRepo, sessionRepo, tokenManager)
	cookieManager := httpapi.NewCookieManager(cfg.CookieSecure)
	handler := httpapi.NewHandler(authService, cookieManager, cfg.AccessTTL, cfg.RefreshTTL)

	mux := http.NewServeMux()
	handler.Register(mux)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":" + cfg.AuthPort,
		Handler:           httpapi.RequestLogger(mux),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server: server,
		db:     db,
		redis:  rdb,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	log := logger.From(ctx)
	errCh := make(chan error, 1)

	go func() {
		log.Info("starting auth service", "addr", a.server.Addr)
		err := a.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error("auth server stopped with error", "err", err)
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			log.Error("failed to shutdown auth server", "err", err)
			return err
		}
		log.Info("auth service stopped")
		return nil
	case err := <-errCh:
		return err
	}
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
	if a.redis != nil {
		_ = a.redis.Close()
	}
}

func newPostgres(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func newRedis(ctx context.Context, cfg config.Config) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
