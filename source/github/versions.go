package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

const (
	refsTagsPrefix = "refs/tags/"
	tagSuffix      = "^{}"
	commentPrefix  = "#"
	newlineChar    = "\n"
	nullChar       = "\x00"
)

func (s *Source) GetVersions(ctx context.Context, project domain.Project) (versions []string, err error) {

	owner, repo := s.extractOwnerRepo(project.RepoURL)
	if owner == "" || repo == "" {
		err = fmt.Errorf("%w: invalid repo URL", errs.ErrGitHubAPI)
		return
	}

	var tags []string
	if tags, err = s.listTags(ctx, owner, repo, project); err != nil {
		return
	}

	versions = make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag != "" {
			versions = append(versions, tag)
		}
	}

	return
}

func (s *Source) listTags(ctx context.Context, owner string, repo string, project domain.Project) (tags []string, err error) {

	gitURL := fmt.Sprintf(gitInfoRefs, owner, repo) + gitService

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, gitURL, nil); err != nil {
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
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	req.Header.Set("Accept", "*/*")

	var resp *http.Response
	if resp, err = s.http.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var body []byte
		body, _ = io.ReadAll(resp.Body)
		err = fmt.Errorf("%w: status %d, body: %s", errs.ErrGitHubAPI, resp.StatusCode, string(body))
		return
	}

	var body []byte
	if body, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	tags = []string{}
	lines := strings.Split(string(body), newlineChar)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, commentPrefix) {
			continue
		}
		if !strings.Contains(line, refsTagsPrefix) {
			continue
		}
		parts := strings.Fields(line)
		for _, part := range parts {
			if strings.HasPrefix(part, refsTagsPrefix) {
				tag := strings.TrimPrefix(part, refsTagsPrefix)
				tag = strings.TrimSuffix(tag, tagSuffix)
				tag = strings.TrimRight(tag, nullChar)
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	}

	return
}
