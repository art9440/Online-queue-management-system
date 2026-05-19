package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const scheduledNotificationsKey = "notifications:scheduled"

type NotificationRepository struct {
	client *goredis.Client
}

func NewNotificationRepository(client *goredis.Client) *NotificationRepository {
	return &NotificationRepository{client: client}
}

func (r *NotificationRepository) CountDue(ctx context.Context, now time.Time) (int64, error) {
	return r.client.ZCount(ctx, scheduledNotificationsKey, "-inf", timestampScore(now)).Result()
}

func timestampScore(value time.Time) string {
	return strconv.FormatInt(value.UTC().Unix(), 10)
}
