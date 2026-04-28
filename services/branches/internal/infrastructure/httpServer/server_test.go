package httpserver

import (
	sharedauth "Online-queue-management-system/libs/auth"
	"Online-queue-management-system/services/branches/internal/application/service"
	branchesdomain "Online-queue-management-system/services/branches/internal/domain"
	"context"
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
	repo.byBusinessID[7] = []branchesdomain.Branch{
		{ID: 1, BusinessID: 7, Name: "Central", Address: "Main street"},
		{ID: 2, BusinessID: 7, Name: "Left Bank", Address: "Second street"},
	}

	rec := serveAuthenticatedBranchesRequest(t, server, accessTokenClaims{
		ID:         42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   string(branchesdomain.RoleBusinessAdmin),
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
	if repo.lastBusinessID != 7 {
		t.Fatalf("expected business id 7, got %d", repo.lastBusinessID)
	}
}

func TestGetBranches_WhenManagerIsAuthorized_ShouldReturnManagerBranch(t *testing.T) {
	server, repo := newTestHTTPServer()
	branchID := int64(11)
	repo.byID[branchID] = []branchesdomain.Branch{
		{ID: branchID, BusinessID: 7, Name: "Central", Address: "Main street"},
	}

	rec := serveAuthenticatedBranchesRequest(t, server, accessTokenClaims{
		ID:         43,
		Login:      "manager@example.com",
		RoleID:     3,
		RoleName:   string(branchesdomain.RoleManager),
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
	if repo.lastBranchID != branchID {
		t.Fatalf("expected branch id %d, got %d", branchID, repo.lastBranchID)
	}
}

func TestGetBranches_WhenRoleIsForbidden_ShouldReturnInternalServerErrorWithForbiddenBody(t *testing.T) {
	server, _ := newTestHTTPServer()

	rec := serveAuthenticatedBranchesRequest(t, server, accessTokenClaims{
		ID:         44,
		Login:      "employee@example.com",
		RoleID:     4,
		RoleName:   string(branchesdomain.RoleEmployee),
		BusinessID: 7,
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), branchesdomain.ErrForbidden.Error()) {
		t.Fatalf("expected forbidden body, got %q", rec.Body.String())
	}
}

func serveAuthenticatedBranchesRequest(t *testing.T, server *HttpServer, claims accessTokenClaims) *httptest.ResponseRecorder {
	t.Helper()

	accessToken := signAccessToken(t, claims)

	req := httptest.NewRequest(http.MethodGet, "/branches", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	rec := httptest.NewRecorder()

	middleware := sharedauth.Middleware(sharedauth.NewTokenParser("branches-access-secret"))
	middleware(http.HandlerFunc(server.GetBranches)).ServeHTTP(rec, req)
	return rec
}

type accessTokenClaims struct {
	ID         int64
	Login      string
	RoleID     int64
	RoleName   string
	BusinessID int64
	BranchID   *int64
}

type accessClaimsDTO struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	RoleName   string `json:"role_name,omitempty"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
	jwt.RegisteredClaims
}

func signAccessToken(t *testing.T, claims accessTokenClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaimsDTO{
		UserID:     claims.ID,
		Login:      claims.Login,
		RoleID:     claims.RoleID,
		RoleName:   claims.RoleName,
		BusinessID: claims.BusinessID,
		BranchID:   claims.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(claims.ID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	signed, err := token.SignedString([]byte("branches-access-secret"))
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return signed
}

func newTestHTTPServer() (*HttpServer, *fakeBranchesRepository) {
	repo := newFakeBranchesRepository()
	svc := service.New(repo)
	return NewHttpServer(svc), repo
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(rec.Body).Decode(target); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
}

type fakeBranchesRepository struct {
	byBusinessID    map[int64][]branchesdomain.Branch
	byID            map[int64][]branchesdomain.Branch
	lastBusinessID  int64
	lastBranchID    int64
	businessIDCalls int
	idCalls         int
	err             error
}

func newFakeBranchesRepository() *fakeBranchesRepository {
	return &fakeBranchesRepository{
		byBusinessID: make(map[int64][]branchesdomain.Branch),
		byID:         make(map[int64][]branchesdomain.Branch),
	}
}

func (r *fakeBranchesRepository) GetByBusinessID(_ context.Context, businessID int64) ([]branchesdomain.Branch, error) {
	r.businessIDCalls++
	r.lastBusinessID = businessID
	if r.err != nil {
		return nil, r.err
	}
	return r.byBusinessID[businessID], nil
}

func (r *fakeBranchesRepository) GetByID(_ context.Context, branchID int64) ([]branchesdomain.Branch, error) {
	r.idCalls++
	r.lastBranchID = branchID
	if r.err != nil {
		return nil, r.err
	}
	branches, ok := r.byID[branchID]
	if !ok {
		return nil, branchesdomain.ErrBranchNotFound
	}
	return branches, nil
}

func TestGetBranches_WhenRepositoryFails_ShouldReturnInternalServerError(t *testing.T) {
	server, repo := newTestHTTPServer()
	repo.err = errors.New("db failed")

	rec := serveAuthenticatedBranchesRequest(t, server, accessTokenClaims{
		ID:         42,
		Login:      "owner@example.com",
		RoleID:     2,
		RoleName:   string(branchesdomain.RoleBusinessAdmin),
		BusinessID: 7,
	})

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusInternalServerError, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "db failed") {
		t.Fatalf("expected repository error body, got %q", rec.Body.String())
	}
}
