package http

import (
	sharedauth "Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/auth/internal/application/service"
	"Online-queue-management-system/services/auth/internal/domain"
	jwtmanager "Online-queue-management-system/services/auth/internal/infrastructure/jwt"
	"Online-queue-management-system/services/auth/internal/mocks"
	"bytes"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestHandleLogin_WhenCredentialsAreValid_ShouldSetTokenCookies(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"login":"owner@example.com",
		"password":"secret-password"
	}`))
	rec := httptest.NewRecorder()

	handler.handleLogin(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusOK, "ok")
	assertCookie(t, rec, "access_token", "access-owner@example.com", "/")
	assertCookie(t, rec, "refresh_token", "refresh-owner@example.com", "/auth")
}

func TestHandleLogin_WhenCredentialsAreInvalid_ShouldReturnUnauthorized(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"login":"owner@example.com",
		"password":"wrong-password"
	}`))
	rec := httptest.NewRecorder()

	handler.handleLogin(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusUnauthorized, "bad credentials")
}

func TestHandleLogin_WhenRequestIsInvalid_ShouldReturnBadRequest(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/login", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	handler.handleLogin(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusBadRequest, "invalid request")
}

func TestHandleRefresh_WhenSessionExists_ShouldRotateCookies(t *testing.T) {
	handler, deps := newTestHandler()
	deps.tokens.RefreshClaims["old-refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "old-jti"}
	deps.sessions.Exists["old-jti:42"] = true

	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/refresh", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "refresh_token", Value: "old-refresh-token"})
	rec := httptest.NewRecorder()

	handler.handleRefresh(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusOK, "ok")
	assertCookie(t, rec, "access_token", "access-owner@example.com", "/")
	assertCookie(t, rec, "refresh_token", "refresh-owner@example.com", "/auth")
	if !deps.sessions.Deleted["old-jti"] {
		t.Fatal("expected old refresh session to be deleted")
	}
}

func TestHandleRefresh_WhenCookieIsMissing_ShouldReturnUnauthorized(t *testing.T) {
	handler, _ := newTestHandler()
	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/refresh", nil)
	rec := httptest.NewRecorder()

	handler.handleRefresh(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusUnauthorized, "unauthorized")
}

func TestHandleLogout_WhenRefreshCookieExists_ShouldDeleteSessionAndClearCookies(t *testing.T) {
	handler, deps := newTestHandler()
	deps.tokens.RefreshClaims["refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "refresh-jti"}
	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/logout", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "refresh_token", Value: "refresh-token"})
	rec := httptest.NewRecorder()

	handler.handleLogout(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusOK, "ok")
	assertClearedCookie(t, rec, "access_token")
	assertClearedCookie(t, rec, "refresh_token")
	if !deps.sessions.Deleted["refresh-jti"] {
		t.Fatal("expected refresh session to be deleted")
	}
}

func TestHandleMe_WhenAccessTokenIsValid_ShouldReturnClaims(t *testing.T) {
	handler, _ := newTestHandler()
	branchID := int64(3)
	manager := jwtmanager.New("access-secret", "refresh-secret", time.Hour, time.Hour)
	accessToken, err := manager.NewAccessToken(&domain.User{
		ID:         42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   "business_admin",
		BusinessID: 7,
		BranchID:   &branchID,
	})
	if err != nil {
		t.Fatalf("new access token: %v", err)
	}
	req := httptest.NewRequest(stdhttp.MethodGet, "/auth/me", nil)
	req.AddCookie(&stdhttp.Cookie{Name: "access_token", Value: accessToken})
	rec := httptest.NewRecorder()

	handler.handleMe(rec, req)

	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", stdhttp.StatusOK, rec.Code, rec.Body.String())
	}

	var body MeResponse
	decodeJSON(t, rec, &body)
	if body.UserID != 42 || body.Login != "owner@example.com" || body.RoleID != 2 || body.BusinessID != 7 {
		t.Fatalf("unexpected me response: %#v", body)
	}
	if body.BranchID == nil || *body.BranchID != branchID {
		t.Fatalf("unexpected branch id: %#v", body.BranchID)
	}
}

func TestRegister_WhenMuxReceivesLoginRequest_ShouldRouteToLoginHandler(t *testing.T) {
	handler, _ := newTestHandler()
	mux := stdhttp.NewServeMux()
	handler.Register(mux)

	req := httptest.NewRequest(stdhttp.MethodPost, "/auth/login", bytes.NewBufferString(`{
		"login":"owner@example.com",
		"password":"secret-password"
	}`))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assertMessageResponse(t, rec, stdhttp.StatusOK, "ok")
}

func newTestHandler() (*Handler, *handlerDeps) {
	users := mocks.NewUserRepository()
	sessions := mocks.NewSessionRepository()
	tokens := mocks.NewTokenManager()

	users.ByLogin["owner@example.com"] = domain.User{
		ID:           42,
		Login:        "owner@example.com",
		PasswordHash: hashPasswordForHandlerTest(),
		RoleID:       2,
		BusinessID:   7,
	}
	users.ByID[42] = users.ByLogin["owner@example.com"]

	authService := service.New(users, sessions, tokens)
	handler := NewHandler(
		authService,
		sharedauth.NewTokenParser("access-secret"),
		NewCookieManager(false),
		time.Minute,
		24*time.Hour,
	)

	return handler, &handlerDeps{
		users:    users,
		sessions: sessions,
		tokens:   tokens,
	}
}

type handlerDeps struct {
	users    *mocks.UserRepository
	sessions *mocks.SessionRepository
	tokens   *mocks.TokenManager
}

func assertMessageResponse(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	t.Helper()

	if rec.Code != expectedStatus {
		t.Fatalf("expected status %d, got %d with body %s", expectedStatus, rec.Code, rec.Body.String())
	}

	var body MessageResponse
	decodeJSON(t, rec, &body)
	if body.Message != expectedMessage {
		t.Fatalf("expected message %q, got %q", expectedMessage, body.Message)
	}
}

func assertCookie(t *testing.T, rec *httptest.ResponseRecorder, name, value, path string) {
	t.Helper()

	cookie := findCookie(t, rec, name)
	if cookie.Value != value {
		t.Fatalf("expected %s cookie value %q, got %q", name, value, cookie.Value)
	}
	if cookie.Path != path {
		t.Fatalf("expected %s cookie path %q, got %q", name, path, cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatalf("expected %s cookie to be http only", name)
	}
}

func assertClearedCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) {
	t.Helper()

	cookie := findCookie(t, rec, name)
	if cookie.Value != "" || cookie.MaxAge != -1 {
		t.Fatalf("expected %s cookie to be cleared, got value=%q maxAge=%d", name, cookie.Value, cookie.MaxAge)
	}
}

func findCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *stdhttp.Cookie {
	t.Helper()

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("expected %s cookie", name)
	return nil
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
}

func hashPasswordForHandlerTest() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret-password"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}
