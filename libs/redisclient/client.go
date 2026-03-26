package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func New(ctx context.Context, cfg config.Config, timeout time.Duration) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, timeout*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}