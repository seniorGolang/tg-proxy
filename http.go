package tgproxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/seniorGolang/tg-proxy/core"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/model/dto"
)

func (p *Proxy) SetPublicRoutes(mux *http.ServeMux, prefix string) {

	base := strings.TrimSuffix(prefix, "/")
	if base == "" {
		base = "/"
	}
	h := p.publicAuthMiddleware

	mux.HandleFunc("GET "+base, h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetAggregateManifestNetHTTP(w, r)
	}))
	mux.HandleFunc("GET "+path.Join(base, "manifest.yml"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetAggregateManifestNetHTTP(w, r)
	}))
	mux.HandleFunc("GET "+path.Join(base, "versions"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetCatalogVersionNetHTTP(w, r)
	}))
	mux.HandleFunc("GET "+path.Join(base, "{version}/manifest.yml"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetAggregateManifestAtVersionNetHTTP(w, r, r.PathValue("version"))
	}))
	mux.HandleFunc("GET "+path.Join(base, "{alias}/{version}/manifest.yml"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetManifestNetHTTP(w, r, r.PathValue("alias"), r.PathValue("version"))
	}))
	mux.HandleFunc("GET "+path.Join(base, "{alias}/{version}/{filename...}"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetFileNetHTTP(w, r, r.PathValue("alias"), r.PathValue("version"), r.PathValue("filename"))
	}))
	mux.HandleFunc("GET "+path.Join(base, "{alias}/versions"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetVersionsNetHTTP(w, r, r.PathValue("alias"))
	}))
}

func (p *Proxy) SetAdminRoutes(mux *http.ServeMux, prefix string) {

	base := strings.TrimSuffix(prefix, "/")
	if base == "" {
		base = "/"
	}
	h := p.adminAuthMiddleware

	mux.HandleFunc("GET "+path.Join(base, "projects"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleListProjectsNetHTTP(w, r)
	}))
	mux.HandleFunc("POST "+path.Join(base, "projects"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleCreateProjectNetHTTP(w, r)
	}))
	mux.HandleFunc("GET "+path.Join(base, "projects/{alias}"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleGetProjectNetHTTP(w, r, r.PathValue("alias"))
	}))
	mux.HandleFunc("PUT "+path.Join(base, "projects/{alias}"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleUpdateProjectNetHTTP(w, r, r.PathValue("alias"))
	}))
	mux.HandleFunc("DELETE "+path.Join(base, "projects/{alias}"), h(func(w http.ResponseWriter, r *http.Request) {
		p.handleDeleteProjectNetHTTP(w, r, r.PathValue("alias"))
	}))
}

func (p *Proxy) handleGetManifestNetHTTP(w http.ResponseWriter, r *http.Request, alias string, version string) {

	startTime := time.Now()

	manifest, statusCode, err := p.handleGetManifest(r.Context(), alias, version)
	if err != nil {
		slog.Error("Failed to get manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(statusCode)
	_, _ = w.Write(manifest)
}

func (p *Proxy) handleGetAggregateManifestNetHTTP(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	manifest, statusCode, err := p.handleGetAggregateManifest(r.Context())
	if err != nil {
		slog.Error("Failed to get aggregate manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Aggregate manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(statusCode)
	_, _ = w.Write(manifest)
}

func (p *Proxy) handleGetAggregateManifestAtVersionNetHTTP(w http.ResponseWriter, r *http.Request, requestedVersion string) {

	startTime := time.Now()

	currentVersion, statusCode, err := p.handleGetCatalogVersion(r.Context())
	if err != nil {
		slog.Error("Failed to get catalog version",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}
	if requestedVersion != currentVersion {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	manifest, statusCode, err := p.handleGetAggregateManifest(r.Context())
	if err != nil {
		slog.Error("Failed to get aggregate manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Aggregate manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
		slog.String(helpers.LogKeyVersion, currentVersion),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(statusCode)
	_, _ = w.Write(manifest)
}

func (p *Proxy) handleGetCatalogVersionNetHTTP(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	version, statusCode, err := p.handleGetCatalogVersion(r.Context())
	if err != nil {
		slog.Error("Failed to get catalog version",
			slog.String(helpers.LogKeyAction, helpers.ActionGetCatalogVersion),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Catalog version request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetCatalogVersion),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode([]string{version})
}

func (p *Proxy) handleGetFileNetHTTP(w http.ResponseWriter, r *http.Request, alias string, version string, filename string) {

	startTime := time.Now()

	stream, statusCode, err := p.handleGetFile(r.Context(), alias, version, filename)
	if err != nil {
		slog.Error("Failed to get file",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}
	defer stream.Close()

	var project domain.Project
	var found bool
	if project, found, _ = p.engine.GetProject(r.Context(), alias); !found {
		slog.Debug("Project not found during file streaming",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
		)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	var src core.Source
	if src, err = p.engine.GetSource(project.SourceName); err != nil {
		slog.Error("Source not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, "Source not found", http.StatusInternalServerError)
		return
	}

	slog.Debug("Fetching file from source",
		slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeyFilename, filename),
		slog.String(helpers.LogKeySource, project.SourceName),
	)

	var resp *http.Response
	if resp, err = src.GetFileResponse(r.Context(), project, version, filename); err != nil {
		slog.Error("Failed to fetch file from source",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, "Failed to fetch file", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	slog.Info("File request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeyFilename, filename),
		slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.String("content_type", resp.Header.Get("Content-Type")),
		slog.String("content_length", resp.Header.Get("Content-Length")),
	)

	p.copyResponseHeadersNetHTTP(w, resp)

	_, _ = io.Copy(w, resp.Body)
}

func (p *Proxy) handleGetVersionsNetHTTP(w http.ResponseWriter, r *http.Request, alias string) {

	startTime := time.Now()

	versions, statusCode, err := p.handleGetVersions(r.Context(), alias)
	if err != nil {
		slog.Error("Failed to get versions",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Versions request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int(helpers.LogKeyVersionsCount, len(versions)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(versions)
}

func (p *Proxy) handleListProjectsNetHTTP(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	projects, total, statusCode, err := p.handleListProjects(r.Context(), limit, offset)
	if err != nil {
		slog.Error("Failed to list projects",
			slog.String(helpers.LogKeyAction, helpers.ActionListProjects),
			slog.Int(helpers.LogKeyLimit, limit),
			slog.Int(helpers.LogKeyOffset, offset),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("List projects request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionListProjects),
		slog.Int(helpers.LogKeyLimit, limit),
		slog.Int(helpers.LogKeyOffset, offset),
		slog.Int64(helpers.LogKeyTotal, total),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("returned_count", len(projects)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"projects": projects,
		"total":    total,
	})
}

func (p *Proxy) handleCreateProjectNetHTTP(w http.ResponseWriter, r *http.Request) {

	startTime := time.Now()

	var req dto.ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("Invalid request body",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := helpers.ValidateStruct(&req); err != nil {
		slog.Debug("Validation failed",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyAlias, req.Alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode, err := p.handleCreateProject(r.Context(), req)
	if err != nil {
		slog.Error("Failed to create project",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyAlias, req.Alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
		slog.String(helpers.LogKeyAlias, req.Alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	}
	if req.SourceName != "" {
		args = append(args, slog.String(helpers.LogKeySource, req.SourceName))
	}
	if req.RepoURL != "" {
		args = append(args, slog.String(helpers.LogKeyRepoURL, req.RepoURL))
	}
	slog.Info("Create project request completed", args...)

	w.WriteHeader(statusCode)
}

func (p *Proxy) handleGetProjectNetHTTP(w http.ResponseWriter, r *http.Request, alias string) {

	startTime := time.Now()

	if err := helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, found, statusCode, err := p.handleGetProject(r.Context(), alias)
	if err != nil {
		slog.Error("Failed to get project",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}
	if !found {
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
		)
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	}
	if project.SourceName != "" {
		args = append(args, slog.String(helpers.LogKeySource, project.SourceName))
	}
	if project.RepoURL != "" {
		args = append(args, slog.String(helpers.LogKeyRepoURL, project.RepoURL))
	}
	slog.Info("Get project request completed", args...)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(project)
}

func (p *Proxy) handleUpdateProjectNetHTTP(w http.ResponseWriter, r *http.Request, alias string) {

	startTime := time.Now()

	if err := helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req dto.ProjectUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("Invalid request body",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := helpers.ValidateStruct(&req); err != nil {
		slog.Debug("Validation failed",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode, err := p.handleUpdateProject(r.Context(), alias, req)
	if err != nil {
		slog.Error("Failed to update project",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	}
	if req.SourceName != nil {
		args = append(args, slog.String(helpers.LogKeySource, *req.SourceName))
	}
	if req.RepoURL != nil {
		args = append(args, slog.String(helpers.LogKeyRepoURL, *req.RepoURL))
	}
	if req.Token != nil {
		args = append(args, slog.String(helpers.LogKeyTokenMasked, helpers.MaskToken(*req.Token)))
	}
	if req.Description != nil {
		args = append(args, slog.String(helpers.LogKeyDescription, *req.Description))
	}
	slog.Info("Update project request completed", args...)

	w.WriteHeader(statusCode)
}

func (p *Proxy) handleDeleteProjectNetHTTP(w http.ResponseWriter, r *http.Request, alias string) {

	startTime := time.Now()

	if err := helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	statusCode, err := p.handleDeleteProject(r.Context(), alias)
	if err != nil {
		slog.Error("Failed to delete project",
			slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, r.Method),
			slog.String(helpers.LogKeyPath, r.URL.Path),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		http.Error(w, helpers.GetErrorMessage(err), statusCode)
		return
	}

	slog.Info("Delete project request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, r.Method),
		slog.String(helpers.LogKeyPath, r.URL.Path),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	)

	w.WriteHeader(statusCode)
}

func (p *Proxy) copyResponseHeadersNetHTTP(w http.ResponseWriter, resp *http.Response) {

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}

	w.Header().Set("Cache-Control", "public, max-age=3600")
}
