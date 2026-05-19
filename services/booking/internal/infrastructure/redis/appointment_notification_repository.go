package redis

import (
	"Online-queue-management-system/services/booking/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	scheduledNotificationsKey = "notifications:scheduled"
	notificationKeyPrefix     = "notification:"
)

var reminderOffsets = []time.Duration{
	24 * time.Hour,
	time.Hour,
}

type AppointmentNotificationRepository struct {
	client *goredis.Client
}

type appointmentNotificationPayload struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	Phone       string `json:"phone,omitempty"`
	Username    string `json:"username,omitempty"`
	Service     string `json:"service,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Employee    string `json:"employee,omitempty"`
	StartTime   string `json:"start_time"`
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Body        string `json:"body,omitempty"`
	HTMLBody    string `json:"html_body,omitempty"`
	Status      string `json:"status"`
}

func NewAppointmentNotificationRepository(client *goredis.Client) *AppointmentNotificationRepository {
	return &AppointmentNotificationRepository{client: client}
}

func (r *AppointmentNotificationRepository) SaveAppointmentNotifications(
	ctx context.Context,
	appointment *domain.CreateAppointmentOutput,
	client *domain.ClientInput,
) error {
	if appointment == nil || client == nil {
		return nil
	}

	now := time.Now()
	notifications := buildAppointmentNotifications(appointment, client, now)
	if len(notifications) == 0 {
		return nil
	}

	pipe := r.client.TxPipeline()
	for i := range notifications {
		notification := &notifications[i]
		raw, err := json.Marshal(notification.payload)
		if err != nil {
			return err
		}

		pipe.Set(ctx, notificationKey(notification.payload.ID), raw, 0)
		pipe.ZAdd(ctx, scheduledNotificationsKey, goredis.Z{
			Score:  float64(notification.scheduledAt.UTC().Unix()),
			Member: notification.payload.ID,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

type scheduledAppointmentNotification struct {
	payload     appointmentNotificationPayload
	scheduledAt time.Time
}

func buildAppointmentNotifications(
	appointment *domain.CreateAppointmentOutput,
	client *domain.ClientInput,
	now time.Time,
) []scheduledAppointmentNotification {
	notifications := make([]scheduledAppointmentNotification, 0, len(reminderOffsets)*2)
	for _, offset := range reminderOffsets {
		scheduledAt := appointment.StartTime.Add(-offset)
		if !scheduledAt.After(now) {
			continue
		}

		if email := valueFromPointer(client.Email); email != "" {
			notifications = append(notifications, scheduledAppointmentNotification{
				payload:     emailNotificationPayload(appointment, email, offset),
				scheduledAt: scheduledAt,
			})
		}

		phone := normalizePhone(client.Phone)
		username := normalizeUsername(valueFromPointer(client.TgUsername))
		if phone != "" || username != "" {
			notifications = append(notifications, scheduledAppointmentNotification{
				payload:     telegramNotificationPayload(appointment, phone, username, offset),
				scheduledAt: scheduledAt,
			})
		}
	}

	return notifications
}

func emailNotificationPayload(
	appointment *domain.CreateAppointmentOutput,
	email string,
	offset time.Duration,
) appointmentNotificationPayload {
	payload := baseNotificationPayload(appointment, "email", offset)
	payload.Email = email
	payload.Subject = "Appointment reminder"
	payload.Body = plainReminderBody(appointment, offset)
	payload.HTMLBody = htmlReminderBody(appointment, offset)
	return payload
}

func telegramNotificationPayload(
	appointment *domain.CreateAppointmentOutput,
	phone string,
	username string,
	offset time.Duration,
) appointmentNotificationPayload {
	payload := baseNotificationPayload(appointment, "telegram", offset)
	payload.Phone = phone
	payload.Username = username
	return payload
}

func baseNotificationPayload(
	appointment *domain.CreateAppointmentOutput,
	channel string,
	offset time.Duration,
) appointmentNotificationPayload {
	return appointmentNotificationPayload{
		ID:          notificationID(appointment.AppointmentID, channel, offset),
		Channel:     channel,
		Service:     appointment.ServiceName,
		Branch:      appointment.BranchName,
		Employee:    strings.TrimSpace(appointment.EmployeeName + " " + appointment.EmployeeSurname),
		StartTime:   appointment.StartTime.UTC().Format(time.RFC3339),
		Description: reminderDescription(offset),
		Status:      "scheduled",
	}
}

func plainReminderBody(appointment *domain.CreateAppointmentOutput, offset time.Duration) string {
	lines := []string{
		"Appointment reminder",
		"",
		reminderDescription(offset),
		"When: " + appointment.StartTime.Format("02.01.2006 15:04"),
		"Service: " + appointment.ServiceName,
		"Branch: " + appointment.BranchName,
		"Employee: " + strings.TrimSpace(appointment.EmployeeName+" "+appointment.EmployeeSurname),
	}
	if appointment.Comment != nil && *appointment.Comment != "" {
		lines = append(lines, "", "Comment: "+*appointment.Comment)
	}

	return strings.Join(lines, "\n")
}

func htmlReminderBody(appointment *domain.CreateAppointmentOutput, offset time.Duration) string {
	comment := ""
	if appointment.Comment != nil && *appointment.Comment != "" {
		comment = fmt.Sprintf("<p><b>Comment:</b> %s</p>", html.EscapeString(*appointment.Comment))
	}

	return fmt.Sprintf(
		`<h2>Appointment reminder</h2><p>%s</p><table><tr><td>When</td><td>%s</td></tr><tr><td>Service</td><td>%s</td></tr><tr><td>Branch</td><td>%s</td></tr><tr><td>Employee</td><td>%s</td></tr></table>%s`,
		html.EscapeString(reminderDescription(offset)),
		html.EscapeString(appointment.StartTime.Format("02.01.2006 15:04")),
		html.EscapeString(appointment.ServiceName),
		html.EscapeString(appointment.BranchName),
		html.EscapeString(strings.TrimSpace(appointment.EmployeeName+" "+appointment.EmployeeSurname)),
		comment,
	)
}

func reminderDescription(offset time.Duration) string {
	switch offset {
	case 24 * time.Hour:
		return "Appointment starts in 24 hours."
	case time.Hour:
		return "Appointment starts in 1 hour."
	default:
		return "Appointment starts soon."
	}
}

func notificationID(appointmentID int64, channel string, offset time.Duration) string {
	return fmt.Sprintf(
		"appointment:%d:%s:%d",
		appointmentID,
		channel,
		int(offset/time.Minute),
	)
}

func normalizeUsername(value string) string {
	username := strings.TrimSpace(value)
	username = strings.TrimPrefix(username, "@")
	if username == "" {
		return ""
	}
	return strings.ToLower(username)
}

func normalizePhone(value string) string {
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
		return ""
	}

	phone := builder.String()
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) < 10 || len(digits) > 15 {
		return ""
	}

	if _, err := strconv.ParseUint(digits, 10, 64); err != nil {
		return ""
	}

	if strings.HasPrefix(phone, "+") {
		return phone
	}

	return "+" + phone
}

func valueFromPointer(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func notificationKey(id string) string {
	return notificationKeyPrefix + id
}
