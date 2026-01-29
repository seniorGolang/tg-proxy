package errs

import "errors"

var (
	ErrSourceNotFound          = errors.New("source not found")
	ErrSourceAlreadyRegistered = errors.New("source already registered")
	ErrInvalidSourceType       = errors.New("invalid source type")
	ErrRepoURLSourceMismatch   = errors.New("repo_url must be on the same domain and scheme as the source")
)
