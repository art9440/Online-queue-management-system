package app

import (
	"Online-queue-management-system/libs/auth"
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/middleware"
	"Online-queue-management-system/services/booking/config"
	"Online-queue-management-system/services/booking/internal/application/service"
	httpserver "Online-queue-management-system/services/booking/internal/infrastructure/httpServer"
	"Online-queue-management-system/services/booking/internal/infrastructure/repos"
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type BookingApp struct {
	svc        *service.BookingService
	httpServer *http.Server
}

func NewApp(
	ctx context.Context,
	cfg config.BookingConfig,
	dbCfg libconfig.DBConfig,
) (*BookingApp, error) {
	log := logger.From(ctx)

	repoPostgres, err := repos.NewBookingRepoPostgres(dbCfg.DSN)
	if err != nil {
		log.Error("error creating booking repo", "err", err)
		return nil, err
	}

	svc := service.New(repoPostgres)

	serverImpl := httpserver.NewHttpServer(svc)

	parser := auth.NewTokenParser(cfg.JWTAccessSecret)

	mux := http.NewServeMux()

	authMiddleware := auth.Middleware(parser)

	mux.Handle(
		"POST /appointments",
		authMiddleware(http.HandlerFunc(serverImpl.CreateAppointment)),
	)

	mux.Handle(
		"GET /appointments",
		authMiddleware(http.HandlerFunc(serverImpl.GetAppointments)),
	)

	mux.Handle(
		"GET /appointments/{id}",
		authMiddleware(http.HandlerFunc(serverImpl.GetAppointmentByID)),
	)

	mux.Handle(
		"PATCH /appointments/{id}/cancel",
		authMiddleware(http.HandlerFunc(serverImpl.CancelAppointment)),
	)

	// public health endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsMux := middleware.CORSMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + cfg.BookingPort,
		Handler: middleware.RequestLogger(corsMux),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &BookingApp{
		httpServer: server,
		svc:        svc,
	}, nil
}

func (a *BookingApp) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info("starting booking service")

	srv := a.httpServer
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("SERVER PANIC", "err", err)
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
		return err
	}

	log.Info("booking service stopped")
	return nil
}
