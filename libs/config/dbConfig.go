package config

import (
	"context"
	"fmt"
)

type DBConfig struct {
	DSN      string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

func LoadDBConfig(ctx context.Context) (*DBConfig, error) {
	host, err := MustGet(ctx, "DB_HOST")
	if err != nil {
		return nil, err
	}

	port, err := GetInt(ctx, "DB_PORT")
	if err != nil {
		return nil, err
	}

	name, err := MustGet(ctx, "POSTGRES_DB")
	if err != nil {
		return nil, err
	}

	user, err := MustGet(ctx, "DB_USER")
	if err != nil {
		return nil, err
	}

	password, err := MustGet(ctx, "DB_PASSWORD")
	if err != nil {
		return nil, err
	}

	sslMode := Get(ctx, "DB_SSLMODE", "disable")
	dsn := Get(ctx, "DB_DSN", "")
	if dsn == "" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			user,
			password,
			host,
			port,
			name,
			sslMode,
		)
	}

	return &DBConfig{
		DSN:      dsn,
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
		SSLMode:  sslMode,
	}, nil
}
