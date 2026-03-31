package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"Online-queue-management-system/services/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type accessClaimsDTO struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
	jwt.RegisteredClaims
}

type refreshClaimsDTO struct {
	jwt.RegisteredClaims
}

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func New(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

func (m *TokenManager) NewAccessToken(user *domain.User) (string, error) {
	now := time.Now()

	claims := accessClaimsDTO{
		UserID:     user.ID,
		Login:      user.Login,
		RoleID:     user.RoleID,
		BusinessID: user.BusinessID,
		BranchID:   user.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.accessSecret)
}

func (m *TokenManager) NewRefreshToken(user *domain.User) (string, string, error) {
	now := time.Now()
	jti := uuid.NewString()

	claims := refreshClaimsDTO{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.refreshSecret)
	if err != nil {
		return "", "", err
	}

	return signed, jti, nil
}

func (m *TokenManager) ParseAccessToken(tokenString string) (*domain.AccessClaims, error) {
	claims := &accessClaimsDTO{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return &domain.AccessClaims{
		UserID:     claims.UserID,
		Login:      claims.Login,
		RoleID:     claims.RoleID,
		BusinessID: claims.BusinessID,
		BranchID:   claims.BranchID,
	}, nil
}

func (m *TokenManager) ParseRefreshToken(tokenString string) (*domain.RefreshClaims, error) {
	claims := &refreshClaimsDTO{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, err
	}

	return &domain.RefreshClaims{
		UserID: userID,
		JTI:    claims.ID,
	}, nil
}