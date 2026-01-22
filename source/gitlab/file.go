package gitlab

import (
	"context"
	"fmt"
	"io"
	"log/slog"
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

	apiURL := s.buildAPIURLForManifest(project.RepoURL, version, filename)

	slog.Debug("GitLab API request",
		slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
		slog.String(helpers.LogKeySource, s.Name()),
		slog.String(helpers.LogKeyRequestURL, apiURL),
		slog.String(helpers.LogKeyRepoURL, project.RepoURL),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeyFilename, filename),
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

	if resp, err = s.http.Do(req); err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Debug("GitLab API error response",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeySource, s.Name()),
			slog.String(helpers.LogKeyRequestURL, apiURL),
			slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
			slog.String(helpers.LogKeyRepoURL, project.RepoURL),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
		)
		resp.Body.Close()
		err = fmt.Errorf("%w: status %d", errs.ErrGitLabAPI, resp.StatusCode)
		return
	}

	return
}
