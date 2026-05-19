package domain

import "errors"

var ErrBadCredentials = errors.New("bad credentials")
var ErrUnauthorized = errors.New("unauthorized")