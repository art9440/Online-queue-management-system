package config

import (
	libconfig "Online-queue-management-system/libs/config"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AuthPort string

	DBCfg libconfig.DBConfig

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTAccessSecret  string
	JWTRefreshSecret string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	CookieSecure bool
}

func Load(ctx context.Context) (Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("parse REDIS_DB: %w", err)
	}

	accessTTL, err := time.ParseDuration(mustEnv("ACCESS_TOKEN_TTL"))
	if err != nil {
		return Config{}, fmt.Errorf("parse ACCESS_TOKEN_TTL: %w", err)
	}

	refreshTTL, err := time.ParseDuration(mustEnv("REFRESH_TOKEN_TTL"))
	if err != nil {
		return Config{}, fmt.Errorf("parse REFRESH_TOKEN_TTL: %w", err)
	}

	cookieSecure, err := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse COOKIE_SECURE: %w", err)
	}

	dbCfg, err := libconfig.LoadDBConfig(ctx)
	if err != nil {
		return Config{}, fmt.Errorf("load db config: %w", err)
	}

	return Config{
		AuthPort: getEnv("AUTH_PORT", "8082"),

		DBCfg: *dbCfg,

		RedisAddr:     mustEnv("REDIS_ADDR"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		JWTAccessSecret:  mustEnv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: mustEnv("JWT_REFRESH_SECRET"),

		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
		CookieSecure: cookieSecure,
	}, nil
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env: " + key)
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
