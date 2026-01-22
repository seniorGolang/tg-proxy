package errs

import "errors"

var (
	ErrManifestParseError   = errors.New("manifest parse error")
	ErrManifestMarshalError = errors.New("manifest marshal error")
)
