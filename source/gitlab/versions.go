package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/source/gitlab/internal"
)

func (s *Source) GetVersions(ctx context.Context, project domain.Project) (versions []string, err error) {

	projectPath := s.extractProjectPath(project.RepoURL)
	apiURL := s.buildAPIURLWithQuery(
		map[string]string{
			"package_type": "generic",
			"package_name": "release",
		},
		"api", "v4", "projects", projectPath, "packages",
	)

	slog.Debug("GitLab API request",
		slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
		slog.String(helpers.LogKeySource, sourceName),
		slog.String(helpers.LogKeyRequestURL, apiURL),
		slog.String(helpers.LogKeyRepoURL, project.RepoURL),
		slog.String("project_path", projectPath),
	)

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil); err != nil {
		return
	}

	var token string
	if project.Token != "" {
		token = project.Token
	} else {
		token = s.token
	}
	if token != "" {
		req.Header.Set(privateTokenHeader, token) //nolint:canonicalheader
	}

	var resp *http.Response
	if resp, err = s.http.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("GitLab API error response",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeySource, sourceName),
			slog.String(helpers.LogKeyRequestURL, apiURL),
			slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
			slog.String(helpers.LogKeyRepoURL, project.RepoURL),
			slog.String("project_path", projectPath),
		)
		err = fmt.Errorf("%w: status %d", errs.ErrGitLabAPI, resp.StatusCode)
		return
	}

	var packages []internal.Package
	if err = json.NewDecoder(resp.Body).Decode(&packages); err != nil {
		return
	}

	versions = make([]string, 0, len(packages))
	for _, pkg := range packages {
		if pkg.Version != "" {
			versions = append(versions, pkg.Version)
		}
	}

	return
}

func (s *Source) buildAPIURLWithQuery(queryParams map[string]string, pathParts ...string) (apiURL string) {

	var parsedBaseURL *url.URL
	var err error
	if parsedBaseURL, err = url.Parse(s.baseURL); err != nil {
		return ""
	}

	// Path и RawPath раздельно, чтобы избежать двойного кодирования при сборке URL.
	pathSegments := make([]string, 0, len(pathParts)+1)
	rawPathSegments := make([]string, 0, len(pathParts)+1)
	basePath := strings.TrimSuffix(parsedBaseURL.Path, "/")
	if basePath != "" {
		pathSegments = append(pathSegments, basePath)
		rawPathSegments = append(rawPathSegments, basePath)
	}

	for _, part := range pathParts {
		if part != "" {
			escapedPart := url.PathEscape(part)
			rawPathSegments = append(rawPathSegments, escapedPart)
			pathSegments = append(pathSegments, part)
		}
	}

	newPath := strings.Join(pathSegments, "/")
	rawPath := strings.Join(rawPathSegments, "/")
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
		rawPath = "/" + rawPath
	}

	parsedBaseURL.Path = newPath
	parsedBaseURL.RawPath = rawPath

	if len(queryParams) > 0 {
		query := parsedBaseURL.Query()
		for key, value := range queryParams {
			query.Set(key, value)
		}
		parsedBaseURL.RawQuery = query.Encode()
	}

	if parsedBaseURL.RawPath != "" && parsedBaseURL.RawPath != parsedBaseURL.Path {
		apiURL = parsedBaseURL.Scheme + "://" + parsedBaseURL.Host + parsedBaseURL.RawPath
		if parsedBaseURL.RawQuery != "" {
			apiURL += "?" + parsedBaseURL.RawQuery
		}
		if parsedBaseURL.Fragment != "" {
			apiURL += "#" + parsedBaseURL.Fragment
		}
	} else {
		apiURL = parsedBaseURL.String()
	}
	return
}

func (s *Source) buildAPIURL(pathParts ...string) (apiURL string) {

	var parsedBaseURL *url.URL
	var err error
	if parsedBaseURL, err = url.Parse(s.baseURL); err != nil {
		return ""
	}

	// Path и RawPath раздельно, чтобы избежать двойного кодирования при сборке URL.
	pathSegments := make([]string, 0, len(pathParts)+1)
	rawPathSegments := make([]string, 0, len(pathParts)+1)
	basePath := strings.TrimSuffix(parsedBaseURL.Path, "/")
	if basePath != "" {
		pathSegments = append(pathSegments, basePath)
		rawPathSegments = append(rawPathSegments, basePath)
	}

	for _, part := range pathParts {
		if part != "" {
			escapedPart := url.PathEscape(part)
			rawPathSegments = append(rawPathSegments, escapedPart)
			pathSegments = append(pathSegments, part)
		}
	}

	newPath := strings.Join(pathSegments, "/")
	rawPath := strings.Join(rawPathSegments, "/")
	if !strings.HasPrefix(newPath, "/") {
		newPath = "/" + newPath
		rawPath = "/" + rawPath
	}

	parsedBaseURL.Path = newPath
	parsedBaseURL.RawPath = rawPath

	if parsedBaseURL.RawPath != "" && parsedBaseURL.RawPath != parsedBaseURL.Path {
		apiURL = parsedBaseURL.Scheme + "://" + parsedBaseURL.Host + parsedBaseURL.RawPath
		if parsedBaseURL.RawQuery != "" {
			apiURL += "?" + parsedBaseURL.RawQuery
		}
		if parsedBaseURL.Fragment != "" {
			apiURL += "#" + parsedBaseURL.Fragment
		}
	} else {
		apiURL = parsedBaseURL.String()
	}
	return
}

func (s *Source) ValidateProject(ctx context.Context, project domain.Project) (err error) {

	projectPath := s.extractProjectPath(project.RepoURL)
	apiURL := s.buildAPIURL("api", "v4", "projects", projectPath)

	slog.Debug("GitLab API request",
		slog.String(helpers.LogKeyAction, "validate_project"),
		slog.String(helpers.LogKeySource, sourceName),
		slog.String(helpers.LogKeyRequestURL, apiURL),
		slog.String(helpers.LogKeyRepoURL, project.RepoURL),
		slog.String("project_path", projectPath),
	)

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil); err != nil {
		return
	}

	var token string
	if project.Token != "" {
		token = project.Token
	} else {
		token = s.token
	}
	if token != "" {
		req.Header.Set(privateTokenHeader, token) //nolint:canonicalheader
	}

	var resp *http.Response
	if resp, err = s.http.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("GitLab API error response",
			slog.String(helpers.LogKeyAction, "validate_project"),
			slog.String(helpers.LogKeySource, sourceName),
			slog.String(helpers.LogKeyRequestURL, apiURL),
			slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
			slog.String(helpers.LogKeyRepoURL, project.RepoURL),
			slog.String("project_path", projectPath),
		)
		err = fmt.Errorf("%w: status %d", errs.ErrGitLabAPI, resp.StatusCode)
		return
	}

	return
}
