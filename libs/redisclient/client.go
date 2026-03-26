package redisclient

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

func New(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Client{rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}
