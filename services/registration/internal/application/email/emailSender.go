package email

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/config"
	"context"
	"fmt"

	"github.com/go-mail/mail/v2" // <-- новый импорт
)

type EmailSender struct {
	dialer *mail.Dialer
	from   string
}

func NewEmailSender(cfg config.Config) *EmailSender {
	// Создаём dialer — API почти такой же
	dialer := mail.NewDialer(cfg.SMTPHost, 587, cfg.SMTPUser, cfg.SMTPPass)

	return &EmailSender{
		dialer: dialer,
		from:   cfg.SMTPUser,
	}
}

func (e *EmailSender) SendEmail(ctx context.Context, msg EmailMessage) error {
	log := logger.From(ctx)

	// Создаём письмо — API полностью совместим
	m := mail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)

	// HTML-письмо
	body := fmt.Sprintf(`
		<h2>Подтверждение регистрации</h2>
		<p>Ваш код подтверждения:</p>
		<h1 style="font-size: 32px; letter-spacing: 5px;">%s</h1>
		<p>Введите этот код для завершения регистрации.</p>
	`, msg.Body)
	m.SetBody("text/html", body)

	// Отправляем
	log.Info("отправка письма", "to", msg.To)

	if err := e.dialer.DialAndSend(m); err != nil {
		log.Error("ошибка отправки", "error", err)
		return fmt.Errorf("не удалось отправить письмо: %w", err)
	}

	log.Info("письмо успешно отправлено")
	return nil
}
