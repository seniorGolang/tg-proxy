package gitlab

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

func (s *Source) GetManifest(ctx context.Context, project domain.Project, version string) (manifest domain.Manifest, err error) {

	apiURL := s.buildAPIURLForManifest(project.RepoURL, version, "manifest.yml")

	slog.Debug("GitLab API request",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
		slog.String(helpers.LogKeySource, sourceName),
		slog.String(helpers.LogKeyRequestURL, apiURL),
		slog.String(helpers.LogKeyRepoURL, project.RepoURL),
		slog.String(helpers.LogKeyVersion, version),
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
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeySource, sourceName),
			slog.String(helpers.LogKeyRequestURL, apiURL),
			slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
			slog.String(helpers.LogKeyRepoURL, project.RepoURL),
			slog.String(helpers.LogKeyVersion, version),
		)
		err = fmt.Errorf("%w: status %d", errs.ErrGitLabAPI, resp.StatusCode)
		return
	}

	var data []byte
	if data, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	var modelManifest model.Manifest
	if err = yaml.Unmarshal(data, &modelManifest); err != nil {
		err = fmt.Errorf("%w: %w", errs.ErrManifestParseError, err)
		return
	}

	manifest = modelManifest.ToDomain()
	return
}

func (s *Source) buildAPIURLForManifest(repoURL string, version string, filename string) (apiURL string) {

	projectPath := s.extractProjectPath(repoURL)
	apiURL = s.buildAPIURL("api", "v4", "projects", projectPath, "packages", "generic", "release", version, filename)
	return
}

func (s *Source) extractProjectPath(repoURL string) (path string) {

	var err error
	var parsedURL *url.URL
	if parsedURL, err = url.Parse(repoURL); err != nil {
		slog.Debug("Failed to parse repo URL",
			slog.String(helpers.LogKeySource, sourceName),
			slog.String(helpers.LogKeyRepoURL, repoURL),
			slog.Any(helpers.LogKeyError, err),
		)
		return ""
	}

	path = strings.TrimPrefix(parsedURL.Path, "/")
	path = strings.TrimSuffix(path, ".git")

	slog.Debug("Extracted project path",
		slog.String(helpers.LogKeySource, sourceName),
		slog.String(helpers.LogKeyRepoURL, repoURL),
		slog.String("project_path", path),
	)

	return
}
