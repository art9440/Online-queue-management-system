package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTick_WhenTelegramNotificationIsDue_ShouldDispatchAndMarkSent(t *testing.T) {
	ctx := context.Background()
	repo := &fakeNotificationRepo{
		notifications: []Notification{{ID: "notification-1", Channel: telegramChannel}},
	}
	dispatcher := &fakeDispatcher{}
	scheduler := New(repo, dispatcher, time.Second, 100)

	if err := scheduler.tick(ctx); err != nil {
		t.Fatalf("tick returned error: %v", err)
	}

	if dispatcher.calls != 1 {
		t.Fatalf("expected one dispatch call, got %d", dispatcher.calls)
	}
	if repo.sentID != "notification-1" {
		t.Fatalf("expected notification to be marked sent, got %q", repo.sentID)
	}
	if repo.failedID != "" {
		t.Fatalf("expected no failed notification, got %q", repo.failedID)
	}
}

func TestTick_WhenTelegramDispatchFails_ShouldMarkFailed(t *testing.T) {
	ctx := context.Background()
	repo := &fakeNotificationRepo{
		notifications: []Notification{{ID: "notification-1", Channel: telegramChannel}},
	}
	dispatcher := &fakeDispatcher{err: errors.New("bot unavailable")}
	scheduler := New(repo, dispatcher, time.Second, 100)

	if err := scheduler.tick(ctx); err == nil {
		t.Fatal("expected tick to return dispatch error")
	}

	if repo.sentID != "" {
		t.Fatalf("expected no sent notification, got %q", repo.sentID)
	}
	if repo.failedID != "notification-1" {
		t.Fatalf("expected notification to be marked failed, got %q", repo.failedID)
	}
	if repo.failReason != "bot unavailable" {
		t.Fatalf("expected failure reason to be saved, got %q", repo.failReason)
	}
}

type fakeNotificationRepo struct {
	notifications []Notification
	sentID        string
	failedID      string
	failReason    string
}

func (r *fakeNotificationRepo) FetchDueByChannel(
	_ context.Context,
	_ time.Time,
	channel string,
	limit int64,
) ([]Notification, error) {
	var result []Notification
	for i := range r.notifications {
		notification := r.notifications[i]
		if notification.Channel == channel {
			result = append(result, notification)
		}
		if int64(len(result)) == limit {
			break
		}
	}
	return result, nil
}

func (r *fakeNotificationRepo) MarkSent(_ context.Context, notificationID string, _ time.Time) error {
	r.sentID = notificationID
	return nil
}

func (r *fakeNotificationRepo) MarkFailed(_ context.Context, notificationID string, _ time.Time, reason string) error {
	r.failedID = notificationID
	r.failReason = reason
	return nil
}

type fakeDispatcher struct {
	calls int
	err   error
}

func (d *fakeDispatcher) Dispatch(_ context.Context, _ *Notification) error {
	d.calls++
	return d.err
}
