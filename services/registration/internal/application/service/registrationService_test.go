package service

import (
	"Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func TestRegister_WhenUserDoesNotExist_ShouldSavePendingRegistrationAndQueueEmail(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	output, err := svc.Register(ctx, RegisterInput{
		Email:        "owner@example.com",
		Password:     "secret-password",
		BusinessName: "Demo Business",
		BusinessType: "service_company",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if output.Status != "pending" {
		t.Fatalf("expected pending status, got %q", output.Status)
	}
	if output.RegistrationID == "" {
		t.Fatal("expected registration id")
	}

	saved, ok := pendingRepo.items[output.RegistrationID]
	if !ok {
		t.Fatalf("expected pending registration %q to be saved", output.RegistrationID)
	}
	if saved.Email != "owner@example.com" || saved.BusinessName != "Demo Business" || saved.BusinessType != "service_company" {
		t.Fatalf("unexpected pending registration: %#v", saved)
	}
	if saved.PasswordHash == "" || saved.PasswordHash == "secret-password" {
		t.Fatalf("expected password to be hashed, got %q", saved.PasswordHash)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(saved.PasswordHash), []byte("secret-password")); err != nil {
		t.Fatalf("saved password hash does not match password: %v", err)
	}
	if len(saved.Code) != 6 {
		t.Fatalf("expected six-digit code, got %q", saved.Code)
	}
	assertQueueLen(t, queue, 1)
}

func TestRegister_WhenUserAlreadyExists_ShouldReturnErrorWithoutSavingPendingRegistration(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	userRepo.existingEmails["owner@example.com"] = true
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	_, err := svc.Register(ctx, RegisterInput{
		Email:        "owner@example.com",
		Password:     "secret-password",
		BusinessName: "Demo Business",
		BusinessType: "service_company",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists error, got %v", err)
	}
	if len(pendingRepo.items) != 0 {
		t.Fatalf("expected no pending registrations, got %d", len(pendingRepo.items))
	}
	assertQueueLen(t, queue, 0)
}

func TestVerify_WhenCodeMatches_ShouldCreateUserWithBusinessAndDeletePendingRegistration(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	pendingRepo.items["registration-1"] = pending.PendingRegistration{
		ID:           "registration-1",
		Email:        "owner@example.com",
		PasswordHash: "hash",
		BusinessName: "Demo Business",
		BusinessType: "service_company",
		Code:         "123456",
	}

	if err := svc.Verify(ctx, VerifyInput{RegistrationID: "registration-1", Code: "123456"}); err != nil {
		t.Fatalf("verify: %v", err)
	}

	if userRepo.createCalls != 1 {
		t.Fatalf("expected one create call, got %d", userRepo.createCalls)
	}
	if userRepo.created[0].ClientSlug == nil || *userRepo.created[0].ClientSlug == "" {
		t.Fatalf("expected generated client slug, got %#v", userRepo.created[0].ClientSlug)
	}
	if !pendingRepo.deleted["registration-1"] {
		t.Fatal("expected pending registration to be deleted")
	}
}

func TestVerify_WhenSlugCollides_ShouldRetryCreateUserWithBusiness(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	userRepo.createErrors = []error{uniqueViolation("businesses_registration_slug_uindex"), nil}
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	pendingRepo.items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	if err := svc.Verify(ctx, VerifyInput{RegistrationID: "registration-1", Code: "123456"}); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if userRepo.createCalls != 2 {
		t.Fatalf("expected retry after slug collision, got %d create calls", userRepo.createCalls)
	}
}

func TestVerify_WhenCodeDoesNotMatch_ShouldReturnErrorWithoutCreatingUser(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	pendingRepo.items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	err := svc.Verify(ctx, VerifyInput{RegistrationID: "registration-1", Code: "000000"})
	if err == nil {
		t.Fatal("expected error")
	}
	if userRepo.createCalls != 0 {
		t.Fatalf("expected no create calls, got %d", userRepo.createCalls)
	}
	if pendingRepo.deleted["registration-1"] {
		t.Fatal("pending registration should not be deleted")
	}
}

func TestResendCode_WhenPendingRegistrationExists_ShouldQueueEmail(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	pendingRepo.items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	if err := svc.ResendCode(ctx, ResendInput{RegistrationID: "registration-1"}); err != nil {
		t.Fatalf("resend code: %v", err)
	}
	assertQueueLen(t, queue, 1)
}

func TestRecoverPassword_WhenUserExists_ShouldSaveRecoveryAndQueueEmail(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	userRepo.existingEmails["owner@example.com"] = true
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	output, err := svc.RecoverPassword(ctx, PasswordRecoveryInput{Email: "owner@example.com"})
	if err != nil {
		t.Fatalf("recover password: %v", err)
	}

	if output.Status != "password_recovery_pending" {
		t.Fatalf("expected password recovery pending status, got %q", output.Status)
	}
	if _, ok := recoveryRepo.items[output.RecoveryID]; !ok {
		t.Fatalf("expected recovery %q to be saved", output.RecoveryID)
	}
	assertQueueLen(t, queue, 1)
}

func TestRecoverPassword_WhenUserDoesNotExist_ShouldReturnPendingStatusWithoutSavingRecovery(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	output, err := svc.RecoverPassword(ctx, PasswordRecoveryInput{Email: "missing@example.com"})
	if err != nil {
		t.Fatalf("recover password: %v", err)
	}

	if output.Status != "password_recovery_pending" {
		t.Fatalf("expected password recovery pending status, got %q", output.Status)
	}
	if output.RecoveryID == "" {
		t.Fatal("expected recovery id")
	}
	if len(recoveryRepo.items) != 0 {
		t.Fatalf("expected no saved recoveries, got %d", len(recoveryRepo.items))
	}
	assertQueueLen(t, queue, 0)
}

func TestConfirmPasswordRecovery_WhenCodeMatches_ShouldUpdatePasswordQueueEmailAndDeleteRecovery(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	recoveryRepo.items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	userRepo.updateResults["owner@example.com"] = true

	if err := svc.ConfirmPasswordRecovery(ctx, PasswordRecoveryVerifyInput{RecoveryID: "recovery-1", Code: "123456"}); err != nil {
		t.Fatalf("confirm password recovery: %v", err)
	}

	hash := userRepo.updatedPasswords["owner@example.com"]
	if hash == "" {
		t.Fatal("expected password hash to be updated")
	}
	if strings.HasPrefix(hash, "tmp-") {
		t.Fatalf("expected password hash, got raw temporary password %q", hash)
	}
	assertQueueLen(t, queue, 1)
	if !recoveryRepo.deleted["recovery-1"] {
		t.Fatal("expected recovery to be deleted")
	}
}

func TestConfirmPasswordRecovery_WhenCodeDoesNotMatch_ShouldReturnErrorWithoutUpdatingPassword(t *testing.T) {
	ctx := context.Background()
	pendingRepo := newFakePendingRepo()
	recoveryRepo := newFakeRecoveryRepo()
	userRepo := newFakeUserRepo()
	queue := newTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue)

	recoveryRepo.items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	err := svc.ConfirmPasswordRecovery(ctx, PasswordRecoveryVerifyInput{RecoveryID: "recovery-1", Code: "000000"})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(userRepo.updatedPasswords) != 0 {
		t.Fatalf("expected no password updates, got %d", len(userRepo.updatedPasswords))
	}
	assertQueueLen(t, queue, 0)
}

func newTestEmailQueue() *email.EmailQueue {
	return email.NewEmailQueue(noopEmailSender{}, config.QueueConfig{
		NumWorkers: 0,
		RateLimit:  time.Hour,
		WrkTimeOut: time.Second,
	})
}

func assertQueueLen(t *testing.T, queue *email.EmailQueue, expected int) {
	t.Helper()

	queueLen, _, _ := queue.GetStats()
	if queueLen != expected {
		t.Fatalf("expected email queue len %d, got %d", expected, queueLen)
	}
}

type noopEmailSender struct{}

func (noopEmailSender) SendEmail(context.Context, email.EmailMessage) error {
	return nil
}

type fakePendingRepo struct {
	items   map[string]pending.PendingRegistration
	deleted map[string]bool
	err     error
}

func newFakePendingRepo() *fakePendingRepo {
	return &fakePendingRepo{
		items:   make(map[string]pending.PendingRegistration),
		deleted: make(map[string]bool),
	}
}

func (r *fakePendingRepo) Save(_ context.Context, item pending.PendingRegistration) error {
	if r.err != nil {
		return r.err
	}
	r.items[item.ID] = item
	return nil
}

func (r *fakePendingRepo) Get(_ context.Context, registrationID string) (pending.PendingRegistration, error) {
	if r.err != nil {
		return pending.PendingRegistration{}, r.err
	}
	item, ok := r.items[registrationID]
	if !ok {
		return pending.PendingRegistration{}, errors.New("pending registration not found")
	}
	return item, nil
}

func (r *fakePendingRepo) Delete(_ context.Context, registrationID string) error {
	if r.err != nil {
		return r.err
	}
	r.deleted[registrationID] = true
	delete(r.items, registrationID)
	return nil
}

type fakeRecoveryRepo struct {
	items   map[string]recovery.PasswordRecovery
	deleted map[string]bool
	err     error
}

func newFakeRecoveryRepo() *fakeRecoveryRepo {
	return &fakeRecoveryRepo{
		items:   make(map[string]recovery.PasswordRecovery),
		deleted: make(map[string]bool),
	}
}

func (r *fakeRecoveryRepo) SaveRecovery(_ context.Context, item recovery.PasswordRecovery) error {
	if r.err != nil {
		return r.err
	}
	r.items[item.ID] = item
	return nil
}

func (r *fakeRecoveryRepo) GetRecovery(_ context.Context, recoveryID string) (recovery.PasswordRecovery, error) {
	if r.err != nil {
		return recovery.PasswordRecovery{}, r.err
	}
	item, ok := r.items[recoveryID]
	if !ok {
		return recovery.PasswordRecovery{}, errors.New("password recovery not found")
	}
	return item, nil
}

func (r *fakeRecoveryRepo) DeleteRecovery(_ context.Context, recoveryID string) error {
	if r.err != nil {
		return r.err
	}
	r.deleted[recoveryID] = true
	delete(r.items, recoveryID)
	return nil
}

type fakeUserRepo struct {
	existingEmails   map[string]bool
	updateResults    map[string]bool
	updatedPasswords map[string]string
	created          []pending.PendingRegistration
	createErrors     []error
	createCalls      int
	err              error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		existingEmails:   make(map[string]bool),
		updateResults:    make(map[string]bool),
		updatedPasswords: make(map[string]string),
	}
}

func (r *fakeUserRepo) CreateUserWithBusiness(_ context.Context, item pending.PendingRegistration) error {
	r.createCalls++
	r.created = append(r.created, item)
	if len(r.createErrors) == 0 {
		return nil
	}

	err := r.createErrors[0]
	r.createErrors = r.createErrors[1:]
	return err
}

func (r *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return r.existingEmails[email], nil
}

func (r *fakeUserRepo) UpdatePasswordByEmail(_ context.Context, email string, passwordHash string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	r.updatedPasswords[email] = passwordHash
	return r.updateResults[email], nil
}

func uniqueViolation(constraint string) error {
	return &pgconn.PgError{
		Code:           "23505",
		ConstraintName: constraint,
	}
}
