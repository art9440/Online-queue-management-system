package repos

import (
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	prefix = "registration:pending"
	ttl    = 10 * time.Minute
)

type RegistrationRepoRedis struct {
	client *redis.Client
}

func NewRegistrationRepoRedis(client *redis.Client) *RegistrationRepoRedis {
	return &RegistrationRepoRedis{
		client: client,
	}
}

func (r *RegistrationRepoRedis) Save(ctx context.Context, p pending.PendingRegistration) error {
	key := fmt.Sprintf("%s:%s", prefix, p.ID)

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal pending: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

func (r *RegistrationRepoRedis) Get(ctx context.Context, id string) (pending.PendingRegistration, error) {
	key := fmt.Sprintf("%s:%s", prefix, id)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return pending.PendingRegistration{}, fmt.Errorf("registration not found")
		}
		return pending.PendingRegistration{}, fmt.Errorf("redis get: %w", err)
	}

	var p pending.PendingRegistration
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return pending.PendingRegistration{}, fmt.Errorf("unmarshal pending: %w", err)
	}

	return p, nil
}

func (r *RegistrationRepoRedis) Delete(ctx context.Context, id string) error {
	key := fmt.Sprintf("%s:%s", prefix, id)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}

	return nil
}
