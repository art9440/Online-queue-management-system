package service

import (
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"context"
)

type PendingRepo interface {
	Save(ctx context.Context, pending pending.PendingRegistration) error
	Get(ctx context.Context, registrationID string) (pending.PendingRegistration, error)
	Delete(ctx context.Context, registrationID string) error
}

type UserRepo interface {
	CreateUserWithBusiness(ctx context.Context, p pending.PendingRegistration) error
	GetUserByEmail(ctx context.Context, email string) (bool, error)
}
