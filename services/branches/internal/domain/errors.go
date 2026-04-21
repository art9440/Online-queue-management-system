package domain

import "errors"

var (
	ErrForbidden      = errors.New("forbidden")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrBranchNotFound = errors.New("branch not found")
)
