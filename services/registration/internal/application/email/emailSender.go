package email

import (
	"Online-queue-management-system/libs/logger"
	"Online-queue-management-system/services/registration/config"
	"context"
	"fmt"
	"net/smtp"
)

type EmailSender struct {
	host     string
	port     string
	user     string
	password string
}

func NewEmailSender(cfg config.Config) *EmailSender {
	return &EmailSender{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPass,
	}
}

func (e *EmailSender) SendEmail(ctx context.Context, msg EmailMessage) error {
	log := logger.From(ctx)
	addr := fmt.Sprintf("%s:%s", e.host, e.port)
	log.Info("smtp addr", "addr", addr)
	auth := smtp.PlainAuth("", e.user, e.password, e.host)
	contentType := "text/plain"

	headers := make(map[string]string)
	headers["From"] = e.user
	headers["To"] = msg.To
	headers["Subject"] = msg.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", contentType)

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	err := smtp.SendMail(addr, auth, e.user, []string{msg.To}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Info("email send succesfully")

	return nil
}
