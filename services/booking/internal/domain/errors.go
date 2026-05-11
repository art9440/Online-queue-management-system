package domain

import "errors"

var (
	ErrInvalidStartTime     = errors.New("invalid start time")
	ErrInvalidClient        = errors.New("invalid client")
	ErrInvalidClientContact = errors.New("invalid client contact")
)
