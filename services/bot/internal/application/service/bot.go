package service

import (
	"Online-queue-management-system/libs/logger"
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

type ChatRepository interface {
	SavePhoneBinding(ctx context.Context, phone string, chatID int64) error
	SaveUsernameBinding(ctx context.Context, username string, chatID int64) error
	FindChatID(ctx context.Context, recipient Recipient) (int64, error)
}

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Recipient struct {
	Phone    string
	Username string
}

type Notification struct {
	Recipient   Recipient
	Business    string
	Service     string
	Branch      string
	Employee    string
	StartTime   time.Time
	Description string
}

type IncomingMessage struct {
	ChatID       int64
	Text         string
	ContactPhone string
	Username     string
}

type Bot struct {
	repo   ChatRepository
	sender TelegramSender
}

func New(repo ChatRepository, sender TelegramSender) *Bot {
	return &Bot{
		repo:   repo,
		sender: sender,
	}
}

func (b *Bot) HandleMessage(ctx context.Context, msg IncomingMessage) error {
	log := logger.From(ctx)

	switch {
	case msg.ChatID == 0:
		return nil
	case msg.ContactPhone != "":
		return b.bindPhone(ctx, msg.ChatID, msg.ContactPhone)
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return b.sendHelp(ctx, msg.ChatID)
	}

	if text == "/start" || text == "/help" {
		return b.sendHelp(ctx, msg.ChatID)
	}

	if username, ok := normalizeUsername(text); ok {
		if err := b.repo.SaveUsernameBinding(ctx, username, msg.ChatID); err != nil {
			return err
		}
		return b.sender.SendMessage(ctx, msg.ChatID, confirmationMessage("Telegram", "@"+username))
	}

	if phone, ok := normalizePhone(text); ok {
		return b.bindPhone(ctx, msg.ChatID, phone)
	}

	if msg.Username != "" {
		log.Info("telegram message has profile username", "username", msg.Username, "chat_id", msg.ChatID)
	}

	return b.sender.SendMessage(ctx, msg.ChatID, "Не смог распознать контакт. Отправьте номер телефона или Telegram username в формате @username.")
}

func (b *Bot) SendNotification(ctx context.Context, notification *Notification) error {
	chatID, err := b.repo.FindChatID(ctx, notification.Recipient)
	if err != nil {
		return err
	}

	return b.sender.SendMessage(ctx, chatID, FormatNotification(notification))
}

func (b *Bot) bindPhone(ctx context.Context, chatID int64, phone string) error {
	normalized, ok := normalizePhone(phone)
	if !ok {
		return b.sender.SendMessage(ctx, chatID, "Не смог распознать номер. Отправьте телефон в формате +79990000000.")
	}

	if err := b.repo.SavePhoneBinding(ctx, normalized, chatID); err != nil {
		return err
	}

	return b.sender.SendMessage(ctx, chatID, confirmationMessage("телефон", normalized))
}

func (b *Bot) sendHelp(ctx context.Context, chatID int64) error {
	return b.sender.SendMessage(ctx, chatID, "Привет! Я бот уведомлений Online Queue.\n\nОтправьте номер телефона или Telegram username, чтобы я мог присылать напоминания о записи.")
}

func FormatNotification(notification *Notification) string {
	startTime := "время не указано"
	if !notification.StartTime.IsZero() {
		startTime = notification.StartTime.Format("02.01.2006 15:04")
	}

	lines := []string{
		"<b>Напоминание о записи</b>",
		"",
		"<b>Когда:</b> " + html.EscapeString(startTime),
	}

	if notification.Business != "" {
		lines = append(lines, "<b>Компания:</b> "+html.EscapeString(notification.Business))
	}
	if notification.Service != "" {
		lines = append(lines, "<b>Услуга:</b> "+html.EscapeString(notification.Service))
	}
	if notification.Branch != "" {
		lines = append(lines, "<b>Филиал:</b> "+html.EscapeString(notification.Branch))
	}
	if notification.Employee != "" {
		lines = append(lines, "<b>Специалист:</b> "+html.EscapeString(notification.Employee))
	}
	if notification.Description != "" {
		lines = append(lines, "", html.EscapeString(notification.Description))
	}

	return strings.Join(lines, "\n")
}

func confirmationMessage(kind, value string) string {
	return fmt.Sprintf("Готово, привязал %s: %s\nТеперь сюда можно отправлять напоминания о записи.", kind, value)
}

func normalizeUsername(value string) (string, bool) {
	username := strings.TrimSpace(value)
	username = strings.TrimPrefix(username, "@")
	if len(username) < 5 || len(username) > 32 {
		return "", false
	}

	for _, char := range username {
		if char == '_' || char >= '0' && char <= '9' || char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z' {
			continue
		}
		return "", false
	}

	return strings.ToLower(username), true
}

func normalizePhone(value string) (string, bool) {
	var builder strings.Builder
	for index, char := range strings.TrimSpace(value) {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
			continue
		}
		if char == '+' && index == 0 {
			builder.WriteRune(char)
			continue
		}
		if char == ' ' || char == '-' || char == '(' || char == ')' {
			continue
		}
		return "", false
	}

	phone := builder.String()
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) < 10 || len(digits) > 15 {
		return "", false
	}

	if _, err := strconv.ParseUint(digits, 10, 64); err != nil {
		return "", false
	}

	if strings.HasPrefix(phone, "+") {
		return phone, true
	}

	return "+" + phone, true
}
