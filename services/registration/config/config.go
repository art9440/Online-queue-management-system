package config

import (
	"Online-queue-management-system/libs/config"
	"context"
)

type Config struct {
	AuthURL  string
	RedisURL string
}

func LoadConfig(ctx context.Context) (*Config, error) {
	authURL, err := config.MustGet(ctx, "AUTH_URL")
	if err != nil {
		return nil, err
	}

	redisURL, err := config.MustGet(ctx, "REDIS_URL")
	if err != nil {
		return nil, err
	}

	return &Config{
		AuthURL:  authURL,
		RedisURL: redisURL,
	}, nil
}

type DBConfig struct {
	DSN      string
	Host     string
	Port     int
	User     string
	Password string
	SSLMode  string
}

func LoadDBConfig(ctx context.Context) (*DBConfig, error) {
	host, err := config.MustGet(ctx, "DB_HOST")
	if err != nil {
		return nil, err
	}

	port, err := config.GetInt(ctx, "DB_PORT")
	if err != nil {
		return nil, err
	}

	user, err := config.MustGet(ctx, "DB_USER")
	if err != nil {
		return nil, err
	}

	ssl, err := config.MustGet(ctx, "DB_SSLMODE")
	if err != nil {
		return nil, err
	}

	password, err := config.MustGet(ctx, "DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	dsn := config.Get(ctx, "DB_DSN", "")
	return &DBConfig{
		DSN:      dsn,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		SSLMode:  ssl,
	}, nil

}
