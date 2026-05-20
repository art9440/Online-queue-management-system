package app

import (
	"Online-queue-management-system/libs/auth"
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/libs/middleware"
	branchesConfig "Online-queue-management-system/services/branches/config"
	"Online-queue-management-system/services/branches/internal/application/service"
	httpserver "Online-queue-management-system/services/branches/internal/infrastructure/httpServer"
	"Online-queue-management-system/services/branches/internal/infrastructure/repos"
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type BranchesApp struct {
	svc        *service.BranchesService
	httpServer *http.Server
}

func NewApp(ctx context.Context, cfg branchesConfig.Config, dbCfg *libconfig.DBConfig) (*BranchesApp, error) {
	log := logger.From(ctx)

	repoPostgres, err := repos.NewBranchesRepoPostgres(dbCfg.DSN)
	if err != nil {
		log.Error("error creating branches repo", "err", err)
		return nil, err
	}
	svc := service.New(repoPostgres)

	serverImpl := httpserver.NewHttpServer(svc)
	log.Info("JWT secret", "secret", cfg.BranchesCfg.JWTAccessSecret)
	parser := auth.NewTokenParser(cfg.BranchesCfg.JWTAccessSecret)
	mux := http.NewServeMux()

	authMiddleware := auth.Middleware(parser)

	mux.Handle("GET /branches", authMiddleware(http.HandlerFunc(serverImpl.GetBranches)))
	mux.Handle("GET /branches/{id}/clients", authMiddleware(http.HandlerFunc(serverImpl.GetBranchClients)))
	mux.Handle("GET /branches/{id}/bookings", authMiddleware(http.HandlerFunc(serverImpl.GetBranchAppointments)))
	mux.Handle("/branches", authMiddleware(http.HandlerFunc(serverImpl.GetBranches)))
	mux.Handle("/branches/{id}/employees",
		authMiddleware(http.HandlerFunc(serverImpl.GetBranchEmployees)))
	mux.Handle("/services",
		authMiddleware(http.HandlerFunc(serverImpl.GetServices)))
	mux.Handle("/services/{serviceId}/branches",
		authMiddleware(http.HandlerFunc(serverImpl.GetBranchesWithService)))
	mux.Handle("/services/{serviceId}/branches/{branchId}/employees",
		authMiddleware(http.HandlerFunc(serverImpl.GetEmployeesForService)))
	mux.Handle("GET /businesses/{id}/registration-slug",
		authMiddleware(http.HandlerFunc(serverImpl.GetBusinessRegistrationSlug)))

	// Public endpoints - no authentication required
	mux.Handle("/public/{registrationSlug}/services",
		http.HandlerFunc(serverImpl.GetPublicServices))
	mux.Handle("/public/{registrationSlug}/services/{serviceId}/branches",
		http.HandlerFunc(serverImpl.GetPublicBranchesWithService))
	mux.Handle("/public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees",
		http.HandlerFunc(serverImpl.GetPublicEmployeesForService))

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	CorsMux := middleware.CORSMiddleware(mux)

	server := &http.Server{
		Addr:    ":" + cfg.BranchesCfg.BranchesPort,
		Handler: middleware.TraceRequests(middleware.RequestLogger(CorsMux)),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &BranchesApp{
		httpServer: server,
		svc:        svc,
	}, nil
}

func (a *BranchesApp) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info("starting branches service")

	srv := a.httpServer
	errCh := make(chan error, 1)

	// запуск сервера
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

	// ждём либо shutdown, либо падение сервера
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")

	case err := <-errCh:
		log.Error("http server crashed", "err", err)
		return err
	}

	// graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("shutting down http server")

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown http server", "err", err)
		return err
	}

	log.Info("branches service stopped")
	return nil
}
