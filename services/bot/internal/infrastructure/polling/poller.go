package polling

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/bot/internal/application/service"
	"Online-queue-management-system/services/bot/internal/infrastructure/telegram"
	"context"
	"time"
)

type Poller struct {
	client   *telegram.Client
	bot      *service.Bot
	timeout  int
	interval time.Duration
}

func New(client *telegram.Client, bot *service.Bot, timeout int, interval time.Duration) *Poller {
	return &Poller{
		client:   client,
		bot:      bot,
		timeout:  timeout,
		interval: interval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	log := logger.From(ctx)
	log.Info("telegram polling started")

	var offset int64
	for {
		select {
		case <-ctx.Done():
			log.Info("telegram polling stopped")
			return
		default:
		}

		updates, err := p.client.GetUpdates(ctx, offset, p.timeout)
		if err != nil {
			log.Error("failed to get telegram updates", "err", err)
			sleep(ctx, p.interval)
			continue
		}

		for i := range updates {
			update := updates[i]
			if update.ID >= offset {
				offset = update.ID + 1
			}
			if update.Message == nil {
				continue
			}

			msg := toIncomingMessage(update.Message)
			if err := p.bot.HandleMessage(ctx, msg); err != nil {
				log.Error("failed to handle telegram message", "err", err, "chat_id", msg.ChatID)
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
