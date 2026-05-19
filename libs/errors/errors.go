package errors

import "errors"

var (
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidBranchID   = errors.New("invalid branch id")
	ErrInvalidEmployeeID = errors.New("invalid employee id")
	ErrInvalidServiceID  = errors.New("invalid service id")
)
