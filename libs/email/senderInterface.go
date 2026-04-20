package email

import (
	"context"
)

type Sender interface {
	SendEmail(ctx context.Context, msg EmailMessage) error
}
