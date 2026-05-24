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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := otel.Tracer("Online-queue-management-system/scheduler").Start(
		ctx,
		"POST "+sendNotificationPath,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("http.request.method", http.MethodPost),
			attribute.String("url.full", d.botURL+sendNotificationPath),
			attribute.String("notification.channel", "telegram"),
		),
	)
	defer span.End()

	body, err := json.Marshal(requestFromNotification(notification))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		d.botURL+sendNotificationPath,
		bytes.NewReader(body),
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := d.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
		return fmt.Errorf("telegram bot returned status %d", resp.StatusCode)
	}

	span.SetAttributes(attribute.Int("http.response.status_code", resp.StatusCode))
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
