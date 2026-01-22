package helpers

import (
	"net/url"
	"path"
	"strings"
)

// BuildURL строит URL из базового URL и частей пути
// baseURL - базовый URL (например, https://api.github.com)
// pathParts - части пути, которые будут объединены и экранированы
// Возвращает полный URL или пустую строку в случае ошибки парсинга baseURL
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

// BuildURLWithQuery строит URL из базового URL, частей пути и query параметров
// baseURL - базовый URL (например, https://api.github.com)
// queryParams - map с query параметрами (ключ - имя параметра, значение - значение параметра)
// pathParts - части пути, которые будут объединены и экранированы
// Возвращает полный URL или пустую строку в случае ошибки парсинга baseURL
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

// NormalizeRepoURL нормализует URL репозитория, удаляя суффикс .git если он присутствует
func NormalizeRepoURL(repoURL string) (normalized string) {

	if repoURL == "" {
		return ""
	}

	normalized = strings.TrimSuffix(repoURL, ".git")
	return
}
