package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

func (s *Source) GetFileStream(ctx context.Context, project domain.Project, version string, filename string) (stream io.ReadCloser, err error) {

	var resp *http.Response
	if resp, err = s.GetFileResponse(ctx, project, version, filename); err != nil {
		return
	}

	stream = resp.Body
	return
}

func (s *Source) GetFileResponse(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error) {

	// Пытаемся получить файл из релиза (assets)
	// Если не найдено, пытаемся получить из содержимого репозитория
	apiURL := s.buildContentsURL(project.RepoURL, version, filename)

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
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	if resp, err = s.http.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		// Если файл не найден в содержимом, пытаемся найти в assets релиза
		return s.getFileFromRelease(ctx, project, version, filename)
	}

	return
}

func (s *Source) getFileFromRelease(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	apiURL := helpers.BuildURL(s.baseURL, "repos", owner, repo, "releases", "tags", version)

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

	// Ищем файл в assets
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

	// Загружаем файл из asset
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
