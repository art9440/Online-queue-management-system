package service

import (
	"Online-queue-management-system/services/auth/internal/domain"
	"Online-queue-management-system/services/auth/internal/mocks"
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestLogin_WhenCredentialsAreValid_ShouldReturnTokensAndSaveRefreshSession(t *testing.T) {
	ctx := context.Background()
	users := mocks.NewUserRepository()
	sessions := mocks.NewSessionRepository()
	tokens := mocks.NewTokenManager()
	auth := New(users, sessions, tokens)

	users.ByLogin["owner@example.com"] = domain.User{
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
	if !sessions.Exists["jti-owner@example.com:42"] {
		t.Fatal("expected refresh session to be saved")
	}
}

func TestLogin_WhenPasswordDoesNotMatch_ShouldReturnBadCredentialsWithoutSavingSession(t *testing.T) {
	ctx := context.Background()
	users := mocks.NewUserRepository()
	sessions := mocks.NewSessionRepository()
	tokens := mocks.NewTokenManager()
	auth := New(users, sessions, tokens)

	users.ByLogin["owner@example.com"] = domain.User{
		ID:           42,
		Login:        "owner@example.com",
		PasswordHash: hashPassword(t, "secret-password"),
	}

	_, err := auth.Login(ctx, "owner@example.com", "wrong-password")
	if !errors.Is(err, domain.ErrBadCredentials) {
		t.Fatalf("expected bad credentials, got %v", err)
	}
	if len(sessions.Exists) != 0 {
		t.Fatalf("expected no saved sessions, got %d", len(sessions.Exists))
	}
}

func TestLogin_WhenUserRepositoryFails_ShouldReturnBadCredentials(t *testing.T) {
	ctx := context.Background()
	users := mocks.NewUserRepository()
	users.Err = errors.New("db is down")
	auth := New(users, mocks.NewSessionRepository(), mocks.NewTokenManager())

	_, err := auth.Login(ctx, "owner@example.com", "secret-password")
	if !errors.Is(err, domain.ErrBadCredentials) {
		t.Fatalf("expected bad credentials, got %v", err)
	}
}

func TestRefresh_WhenSessionExists_ShouldRotateRefreshSessionAndReturnNewTokens(t *testing.T) {
	ctx := context.Background()
	users := mocks.NewUserRepository()
	sessions := mocks.NewSessionRepository()
	tokens := mocks.NewTokenManager()
	auth := New(users, sessions, tokens)

	users.ByID[42] = domain.User{
		ID:         42,
		Login:      "owner@example.com",
		RoleID:     2,
		BusinessID: 7,
	}
	tokens.RefreshClaims["old-refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "old-jti"}
	sessions.Exists["old-jti:42"] = true

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
	if !sessions.Deleted["old-jti"] {
		t.Fatal("expected old refresh session to be deleted")
	}
	if !sessions.Exists["jti-owner@example.com:42"] {
		t.Fatal("expected new refresh session to be saved")
	}
}

func TestRefresh_WhenRefreshTokenIsInvalid_ShouldReturnUnauthorized(t *testing.T) {
	ctx := context.Background()
	auth := New(mocks.NewUserRepository(), mocks.NewSessionRepository(), mocks.NewTokenManager())

	_, err := auth.Refresh(ctx, "invalid-refresh-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestRefresh_WhenSessionDoesNotExist_ShouldReturnUnauthorized(t *testing.T) {
	ctx := context.Background()
	tokens := mocks.NewTokenManager()
	tokens.RefreshClaims["old-refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "old-jti"}
	auth := New(mocks.NewUserRepository(), mocks.NewSessionRepository(), tokens)

	_, err := auth.Refresh(ctx, "old-refresh-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestLogout_WhenRefreshTokenIsValid_ShouldDeleteRefreshSession(t *testing.T) {
	ctx := context.Background()
	sessions := mocks.NewSessionRepository()
	tokens := mocks.NewTokenManager()
	tokens.RefreshClaims["refresh-token"] = domain.RefreshClaims{UserID: 42, JTI: "refresh-jti"}
	auth := New(mocks.NewUserRepository(), sessions, tokens)

	if err := auth.Logout(ctx, "refresh-token"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if !sessions.Deleted["refresh-jti"] {
		t.Fatal("expected refresh session to be deleted")
	}
}

func TestLogout_WhenRefreshTokenIsInvalid_ShouldIgnoreError(t *testing.T) {
	ctx := context.Background()
	sessions := mocks.NewSessionRepository()
	auth := New(mocks.NewUserRepository(), sessions, mocks.NewTokenManager())

	if err := auth.Logout(ctx, "invalid-refresh-token"); err != nil {
		t.Fatalf("logout should ignore invalid refresh token, got %v", err)
	}
	if len(sessions.Deleted) != 0 {
		t.Fatalf("expected no deleted sessions, got %d", len(sessions.Deleted))
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
