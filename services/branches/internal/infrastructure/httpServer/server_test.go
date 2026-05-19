package httpserver

import (
	sharedauth "Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/application/service"
	branchesdomain "Online-queue-management-system/services/branches/internal/domain"
	branchesdto "Online-queue-management-system/services/branches/internal/infrastructure/dto"
	"Online-queue-management-system/services/branches/internal/mocks"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetBranches_WhenUserIsMissingFromContext_ShouldReturnUnauthorized(t *testing.T) {
	server, _ := newTestHTTPServer()
	req := httptest.NewRequest(http.MethodGet, "/branches", http.NoBody)
	rec := httptest.NewRecorder()

	server.GetBranches(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "unauthorized") {
		t.Fatalf("expected unauthorized body, got %q", rec.Body.String())
	}
}

func TestGetBranches_WhenBusinessAdminIsAuthorized_ShouldReturnBusinessBranches(t *testing.T) {
	server, repo := newTestHTTPServer()
	repo.ByBusinessID[7] = []branchesdomain.Branch{
		{ID: 1, BusinessID: 7, Name: "Central", Address: "Main street"},
		{ID: 2, BusinessID: 7, Name: "Left Bank", Address: "Second street"},
	}

	rec := serveAuthenticatedBranchesRequest(t, server, sharedauth.AccessClaims{
		UserID:     42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   string(sharedauth.RoleBusinessAdmin),
		BusinessID: 7,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body []branchesdomain.Branch
	decodeJSON(t, rec, &body)
	if len(body) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(body))
	}
	if repo.LastBusinessID != 7 {
		t.Fatalf("expected business id 7, got %d", repo.LastBusinessID)
	}
}

func TestGetBranches_WhenManagerIsAuthorized_ShouldReturnManagerBranch(t *testing.T) {
	server, repo := newTestHTTPServer()
	branchID := int64(11)
	repo.ByID[branchID] = []branchesdomain.Branch{
		{ID: branchID, BusinessID: 7, Name: "Central", Address: "Main street"},
	}

	rec := serveAuthenticatedBranchesRequest(t, server, sharedauth.AccessClaims{
		UserID:     43,
		Login:      "manager@example.com",
		RoleID:     3,
		RoleName:   string(sharedauth.RoleManager),
		BusinessID: 7,
		BranchID:   &branchID,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body []branchesdomain.Branch
	decodeJSON(t, rec, &body)
	if len(body) != 1 || body[0].ID != branchID {
		t.Fatalf("unexpected branches response: %#v", body)
	}
	if repo.LastBranchID != branchID {
		t.Fatalf("expected branch id %d, got %d", branchID, repo.LastBranchID)
	}
}

func TestGetBranches_WhenRoleIsForbidden_ShouldReturnInternalServerErrorWithForbiddenBody(t *testing.T) {
	server, _ := newTestHTTPServer()

	rec := serveAuthenticatedBranchesRequest(t, server, sharedauth.AccessClaims{
		UserID:     44,
		Login:      "employee@example.com",
		RoleID:     4,
		RoleName:   string(sharedauth.RoleEmployee),
		BusinessID: 7,
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), branchesdomain.ErrForbidden.Error()) {
		t.Fatalf("expected forbidden body, got %q", rec.Body.String())
	}
}

func serveAuthenticatedBranchesRequest(t *testing.T, server *HttpServer, claims sharedauth.AccessClaims) *httptest.ResponseRecorder {
	t.Helper()

	return serveAuthenticatedRequest(t, http.HandlerFunc(server.GetBranches), "/branches", claims)
}

func serveAuthenticatedRequest(
	t *testing.T,
	handler http.Handler,
	path string,
	claims sharedauth.AccessClaims,
) *httptest.ResponseRecorder {
	t.Helper()

	accessToken := signAccessToken(t, claims)

	req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	rec := httptest.NewRecorder()

	middleware := sharedauth.Middleware(sharedauth.NewTokenParser("branches-access-secret"))
	middleware(handler).ServeHTTP(rec, req)
	return rec
}

func signAccessToken(t *testing.T, claims sharedauth.AccessClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     claims.UserID,
		"login":       claims.Login,
		"role_id":     claims.RoleID,
		"role_name":   claims.RoleName,
		"business_id": claims.BusinessID,
		"branch_id":   claims.BranchID,
		"sub":         strconv.FormatInt(claims.UserID, 10),
		"iat":         jwt.NewNumericDate(time.Now()).Unix(),
		"exp":         jwt.NewNumericDate(time.Now().Add(time.Hour)).Unix(),
	})
	signed, err := token.SignedString([]byte("branches-access-secret"))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return signed
}

func newTestHTTPServer() (*HttpServer, *mocks.BranchesRepository) {
	repo := mocks.NewBranchesRepository()
	svc := service.New(repo)
	return NewHttpServer(svc), repo
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
}

func TestGetBranches_WhenRepositoryFails_ShouldReturnInternalServerError(t *testing.T) {
	server, repo := newTestHTTPServer()
	repo.Err = errors.New("db failed")

	rec := serveAuthenticatedBranchesRequest(t, server, sharedauth.AccessClaims{
		UserID:     42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   string(sharedauth.RoleBusinessAdmin),
		BusinessID: 7,
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "db failed") {
		t.Fatalf("expected repository error body, got %q", rec.Body.String())
	}
}

func TestGetBranchClients_WhenBusinessAdminIsAuthorized_ShouldReturnClients(t *testing.T) {
	server, repo := newTestHTTPServer()
	repo.BranchBusiness[11] = 7
	repo.ClientsByBranchID[11] = []branchesdomain.Client{
		{ID: 5, Name: "Alex", Surname: "Stone"},
	}

	rec := serveAuthenticatedRequest(t, http.HandlerFunc(server.GetBranchClients), "/branches/11/clients", sharedauth.AccessClaims{
		UserID:     42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   string(branchesdomain.RoleBusinessAdmin),
		BusinessID: 7,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body []branchesdto.ClientResponse
	decodeJSON(t, rec, &body)
	if len(body) != 1 || body[0].ID != 5 {
		t.Fatalf("unexpected clients response: %#v", body)
	}
}

func TestGetBranchAppointments_WhenManagerIsAuthorized_ShouldReturnAppointments(t *testing.T) {
	server, repo := newTestHTTPServer()
	branchID := int64(11)
	date := time.Date(2026, time.May, 18, 0, 0, 0, 0, time.UTC)
	repo.SetAppointments(branchID, date, []branchesdomain.Appointment{
		{ID: 21, BranchID: branchID, Client: branchesdomain.Client{ID: 5, Name: "Alex", Surname: "Stone"}},
	})

	rec := serveAuthenticatedRequest(t, http.HandlerFunc(server.GetBranchAppointments), "/branches/11/bookings?date=2026-05-18", sharedauth.AccessClaims{
		UserID:     43,
		Login:      "manager@example.com",
		RoleID:     3,
		RoleName:   string(branchesdomain.RoleManager),
		BusinessID: 7,
		BranchID:   &branchID,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var body []branchesdto.AppointmentResponse
	decodeJSON(t, rec, &body)
	if len(body) != 1 || body[0].ID != 21 {
		t.Fatalf("unexpected appointments response: %#v", body)
	}
	if repo.LastAppointmentDate.Format(time.DateOnly) != "2026-05-18" {
		t.Fatalf("expected appointment date 2026-05-18, got %s", repo.LastAppointmentDate.Format(time.DateOnly))
	}
}

func TestGetBranchAppointments_WhenDateIsInvalid_ShouldReturnBadRequest(t *testing.T) {
	server, _ := newTestHTTPServer()
	branchID := int64(11)

	rec := serveAuthenticatedRequest(t, http.HandlerFunc(server.GetBranchAppointments), "/branches/11/bookings?date=18-05-2026", sharedauth.AccessClaims{
		UserID:     43,
		Login:      "manager@example.com",
		RoleID:     3,
		RoleName:   string(branchesdomain.RoleManager),
		BusinessID: 7,
		BranchID:   &branchID,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestGetBranchClients_WhenManagerRequestsAnotherBranch_ShouldReturnForbidden(t *testing.T) {
	server, _ := newTestHTTPServer()
	branchID := int64(11)

	rec := serveAuthenticatedRequest(t, http.HandlerFunc(server.GetBranchClients), "/branches/12/clients", sharedauth.AccessClaims{
		UserID:     43,
		Login:      "manager@example.com",
		RoleID:     3,
		RoleName:   string(branchesdomain.RoleManager),
		BusinessID: 7,
		BranchID:   &branchID,
	})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}
