package redis

import (
	"Online-queue-management-system/services/scheduler/internal/application/service"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	scheduledNotificationsKey = "notifications:scheduled"
	notificationKeyPrefix     = "notification:"
)

type NotificationRepository struct {
	client *goredis.Client
}

type notificationPayload struct {
	ID          string `json:"id"`
	Channel     string `json:"channel"`
	Phone       string `json:"phone"`
	Username    string `json:"username"`
	Business    string `json:"business"`
	Service     string `json:"service"`
	Branch      string `json:"branch"`
	Employee    string `json:"employee"`
	StartTime   string `json:"start_time"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SentAt      string `json:"sent_at,omitempty"`
	FailedAt    string `json:"failed_at,omitempty"`
	FailReason  string `json:"fail_reason,omitempty"`
}

func NewNotificationRepository(client *goredis.Client) *NotificationRepository {
	return &NotificationRepository{client: client}
}

func (r *NotificationRepository) FetchDueByChannel(
	ctx context.Context,
	now time.Time,
	channel string,
	limit int64,
) ([]service.Notification, error) {
	ids, err := r.client.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     scheduledNotificationsKey,
		Start:   "-inf",
		Stop:    timestampScore(now),
		ByScore: true,
		Offset:  0,
		Count:   limit,
	}).Result()
	if err != nil {
		return nil, err
	}

	notifications := make([]service.Notification, 0, len(ids))
	for _, id := range ids {
		notification, err := r.getNotification(ctx, id)
		if err != nil {
			if errors.Is(err, goredis.Nil) {
				_ = r.client.ZRem(ctx, scheduledNotificationsKey, id).Err()
				continue
			}
			return nil, err
		}

		if notification.Channel != channel {
			continue
		}

		removed, err := r.client.ZRem(ctx, scheduledNotificationsKey, id).Result()
		if err != nil {
			return nil, err
		}
		if removed == 0 {
			continue
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkSent(ctx context.Context, notificationID string, sentAt time.Time) error {
	payload, err := r.getPayload(ctx, notificationID)
	if err != nil {
		return err
	}

	payload.Status = "sent"
	payload.SentAt = sentAt.UTC().Format(time.RFC3339)
	payload.FailedAt = ""
	payload.FailReason = ""

	return r.savePayload(ctx, &payload)
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, notificationID string, failedAt time.Time, reason string) error {
	payload, err := r.getPayload(ctx, notificationID)
	if err != nil {
		return err
	}

	payload.Status = "failed"
	payload.FailedAt = failedAt.UTC().Format(time.RFC3339)
	payload.FailReason = reason

	return r.savePayload(ctx, &payload)
}

func (r *NotificationRepository) getNotification(ctx context.Context, id string) (service.Notification, error) {
	payload, err := r.getPayload(ctx, id)
	if err != nil {
		return service.Notification{}, err
	}

	startTime, err := parseStartTime(payload.StartTime)
	if err != nil {
		return service.Notification{}, err
	}

	return service.Notification{
		ID:          firstNonEmpty(payload.ID, id),
		Channel:     payload.Channel,
		Phone:       payload.Phone,
		Username:    payload.Username,
		Business:    payload.Business,
		Service:     payload.Service,
		Branch:      payload.Branch,
		Employee:    payload.Employee,
		StartTime:   startTime,
		Description: payload.Description,
	}, nil
}

func (r *NotificationRepository) getPayload(ctx context.Context, id string) (notificationPayload, error) {
	raw, err := r.client.Get(ctx, notificationKey(id)).Result()
	if err != nil {
		return notificationPayload{}, err
	}

	var payload notificationPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return notificationPayload{}, err
	}
	if payload.ID == "" {
		payload.ID = id
	}

	return payload, nil
}

func (r *NotificationRepository) savePayload(ctx context.Context, payload *notificationPayload) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, notificationKey(payload.ID), raw, 0).Err()
}

func parseStartTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.RFC3339, value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func notificationKey(id string) string {
	return notificationKeyPrefix + id
}

func timestampScore(value time.Time) string {
	return strconv.FormatInt(value.UTC().Unix(), 10)
}
