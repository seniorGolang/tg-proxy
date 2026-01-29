package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

	manifest, err = s.getManifestFromRelease(ctx, project, version)
	if err == nil {
		return manifest, nil
	}

	apiURL := s.buildContentsURL(project.RepoURL, version, "manifest.yml")

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
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	var resp *http.Response
	if resp, err = s.http.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%w: status %d", errs.ErrGitHubAPI, resp.StatusCode)
		return
	}

	var contentsResponse struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	if err = s.decodeJSONResponse(resp.Body, &contentsResponse); err != nil {
		return
	}

	var data []byte
	if contentsResponse.Encoding == "base64" {
		if data, err = base64.StdEncoding.DecodeString(contentsResponse.Content); err != nil {
			err = fmt.Errorf("failed to decode base64 content: %w", err)
			return
		}
	} else {
		data = []byte(contentsResponse.Content)
	}

	var modelManifest model.Manifest
	if err = yaml.Unmarshal(data, &modelManifest); err != nil {
		err = fmt.Errorf("%w: %w", errs.ErrManifestParseError, err)
		return
	}

	manifest = modelManifest.ToDomain()
	return
}

func (s *Source) getManifestFromRelease(ctx context.Context, project domain.Project, version string) (manifest domain.Manifest, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	if owner == "" || repo == "" {
		err = fmt.Errorf("%w: invalid repo URL", errs.ErrGitHubAPI)
		return
	}

	tag := ensureVersionPrefix(version)
	for _, name := range []string{"manifest.yaml", "manifest.yml"} {
		directURL := fmt.Sprintf(releasesURL+"/%s", owner, repo, tag, name)
		var req *http.Request
		if req, err = http.NewRequestWithContext(ctx, http.MethodGet, directURL, nil); err != nil {
			return
		}
		var token string
		if project.Token != "" {
			token = project.Token
		} else {
			token = s.token
		}
		if token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
		req.Header.Set("Accept", "application/octet-stream")

		var resp *http.Response
		if resp, err = s.http.Do(req); err != nil {
			return
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			continue
		}
		var data []byte
		if data, err = io.ReadAll(resp.Body); err != nil {
			_ = resp.Body.Close()
			return
		}
		_ = resp.Body.Close()

		var modelManifest model.Manifest
		if err = yaml.Unmarshal(data, &modelManifest); err != nil {
			err = fmt.Errorf("%w: %w", errs.ErrManifestParseError, err)
			return
		}
		manifest = modelManifest.ToDomain()
		return manifest, nil
	}

	err = fmt.Errorf("%w: manifest not found in release %s", errs.ErrManifestParseError, version)
	return
}

func (s *Source) buildContentsURL(repoURL string, ref string, path string) (apiURL string) {

	owner, repo := s.extractOwnerRepo(repoURL)
	apiURL = helpers.BuildURLWithQuery(apiBaseURL, map[string]string{"ref": ref}, "repos", owner, repo, "contents", path)
	return
}

func (s *Source) extractOwnerRepo(repoURL string) (owner string, repo string) {

	var parsedURL *url.URL
	var err error
	if parsedURL, err = url.Parse(repoURL); err != nil {
		return "", ""
	}

	path := strings.TrimPrefix(parsedURL.Path, "/")
	owner, rest, _ := strings.Cut(path, "/")
	if rest != "" {
		repo, _, _ = strings.Cut(rest, "/")
		repo = strings.TrimSuffix(repo, ".git")
	}

	return
}

func (s *Source) decodeJSONResponse(body io.Reader, target interface{}) (err error) {

	var data []byte
	if data, err = io.ReadAll(body); err != nil {
		return
	}

	if err = json.Unmarshal(data, target); err != nil {
		return
	}

	return
}
