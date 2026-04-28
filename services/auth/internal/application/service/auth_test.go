package service

import (
	"Online-queue-management-system/services/auth/internal/domain"
	"context"
	"errors"
	"strconv"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLogin_WhenCredentialsAreValid_ShouldReturnTokensAndSaveRefreshSession(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	tokens := newFakeTokenManager()
	auth := New(users, sessions, tokens)

	users.byLogin["owner@example.com"] = domain.User{
		ID:           42,
		Login:        "owner@example.com",
		PasswordHash: hashPassword(t, "secret-password"),
		RoleID:       2,
		BusinessID:   7,
	}

	got, err := auth.Login(ctx, "owner@example.com", "secret-password")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if got.AccessToken != "access-owner@example.com" {
		t.Fatalf("unexpected access token %q", got.AccessToken)
	}
	if got.RefreshToken != "refresh-owner@example.com" {
		t.Fatalf("unexpected refresh token %q", got.RefreshToken)
	}
	if !sessions.exists["jti-owner@example.com:42"] {
		t.Fatal("expected refresh session to be saved")
	}
}

func TestLogin_WhenPasswordDoesNotMatch_ShouldReturnBadCredentialsWithoutSavingSession(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	tokens := newFakeTokenManager()
	auth := New(users, sessions, tokens)

	users.byLogin["owner@example.com"] = domain.User{
		ID:           42,
		Login:        "owner@example.com",
		PasswordHash: hashPassword(t, "secret-password"),
	}

	_, err := auth.Login(ctx, "owner@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrBadCredentials) {
		t.Fatalf("expected bad credentials, got %v", err)
	}
	if len(sessions.exists) != 0 {
		t.Fatalf("expected no saved sessions, got %d", len(sessions.exists))
	}
}

func TestLogin_WhenUserRepositoryFails_ShouldReturnBadCredentials(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepository()
	users.err = errors.New("db is down")
	auth := New(users, newFakeSessionRepository(), newFakeTokenManager())

	_, err := auth.Login(ctx, "owner@example.com", "secret-password")
	if !errors.Is(err, domain.ErrBadCredentials) {
		t.Fatalf("expected bad credentials, got %v", err)
	}
}

func TestRefresh_WhenSessionExists_ShouldRotateRefreshSessionAndReturnNewTokens(t *testing.T) {
	ctx := context.Background()
	users := newFakeUserRepository()
	sessions := newFakeSessionRepository()
	tokens := newFakeTokenManager()
	auth := New(users, sessions, tokens)

	users.byID[42] = domain.User{
		ID:         42,
		Login:      "owner@example.com",
		RoleID:     2,
		BusinessID: 7,
	}
	tokens.refreshClaims["old-refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "old-jti"}
	sessions.exists["old-jti:42"] = true

	got, err := auth.Refresh(ctx, "old-refresh-token")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if got.AccessToken != "access-owner@example.com" {
		t.Fatalf("unexpected access token %q", got.AccessToken)
	}
	if got.RefreshToken != "refresh-owner@example.com" {
		t.Fatalf("unexpected refresh token %q", got.RefreshToken)
	}
	if !sessions.deleted["old-jti"] {
		t.Fatal("expected old refresh session to be deleted")
	}
	if !sessions.exists["jti-owner@example.com:42"] {
		t.Fatal("expected new refresh session to be saved")
	}
}

func TestRefresh_WhenRefreshTokenIsInvalid_ShouldReturnUnauthorized(t *testing.T) {
	ctx := context.Background()
	auth := New(newFakeUserRepository(), newFakeSessionRepository(), newFakeTokenManager())

	_, err := auth.Refresh(ctx, "invalid-refresh-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestRefresh_WhenSessionDoesNotExist_ShouldReturnUnauthorized(t *testing.T) {
	ctx := context.Background()
	tokens := newFakeTokenManager()
	tokens.refreshClaims["old-refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "old-jti"}
	auth := New(newFakeUserRepository(), newFakeSessionRepository(), tokens)

	_, err := auth.Refresh(ctx, "old-refresh-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestLogout_WhenRefreshTokenIsValid_ShouldDeleteRefreshSession(t *testing.T) {
	ctx := context.Background()
	sessions := newFakeSessionRepository()
	tokens := newFakeTokenManager()
	tokens.refreshClaims["refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "refresh-jti"}
	auth := New(newFakeUserRepository(), sessions, tokens)

	if err := auth.Logout(ctx, "refresh-token"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !sessions.deleted["refresh-jti"] {
		t.Fatal("expected refresh session to be deleted")
	}
}

func TestLogout_WhenRefreshTokenIsInvalid_ShouldIgnoreError(t *testing.T) {
	ctx := context.Background()
	sessions := newFakeSessionRepository()
	auth := New(newFakeUserRepository(), sessions, newFakeTokenManager())

	if err := auth.Logout(ctx, "invalid-refresh-token"); err != nil {
		t.Fatalf("logout should ignore invalid refresh token, got %v", err)
	}
	if len(sessions.deleted) != 0 {
		t.Fatalf("expected no deleted sessions, got %d", len(sessions.deleted))
	}
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(hash)
}

type fakeUserRepository struct {
	byLogin map[string]domain.User
	byID    map[int64]domain.User
	err     error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		byLogin: make(map[string]domain.User),
		byID:    make(map[int64]domain.User),
	}
}

func (r *fakeUserRepository) GetByLogin(_ context.Context, login string) (*domain.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	user, ok := r.byLogin[login]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *fakeUserRepository) GetByID(_ context.Context, id int64) (*domain.User, error) {
	if r.err != nil {
		return nil, r.err
	}
	user, ok := r.byID[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

type fakeSessionRepository struct {
	exists  map[string]bool
	deleted map[string]bool
	err     error
}

func newFakeSessionRepository() *fakeSessionRepository {
	return &fakeSessionRepository{
		exists:  make(map[string]bool),
		deleted: make(map[string]bool),
	}
}

func (r *fakeSessionRepository) SaveRefreshSession(_ context.Context, jti string, userID int64) error {
	if r.err != nil {
		return r.err
	}
	r.exists[sessionKey(jti, userID)] = true
	return nil
}

func (r *fakeSessionRepository) RefreshSessionExists(_ context.Context, jti string, userID int64) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	return r.exists[sessionKey(jti, userID)], nil
}

func (r *fakeSessionRepository) DeleteRefreshSession(_ context.Context, jti string) error {
	if r.err != nil {
		return r.err
	}
	r.deleted[jti] = true
	return nil
}

func sessionKey(jti string, userID int64) string {
	return jti + ":" + strconv.FormatInt(userID, 10)
}

type fakeTokenManager struct {
	refreshClaims map[string]domain.RefreshClaims
	err           error
}

func newFakeTokenManager() *fakeTokenManager {
	return &fakeTokenManager{refreshClaims: make(map[string]domain.RefreshClaims)}
}

func (m *fakeTokenManager) NewAccessToken(user *domain.User) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return "access-" + user.Login, nil
}

func (m *fakeTokenManager) NewRefreshToken(user *domain.User) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return "refresh-" + user.Login, "jti-" + user.Login, nil
}

func (m *fakeTokenManager) ParseRefreshToken(token string) (*domain.RefreshClaims, error) {
	if m.err != nil {
		return nil, m.err
	}
	claims, ok := m.refreshClaims[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}
