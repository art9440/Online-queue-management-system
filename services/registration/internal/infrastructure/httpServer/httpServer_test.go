package httpserver

import (
	"Online-queue-management-system/libs/config"
	"Online-queue-management-system/libs/email"
	"Online-queue-management-system/services/registration/internal/application/service"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRegister_WhenRequestIsValid_ShouldReturnPendingResponse(t *testing.T) {
	server, _ := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{
		"email":"owner@example.com",
		"password":"secret-password",
		"business_name":"Demo Business",
		"business_type":"service_company"
	}`))
	rec := httptest.NewRecorder()

	server.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body map[string]string
	decodeJSON(t, rec.Body, &body)
	if body["status"] != "pending" {
		t.Fatalf("expected pending status, got %q", body["status"])
	}
	if body["registration_id"] == "" {
		t.Fatal("expected registration_id")
	}
}

func TestRegister_WhenBodyIsInvalidJSON_ShouldReturnBadRequest(t *testing.T) {
	server, _ := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{`))
	rec := httptest.NewRecorder()

	server.Register(rec, req)

	assertErrorResponse(t, rec, http.StatusBadRequest, "invalid request")
}

func TestVerify_WhenCodeIsInvalid_ShouldReturnBadRequest(t *testing.T) {
	server, repos := newTestServer()
	repos.pending.items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	req := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(`{
		"registration_id":"registration-1",
		"code":"000000"
	}`))
	rec := httptest.NewRecorder()

	server.Verify(rec, req)

	assertErrorResponse(t, rec, http.StatusBadRequest, "invalid code")
}

func TestResendCode_WhenPendingRegistrationExists_ShouldReturnResendedStatus(t *testing.T) {
	server, repos := newTestServer()
	repos.pending.items["registration-1"] = pending.PendingRegistration{
		ID:    "registration-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	req := httptest.NewRequest(http.MethodPost, "/resend", strings.NewReader(`{"registration_id":"registration-1"}`))
	rec := httptest.NewRecorder()

	server.ResendCode(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body map[string]string
	decodeJSON(t, rec.Body, &body)
	if body["status"] != "resended" {
		t.Fatalf("expected resended status, got %q", body["status"])
	}
}

func TestPasswordRecovery_WhenEmailIsMissing_ShouldReturnBadRequest(t *testing.T) {
	server, _ := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/password-recovery", strings.NewReader(`{"email":""}`))
	rec := httptest.NewRecorder()

	server.PasswordRecovery(rec, req)

	assertErrorResponse(t, rec, http.StatusBadRequest, "email is required")
}

func TestConfirmPasswordRecovery_WhenCodeMatches_ShouldReturnCompletedStatus(t *testing.T) {
	server, repos := newTestServer()
	repos.recovery.items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	repos.user.updateResults["owner@example.com"] = true
	req := httptest.NewRequest(http.MethodPost, "/password-recovery/confirm", strings.NewReader(`{
		"recovery_id":"recovery-1",
		"code":"123456"
	}`))
	rec := httptest.NewRecorder()

	server.ConfirmPasswordRecovery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body map[string]string
	decodeJSON(t, rec.Body, &body)
	if body["status"] != "password_recovery_completed" {
		t.Fatalf("expected password recovery completed status, got %q", body["status"])
	}
}

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedError string) {
	t.Helper()

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var body map[string]string
	decodeJSON(t, rec.Body, &body)
	if body["error"] != expectedError {
		t.Fatalf("expected error %q, got %q", expectedError, body["error"])
	}
}

func decodeJSON(t *testing.T, body *bytes.Buffer, target any) {
	t.Helper()

	if err := json.NewDecoder(body).Decode(target); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
}

func newTestServer() (*HttpServer, *testRepos) {
	repos := &testRepos{
		pending:  newFakePendingRepo(),
		recovery: newFakeRecoveryRepo(),
		user:     newFakeUserRepo(),
	}
	queue := email.NewEmailQueue(noopEmailSender{}, config.QueueConfig{
		NumWorkers: 0,
		RateLimit:  time.Hour,
		WrkTimeOut: time.Second,
	})
	svc := service.NewRegistrationService(repos.pending, repos.recovery, repos.user, queue)
	return NewHttpServer(svc), repos
}

type testRepos struct {
	pending  *fakePendingRepo
	recovery *fakeRecoveryRepo
	user     *fakeUserRepo
}

type noopEmailSender struct{}

func (noopEmailSender) SendEmail(context.Context, email.EmailMessage) error {
	return nil
}

type fakePendingRepo struct {
	items map[string]pending.PendingRegistration
}

func newFakePendingRepo() *fakePendingRepo {
	return &fakePendingRepo{items: make(map[string]pending.PendingRegistration)}
}

func (r *fakePendingRepo) Save(_ context.Context, item pending.PendingRegistration) error {
	r.items[item.ID] = item
	return nil
}

func (r *fakePendingRepo) Get(_ context.Context, registrationID string) (pending.PendingRegistration, error) {
	item, ok := r.items[registrationID]
	if !ok {
		return pending.PendingRegistration{}, errors.New("pending registration not found")
	}
	return item, nil
}

func (r *fakePendingRepo) Delete(_ context.Context, registrationID string) error {
	delete(r.items, registrationID)
	return nil
}

type fakeRecoveryRepo struct {
	items map[string]recovery.PasswordRecovery
}

func newFakeRecoveryRepo() *fakeRecoveryRepo {
	return &fakeRecoveryRepo{items: make(map[string]recovery.PasswordRecovery)}
}

func (r *fakeRecoveryRepo) SaveRecovery(_ context.Context, item recovery.PasswordRecovery) error {
	r.items[item.ID] = item
	return nil
}

func (r *fakeRecoveryRepo) GetRecovery(_ context.Context, recoveryID string) (recovery.PasswordRecovery, error) {
	item, ok := r.items[recoveryID]
	if !ok {
		return recovery.PasswordRecovery{}, errors.New("password recovery not found")
	}
	return item, nil
}

func (r *fakeRecoveryRepo) DeleteRecovery(_ context.Context, recoveryID string) error {
	delete(r.items, recoveryID)
	return nil
}

type fakeUserRepo struct {
	existingEmails   map[string]bool
	updateResults    map[string]bool
	updatedPasswords map[string]string
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		existingEmails:   make(map[string]bool),
		updateResults:    make(map[string]bool),
		updatedPasswords: make(map[string]string),
	}
}

func (r *fakeUserRepo) CreateUserWithBusiness(_ context.Context, item pending.PendingRegistration) error {
	r.existingEmails[item.Email] = true
	return nil
}

func (r *fakeUserRepo) GetUserByEmail(_ context.Context, email string) (bool, error) {
	return r.existingEmails[email], nil
}

func (r *fakeUserRepo) UpdatePasswordByEmail(_ context.Context, email string, passwordHash string) (bool, error) {
	r.updatedPasswords[email] = passwordHash
	return r.updateResults[email], nil
}
