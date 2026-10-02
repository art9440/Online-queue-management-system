package main

import (
	libconfig "Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	log := logger.New(logger.Config{
		Level:   slog.LevelInfo,
		JSON:    true,
		Service: "migrate",
	})
	slog.SetDefault(log)

	ctx := logger.With(context.Background(), log)
	if err := run(ctx); err != nil {
		log.Error("migrations failed", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	log := logger.From(ctx)

	dbCfg, err := libconfig.LoadDBConfig(ctx)
	if err != nil {
		return fmt.Errorf("load db config: %w", err)
	}
	migrationsDir := libconfig.Get(ctx, "MIGRATIONS_DIR", "migrations")

	db, err := waitForDB(ctx, dbCfg.DSN, 30, 2*time.Second)
	if err != nil {
		return fmt.Errorf("database is not ready: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Warn("db close failed", "err", err)
		}
	}()

	goose.SetLogger(gooseLogger{log: log})
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}

	log.Info("running migrations", "dir", migrationsDir)
	if err := goose.UpContext(ctx, db, migrationsDir); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		return fmt.Errorf("goose up: %w", err)
	}

	version, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		return fmt.Errorf("goose get version: %w", err)
	}

	log.Info("migrations applied successfully", "version", version)
	return nil
}

func waitForDB(ctx context.Context, dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	log := logger.From(ctx)
	var lastErr error

	for i := 1; i <= attempts; i++ {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			lastErr = err
			log.Warn("db open failed", "attempt", i, "attempts", attempts, "err", err)
			time.Sleep(delay)
			continue
		}

		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = db.PingContext(pingCtx)
		cancel()

		if err == nil {
			log.Info("database is ready", "attempt", i)
			return db, nil
		}

		lastErr = err
		_ = db.Close()
		log.Warn("db ping failed", "attempt", i, "attempts", attempts, "err", err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("could not connect to db after %d attempts: %w", attempts, lastErr)
}

// gooseLogger routes goose output (applied migrations) into the structured stdout log.
type gooseLogger struct {
	log *slog.Logger
}

func (l gooseLogger) Printf(format string, v ...any) {
	l.log.Info(strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (l gooseLogger) Fatalf(format string, v ...any) {
	l.log.Error(strings.TrimSpace(fmt.Sprintf(format, v...)))
	os.Exit(1)
}
