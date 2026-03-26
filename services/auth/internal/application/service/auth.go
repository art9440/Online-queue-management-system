package service

import (
	"context"

	"Online-queue-management-system/services/auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type SessionRepository interface {
	SaveRefreshSession(ctx context.Context, jti string, userID int64) error
	RefreshSessionExists(ctx context.Context, jti string, userID int64) (bool, error)
	DeleteRefreshSession(ctx context.Context, jti string) error
}

type TokenManager interface {
	NewAccessToken(user *domain.User) (string, error)
	NewRefreshToken(user *domain.User) (token string, jti string, err error)
	ParseAccessToken(token string) (*domain.AccessClaims, error)
	ParseRefreshToken(token string) (*domain.RefreshClaims, error)
}

type AuthService struct {
	users    UserRepository
	sessions SessionRepository
	tokens   TokenManager
}

func New(
	users UserRepository,
	sessions SessionRepository,
	tokens TokenManager,
) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		tokens:   tokens,
	}
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*domain.Tokens, error) {
	user, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		return nil, domain.ErrBadCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrBadCredentials
	}

	accessToken, err := s.tokens.NewAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, jti, err := s.tokens.NewRefreshToken(user)
	if err != nil {
		return nil, err
	}

	if err := s.sessions.SaveRefreshSession(ctx, jti, user.ID); err != nil {
		return nil, err
	}

	return &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*domain.Tokens, error) {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	ok, err := s.sessions.RefreshSessionExists(ctx, claims.JTI, claims.UserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if err := s.sessions.DeleteRefreshSession(ctx, claims.JTI); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	accessToken, err := s.tokens.NewAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, newJTI, err := s.tokens.NewRefreshToken(user)
	if err != nil {
		return nil, err
	}

	if err := s.sessions.SaveRefreshSession(ctx, newJTI, user.ID); err != nil {
		return nil, err
	}

	return &domain.Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokens.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil
	}

	return s.sessions.DeleteRefreshSession(ctx, claims.JTI)
}

func (s *AuthService) Me(ctx context.Context, accessToken string) (*domain.AccessClaims, error) {
	claims, err := s.tokens.ParseAccessToken(accessToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return claims, nil
}