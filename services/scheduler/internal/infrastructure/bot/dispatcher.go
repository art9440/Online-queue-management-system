package bot

import (
	"Online-queue-management-system/services/scheduler/internal/application/service"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sendNotificationPath = "/telegram/notifications"

type Dispatcher struct {
	botURL     string
	httpClient *http.Client
}

type sendNotificationRequest struct {
	Phone       string `json:"phone,omitempty"`
	Username    string `json:"username,omitempty"`
	Business    string `json:"business,omitempty"`
	Service     string `json:"service,omitempty"`
	Branch      string `json:"branch,omitempty"`
	Employee    string `json:"employee,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	Description string `json:"description,omitempty"`
}

func NewDispatcher(botURL string) *Dispatcher {
	return &Dispatcher{
		botURL: strings.TrimRight(botURL, "/"),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, notification *service.Notification) error {
	body, err := json.Marshal(requestFromNotification(notification))
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		d.botURL+sendNotificationPath,
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("telegram bot returned status %d", resp.StatusCode)
	}

	return nil
}

func requestFromNotification(notification *service.Notification) sendNotificationRequest {
	var startTime string
	if !notification.StartTime.IsZero() {
		startTime = notification.StartTime.Format(time.RFC3339)
	}

	return sendNotificationRequest{
		Phone:       notification.Phone,
		Username:    notification.Username,
		Business:    notification.Business,
		Service:     notification.Service,
		Branch:      notification.Branch,
		Employee:    notification.Employee,
		StartTime:   startTime,
		Description: notification.Description,
	}
}
