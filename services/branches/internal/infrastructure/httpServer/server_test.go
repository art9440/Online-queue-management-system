package httpserver

import (
	"Online-queue-management-system/libs/auth"
	sharedauth "Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/application/service"
	branchesdomain "Online-queue-management-system/services/branches/internal/domain"
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
	req := httptest.NewRequest(http.MethodGet, "/branches", nil)
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
		RoleName:   string(auth.RoleBusinessAdmin),
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
		RoleName:   string(auth.RoleManager),
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
		RoleName:   string(auth.RoleEmployee),
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

	accessToken := signAccessToken(t, claims)

	req := httptest.NewRequest(http.MethodGet, "/branches", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	rec := httptest.NewRecorder()

	middleware := sharedauth.Middleware(sharedauth.NewTokenParser("branches-access-secret"))
	middleware(http.HandlerFunc(server.GetBranches)).ServeHTTP(rec, req)
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
		RoleName:   string(auth.RoleBusinessAdmin),
		BusinessID: 7,
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "db failed") {
		t.Fatalf("expected repository error body, got %q", rec.Body.String())
	}
}
