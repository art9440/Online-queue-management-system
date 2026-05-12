package service

import (
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"Online-queue-management-system/services/registration/internal/mocks"
	"context"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRegister_WhenUserDoesNotExist_ShouldSavePendingRegistrationAndQueueEmail(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

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

	saved, ok := pendingRepo.Items[output.RegistrationID]
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

func TestVerify_WhenUserAlreadyExists_ShouldReturnNilAndDeletePendingRegistration(t *testing.T) {
	ctx := context.Background()

	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()

	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "test")

	regID := "reg-1"

	pendingRepo.Items[regID] = pending.PendingRegistration{
		ID:           regID,
		Email:        "owner@example.com",
		PasswordHash: "hashed-password",
		BusinessName: "Demo Business",
		BusinessType: "service_company",
		Code:         "123456",
	}

	userRepo.CreateErrors = []error{
		mocks.UniqueViolation("users_login_key"),
	}

	err := svc.Verify(ctx, VerifyInput{
		RegistrationID: regID,
		Code:           "123456",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if _, ok := pendingRepo.Items[regID]; ok {
		t.Fatal("expected pending registration to be deleted")
	}

	if !pendingRepo.Deleted[regID] {
		t.Fatal("expected pending registration to be marked as deleted")
	}
}

func TestVerify_WhenCodeMatches_ShouldCreateUserWithBusinessAndDeletePendingRegistration(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	pendingRepo.Items["registration-1"] = pending.PendingRegistration{
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

	if userRepo.CreateCalls != 1 {
		t.Fatalf("expected one create call, got %d", userRepo.CreateCalls)
	}
	if userRepo.Created[0].ClientSlug == nil || *userRepo.Created[0].ClientSlug == "" {
		t.Fatalf("expected generated client slug, got %#v", userRepo.Created[0].ClientSlug)
	}
	if !pendingRepo.Deleted["registration-1"] {
		t.Fatal("expected pending registration to be deleted")
	}
}

func TestVerify_WhenSlugCollides_ShouldRetryCreateUserWithBusiness(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	userRepo.CreateErrors = []error{mocks.UniqueViolation("businesses_registration_slug_uindex"), nil}
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	pendingRepo.Items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	if err := svc.Verify(ctx, VerifyInput{RegistrationID: "registration-1", Code: "123456"}); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if userRepo.CreateCalls != 2 {
		t.Fatalf("expected retry after slug collision, got %d create calls", userRepo.CreateCalls)
	}
}

func TestVerify_WhenCodeDoesNotMatch_ShouldReturnErrorWithoutCreatingUser(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	pendingRepo.Items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	err := svc.Verify(ctx, VerifyInput{RegistrationID: "registration-1", Code: "000000"})
	if err == nil {
		t.Fatal("expected error")
	}
	if userRepo.CreateCalls != 0 {
		t.Fatalf("expected no create calls, got %d", userRepo.CreateCalls)
	}
	if pendingRepo.Deleted["registration-1"] {
		t.Fatal("pending registration should not be deleted")
	}
}

func TestResendCode_WhenPendingRegistrationExists_ShouldQueueEmail(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	pendingRepo.Items["registration-1"] = pending.PendingRegistration{
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
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	userRepo.ExistingEmails["owner@example.com"] = true
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	output, err := svc.RecoverPassword(ctx, PasswordRecoveryInput{Email: "owner@example.com"})
	if err != nil {
		t.Fatalf("recover password: %v", err)
	}

	if output.Status != "password_recovery_pending" {
		t.Fatalf("expected password recovery pending status, got %q", output.Status)
	}
	if _, ok := recoveryRepo.Items[output.RecoveryID]; !ok {
		t.Fatalf("expected recovery %q to be saved", output.RecoveryID)
	}
	assertQueueLen(t, queue, 1)
}

func TestRecoverPassword_WhenUserDoesNotExist_ShouldReturnPendingStatusWithoutSavingRecovery(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

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
	if len(recoveryRepo.Items) != 0 {
		t.Fatalf("expected no saved recoveries, got %d", len(recoveryRepo.Items))
	}
	assertQueueLen(t, queue, 0)
}

func TestConfirmPasswordRecovery_WhenCodeMatches_ShouldUpdatePasswordQueueEmailAndDeleteRecovery(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	recoveryRepo.Items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	userRepo.UpdateResults["owner@example.com"] = true

	if err := svc.ConfirmPasswordRecovery(ctx, PasswordRecoveryVerifyInput{RecoveryID: "recovery-1", Code: "123456"}); err != nil {
		t.Fatalf("confirm password recovery: %v", err)
	}

	hash := userRepo.UpdatedPasswords["owner@example.com"]
	if hash == "" {
		t.Fatal("expected password hash to be updated")
	}
	if strings.HasPrefix(hash, "tmp-") {
		t.Fatalf("expected password hash, got raw temporary password %q", hash)
	}
	assertQueueLen(t, queue, 1)
	if !recoveryRepo.Deleted["recovery-1"] {
		t.Fatal("expected recovery to be deleted")
	}
}

func TestConfirmPasswordRecovery_WhenCodeDoesNotMatch_ShouldReturnErrorWithoutUpdatingPassword(t *testing.T) {
	ctx := context.Background()
	pendingRepo := mocks.NewPendingRepo()
	recoveryRepo := mocks.NewRecoveryRepo()
	userRepo := mocks.NewUserRepo()
	queue := mocks.NewTestEmailQueue()
	svc := NewRegistrationService(pendingRepo, recoveryRepo, userRepo, queue, "")

	recoveryRepo.Items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}

	err := svc.ConfirmPasswordRecovery(ctx, PasswordRecoveryVerifyInput{RecoveryID: "recovery-1", Code: "000000"})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(userRepo.UpdatedPasswords) != 0 {
		t.Fatalf("expected no password updates, got %d", len(userRepo.UpdatedPasswords))
	}
	assertQueueLen(t, queue, 0)
}

func assertQueueLen(t *testing.T, queue *email.EmailQueue, expected int) {
	t.Helper()

	queueLen, _, _ := queue.GetStats()
	if queueLen != expected {
		t.Fatalf("expected email queue len %d, got %d", expected, queueLen)
	}
}
