package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := buildDSN()
	migrationsDir := getEnv("MIGRATIONS_DIR", "migrations")

	db, err := waitForDB(dsn, 30, 2*time.Second)
	if err != nil {
		log.Fatalf("database is not ready: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose set dialect: %v", err)
	}

	log.Println("running migrations...")
	if err := goose.Up(db, migrationsDir); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		log.Fatalf("goose up: %v", err)
	}

	log.Println("migrations applied successfully")
}

func buildDSN() string {
	host := mustEnv("DB_HOST")
	port := mustEnv("DB_PORT")
	user := mustEnv("DB_USER")
	password := mustEnv("DB_PASSWORD")
	dbname := mustEnv("POSTGRES_DB")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode,
	)
}

func waitForDB(dsn string, attempts int, delay time.Duration) (*sql.DB, error) {
	var lastErr error

	for i := 1; i <= attempts; i++ {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			lastErr = err
			log.Printf("db open failed (attempt %d/%d): %v", i, attempts, err)
			time.Sleep(delay)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err = db.PingContext(ctx)
		cancel()

		if err == nil {
			log.Printf("database is ready on attempt %d/%d", i, attempts)
			return db, nil
		}

		lastErr = err
		_ = db.Close()
		log.Printf("db ping failed (attempt %d/%d): %v", i, attempts, err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("could not connect to db after %d attempts: %w", attempts, lastErr)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("environment variable %s is required", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}