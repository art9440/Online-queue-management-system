package app

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/redisclient"
	"Online-queue-management-system/services/registration/config"
	"Online-queue-management-system/services/registration/internal/application/email"
	"Online-queue-management-system/services/registration/internal/application/queue"
	"Online-queue-management-system/services/registration/internal/application/service"
	httpserver "Online-queue-management-system/services/registration/internal/infrastructure/httpServer"
	"Online-queue-management-system/services/registration/internal/infrastructure/repos"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type App struct {
	svc        *service.RegistrationService
	httpServer *http.Server
	emailQueue *queue.EmailQueue
}

func NewApp(ctx context.Context, cfg config.Config, dbCfg config.DBConfig) (*App, error) {
	log := logger.From(ctx)
	redisClient, err := redisclient.New(ctx, cfg, 5*time.Second)

	if err := waitForRedis(ctx, redisClient); err != nil {
		log.Error("redis not ready", "err", err)
		return nil, fmt.Errorf("redis not ready: %w", err)
	}

	repoRedis := repos.NewRegistrationRepoRedis(redisClient)
	repoPostgres, err := repos.NewRegistrationRepoPostgres(dbCfg.DSN)
	if err != nil {
		log.Error("error creating registration repo", "err", err)
		return nil, err
	}
	emailSender := email.NewEmailSender(cfg)
	emailQueue := queue.NewEmailQueue(emailSender, 10)
	svc := service.NewRegistrationService(repoRedis, repoPostgres, emailQueue)

	serverImpl := httpserver.NewHttpServer(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("/register", httpserver.RecoverMiddleware(serverImpl.Register))
	mux.HandleFunc("/verify", httpserver.RecoverMiddleware(serverImpl.Verify))

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	//для теста
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	httpServer := &http.Server{
		Addr:    ":" + cfg.RegistrationPort,
		Handler: mux,
	}

	return &App{svc: svc,
		httpServer: httpServer,
		emailQueue: emailQueue}, nil
}

func (a *App) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info("starting registration service")

	srv := a.httpServer

	errCh := make(chan error, 1)

	// запускаем сервер
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("SERVER PANIC:", err)
			}
		}()
		log.Info("http server started", "addr", srv.Addr)

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("ListenAndServe failed", "err", err)
			errCh <- err
		}

	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")

	case err := <-errCh:
		log.Error("http server crashed", "err", err)
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("shutting down http server")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown http server", "err", err)
		return fmt.Errorf("shutdown failed: %w", err)
	}
	log.Info("http server stopped")

	log.Info("shutting down email queue", "pending_emails", func() int {
		len, _, _ := a.emailQueue.GetStats()
		return len
	}())

	a.emailQueue.Shutdown()
	log.Info("email queue stopped")

	log.Info("registration service stopped")
	return nil
}
