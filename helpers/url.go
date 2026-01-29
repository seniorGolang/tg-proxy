package helpers

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/seniorGolang/tg-proxy/errs"
)

func BuildURL(baseURL string, pathParts ...string) (resultURL string) {

	var parsedURL *url.URL
	var err error
	if parsedURL, err = url.Parse(baseURL); err != nil {
		return ""
	}

	escapedParts := make([]string, 0, len(pathParts))
	for _, part := range pathParts {
		if part != "" {
			escapedParts = append(escapedParts, url.PathEscape(part))
		}
	}

	basePath := strings.TrimSuffix(parsedURL.Path, "/")
	newPath := path.Join(append([]string{basePath}, escapedParts...)...)
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
	}
	parsedURL.Path = newPath

	resultURL = parsedURL.String()
	return
}

func BuildURLWithQuery(baseURL string, queryParams map[string]string, pathParts ...string) (resultURL string) {

	var parsedURL *url.URL
	var err error
	if parsedURL, err = url.Parse(baseURL); err != nil {
		return ""
	}

	escapedParts := make([]string, 0, len(pathParts))
	for _, part := range pathParts {
		if part != "" {
			escapedParts = append(escapedParts, url.PathEscape(part))
		}
	}

	basePath := strings.TrimSuffix(parsedURL.Path, "/")
	newPath := path.Join(append([]string{basePath}, escapedParts...)...)
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
	}
	parsedURL.Path = newPath

	if len(queryParams) > 0 {
		query := parsedURL.Query()
		for key, value := range queryParams {
			query.Set(key, value)
		}
		parsedURL.RawQuery = query.Encode()
	}

	resultURL = parsedURL.String()
	return
}

func NormalizeRepoURL(repoURL string) (normalized string) {

	if repoURL == "" {
		return ""
	}

	normalized = strings.TrimSuffix(repoURL, ".git")
	return
}

// ValidateRepoURLMatchesSource проверяет, что repoURL имеет ту же схему и хост, что и sourceBaseURL.
// Возвращает errs.ErrRepoURLSourceMismatch при несовпадении или ошибке разбора URL.
func ValidateRepoURLMatchesSource(repoURL string, sourceBaseURL string) (err error) {

	repoParsed, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrRepoURLSourceMismatch, err)
	}
	sourceParsed, err := url.Parse(sourceBaseURL)
	if err != nil {
		return fmt.Errorf("%w: %v", errs.ErrRepoURLSourceMismatch, err)
	}

	repoScheme := strings.ToLower(repoParsed.Scheme)
	repoHost := strings.ToLower(repoParsed.Host)
	sourceScheme := strings.ToLower(sourceParsed.Scheme)
	sourceHost := strings.ToLower(sourceParsed.Host)

	if repoScheme != sourceScheme || repoHost != sourceHost {
		return errs.ErrRepoURLSourceMismatch
	}

	return nil
}
