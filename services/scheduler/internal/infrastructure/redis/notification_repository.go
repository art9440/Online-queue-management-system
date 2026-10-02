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
	scheduledNotificationsKey  = "notifications:scheduled"
	processingNotificationsKey = "notifications:processing"
	notificationKeyPrefix      = "notification:"

	// Sent and failed notifications are kept for inspection, then expire.
	finishedNotificationTTL = 7 * 24 * time.Hour
)

// claimDueScript atomically moves due notifications from the schedule to the processing set,
// scored by the lease deadline. Several scheduler replicas never get the same notification,
// and a notification is never "in flight" outside Redis (12-factor VI, VIII).
var claimDueScript = goredis.NewScript(`
local ids = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1], 'LIMIT', 0, tonumber(ARGV[2]))
for _, id in ipairs(ids) do
	redis.call('ZREM', KEYS[1], id)
	redis.call('ZADD', KEYS[2], ARGV[3], id)
end
return ids
`)

// requeueExpiredScript returns notifications whose lease has expired (their scheduler died)
// back to the schedule, due immediately.
var requeueExpiredScript = goredis.NewScript(`
local ids = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
for _, id in ipairs(ids) do
	redis.call('ZREM', KEYS[1], id)
	redis.call('ZADD', KEYS[2], ARGV[1], id)
end
return #ids
`)

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
	Email       string `json:"email"`
	Subject     string `json:"subject"`
	Body        string `json:"body"`
	HTMLBody    string `json:"html_body"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts,omitempty"`
	SentAt      string `json:"sent_at,omitempty"`
	FailedAt    string `json:"failed_at,omitempty"`
	FailReason  string `json:"fail_reason,omitempty"`
}

func NewNotificationRepository(client *goredis.Client) *NotificationRepository {
	return &NotificationRepository{client: client}
}

// ClaimDue takes up to limit due notifications for this scheduler until leaseUntil.
func (r *NotificationRepository) ClaimDue(
	ctx context.Context,
	now time.Time,
	limit int64,
	leaseUntil time.Time,
) ([]service.Notification, error) {
	ids, err := claimDueScript.Run(
		ctx,
		r.client,
		[]string{scheduledNotificationsKey, processingNotificationsKey},
		timestampScore(now),
		limit,
		timestampScore(leaseUntil),
	).StringSlice()
	if err != nil {
		return nil, err
	}

	notifications := make([]service.Notification, 0, len(ids))
	for _, id := range ids {
		notification, err := r.getNotification(ctx, id)
		if errors.Is(err, goredis.Nil) {
			// Payload is gone: nothing to send.
			_ = r.client.ZRem(ctx, processingNotificationsKey, id).Err()
			continue
		}
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// RequeueExpired returns notifications with an expired lease to the schedule.
func (r *NotificationRepository) RequeueExpired(ctx context.Context, now time.Time) (int64, error) {
	return requeueExpiredScript.Run(
		ctx,
		r.client,
		[]string{processingNotificationsKey, scheduledNotificationsKey},
		timestampScore(now),
	).Int64()
}

// Release gives claimed but unprocessed notifications back to the schedule, due at the given time.
func (r *NotificationRepository) Release(ctx context.Context, ids []string, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		for _, id := range ids {
			pipe.ZRem(ctx, processingNotificationsKey, id)
			pipe.ZAdd(ctx, scheduledNotificationsKey, goredis.Z{Score: float64(at.UTC().Unix()), Member: id})
		}
		return nil
	})
	return err
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

	return r.finish(ctx, &payload)
}

func (r *NotificationRepository) MarkFailed(ctx context.Context, notificationID string, failedAt time.Time, reason string) error {
	payload, err := r.getPayload(ctx, notificationID)
	if err != nil {
		return err
	}

	payload.Status = "failed"
	payload.Attempts++
	payload.FailedAt = failedAt.UTC().Format(time.RFC3339)
	payload.FailReason = reason

	return r.finish(ctx, &payload)
}

// Retry schedules another attempt at retryAt and records the failure.
func (r *NotificationRepository) Retry(ctx context.Context, notificationID string, retryAt time.Time, reason string) error {
	payload, err := r.getPayload(ctx, notificationID)
	if err != nil {
		return err
	}

	payload.Status = "retrying"
	payload.Attempts++
	payload.FailedAt = time.Now().UTC().Format(time.RFC3339)
	payload.FailReason = reason

	raw, err := json.Marshal(&payload)
	if err != nil {
		return err
	}

	_, err = r.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.Set(ctx, notificationKey(payload.ID), raw, 0)
		pipe.ZRem(ctx, processingNotificationsKey, payload.ID)
		pipe.ZAdd(ctx, scheduledNotificationsKey, goredis.Z{Score: float64(retryAt.UTC().Unix()), Member: payload.ID})
		return nil
	})
	return err
}

// finish stores the final status and removes the notification from processing.
func (r *NotificationRepository) finish(ctx context.Context, payload *notificationPayload) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = r.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.Set(ctx, notificationKey(payload.ID), raw, finishedNotificationTTL)
		pipe.ZRem(ctx, processingNotificationsKey, payload.ID)
		return nil
	})
	return err
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
		Email:       payload.Email,
		Subject:     payload.Subject,
		Body:        payload.Body,
		HTMLBody:    payload.HTMLBody,
		Attempts:    payload.Attempts,
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
