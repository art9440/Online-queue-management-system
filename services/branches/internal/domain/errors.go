package domain

import "errors"

var (
	ErrForbidden               = errors.New("forbidden")
	ErrUnauthorized            = errors.New("unauthorized")
	ErrBusinessNotFound        = errors.New("business not found")
	ErrBranchNotFound          = errors.New("branch not found")
	ErrInvalidBusinessID       = errors.New("invalid business id")
	ErrInvalidBranchID         = errors.New("invalid branch id")
	ErrInvalidServiceID        = errors.New("invalid service id")
	ErrInvalidRegistrationSlug = errors.New("invalid registration slug")
	ErrRegistrationSlugNotSet  = errors.New("registration slug is not set")
)
