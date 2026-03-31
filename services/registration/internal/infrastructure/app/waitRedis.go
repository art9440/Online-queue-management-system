package app

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func waitForRedis(ctx context.Context, client *redis.Client) error {
	for i := 0; i < 10; i++ {
		err := client.Ping(ctx).Err()
		if err == nil {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("redis not available after retries")
}
