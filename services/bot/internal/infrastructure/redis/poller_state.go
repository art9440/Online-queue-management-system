package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	pollerLeaderKey = "telegram:poller:leader"
	pollerOffsetKey = "telegram:poller:offset"
)

// Telegram allows only one getUpdates consumer per bot token (others get 409 Conflict),
// so bot replicas elect a single poller through this lock (12-factor VIII). The value is the
// owner's id: only the owner may renew or release it, and a crashed owner's lock simply expires.
var (
	renewLockScript = goredis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)
	releaseLockScript = goredis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
	return redis.call('DEL', KEYS[1])
end
return 0
`)
)

type LeaderLock struct {
	client *goredis.Client
	owner  string
	ttl    time.Duration
}

func NewLeaderLock(client *goredis.Client, owner string, ttl time.Duration) *LeaderLock {
	return &LeaderLock{client: client, owner: owner, ttl: ttl}
}

// TryAcquire takes the lock if nobody holds it.
func (l *LeaderLock) TryAcquire(ctx context.Context) (bool, error) {
	err := l.client.SetArgs(ctx, pollerLeaderKey, l.owner, goredis.SetArgs{Mode: "NX", TTL: l.ttl}).Err()
	if errors.Is(err, goredis.Nil) {
		return false, nil // held by another replica
	}
	return err == nil, err
}

// Renew extends the lock; false means it was lost (expired and taken by another replica).
func (l *LeaderLock) Renew(ctx context.Context) (bool, error) {
	n, err := renewLockScript.Run(ctx, l.client, []string{pollerLeaderKey}, l.owner, l.ttl.Milliseconds()).Int64()
	return n == 1, err
}

// Release frees the lock right away so another replica does not wait for the TTL.
func (l *LeaderLock) Release(ctx context.Context) error {
	return releaseLockScript.Run(ctx, l.client, []string{pollerLeaderKey}, l.owner).Err()
}

// OffsetStore keeps the Telegram update offset outside the process (12-factor VI),
// so a restarted or new leader continues exactly where the previous one stopped.
type OffsetStore struct {
	client *goredis.Client
}

func NewOffsetStore(client *goredis.Client) *OffsetStore {
	return &OffsetStore{client: client}
}

func (s *OffsetStore) Load(ctx context.Context) (int64, error) {
	offset, err := s.client.Get(ctx, pollerOffsetKey).Int64()
	if errors.Is(err, goredis.Nil) {
		return 0, nil
	}
	return offset, err
}

func (s *OffsetStore) Save(ctx context.Context, offset int64) error {
	return s.client.Set(ctx, pollerOffsetKey, offset, 0).Err()
}
