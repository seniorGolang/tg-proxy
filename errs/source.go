package errs

import "errors"

var (
	ErrSourceNotFound          = errors.New("source not found")
	ErrSourceAlreadyRegistered = errors.New("source already registered")
	ErrInvalidSourceType       = errors.New("invalid source type")
)
