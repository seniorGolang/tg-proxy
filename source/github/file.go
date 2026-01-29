package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

func ensureVersionPrefix(version string) (tag string) {

	tag = version
	if tag != "" && !strings.HasPrefix(tag, versionPrefix) {
		tag = versionPrefix + tag
	}
	return tag
}

func (s *Source) GetFileStream(ctx context.Context, project domain.Project, version string, filename string) (stream io.ReadCloser, err error) {

	var resp *http.Response
	if resp, err = s.GetFileResponse(ctx, project, version, filename); err != nil {
		return
	}

	stream = resp.Body
	return
}

func (s *Source) GetFileResponse(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	if owner == "" || repo == "" {
		err = fmt.Errorf("%w: invalid repo URL", errs.ErrGitHubAPI)
		return
	}

	tag := ensureVersionPrefix(version)
	directURL := fmt.Sprintf(releasesURL+"/%s", owner, repo, tag, filename)

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

	if resp, err = s.http.Do(req); err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		return
	}

	_ = resp.Body.Close()
	return s.getFileFromRelease(ctx, project, version, filename)
}

func (s *Source) getFileFromRelease(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	apiURL := helpers.BuildURL(apiBaseURL, "repos", owner, repo, "releases", "tags", version)

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

	var releaseResp *http.Response
	if releaseResp, err = s.http.Do(req); err != nil {
		return
	}
	defer releaseResp.Body.Close()

	if releaseResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%w: status %d", errs.ErrGitHubAPI, releaseResp.StatusCode)
		return
	}

	var release struct {
		Assets []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"assets"`
	}
	if err = json.NewDecoder(releaseResp.Body).Decode(&release); err != nil {
		return
	}

	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == filename {
			assetURL = asset.URL
			break
		}
	}

	if assetURL == "" {
		err = fmt.Errorf("%w: file %s not found in release %s", errs.ErrFileNotFound, filename, version)
		return
	}

	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil); err != nil {
		return
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}
	req.Header.Set("Accept", "application/octet-stream")

	if resp, err = s.http.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		err = fmt.Errorf("%w: status %d", errs.ErrGitHubAPI, resp.StatusCode)
		return
	}

	return
}
