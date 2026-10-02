package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testConfig() Config {
	return Config{
		PollInterval: time.Second,
		BatchSize:    100,
		Lease:        time.Minute,
		MaxAttempts:  3,
		RetryDelay:   time.Minute,
	}
}

func TestTick_WhenTelegramNotificationIsDue_ShouldDispatchAndMarkSent(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{{ID: "n-1", Channel: telegramChannel}}}
	telegram := &fakeDispatcher{}
	scheduler := New(repo, telegram, nil, testConfig())

	if err := scheduler.tick(context.Background()); err != nil {
		t.Fatalf("tick returned error: %v", err)
	}

	if telegram.calls != 1 {
		t.Fatalf("expected one dispatch call, got %d", telegram.calls)
	}
	assertIDs(t, "sent", repo.sent, "n-1")
	assertIDs(t, "failed", repo.failed)
	assertIDs(t, "retried", repo.retried)
}

func TestTick_WhenEmailNotificationIsDue_ShouldDispatchAndMarkSent(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{{ID: "n-1", Channel: emailChannel, Email: "client@example.com"}}}
	email := &fakeDispatcher{}
	scheduler := New(repo, &fakeDispatcher{}, email, testConfig())

	if err := scheduler.tick(context.Background()); err != nil {
		t.Fatalf("tick returned error: %v", err)
	}

	if email.calls != 1 {
		t.Fatalf("expected one email dispatch, got %d", email.calls)
	}
	assertIDs(t, "sent", repo.sent, "n-1")
}

func TestTick_WhenDispatchFailsAndAttemptsLeft_ShouldScheduleRetry(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{{ID: "n-1", Channel: telegramChannel, Attempts: 1}}}
	scheduler := New(repo, &fakeDispatcher{err: errors.New("bot unavailable")}, nil, testConfig())

	before := time.Now()
	if err := scheduler.tick(context.Background()); err == nil {
		t.Fatal("expected tick to return dispatch error")
	}

	assertIDs(t, "retried", repo.retried, "n-1")
	assertIDs(t, "failed", repo.failed)
	// Second failure: delay is RetryDelay * attempt number.
	wantAt := before.Add(2 * time.Minute)
	if repo.retryAt.Before(wantAt) || repo.retryAt.After(wantAt.Add(5*time.Second)) {
		t.Fatalf("retry scheduled at %v, want about %v", repo.retryAt, wantAt)
	}
	if repo.failReason != "bot unavailable" {
		t.Fatalf("expected failure reason to be saved, got %q", repo.failReason)
	}
}

func TestTick_WhenDispatchFailsOnLastAttempt_ShouldMarkFailed(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{{ID: "n-1", Channel: emailChannel, Attempts: 2}}}
	scheduler := New(repo, nil, &fakeDispatcher{err: errors.New("smtp unavailable")}, testConfig())

	if err := scheduler.tick(context.Background()); err == nil {
		t.Fatal("expected tick to return dispatch error")
	}

	assertIDs(t, "failed", repo.failed, "n-1")
	assertIDs(t, "retried", repo.retried)
}

func TestTick_WhenChannelIsUnknown_ShouldMarkFailed(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{{ID: "n-1", Channel: "sms"}}}
	scheduler := New(repo, &fakeDispatcher{}, &fakeDispatcher{}, testConfig())

	if err := scheduler.tick(context.Background()); err == nil {
		t.Fatal("expected error for unknown channel")
	}

	assertIDs(t, "failed", repo.failed, "n-1")
}

func TestTick_WhenShuttingDown_ShouldReleaseClaimedNotifications(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{
		{ID: "n-1", Channel: emailChannel},
		{ID: "n-2", Channel: emailChannel},
	}}
	email := &fakeDispatcher{}
	scheduler := New(repo, nil, email, testConfig())

	ctx, cancel := context.WithCancel(context.Background())
	repo.onClaim = cancel // SIGTERM arrives right after the batch is claimed

	_ = scheduler.tick(ctx)

	if email.calls != 0 {
		t.Fatalf("expected no dispatch after shutdown, got %d", email.calls)
	}
	assertIDs(t, "released", repo.released, "n-1", "n-2")
}

func TestTick_WhenLeaseIsAlmostOver_ShouldReleaseTheRest(t *testing.T) {
	repo := &fakeNotificationRepo{notifications: []Notification{
		{ID: "n-1", Channel: emailChannel},
		{ID: "n-2", Channel: emailChannel},
		{ID: "n-3", Channel: emailChannel},
	}}
	cfg := testConfig()
	cfg.Lease = 200 * time.Millisecond // stop processing after 150ms
	email := &fakeDispatcher{delay: 100 * time.Millisecond}
	scheduler := New(repo, nil, email, cfg)

	_ = scheduler.tick(context.Background())

	assertIDs(t, "sent", repo.sent, "n-1", "n-2")
	assertIDs(t, "released", repo.released, "n-3")
}

func TestTick_ShouldRequeueExpiredLeasesFirst(t *testing.T) {
	repo := &fakeNotificationRepo{}
	scheduler := New(repo, nil, nil, testConfig())

	if err := scheduler.tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !repo.requeueCalled {
		t.Fatal("expected expired leases to be requeued")
	}
}

func assertIDs(t *testing.T, what string, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", what, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: got %v, want %v", what, got, want)
		}
	}
}

type fakeNotificationRepo struct {
	notifications []Notification
	onClaim       func()

	requeueCalled bool
	sent          []string
	failed        []string
	retried       []string
	released      []string
	retryAt       time.Time
	failReason    string
}

func (r *fakeNotificationRepo) ClaimDue(context.Context, time.Time, int64, time.Time) ([]Notification, error) {
	if r.onClaim != nil {
		r.onClaim()
	}
	return r.notifications, nil
}

func (r *fakeNotificationRepo) RequeueExpired(context.Context, time.Time) (int64, error) {
	r.requeueCalled = true
	return 0, nil
}

func (r *fakeNotificationRepo) Release(_ context.Context, ids []string, _ time.Time) error {
	r.released = append(r.released, ids...)
	return nil
}

func (r *fakeNotificationRepo) MarkSent(_ context.Context, id string, _ time.Time) error {
	r.sent = append(r.sent, id)
	return nil
}

func (r *fakeNotificationRepo) MarkFailed(_ context.Context, id string, _ time.Time, reason string) error {
	r.failed = append(r.failed, id)
	r.failReason = reason
	return nil
}

func (r *fakeNotificationRepo) Retry(_ context.Context, id string, retryAt time.Time, reason string) error {
	r.retried = append(r.retried, id)
	r.retryAt = retryAt
	r.failReason = reason
	return nil
}

type fakeDispatcher struct {
	calls int
	err   error
	delay time.Duration
}

func (d *fakeDispatcher) Dispatch(context.Context, *Notification) error {
	d.calls++
	time.Sleep(d.delay)
	return d.err
}
