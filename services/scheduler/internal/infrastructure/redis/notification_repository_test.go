package redis

import (
	"Online-queue-management-system/services/scheduler/internal/application/service"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestClaimDue_MovesOnlyDueNotificationsToProcessing(t *testing.T) {
	client, repo := newTestRepo(t)
	now := time.Now()
	schedule(t, client, "due", "email", now.Add(-time.Minute))
	schedule(t, client, "future", "email", now.Add(time.Hour))

	claimed, err := repo.ClaimDue(context.Background(), now, 10, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	if len(claimed) != 1 || claimed[0].ID != "due" {
		t.Fatalf("unexpected claimed notifications: %+v", claimed)
	}
	assertMembers(t, client, scheduledNotificationsKey, "future")
	assertMembers(t, client, processingNotificationsKey, "due")

	// A second scheduler replica gets nothing: the notification is already taken.
	again, err := repo.ClaimDue(context.Background(), now, 10, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(again) != 0 {
		t.Fatalf("notification claimed twice: %+v", again)
	}
}

func TestRequeueExpired_ReturnsNotificationsOfCrashedScheduler(t *testing.T) {
	client, repo := newTestRepo(t)
	now := time.Now()
	schedule(t, client, "n-1", "email", now.Add(-time.Minute))

	// Claimed by a scheduler that then crashed: the lease expires.
	if _, err := repo.ClaimDue(context.Background(), now, 10, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	requeued, err := repo.RequeueExpired(context.Background(), now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if requeued != 1 {
		t.Fatalf("expected 1 requeued notification, got %d", requeued)
	}
	assertMembers(t, client, scheduledNotificationsKey, "n-1")
	assertMembers(t, client, processingNotificationsKey)
}

func TestRetryAndFinish(t *testing.T) {
	client, repo := newTestRepo(t)
	ctx := context.Background()
	now := time.Now()
	schedule(t, client, "n-1", "email", now.Add(-time.Minute))
	schedule(t, client, "n-2", "email", now.Add(-time.Minute))
	if _, err := repo.ClaimDue(ctx, now, 10, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	retryAt := now.Add(5 * time.Minute)
	if err := repo.Retry(ctx, "n-1", retryAt, "smtp down"); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkSent(ctx, "n-2", now); err != nil {
		t.Fatal(err)
	}

	assertMembers(t, client, processingNotificationsKey)
	assertMembers(t, client, scheduledNotificationsKey, "n-1")
	score, _ := client.ZScore(ctx, scheduledNotificationsKey, "n-1").Result()
	if int64(score) != retryAt.Unix() {
		t.Fatalf("retry score = %v, want %v", int64(score), retryAt.Unix())
	}

	retried := payload(t, client, "n-1")
	if retried.Status != "retrying" || retried.Attempts != 1 || retried.FailReason != "smtp down" {
		t.Fatalf("unexpected retried payload: %+v", retried)
	}
	if ttl := client.TTL(ctx, notificationKey("n-1")).Val(); ttl != -1 {
		t.Fatalf("pending notification must not expire, ttl=%v", ttl)
	}

	if sent := payload(t, client, "n-2"); sent.Status != "sent" {
		t.Fatalf("unexpected sent payload: %+v", sent)
	}
	if ttl := client.TTL(ctx, notificationKey("n-2")).Val(); ttl <= 0 {
		t.Fatalf("finished notification must expire, ttl=%v", ttl)
	}
}

func TestClaimDue_DropsNotificationWithoutPayload(t *testing.T) {
	client, repo := newTestRepo(t)
	now := time.Now()
	if err := client.ZAdd(context.Background(), scheduledNotificationsKey, goredis.Z{
		Score: float64(now.Add(-time.Minute).Unix()), Member: "orphan",
	}).Err(); err != nil {
		t.Fatal(err)
	}

	claimed, err := repo.ClaimDue(context.Background(), now, 10, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(claimed) != 0 {
		t.Fatalf("expected nothing to dispatch, got %+v", claimed)
	}
	assertMembers(t, client, processingNotificationsKey)
}

// Two scheduler replicas on one Redis: every notification is dispatched exactly once.
func TestTwoSchedulers_DispatchEachNotificationOnce(t *testing.T) {
	client, repo := newTestRepo(t)
	now := time.Now()
	const total = 50
	for i := range total {
		schedule(t, client, fmt.Sprintf("n-%d", i), "email", now.Add(-time.Minute))
	}

	dispatcher := &countingDispatcher{seen: map[string]int{}}
	cfg := service.Config{PollInterval: 20 * time.Millisecond, BatchSize: 7, Lease: time.Minute, MaxAttempts: 3, RetryDelay: time.Minute}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for range 2 {
		scheduler := service.New(repo, nil, dispatcher, cfg)
		go func() { _ = scheduler.Run(ctx) }()
	}

	deadline := time.Now().Add(5 * time.Second)
	for dispatcher.total() < total && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)

	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	if len(dispatcher.seen) != total {
		t.Fatalf("dispatched %d distinct notifications, want %d", len(dispatcher.seen), total)
	}
	for id, n := range dispatcher.seen {
		if n != 1 {
			t.Fatalf("notification %s dispatched %d times", id, n)
		}
	}
}

func newTestRepo(t *testing.T) (*goredis.Client, *NotificationRepository) {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client, NewNotificationRepository(client)
}

func schedule(t *testing.T, client *goredis.Client, id, channel string, at time.Time) {
	t.Helper()
	ctx := context.Background()
	raw, _ := json.Marshal(notificationPayload{ID: id, Channel: channel, Email: "c@example.com", Status: "scheduled"})
	if err := client.Set(ctx, notificationKey(id), raw, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.ZAdd(ctx, scheduledNotificationsKey, goredis.Z{Score: float64(at.Unix()), Member: id}).Err(); err != nil {
		t.Fatal(err)
	}
}

func assertMembers(t *testing.T, client *goredis.Client, key string, want ...string) {
	t.Helper()
	got, err := client.ZRange(context.Background(), key, 0, -1).Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", key, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: got %v, want %v", key, got, want)
		}
	}
}

func payload(t *testing.T, client *goredis.Client, id string) notificationPayload {
	t.Helper()
	raw, err := client.Get(context.Background(), notificationKey(id)).Result()
	if err != nil {
		t.Fatal(err)
	}
	var p notificationPayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

type countingDispatcher struct {
	mu   sync.Mutex
	seen map[string]int
}

func (d *countingDispatcher) Dispatch(_ context.Context, n *service.Notification) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen[n.ID]++
	return nil
}

func (d *countingDispatcher) total() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}
