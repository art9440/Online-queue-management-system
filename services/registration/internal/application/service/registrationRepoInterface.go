package service

import (
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"context"
)

type PendingRepo interface {
	Save(ctx context.Context, pending *pending.PendingRegistration) error
	Get(ctx context.Context, registrationID string) (pending.PendingRegistration, error)
	Delete(ctx context.Context, registrationID string) error
	GetAndValidate(ctx context.Context, id, code string) (pending.PendingRegistration, error)
}

type RecoveryRepo interface {
	SaveRecovery(ctx context.Context, recovery recovery.PasswordRecovery) error
	GetRecovery(ctx context.Context, recoveryID string) (recovery.PasswordRecovery, error)
	DeleteRecovery(ctx context.Context, recoveryID string) error
}

type UserRepo interface {
	CreateUserWithBusiness(ctx context.Context, p *pending.PendingRegistration) error
	GetUserByEmail(ctx context.Context, email string) (bool, error)
	UpdatePasswordByEmail(ctx context.Context, email, passwordHash string) (bool, error)
}

// EmailQueue accepts emails for asynchronous delivery; Enqueue fails if the email
// could not be stored durably, so the caller can report the error instead of losing it.
type EmailQueue interface {
	Enqueue(ctx context.Context, msg email.EmailMessage) error
}
