package email

import (
	"Online-queue-management-system/libs/config"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func TestEmailQueue_SendsAndAcknowledges(t *testing.T) {
	client := newTestRedis(t)
	sender := &fakeSender{}
	queue := startQueue(t, client, sender, 2, 5)

	enqueue(t, queue, "a@example.com", "b@example.com")

	waitFor(t, func() bool { return sender.sentCount() == 2 })
	waitFor(t, func() bool { return streamLen(t, client, StreamKey) == 0 })
	assertNoPending(t, client)
}

func TestEmailQueue_RetriesFailedEmail(t *testing.T) {
	client := newTestRedis(t)
	sender := &fakeSender{failures: 2}
	queue := startQueue(t, client, sender, 1, 5)

	enqueue(t, queue, "retry@example.com")

	waitFor(t, func() bool { return sender.sentCount() == 1 })
	if got := sender.attemptCount(); got != 3 {
		t.Fatalf("expected 3 attempts (2 failures + success), got %d", got)
	}
	waitFor(t, func() bool { return streamLen(t, client, StreamKey) == 0 })
	assertNoPending(t, client)
}

func TestEmailQueue_MovesToDeadLetterAfterMaxAttempts(t *testing.T) {
	client := newTestRedis(t)
	sender := &fakeSender{failures: 100}
	queue := startQueue(t, client, sender, 1, 2)

	enqueue(t, queue, "broken@example.com")

	waitFor(t, func() bool { return streamLen(t, client, DeadStreamKey) == 1 })
	waitFor(t, func() bool { return streamLen(t, client, StreamKey) == 0 })
	assertNoPending(t, client)

	dead, err := client.XRange(context.Background(), DeadStreamKey, "-", "+").Result()
	if err != nil {
		t.Fatal(err)
	}
	if dead[0].Values[fieldTo] != "broken@example.com" || dead[0].Values["error"] == "" {
		t.Fatalf("unexpected dead letter: %v", dead[0].Values)
	}
}

// An email taken by a process that died before sending must be delivered by another process.
func TestEmailQueue_ReclaimsEmailOfCrashedConsumer(t *testing.T) {
	client := newTestRedis(t)
	ctx := context.Background()

	if err := client.XGroupCreateMkStream(ctx, StreamKey, GroupName, "0").Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.XAdd(ctx, &goredis.XAddArgs{
		Stream: StreamKey,
		Values: map[string]any{fieldTo: "orphan@example.com", fieldSubject: "s"},
	}).Err(); err != nil {
		t.Fatal(err)
	}
	// The "crashed" consumer reads the email and never acknowledges it.
	if err := client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group: GroupName, Consumer: "crashed", Streams: []string{StreamKey, ">"}, Count: 1,
	}).Err(); err != nil {
		t.Fatal(err)
	}

	sender := &fakeSender{}
	startQueue(t, client, sender, 1, 5)

	waitFor(t, func() bool { return sender.sentCount() == 1 })
	if sender.sent[0].To != "orphan@example.com" {
		t.Fatalf("unexpected email: %+v", sender.sent[0])
	}
	assertNoPending(t, client)
}

// Two replicas share one consumer group: every email is sent exactly once.
func TestEmailQueue_ReplicasDeliverEachEmailOnce(t *testing.T) {
	client := newTestRedis(t)
	sender := &fakeSender{}
	first := startQueue(t, client, sender, 2, 5)
	startQueue(t, client, sender, 2, 5)

	recipients := []string{"1@x.io", "2@x.io", "3@x.io", "4@x.io", "5@x.io", "6@x.io"}
	enqueue(t, first, recipients...)

	waitFor(t, func() bool { return sender.sentCount() == len(recipients) })
	time.Sleep(100 * time.Millisecond)

	seen := map[string]int{}
	for _, msg := range sender.sentMessages() {
		seen[msg.To]++
	}
	for _, to := range recipients {
		if seen[to] != 1 {
			t.Fatalf("email to %s sent %d times", to, seen[to])
		}
	}
}

func TestEmailQueue_EnqueueFailsWhenRedisIsDown(t *testing.T) {
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	queue := startQueue(t, client, &fakeSender{}, 0, 5)
	server.Close()

	if err := queue.Enqueue(context.Background(), EmailMessage{To: "x@example.com"}); err == nil {
		t.Fatal("expected error when redis is unavailable")
	}
}

func newTestRedis(t *testing.T) *goredis.Client {
	t.Helper()
	server := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func startQueue(t *testing.T, client *goredis.Client, sender Sender, workers, maxAttempts int) *EmailQueue {
	t.Helper()
	queue, err := NewEmailQueue(context.Background(), client, sender, config.QueueConfig{
		NumWorkers:  workers,
		RateLimit:   time.Millisecond,
		WrkTimeOut:  time.Second,
		RetryAfter:  20 * time.Millisecond,
		MaxAttempts: maxAttempts,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(queue.Shutdown)
	return queue
}

func enqueue(t *testing.T, queue *EmailQueue, recipients ...string) {
	t.Helper()
	for _, to := range recipients {
		if err := queue.Enqueue(context.Background(), EmailMessage{To: to, Subject: "Код", Body: "123456"}); err != nil {
			t.Fatal(err)
		}
	}
}

func streamLen(t *testing.T, client *goredis.Client, key string) int64 {
	t.Helper()
	n, err := client.XLen(context.Background(), key).Result()
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func assertNoPending(t *testing.T, client *goredis.Client) {
	t.Helper()
	waitFor(t, func() bool {
		pending, err := client.XPending(context.Background(), StreamKey, GroupName).Result()
		return err == nil && pending.Count == 0
	})
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met in time")
}

type fakeSender struct {
	mu       sync.Mutex
	failures int
	attempts int
	sent     []EmailMessage
}

func (s *fakeSender) SendEmail(_ context.Context, msg EmailMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attempts++
	if s.failures > 0 {
		s.failures--
		return errors.New("smtp unavailable")
	}
	s.sent = append(s.sent, msg)
	return nil
}

func (s *fakeSender) sentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

func (s *fakeSender) attemptCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.attempts
}

func (s *fakeSender) sentMessages() []EmailMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]EmailMessage(nil), s.sent...)
}
