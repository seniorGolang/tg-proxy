package errs

import "errors"

var (
	ErrVersionNotFound = errors.New("version not found")
	ErrVersionMismatch = errors.New("version mismatch")
)
