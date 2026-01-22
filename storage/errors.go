package storage

import "github.com/seniorGolang/tg-proxy/errs"

var (
	ErrProjectNotFound      = errs.ErrProjectNotFound
	ErrProjectAlreadyExists = errs.ErrProjectAlreadyExists
	ErrInvalidProject       = errs.ErrInvalidProject
	ErrStorageError         = errs.ErrStorageError
)
