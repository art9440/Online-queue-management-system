package service

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"errors"
	"time"
)

const (
	telegramChannel = "telegram"
	emailChannel    = "email"
)

type NotificationRepository interface {
	FetchDueByChannel(ctx context.Context, now time.Time, channel string, limit int64) ([]Notification, error)
	MarkSent(ctx context.Context, notificationID string, sentAt time.Time) error
	MarkFailed(ctx context.Context, notificationID string, failedAt time.Time, reason string) error
}

type Dispatcher interface {
	Dispatch(ctx context.Context, notification *Notification) error
}

type Scheduler struct {
	repo               NotificationRepository
	telegramDispatcher Dispatcher
	emailDispatcher    Dispatcher
	pollInterval       time.Duration
	batchSize          int
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
}

func New(
	repo NotificationRepository,
	telegramDispatcher Dispatcher,
	emailDispatcher Dispatcher,
	pollInterval time.Duration,
	batchSize int,
) *Scheduler {
	return &Scheduler{
		repo:               repo,
		telegramDispatcher: telegramDispatcher,
		emailDispatcher:    emailDispatcher,
		pollInterval:       pollInterval,
		batchSize:          batchSize,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	log := logger.From(ctx)
	log.Info(
		"scheduler started",
		"poll_interval", s.pollInterval.String(),
		"batch_size", s.batchSize,
	)

	ticker := time.NewTicker(s.pollInterval)
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
	var dispatchErrors error

	telegramProcessed, err := s.dispatchChannel(ctx, now, telegramChannel, s.telegramDispatcher)
	if err != nil {
		dispatchErrors = errors.Join(dispatchErrors, err)
	}

	emailProcessed, err := s.dispatchChannel(ctx, now, emailChannel, s.emailDispatcher)
	if err != nil {
		dispatchErrors = errors.Join(dispatchErrors, err)
	}

	log.Info(
		"scheduler tick completed",
		"telegram_notifications", telegramProcessed,
		"email_notifications", emailProcessed,
		"batch_size", s.batchSize,
	)
	return dispatchErrors
}

func (s *Scheduler) dispatchChannel(
	ctx context.Context,
	now time.Time,
	channel string,
	dispatcher Dispatcher,
) (int, error) {
	log := logger.From(ctx)

	if dispatcher == nil {
		return 0, nil
	}

	notifications, err := s.repo.FetchDueByChannel(ctx, now, channel, int64(s.batchSize))
	if err != nil {
		return 0, err
	}

	var dispatchErrors error
	for i := range notifications {
		notification := notifications[i]
		if err := dispatcher.Dispatch(ctx, &notification); err != nil {
			log.Error("failed to dispatch notification", "notification_id", notification.ID, "channel", channel, "err", err)
			if markErr := s.repo.MarkFailed(ctx, notification.ID, time.Now(), err.Error()); markErr != nil {
				log.Error("failed to mark notification as failed", "notification_id", notification.ID, "err", markErr)
				dispatchErrors = errors.Join(dispatchErrors, markErr)
			}
			dispatchErrors = errors.Join(dispatchErrors, err)
			continue
		}

		if err := s.repo.MarkSent(ctx, notification.ID, time.Now()); err != nil {
			log.Error("failed to mark notification as sent", "notification_id", notification.ID, "err", err)
			dispatchErrors = errors.Join(dispatchErrors, err)
			continue
		}

		log.Info("notification dispatched", "notification_id", notification.ID, "channel", channel)
	}

	return len(notifications), dispatchErrors
}
