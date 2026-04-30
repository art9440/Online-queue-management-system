package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrNotFound          = errors.New("pending registration not found")
	ErrInvalidCode       = errors.New("invalid verification code")
)
