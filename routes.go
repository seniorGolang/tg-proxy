package tgproxy

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/seniorGolang/tg-proxy/core"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
	"github.com/seniorGolang/tg-proxy/model/dto"
)

// normalizePrefix нормализует префикс пути, убирая завершающий слэш.
// Возвращает нормализованный префикс и префикс для обрезки пути.
// Если входной префикс равен "/" или пустой, возвращает "/" для обоих значений.
func normalizePrefix(prefix string) (normalizedPrefix string, trimPrefix string) {

	normalizedPrefix = strings.TrimSuffix(prefix, "/")
	if normalizedPrefix == "" {
		normalizedPrefix = "/"
	}

	if normalizedPrefix == "/" {
		trimPrefix = "/"
	} else {
		trimPrefix = normalizedPrefix + "/"
	}

	return normalizedPrefix, trimPrefix
}

// SetPublicRoutes регистрирует публичные роуты в стандартном net/http
// Публичные роуты: получение манифестов, файлов, списка версий
// mux - HTTP ServeMux для регистрации роутов
// prefix - префикс для всех роутов (например, "/api/v1/proxy")
func (p *Proxy) SetPublicRoutes(mux *http.ServeMux, prefix string) {

	normalizedPrefix, trimPrefix := normalizePrefix(prefix)

	mux.HandleFunc(trimPrefix, p.publicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		p.handlePublicRoutesNetHTTP(w, r, normalizedPrefix)
	}))
}

func (p *Proxy) handlePublicRoutesNetHTTP(w http.ResponseWriter, r *http.Request, prefix string) {

	_, trimPrefix := normalizePrefix(prefix)
	path := strings.TrimPrefix(r.URL.Path, trimPrefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	alias := parts[0]

	// Манифест: /<alias>/<version>/manifest.yml
	if len(parts) >= 3 && parts[2] == "manifest.yml" {
		version := parts[1]
		p.handleGetManifestNetHTTP(w, r, alias, version)
		return
	}

	// Файлы: /<alias>/<version>/<filename>
	if len(parts) >= 3 {
		version := parts[1]
		filename := strings.Join(parts[2:], "/")
		p.handleGetFileNetHTTP(w, r, alias, version, filename)
		return
	}

	// Версии: /<alias>/versions
	if len(parts) >= 2 && parts[1] == "versions" {
		p.handleGetVersionsNetHTTP(w, r, alias)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

// SetAdminRoutes регистрирует админские роуты в стандартном net/http
// Админские роуты: управление проектами (CRUD операции)
// mux - HTTP ServeMux для регистрации роутов
// prefix - префикс для всех роутов (например, "/api/v1/admin")
func (p *Proxy) SetAdminRoutes(mux *http.ServeMux, prefix string) {

	normalizedPrefix, trimPrefix := normalizePrefix(prefix)

	mux.HandleFunc(trimPrefix, p.adminAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		p.handleAdminRoutesNetHTTP(w, r, normalizedPrefix)
	}))
}

func (p *Proxy) handleAdminRoutesNetHTTP(w http.ResponseWriter, r *http.Request, prefix string) {

	_, trimPrefix := normalizePrefix(prefix)
	path := strings.TrimPrefix(r.URL.Path, trimPrefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] != "projects" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if len(parts) == 1 {
		p.handleProjectsNetHTTP(w, r)
		return
	}

	if len(parts) >= 2 {
		alias := parts[1]
		p.handleProjectNetHTTP(w, r, alias)
		return
	}

	http.Error(w, "Not found", http.StatusNotFound)
}

// SetPublicRoutesFiber регистрирует публичные роуты в go-fiber
// Публичные роуты: получение манифестов, файлов, списка версий
// app - Fiber приложение для регистрации роутов
// prefix - префикс для всех роутов (например, "/api/v1/proxy")
func (p *Proxy) SetPublicRoutesFiber(app *fiber.App, prefix string) {

	group := app.Group(prefix, p.publicFiberAuthMiddleware)
	group.Get("/:alias/:version/manifest.yml", p.handleGetManifestFiber)
	group.Get("/:alias/:version/*", p.handleGetFileFiber)
	group.Get("/:alias/versions", p.handleGetVersionsFiber)
}

// SetAdminRoutesFiber регистрирует админские роуты в go-fiber
// Админские роуты: управление проектами (CRUD операции)
// app - Fiber приложение для регистрации роутов
// prefix - префикс для всех роутов (например, "/api/v1/admin")
func (p *Proxy) SetAdminRoutesFiber(app *fiber.App, prefix string) {

	group := app.Group(prefix, p.adminFiberAuthMiddleware)
	group.Get("/projects", p.handleListProjectsFiber)
	group.Post("/projects", p.handleCreateProjectFiber)
	group.Get("/projects/:alias", p.handleGetProjectFiber)
	group.Put("/projects/:alias", p.handleUpdateProjectFiber)
	group.Delete("/projects/:alias", p.handleDeleteProjectFiber)
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

func (p *Proxy) handleProjectsNetHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		p.handleListProjectsNetHTTP(w, r)
		return
	}
	if r.Method == http.MethodPost {
		p.handleCreateProjectNetHTTP(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (p *Proxy) handleProjectNetHTTP(w http.ResponseWriter, r *http.Request, alias string) {

	if r.Method == http.MethodGet {
		p.handleGetProjectNetHTTP(w, r, alias)
		return
	}
	if r.Method == http.MethodPut {
		p.handleUpdateProjectNetHTTP(w, r, alias)
		return
	}
	if r.Method == http.MethodDelete {
		p.handleDeleteProjectNetHTTP(w, r, alias)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

func (p *Proxy) handleGetManifestFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")
	version := c.Params("version")

	manifest, statusCode, err := p.handleGetManifest(c.Context(), alias, version)
	if err != nil {
		slog.Error("Failed to get manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	slog.Info("Manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	c.Set("Content-Type", "application/x-yaml")
	return c.Status(statusCode).Send(manifest)
}

func (p *Proxy) handleGetFileFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")
	version := c.Params("version")
	filename := strings.TrimPrefix(c.Params("*"), "/")

	var project domain.Project
	var found bool
	if project, found, err = p.engine.GetProject(c.Context(), alias); err != nil {
		slog.Error("Failed to get project for file",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}
	if !found {
		slog.Debug("Project not found during file streaming",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var src core.Source
	if src, err = p.engine.GetSource(project.SourceName); err != nil {
		slog.Error("Source not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Source not found",
		})
	}

	slog.Debug("Fetching file from source",
		slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeyFilename, filename),
		slog.String(helpers.LogKeySource, project.SourceName),
	)

	var resp *http.Response
	if resp, err = src.GetFileResponse(c.Context(), project, version, filename); err != nil {
		slog.Error("Failed to fetch file from source",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "Failed to fetch file",
		})
	}
	defer resp.Body.Close()

	slog.Info("File request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeyFilename, filename),
		slog.Int(helpers.LogKeyStatusCode, resp.StatusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.String("content_type", resp.Header.Get("Content-Type")),
		slog.String("content_length", resp.Header.Get("Content-Length")),
	)

	p.copyResponseHeaders(c, resp)

	_, err = io.Copy(c.Response().BodyWriter(), resp.Body)
	return
}

func (p *Proxy) handleGetVersionsFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")

	versions, statusCode, err := p.handleGetVersions(c.Context(), alias)
	if err != nil {
		slog.Error("Failed to get versions",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	slog.Info("Versions request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int(helpers.LogKeyVersionsCount, len(versions)),
	)

	return c.Status(statusCode).JSON(versions)
}

func (p *Proxy) handleListProjectsFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, parseErr := strconv.Atoi(limitStr); parseErr == nil {
			limit = parsedLimit
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, parseErr := strconv.Atoi(offsetStr); parseErr == nil {
			offset = parsedOffset
		}
	}

	projects, total, statusCode, err := p.handleListProjects(c.Context(), limit, offset)
	if err != nil {
		slog.Error("Failed to list projects",
			slog.String(helpers.LogKeyAction, helpers.ActionListProjects),
			slog.Int(helpers.LogKeyLimit, limit),
			slog.Int(helpers.LogKeyOffset, offset),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	slog.Info("List projects request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionListProjects),
		slog.Int(helpers.LogKeyLimit, limit),
		slog.Int(helpers.LogKeyOffset, offset),
		slog.Int64(helpers.LogKeyTotal, total),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("returned_count", len(projects)),
	)

	return c.Status(statusCode).JSON(fiber.Map{
		"projects": projects,
		"total":    total,
	})
}

func (p *Proxy) handleCreateProjectFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()

	var req dto.ProjectCreateRequest
	if err = c.BodyParser(&req); err != nil {
		slog.Debug("Invalid request body",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err = helpers.ValidateStruct(&req); err != nil {
		slog.Debug("Validation failed",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyAlias, req.Alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	statusCode, err := p.handleCreateProject(c.Context(), req)
	if err != nil {
		slog.Error("Failed to create project",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyAlias, req.Alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
		slog.String(helpers.LogKeyAlias, req.Alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	}
	if req.SourceName != "" {
		args = append(args, slog.String(helpers.LogKeySource, req.SourceName))
	}
	if req.RepoURL != "" {
		args = append(args, slog.String(helpers.LogKeyRepoURL, req.RepoURL))
	}
	slog.Info("Create project request completed", args...)

	return c.SendStatus(statusCode)
}

func (p *Proxy) handleGetProjectFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")

	if err = helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	project, found, statusCode, err := p.handleGetProject(c.Context(), alias)
	if err != nil {
		slog.Error("Failed to get project",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}
	if !found {
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
		)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionGetProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	}
	if project.SourceName != "" {
		args = append(args, slog.String(helpers.LogKeySource, project.SourceName))
	}
	if project.RepoURL != "" {
		args = append(args, slog.String(helpers.LogKeyRepoURL, project.RepoURL))
	}
	slog.Info("Get project request completed", args...)

	return c.Status(statusCode).JSON(project)
}

func (p *Proxy) handleUpdateProjectFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")

	if err = helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var req dto.ProjectUpdateRequest
	if err = c.BodyParser(&req); err != nil {
		slog.Debug("Invalid request body",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err = helpers.ValidateStruct(&req); err != nil {
		slog.Debug("Validation failed",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	statusCode, err := p.handleUpdateProject(c.Context(), alias, req)
	if err != nil {
		slog.Error("Failed to update project",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	args := []any{
		slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
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

	return c.SendStatus(statusCode)
}

func (p *Proxy) handleDeleteProjectFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	alias := c.Params("alias")

	if err = helpers.ValidateAlias(alias); err != nil {
		slog.Debug("Invalid alias",
			slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	statusCode, err := p.handleDeleteProject(c.Context(), alias)
	if err != nil {
		slog.Error("Failed to delete project",
			slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyStatusCode, statusCode),
			slog.String(helpers.LogKeyMethod, c.Method()),
			slog.String(helpers.LogKeyPath, c.Path()),
			slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
			slog.Any(helpers.LogKeyError, err),
		)
		return c.Status(statusCode).JSON(fiber.Map{
			"error": helpers.GetErrorMessage(err),
		})
	}

	slog.Info("Delete project request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
		slog.String(helpers.LogKeyAlias, alias),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	)

	return c.SendStatus(statusCode)
}

func (p *Proxy) copyResponseHeaders(c *fiber.Ctx, resp *http.Response) {

	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		c.Set("Content-Type", contentType)
	} else {
		c.Set("Content-Type", "application/octet-stream")
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Set("Content-Length", contentLength)
	}

	c.Set("Cache-Control", "public, max-age=3600")
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
