package config

import (
	"Online-queue-management-system/libs/config"
	"context"
)

type BookingConfig struct {
	BookingPort     string
	JWTAccessSecret string
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

	return &BookingConfig{
		BookingPort:     bookingPort,
		JWTAccessSecret: jwtAccessSecret,
	}, nil
}
