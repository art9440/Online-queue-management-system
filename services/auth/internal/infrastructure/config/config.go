package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AuthPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTAccessSecret  string
	JWTRefreshSecret string

	AccessTTL  time.Duration
	RefreshTTL time.Duration

	CookieSecure bool
}

func Load() (Config, error) {
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

	return Config{
		AuthPort: getEnv("AUTH_PORT", "8082"),

		DBHost:     mustEnv("DB_HOST"),
		DBPort:     mustEnv("DB_PORT"),
		DBUser:     mustEnv("DB_USER"),
		DBPassword: mustEnv("DB_PASSWORD"),
		DBName:     mustEnv("DB_NAME"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

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