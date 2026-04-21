package auth

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type accessClaimsDTO struct {
	UserID     int64  `json:"user_id"`
	Login      string `json:"login"`
	RoleID     int64  `json:"role_id"`
	RoleName   string `json:"role_name,omitempty"`
	BusinessID int64  `json:"business_id"`
	BranchID   *int64 `json:"branch_id,omitempty"`
	jwt.RegisteredClaims
}

type TokenParser struct {
	secret []byte
}

func NewTokenParser(secret string) *TokenParser {
	return &TokenParser{
		secret: []byte(secret),
	}
}

func (p *TokenParser) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &accessClaimsDTO{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return p.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &AccessClaims{
		UserID:     claims.UserID,
		Login:      claims.Login,
		RoleID:     claims.RoleID,
		RoleName:   claims.RoleName,
		BusinessID: claims.BusinessID,
		BranchID:   claims.BranchID,
	}, nil
}
