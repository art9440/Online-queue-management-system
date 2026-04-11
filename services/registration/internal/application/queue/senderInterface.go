package queue

import (
	"Online-queue-management-system/services/registration/internal/application/email"
	"context"
)

type Sender interface {
	SendEmail(ctx context.Context, msg email.EmailMessage) error
}
