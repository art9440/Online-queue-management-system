package httpserver

import (
	"Online-queue-management-system/services/registration/internal/application/service"
	"Online-queue-management-system/services/registration/internal/domain/pending"
	"Online-queue-management-system/services/registration/internal/domain/recovery"
	"Online-queue-management-system/services/registration/internal/mocks"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
	repos.pending.Items["registration-1"] = pending.PendingRegistration{
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
	repos.pending.Items["registration-1"] = pending.PendingRegistration{
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
	repos.recovery.Items["recovery-1"] = recovery.PasswordRecovery{
		ID:    "recovery-1",
		Email: "owner@example.com",
		Code:  "123456",
	}
	repos.user.UpdateResults["owner@example.com"] = true
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
		pending:  mocks.NewPendingRepo(),
		recovery: mocks.NewRecoveryRepo(),
		user:     mocks.NewUserRepo(),
	}
	queue := mocks.NewTestEmailQueue()
	svc := service.NewRegistrationService(repos.pending, repos.recovery, repos.user, queue, "")
	return NewHttpServer(svc), repos
}

type testRepos struct {
	pending  *mocks.PendingRepo
	recovery *mocks.RecoveryRepo
	user     *mocks.UserRepo
}
