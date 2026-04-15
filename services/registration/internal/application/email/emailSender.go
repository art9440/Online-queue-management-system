package email

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/config"
	"context"
	"fmt"

	"github.com/go-mail/mail/v2"
)

type EmailSender struct {
	dialer *mail.Dialer
	from   string
}

func NewEmailSender(cfg config.EmailSenderConfig) *EmailSender {
	dialer := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	dialer.StartTLSPolicy = mail.MandatoryStartTLS
	dialer.Timeout = cfg.SendTimeOut

	return &EmailSender{
		dialer: dialer,
		from:   cfg.SMTPUser,
	}
}

func (e *EmailSender) SendEmail(ctx context.Context, msg EmailMessage) error {
	log := logger.From(ctx)

	m := mail.NewMessage()
	m.SetHeader("From", e.from)
	m.SetHeader("To", msg.To)
	m.SetHeader("Subject", msg.Subject)

	body := fmt.Sprintf(`
		<h2>Подтверждение регистрации</h2>
		<p>Ваш код подтверждения:</p>
		<h1 style="font-size: 32px; letter-spacing: 5px;">%s</h1>
		<p>Введите этот код для завершения регистрации.</p>
	`, msg.Body)

	m.SetBody("text/html", body)

	log.Info("sending email", "to", msg.To)

	s, err := e.dialer.Dial()
	if err != nil {
		log.Error("failed to dial SMTP", "error", err)
		return fmt.Errorf("failed to dial SMTP: %w", err)
	}
	defer s.Close()

	if err := mail.Send(s, m); err != nil {
		log.Error("error sending email", "error", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Info("email sent successfully")
	return nil
}
