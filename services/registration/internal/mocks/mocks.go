package mocks

import (
	"Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/registration/internal/domain"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type NoopEmailSender struct{}

func (NoopEmailSender) SendEmail(context.Context, email.EmailMessage) error {
	return nil
}

func NewTestEmailQueue() *email.EmailQueue {
	return email.NewEmailQueue(NoopEmailSender{}, config.QueueConfig{
		NumWorkers: 0,
		RateLimit:  time.Hour,
		WrkTimeOut: time.Second,
	})
}

type PendingRepo struct {
	Items   map[string]pending.PendingRegistration
	Deleted map[string]bool
	Err     error
}

func NewPendingRepo() *PendingRepo {
	return &PendingRepo{
		Items:   make(map[string]pending.PendingRegistration),
		Deleted: make(map[string]bool),
	}
}

func (r *PendingRepo) Save(_ context.Context, item pending.PendingRegistration) error {
	if r.Err != nil {
		return r.Err
	}
	r.Items[item.ID] = item
	return nil
}

func (r *PendingRepo) Get(_ context.Context, registrationID string) (pending.PendingRegistration, error) {
	if r.Err != nil {
		return pending.PendingRegistration{}, r.Err
	}
	item, ok := r.Items[registrationID]
	if !ok {
		return pending.PendingRegistration{}, errors.New("pending registration not found")
	}
	return item, nil
}

func (r *PendingRepo) Delete(_ context.Context, registrationID string) error {
	if r.Err != nil {
		return r.Err
	}
	r.Deleted[registrationID] = true
	delete(r.Items, registrationID)
	return nil
}

func (r *PendingRepo) GetAndValidate(_ context.Context, registrationID string, code string) (pending.PendingRegistration, error) {
	if r.Err != nil {
		return pending.PendingRegistration{}, r.Err
	}

	item, ok := r.Items[registrationID]
	if !ok {
		return pending.PendingRegistration{}, domain.ErrNotFound
	}

	if item.Code != code {
		return pending.PendingRegistration{}, domain.ErrInvalidCode
	}

	delete(r.Items, registrationID)
	r.Deleted[registrationID] = true

	return item, nil
}

type RecoveryRepo struct {
	Items   map[string]recovery.PasswordRecovery
	Deleted map[string]bool
	Err     error
}

func NewRecoveryRepo() *RecoveryRepo {
	return &RecoveryRepo{
		Items:   make(map[string]recovery.PasswordRecovery),
		Deleted: make(map[string]bool),
	}
}

func (r *RecoveryRepo) SaveRecovery(_ context.Context, item recovery.PasswordRecovery) error {
	if r.Err != nil {
		return r.Err
	}
	r.Items[item.ID] = item
	return nil
}

func (r *RecoveryRepo) GetRecovery(_ context.Context, recoveryID string) (recovery.PasswordRecovery, error) {
	if r.Err != nil {
		return recovery.PasswordRecovery{}, r.Err
	}
	item, ok := r.Items[recoveryID]
	if !ok {
		return recovery.PasswordRecovery{}, errors.New("password recovery not found")
	}
	return item, nil
}

func (r *RecoveryRepo) DeleteRecovery(_ context.Context, recoveryID string) error {
	if r.Err != nil {
		return r.Err
	}
	r.Deleted[recoveryID] = true
	delete(r.Items, recoveryID)
	return nil
}

type UserRepo struct {
	ExistingEmails   map[string]bool
	UpdateResults    map[string]bool
	UpdatedPasswords map[string]string
	Created          []pending.PendingRegistration
	CreateErrors     []error
	CreateCalls      int
	Err              error
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		ExistingEmails:   make(map[string]bool),
		UpdateResults:    make(map[string]bool),
		UpdatedPasswords: make(map[string]string),
	}
}

func (r *UserRepo) CreateUserWithBusiness(_ context.Context, item pending.PendingRegistration) error {
	r.CreateCalls++
	r.Created = append(r.Created, item)
	if len(r.CreateErrors) == 0 {
		return nil
	}

	err := r.CreateErrors[0]
	r.CreateErrors = r.CreateErrors[1:]
	return err
}

func (r *UserRepo) GetUserByEmail(_ context.Context, email string) (bool, error) {
	if r.Err != nil {
		return false, r.Err
	}
	return r.ExistingEmails[email], nil
}

func (r *UserRepo) UpdatePasswordByEmail(_ context.Context, email string, passwordHash string) (bool, error) {
	if r.Err != nil {
		return false, r.Err
	}
	r.UpdatedPasswords[email] = passwordHash
	return r.UpdateResults[email], nil
}

func UniqueViolation(constraint string) error {
	return &pgconn.PgError{
		Code:           "23505",
		ConstraintName: constraint,
	}
}
