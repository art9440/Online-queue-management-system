package service

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	telegramChannel = "telegram"
	emailChannel    = "email"
)

type NotificationRepository interface {
	// ClaimDue atomically takes due notifications for this scheduler until leaseUntil.
	ClaimDue(ctx context.Context, now time.Time, limit int64, leaseUntil time.Time) ([]Notification, error)
	// RequeueExpired returns notifications of crashed schedulers (expired lease) to the schedule.
	RequeueExpired(ctx context.Context, now time.Time) (int64, error)
	// Release returns claimed but unprocessed notifications to the schedule.
	Release(ctx context.Context, ids []string, at time.Time) error
	MarkSent(ctx context.Context, notificationID string, sentAt time.Time) error
	MarkFailed(ctx context.Context, notificationID string, failedAt time.Time, reason string) error
	Retry(ctx context.Context, notificationID string, retryAt time.Time, reason string) error
}

type Dispatcher interface {
	Dispatch(ctx context.Context, notification *Notification) error
}

type Config struct {
	PollInterval time.Duration
	BatchSize    int
	// Lease: how long claimed notifications belong to this scheduler before others may take them.
	Lease time.Duration
	// MaxAttempts: failed deliveries before the notification is marked failed.
	MaxAttempts int
	// RetryDelay: base delay before the next attempt (multiplied by the attempt number).
	RetryDelay time.Duration
}

type Scheduler struct {
	repo        NotificationRepository
	dispatchers map[string]Dispatcher
	cfg         Config
}

type Notification struct {
	ID          string
	Channel     string
	Phone       string
	Username    string
	Business    string
	Service     string
	Branch      string
	Employee    string
	StartTime   time.Time
	Description string
	Email       string
	Subject     string
	Body        string
	HTMLBody    string
	// Attempts: failed deliveries so far.
	Attempts int
}

func New(
	repo NotificationRepository,
	telegramDispatcher Dispatcher,
	emailDispatcher Dispatcher,
	cfg Config,
) *Scheduler {
	dispatchers := map[string]Dispatcher{}
	if telegramDispatcher != nil {
		dispatchers[telegramChannel] = telegramDispatcher
	}
	if emailDispatcher != nil {
		dispatchers[emailChannel] = emailDispatcher
	}

	return &Scheduler{
		repo:        repo,
		dispatchers: dispatchers,
		cfg:         cfg,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info(
		"scheduler started",
		"poll_interval", s.cfg.PollInterval.String(),
		"batch_size", s.cfg.BatchSize,
		"lease", s.cfg.Lease.String(),
		"max_attempts", s.cfg.MaxAttempts,
	)

	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()

	if err := s.tick(ctx); err != nil {
		log.Error("scheduler tick failed", "err", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("scheduler stopped")
			return nil
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				log.Error("scheduler tick failed", "err", err)
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	log := logger.From(ctx)
	now := time.Now()

	requeued, err := s.repo.RequeueExpired(ctx, now)
	if err != nil {
		return err
	}
	if requeued > 0 {
		log.Warn("requeued notifications with expired lease", "count", requeued)
	}

	claimedAt := time.Now()
	leaseUntil := claimedAt.Add(s.cfg.Lease)
	notifications, err := s.repo.ClaimDue(ctx, claimedAt, int64(s.cfg.BatchSize), leaseUntil)
	if err != nil {
		return err
	}

	// Stop early enough that the notification being sent finishes before the lease expires;
	// otherwise another replica could take it and send it twice.
	processUntil := leaseUntil.Add(-s.cfg.Lease / 4)

	var (
		dispatchErrors error
		sent, failed   int
	)
	for i := range notifications {
		if ctx.Err() != nil || time.Now().After(processUntil) {
			s.release(ctx, notifications[i:])
			break
		}

		notification := notifications[i]
		if err := s.dispatch(ctx, &notification); err != nil {
			failed++
			dispatchErrors = errors.Join(dispatchErrors, err)
			continue
		}
		sent++
	}

	log.Info(
		"scheduler tick completed",
		"claimed", len(notifications),
		"sent", sent,
		"failed", failed,
	)
	return dispatchErrors
}

func (s *Scheduler) dispatch(ctx context.Context, notification *Notification) error {
	log := logger.From(ctx).With("notification_id", notification.ID, "channel", notification.Channel)
	// Finish the current notification even during shutdown (12-factor IX); dispatchers have their own timeouts.
	ctx = context.WithoutCancel(ctx)

	dispatcher, ok := s.dispatchers[notification.Channel]
	if !ok {
		err := fmt.Errorf("unknown notification channel %q", notification.Channel)
		log.Error("cannot dispatch notification", "err", err)
		return errors.Join(err, s.repo.MarkFailed(ctx, notification.ID, time.Now(), err.Error()))
	}

	if err := dispatcher.Dispatch(ctx, notification); err != nil {
		attempt := notification.Attempts + 1
		if attempt >= s.cfg.MaxAttempts {
			log.Error("notification failed, no attempts left", "attempt", attempt, "err", err)
			return errors.Join(err, s.repo.MarkFailed(ctx, notification.ID, time.Now(), err.Error()))
		}

		retryAt := time.Now().Add(s.cfg.RetryDelay * time.Duration(attempt))
		log.Warn("notification failed, will retry", "attempt", attempt, "retry_at", retryAt, "err", err)
		return errors.Join(err, s.repo.Retry(ctx, notification.ID, retryAt, err.Error()))
	}

	if err := s.repo.MarkSent(ctx, notification.ID, time.Now()); err != nil {
		log.Error("failed to mark notification as sent", "err", err)
		return err
	}

	log.Info("notification dispatched")
	return nil
}

// release gives unprocessed notifications back so another replica can take them right away.
func (s *Scheduler) release(ctx context.Context, notifications []Notification) {
	ids := make([]string, 0, len(notifications))
	for i := range notifications {
		ids = append(ids, notifications[i].ID)
	}

	if err := s.repo.Release(context.WithoutCancel(ctx), ids, time.Now()); err != nil {
		logger.From(ctx).Error("failed to release notifications", "count", len(ids), "err", err)
		return
	}
	logger.From(ctx).Info("released unprocessed notifications", "count", len(ids))
}
