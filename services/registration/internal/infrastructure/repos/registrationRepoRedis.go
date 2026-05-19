package repos

import (
	"Online-queue-management-system/services/registration/internal/domain"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	pendingPrefix  = "registration:pending"
	recoveryPrefix = "registration:recovery"
	ttl            = 10 * time.Minute
)

type RegistrationRepoRedis struct {
	client *redis.Client
}

func NewRegistrationRepoRedis(client *redis.Client) *RegistrationRepoRedis {
	return &RegistrationRepoRedis{
		client: client,
	}
}

func (r *RegistrationRepoRedis) Save(ctx context.Context, p *pending.PendingRegistration) error {
	key := fmt.Sprintf("%s:%s", pendingPrefix, p.ID)

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
	key := fmt.Sprintf("%s:%s", pendingPrefix, id)

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
	key := fmt.Sprintf("%s:%s", pendingPrefix, id)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis delete: %w", err)
	}

	return nil
}

func (r *RegistrationRepoRedis) GetAndValidate(ctx context.Context, id, code string) (pending.PendingRegistration, error) {
	key := fmt.Sprintf("%s:%s", pendingPrefix, id)

	script := redis.NewScript(`
		local val = redis.call("GET", KEYS[1])
		if not val then
			return nil
		end

		local data = cjson.decode(val)

		if tostring(data.code) ~= ARGV[1] then
			return "INVALID_CODE"
		end

		redis.call("DEL", KEYS[1])
		return val
	`)

	res, err := script.Run(ctx, r.client, []string{key}, code).Result()
	if err != nil {
		return pending.PendingRegistration{}, err
	}

	if res == nil {
		return pending.PendingRegistration{}, domain.ErrNotFound
	}

	if str, ok := res.(string); ok && str == "INVALID_CODE" {
		return pending.PendingRegistration{}, domain.ErrInvalidCode
	}

	var p pending.PendingRegistration
	if err := json.Unmarshal([]byte(res.(string)), &p); err != nil {
		return pending.PendingRegistration{}, err
	}

	return p, nil
}

func (r *RegistrationRepoRedis) SaveRecovery(ctx context.Context, item recovery.PasswordRecovery) error {
	key := fmt.Sprintf("%s:%s", recoveryPrefix, item.ID)

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal recovery: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set recovery: %w", err)
	}

	return nil
}

func (r *RegistrationRepoRedis) GetRecovery(ctx context.Context, recoveryID string) (recovery.PasswordRecovery, error) {
	key := fmt.Sprintf("%s:%s", recoveryPrefix, recoveryID)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return recovery.PasswordRecovery{}, fmt.Errorf("recovery not found")
		}
		return recovery.PasswordRecovery{}, fmt.Errorf("redis get recovery: %w", err)
	}

	var item recovery.PasswordRecovery
	if err := json.Unmarshal([]byte(val), &item); err != nil {
		return recovery.PasswordRecovery{}, fmt.Errorf("unmarshal recovery: %w", err)
	}

	return item, nil
}

func (r *RegistrationRepoRedis) DeleteRecovery(ctx context.Context, recoveryID string) error {
	key := fmt.Sprintf("%s:%s", recoveryPrefix, recoveryID)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete recovery: %w", err)
	}

	return nil
}
