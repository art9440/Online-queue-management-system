package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestLeaderLock_OnlyOneOwner(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	first := NewLeaderLock(client, "first", time.Minute)
	second := NewLeaderLock(client, "second", time.Minute)

	if ok, err := first.TryAcquire(ctx); err != nil || !ok {
		t.Fatalf("first acquire: ok=%v err=%v", ok, err)
	}
	if ok, err := second.TryAcquire(ctx); err != nil || ok {
		t.Fatalf("second must not acquire a held lock: ok=%v err=%v", ok, err)
	}

	// Only the owner can renew or release.
	if ok, _ := second.Renew(ctx); ok {
		t.Fatal("non-owner renewed the lock")
	}
	if err := second.Release(ctx); err != nil {
		t.Fatal(err)
	}
	if got, _ := client.Get(ctx, pollerLeaderKey).Result(); got != "first" {
		t.Fatalf("non-owner released the lock, owner is now %q", got)
	}
	if ok, err := first.Renew(ctx); err != nil || !ok {
		t.Fatalf("owner renew: ok=%v err=%v", ok, err)
	}

	if err := first.Release(ctx); err != nil {
		t.Fatal(err)
	}
	if ok, _ := second.TryAcquire(ctx); !ok {
		t.Fatal("lock must be free after the owner released it")
	}
	_ = server
}

func TestLeaderLock_ExpiredLockIsTakenOver(t *testing.T) {
	server, client := newTestRedis(t)
	ctx := context.Background()
	crashed := NewLeaderLock(client, "crashed", 30*time.Second)
	follower := NewLeaderLock(client, "follower", 30*time.Second)

	if ok, _ := crashed.TryAcquire(ctx); !ok {
		t.Fatal("acquire failed")
	}
	server.FastForward(31 * time.Second) // the owner died and stopped renewing

	if ok, _ := follower.TryAcquire(ctx); !ok {
		t.Fatal("follower must take over an expired lock")
	}
	if ok, _ := crashed.Renew(ctx); ok {
		t.Fatal("old owner must not renew a lock taken by another replica")
	}
}

func TestOffsetStore(t *testing.T) {
	_, client := newTestRedis(t)
	ctx := context.Background()
	store := NewOffsetStore(client)

	if offset, err := store.Load(ctx); err != nil || offset != 0 {
		t.Fatalf("empty store: offset=%d err=%v", offset, err)
	}
	if err := store.Save(ctx, 42); err != nil {
		t.Fatal(err)
	}
	if offset, _ := store.Load(ctx); offset != 42 {
		t.Fatalf("offset = %d, want 42", offset)
	}
}

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *goredis.Client) {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return server, client
}
