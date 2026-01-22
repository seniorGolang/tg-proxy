package errs

import "errors"

var (
	ErrGitLabAPI = errors.New("gitlab api error")
	ErrGitHubAPI = errors.New("github api error")
)
