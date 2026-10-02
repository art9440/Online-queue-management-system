package polling

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/bot/internal/application/service"
	"Online-queue-management-system/services/bot/internal/infrastructure/telegram"
	"context"
	"time"
)

type UpdatesClient interface {
	GetUpdates(ctx context.Context, offset int64, timeout int) ([]telegram.Update, error)
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, msg service.IncomingMessage) error
}

// LeaderLock elects the single replica allowed to call getUpdates.
type LeaderLock interface {
	TryAcquire(ctx context.Context) (bool, error)
	Renew(ctx context.Context) (bool, error)
	Release(ctx context.Context) error
}

type OffsetStore interface {
	Load(ctx context.Context) (int64, error)
	Save(ctx context.Context, offset int64) error
}

type Poller struct {
	client   UpdatesClient
	handler  MessageHandler
	lock     LeaderLock
	offsets  OffsetStore
	timeout  int
	interval time.Duration
	// lockTTL drives how often the lock is renewed and how often followers try to take it.
	lockTTL time.Duration
}

func New(
	client UpdatesClient,
	handler MessageHandler,
	lock LeaderLock,
	offsets OffsetStore,
	timeout int,
	interval time.Duration,
	lockTTL time.Duration,
) *Poller {
	return &Poller{
		client:   client,
		handler:  handler,
		lock:     lock,
		offsets:  offsets,
		timeout:  timeout,
		interval: interval,
		lockTTL:  lockTTL,
	}
}

// Run waits for leadership and polls Telegram while holding it. Replicas that are not the
// leader keep serving HTTP (sending notifications) and take over if the leader disappears.
func (p *Poller) Run(ctx context.Context) {
	log := logger.From(ctx)
	log.Info("telegram poller started, waiting for leadership")

	for ctx.Err() == nil {
		acquired, err := p.lock.TryAcquire(ctx)
		if err != nil {
			log.Error("failed to acquire telegram poller lock", "err", err)
		}
		if acquired {
			p.lead(ctx)
		}
		sleep(ctx, p.lockTTL/3)
	}

	log.Info("telegram poller stopped")
}

func (p *Poller) lead(ctx context.Context) {
	log := logger.From(ctx)
	log.Info("became telegram poller leader")

	leaderCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go p.keepLock(leaderCtx, cancel)
	p.poll(leaderCtx)

	if err := p.lock.Release(context.WithoutCancel(ctx)); err != nil {
		log.Warn("failed to release telegram poller lock", "err", err)
	}
	log.Info("stopped being telegram poller leader")
}

// keepLock renews the lock and stops polling as soon as leadership is lost.
func (p *Poller) keepLock(ctx context.Context, stopPolling context.CancelFunc) {
	ticker := time.NewTicker(p.lockTTL / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := p.lock.Renew(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil || !ok {
				logger.From(ctx).Warn("lost telegram poller leadership", "err", err)
				stopPolling()
				return
			}
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	log := logger.From(ctx)

	offset, err := p.offsets.Load(ctx)
	if err != nil {
		log.Error("failed to load telegram offset", "err", err)
		return
	}

	for ctx.Err() == nil {
		updates, err := p.client.GetUpdates(ctx, offset, p.timeout)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			log.Error("failed to get telegram updates", "err", err)
			sleep(ctx, p.interval)
			continue
		}

		for i := range updates {
			update := updates[i]
			if update.Message != nil {
				msg := toIncomingMessage(update.Message)
				if err := p.handler.HandleMessage(ctx, msg); err != nil {
					log.Error("failed to handle telegram message", "err", err, "chat_id", msg.ChatID)
				}
			}

			if update.ID >= offset {
				offset = update.ID + 1
				if err := p.offsets.Save(context.WithoutCancel(ctx), offset); err != nil {
					log.Error("failed to save telegram offset", "err", err)
				}
			}
		}
	}
}

func toIncomingMessage(message *telegram.Message) service.IncomingMessage {
	msg := service.IncomingMessage{
		ChatID: message.Chat.ID,
		Text:   message.Text,
	}
	if message.Contact != nil {
		msg.ContactPhone = message.Contact.PhoneNumber
	}
	if message.From != nil {
		msg.Username = message.From.Username
	}
	return msg
}

func sleep(ctx context.Context, interval time.Duration) {
	timer := time.NewTimer(interval)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
