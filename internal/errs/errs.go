package errs

import "errors"

var (
	ErrLinkNotFound      = errors.New("link not found")
	ErrCodeAlreadyExists = errors.New("short code already exists")
)
