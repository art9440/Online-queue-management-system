package redis

import (
	"Online-queue-management-system/services/bot/internal/application/service"
	"context"
	"errors"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
)

const (
	phoneBindingPrefix    = "telegram:phone:"
	usernameBindingPrefix = "telegram:username:"
	chatBindingPrefix     = "telegram:chat:"
)

type ChatRepository struct {
	client *goredis.Client
}

func NewChatRepository(client *goredis.Client) *ChatRepository {
	return &ChatRepository{client: client}
}

func (r *ChatRepository) SavePhoneBinding(ctx context.Context, phone string, chatID int64) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, phoneBindingPrefix+phone, chatID, 0)
	pipe.HSet(ctx, chatBindingPrefix+formatChatID(chatID), "phone", phone)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *ChatRepository) SaveUsernameBinding(ctx context.Context, username string, chatID int64) error {
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, usernameBindingPrefix+username, chatID, 0)
	pipe.HSet(ctx, chatBindingPrefix+formatChatID(chatID), "username", username)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *ChatRepository) FindChatID(ctx context.Context, recipient service.Recipient) (int64, error) {
	keys := recipientKeys(recipient)
	for _, key := range keys {
		chatID, err := r.client.Get(ctx, key).Int64()
		if err == nil {
			return chatID, nil
		}
		if !errors.Is(err, goredis.Nil) {
			return 0, err
		}
	}

	return 0, errors.New("telegram chat binding not found")
}

func recipientKeys(recipient service.Recipient) []string {
	keys := make([]string, 0, 2)
	if recipient.Phone != "" {
		keys = append(keys, phoneBindingPrefix+recipient.Phone)
	}
	if recipient.Username != "" {
		keys = append(keys, usernameBindingPrefix+recipient.Username)
	}
	return keys
}

func formatChatID(chatID int64) string {
	return strconv.FormatInt(chatID, 10)
}
