package redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *goredis.Client
	ttl time.Duration
}

func NewSessionRepository(rdb *goredis.Client, ttl time.Duration) *SessionRepository {
	return &SessionRepository{
		rdb: rdb,
		ttl: ttl,
	}
}

func (r *SessionRepository) SaveRefreshSession(ctx context.Context, jti string, userID int64) error {
	key := "auth:refresh:" + jti
	return r.rdb.Set(ctx, key, strconv.FormatInt(userID, 10), r.ttl).Err()
}

func (r *SessionRepository) RefreshSessionExists(ctx context.Context, jti string, userID int64) (bool, error) {
	key := "auth:refresh:" + jti

	value, err := r.rdb.Get(ctx, key).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return value == strconv.FormatInt(userID, 10), nil
}

func (r *SessionRepository) DeleteRefreshSession(ctx context.Context, jti string) error {
	key := "auth:refresh:" + jti
	return r.rdb.Del(ctx, key).Err()
}