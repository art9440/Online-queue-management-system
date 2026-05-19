package service

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"time"
)

type DueNotificationRepository interface {
	CountDue(ctx context.Context, now time.Time) (int64, error)
}

type Scheduler struct {
	repo         DueNotificationRepository
	pollInterval time.Duration
	batchSize    int
}

func New(repo DueNotificationRepository, pollInterval time.Duration, batchSize int) *Scheduler {
	return &Scheduler{
		repo:         repo,
		pollInterval: pollInterval,
		batchSize:    batchSize,
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

	dueCount, err := s.repo.CountDue(ctx, time.Now())
	if err != nil {
		return err
	}

	log.Info("scheduler tick completed", "due_notifications", dueCount, "batch_size", s.batchSize)
	return nil
}
