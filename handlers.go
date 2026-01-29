package tgproxy

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/model/dto"
)

func (p *Proxy) handleGetManifest(ctx context.Context, alias string, version string) (manifest []byte, statusCode int, err error) {

	if manifest, err = p.engine.GetManifest(ctx, alias, version, p.baseURL); err != nil {
		if errors.Is(err, errs.ErrProjectNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		if errors.Is(err, errs.ErrVersionNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		if errors.Is(err, errs.ErrVersionMismatch) {
			statusCode = http.StatusBadRequest
			return
		}
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleGetAggregateManifest(ctx context.Context) (manifest []byte, statusCode int, err error) {

	if manifest, err = p.engine.GetAggregateManifest(ctx, p.baseURL); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleGetCatalogVersion(ctx context.Context) (version string, statusCode int, err error) {

	if version, err = p.engine.GetCatalogVersion(ctx); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleGetFile(ctx context.Context, alias string, version string, filename string) (stream io.ReadCloser, statusCode int, err error) {

	if stream, err = p.engine.GetFile(ctx, alias, version, filename); err != nil {
		if errors.Is(err, errs.ErrProjectNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		if errors.Is(err, errs.ErrVersionNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		if errors.Is(err, errs.ErrVersionMismatch) {
			statusCode = http.StatusBadRequest
			return
		}
		statusCode = http.StatusBadGateway
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleGetVersions(ctx context.Context, alias string) (versions []string, statusCode int, err error) {

	if versions, err = p.engine.GetVersions(ctx, alias); err != nil {
		if errors.Is(err, errs.ErrProjectNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleCreateProject(ctx context.Context, req dto.ProjectCreateRequest) (statusCode int, err error) {

	project := req.ToDomain()

	src, err := p.engine.GetSource(project.SourceName)
	if err != nil {
		if errors.Is(err, errs.ErrSourceNotFound) {
			return http.StatusBadRequest, err
		}
		return http.StatusInternalServerError, err
	}
	_, sourceURL := src.Info()
	if err = helpers.ValidateRepoURLMatchesSource(project.RepoURL, sourceURL); err != nil {
		if errors.Is(err, errs.ErrRepoURLSourceMismatch) {
			return http.StatusBadRequest, err
		}
		return http.StatusInternalServerError, err
	}

	if err = p.engine.CreateProject(ctx, project); err != nil {
		if errors.Is(err, errs.ErrProjectAlreadyExists) {
			statusCode = http.StatusConflict
			return
		}
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusCreated
	return
}

func (p *Proxy) handleGetProject(ctx context.Context, alias string) (project dto.ProjectResponse, found bool, statusCode int, err error) {

	slog.Info("Getting project",
		slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
		slog.String(helpers.LogKeyAlias, alias),
	)

	slog.Debug("Project request details",
		slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
		slog.String(helpers.LogKeyAlias, alias),
	)

	var domainProject domain.Project
	if domainProject, found, err = p.engine.GetProject(ctx, alias); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}
	if !found {
		statusCode = http.StatusNotFound
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
		)
		return
	}

	project = dto.FromDomain(domainProject)

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleUpdateProject(ctx context.Context, alias string, req dto.ProjectUpdateRequest) (statusCode int, err error) {

	var currentProject domain.Project
	var found bool
	if currentProject, found, err = p.engine.GetProject(ctx, alias); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}
	if !found {
		statusCode = http.StatusNotFound
		return
	}

	updateProject := req.ToDomain(alias)

	if req.SourceName != nil {
		if _, err = p.engine.GetSource(*req.SourceName); err != nil {
			if errors.Is(err, errs.ErrSourceNotFound) {
				return http.StatusBadRequest, err
			}
			return http.StatusInternalServerError, err
		}
	}

	if req.RepoURL != nil {
		currentProject.RepoURL = updateProject.RepoURL
	}
	if req.Token != nil {
		currentProject.Token = updateProject.Token
	}
	if req.Description != nil {
		currentProject.Description = updateProject.Description
	}
	if req.SourceName != nil {
		currentProject.SourceName = updateProject.SourceName
	}

	src, err := p.engine.GetSource(currentProject.SourceName)
	if err != nil {
		if errors.Is(err, errs.ErrSourceNotFound) {
			return http.StatusBadRequest, err
		}
		return http.StatusInternalServerError, err
	}
	_, sourceURL := src.Info()
	if err = helpers.ValidateRepoURLMatchesSource(currentProject.RepoURL, sourceURL); err != nil {
		if errors.Is(err, errs.ErrRepoURLSourceMismatch) {
			return http.StatusBadRequest, err
		}
		return http.StatusInternalServerError, err
	}

	if err = p.engine.UpdateProject(ctx, alias, currentProject); err != nil {
		if errors.Is(err, errs.ErrProjectNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusOK
	return
}

func (p *Proxy) handleDeleteProject(ctx context.Context, alias string) (statusCode int, err error) {

	if err = p.engine.DeleteProject(ctx, alias); err != nil {
		if errors.Is(err, errs.ErrProjectNotFound) {
			statusCode = http.StatusNotFound
			return
		}
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusNoContent
	return
}

func (p *Proxy) handleListProjects(ctx context.Context, limit int, offset int) (projects []dto.ProjectResponse, total int64, statusCode int, err error) {

	var domainProjects []domain.Project
	if domainProjects, total, err = p.engine.ListProjects(ctx, limit, offset); err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	projects = make([]dto.ProjectResponse, len(domainProjects))
	for i := range domainProjects {
		projects[i] = dto.FromDomain(domainProjects[i])
	}

	statusCode = http.StatusOK
	return
}
