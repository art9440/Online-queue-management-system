package email

import (
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/scheduler/internal/application/service"
	"context"
	"errors"
	"fmt"
	"html"
	"strings"
)

const defaultReminderSubject = "Напоминание о записи"

type Sender interface {
	SendEmail(ctx context.Context, msg email.EmailMessage) error
}

type Dispatcher struct {
	sender Sender
}

func NewDispatcher(sender Sender) *Dispatcher {
	return &Dispatcher{sender: sender}
}

func (d *Dispatcher) Dispatch(ctx context.Context, notification *service.Notification) error {
	if notification.Email == "" {
		return errors.New("email recipient is required")
	}

	return d.sender.SendEmail(ctx, email.EmailMessage{
		To:       notification.Email,
		Subject:  subject(notification),
		Body:     plainBody(notification),
		HTMLBody: htmlBody(notification),
	})
}

func subject(notification *service.Notification) string {
	if notification.Subject != "" {
		return notification.Subject
	}
	return defaultReminderSubject
}

func plainBody(notification *service.Notification) string {
	if notification.Body != "" {
		return notification.Body
	}

	lines := []string{"Напоминание о записи"}
	if !notification.StartTime.IsZero() {
		lines = append(lines, "Когда: "+notification.StartTime.Format("02.01.2006 15:04"))
	}
	if notification.Business != "" {
		lines = append(lines, "Компания: "+notification.Business)
	}
	if notification.Service != "" {
		lines = append(lines, "Услуга: "+notification.Service)
	}
	if notification.Branch != "" {
		lines = append(lines, "Филиал: "+notification.Branch)
	}
	if notification.Employee != "" {
		lines = append(lines, "Специалист: "+notification.Employee)
	}
	if notification.Description != "" {
		lines = append(lines, "", notification.Description)
	}

	return strings.Join(lines, "\n")
}

func htmlBody(notification *service.Notification) string {
	if notification.HTMLBody != "" {
		return notification.HTMLBody
	}

	rows := make([]string, 0, 5)
	if !notification.StartTime.IsZero() {
		rows = append(rows, row("Когда", notification.StartTime.Format("02.01.2006 15:04")))
	}
	if notification.Business != "" {
		rows = append(rows, row("Компания", notification.Business))
	}
	if notification.Service != "" {
		rows = append(rows, row("Услуга", notification.Service))
	}
	if notification.Branch != "" {
		rows = append(rows, row("Филиал", notification.Branch))
	}
	if notification.Employee != "" {
		rows = append(rows, row("Специалист", notification.Employee))
	}

	description := ""
	if notification.Description != "" {
		description = fmt.Sprintf(
			`<p style="margin: 20px 0 0; color: #475569;">%s</p>`,
			html.EscapeString(notification.Description),
		)
	}

	return fmt.Sprintf(`
<div style="font-family: Arial, sans-serif; max-width: 560px; margin: 0 auto; color: #0f172a;">
  <div style="border: 1px solid #dbe3ef; border-radius: 8px; overflow: hidden;">
    <div style="background: #0f766e; color: #ffffff; padding: 18px 22px;">
      <h2 style="margin: 0; font-size: 22px;">Напоминание о записи</h2>
    </div>
    <div style="padding: 20px 22px;">
      <table style="width: 100%%; border-collapse: collapse;">%s</table>
      %s
    </div>
  </div>
</div>`, strings.Join(rows, ""), description)
}

func row(label, value string) string {
	return fmt.Sprintf(
		`<tr><td style="padding: 8px 0; color: #64748b; width: 130px;">%s</td><td style="padding: 8px 0; font-weight: 600;">%s</td></tr>`,
		html.EscapeString(label),
		html.EscapeString(value),
	)
}
