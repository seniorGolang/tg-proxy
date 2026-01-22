package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/source/github/internal"
)

func (s *Source) GetVersions(ctx context.Context, project domain.Project) (versions []string, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	apiURL := helpers.BuildURL(s.baseURL, "repos", owner, repo, "releases")

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

	var releases []internal.Release
	if err = json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return
	}

	versions = make([]string, 0, len(releases))
	for _, release := range releases {
		if release.TagName != "" {
			versions = append(versions, release.TagName)
		}
	}

	return
}

func (s *Source) ValidateProject(ctx context.Context, project domain.Project) (err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	apiURL := helpers.BuildURL(s.baseURL, "repos", owner, repo)

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

	return
}
