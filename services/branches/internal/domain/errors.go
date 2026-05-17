package domain

import "errors"

var (
	ErrForbidden             = errors.New("forbidden")
	ErrUnauthorized          = errors.New("unauthorized")
	ErrBranchNotFound        = errors.New("branch not found")
	ErrInvalidBranchID       = errors.New("invalid branch id")
	ErrInvalidServiceID      = errors.New("invalid service id")
	ErrInvalidRegistrationSlug = errors.New("invalid registration slug")
)
