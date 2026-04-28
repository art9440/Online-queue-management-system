package mocks

import (
	"Online-queue-management-system/services/auth/internal/domain"
	"context"
	"errors"
	"strconv"
)

type UserRepository struct {
	ByLogin map[string]domain.User
	ByID    map[int64]domain.User
	Err     error
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		ByLogin: make(map[string]domain.User),
		ByID:    make(map[int64]domain.User),
	}
}

func (r *UserRepository) GetByLogin(_ context.Context, login string) (*domain.User, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	user, ok := r.ByLogin[login]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *UserRepository) GetByID(_ context.Context, id int64) (*domain.User, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	user, ok := r.ByID[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

type SessionRepository struct {
	Exists  map[string]bool
	Deleted map[string]bool
	Err     error
}

func NewSessionRepository() *SessionRepository {
	return &SessionRepository{
		Exists:  make(map[string]bool),
		Deleted: make(map[string]bool),
	}
}

func (r *SessionRepository) SaveRefreshSession(_ context.Context, jti string, userID int64) error {
	if r.Err != nil {
		return r.Err
	}
	r.Exists[SessionKey(jti, userID)] = true
	return nil
}

func (r *SessionRepository) RefreshSessionExists(_ context.Context, jti string, userID int64) (bool, error) {
	if r.Err != nil {
		return false, r.Err
	}
	return r.Exists[SessionKey(jti, userID)], nil
}

func (r *SessionRepository) DeleteRefreshSession(_ context.Context, jti string) error {
	if r.Err != nil {
		return r.Err
	}
	r.Deleted[jti] = true
	return nil
}

func SessionKey(jti string, userID int64) string {
	return jti + ":" + strconv.FormatInt(userID, 10)
}

type TokenManager struct {
	RefreshClaims map[string]domain.RefreshClaims
	Err           error
}

func NewTokenManager() *TokenManager {
	return &TokenManager{RefreshClaims: make(map[string]domain.RefreshClaims)}
}

func (m *TokenManager) NewAccessToken(user *domain.User) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return "access-" + user.Login, nil
}

func (m *TokenManager) NewRefreshToken(user *domain.User) (string, string, error) {
	if m.Err != nil {
		return "", "", m.Err
	}
	return "refresh-" + user.Login, "jti-" + user.Login, nil
}

func (m *TokenManager) ParseRefreshToken(token string) (*domain.RefreshClaims, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	claims, ok := m.RefreshClaims[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}
