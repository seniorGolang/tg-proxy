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

// ValidateRepoURLMatchesSource проверяет схему и хост repoURL и sourceBaseURL. Возвращает errs.ErrRepoURLSourceMismatch при несовпадении.
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

// ParseManifestRefURL извлекает alias и version из URL манифеста прокси (baseURL/alias/version/...). ok=false, если refURL не наш.
func ParseManifestRefURL(baseURL string, refURL string) (alias string, version string, ok bool) {

	if baseURL == "" || refURL == "" {
		return "", "", false
	}

	baseParsed, err := url.Parse(baseURL)
	if err != nil {
		return "", "", false
	}
	refParsed, err := url.Parse(refURL)
	if err != nil {
		return "", "", false
	}

	if baseParsed.Scheme != refParsed.Scheme || baseParsed.Host != refParsed.Host {
		return "", "", false
	}

	basePath := strings.TrimSuffix(path.Clean(baseParsed.Path), "/")
	refPath := path.Clean(refParsed.Path)
	if basePath != "" && !strings.HasPrefix(refPath, basePath+"/") {
		return "", "", false
	}
	suffix := refPath
	if basePath != "" {
		suffix = strings.TrimPrefix(refPath, basePath+"/")
	}
	parts := strings.Split(suffix, "/")
	if len(parts) < 3 {
		return "", "", false
	}
	alias = parts[0]
	version = parts[1]
	return alias, version, true
}
