package config

import (
	"Online-queue-management-system/libs/config"
	"context"
	"fmt"
)

type BookingConfig struct {
	BookingPort     string
	JWTAccessSecret string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	GoogleClientID  string
	GoogleSecret    string
	GoogleRedirect  string
}

func LoadConfig(ctx context.Context) (*BookingConfig, error) {
	bookingPort, err := config.MustGet(ctx, "BOOKING_PORT")
	if err != nil {
		return nil, err
	}
	jwtAccessSecret, err := config.MustGet(ctx, "JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}
	redisAddr := config.Get(ctx, "REDIS_ADDR", "")
	if redisAddr == "" {
		redisAddr = fmt.Sprintf(
			"%s:%s",
			config.Get(ctx, "REDIS_HOST", "redis"),
			config.Get(ctx, "REDIS_PORT", "6379"),
		)
	}
	redisPassword := config.Get(ctx, "REDIS_PASSWORD", "")
	redisDB := config.GetIntDefault(ctx, "REDIS_DB", 0)

	return &BookingConfig{
		BookingPort:     bookingPort,
		JWTAccessSecret: jwtAccessSecret,
		RedisAddr:       redisAddr,
		RedisPassword:   redisPassword,
		RedisDB:         redisDB,
		GoogleClientID:  config.Get(ctx, "GOOGLE_CLIENT_ID", ""),
		GoogleSecret:    config.Get(ctx, "GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirect:  config.Get(ctx, "GOOGLE_REDIRECT_URL", ""),
	}, nil
}
