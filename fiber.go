package tgproxy

import (
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

func (p *Proxy) SetPublicRoutesFiber(app *fiber.App, prefix string) {

	base := strings.TrimSuffix(prefix, "/")
	if base == "" {
		base = "/"
	}
	p.publicPrefix = base
	group := app.Group(prefix, p.publicFiberAuthMiddleware)
	group.Get("/", p.handleGetAggregateManifestFiber)
	group.Get("/manifest.yml", p.handleGetAggregateManifestFiber)
	group.Get("/versions", p.handleGetCatalogVersionFiber)
	group.Get("/:version/manifest.yml", p.handleGetAggregateManifestAtVersionFiber)
	group.Get("/:alias/:version/manifest.yml", p.handleGetManifestFiber)
	group.Get("/:alias/versions", p.handleGetVersionsFiber)
	group.Get("/:alias/:version/*", p.handleGetFileFiber)
}

func (p *Proxy) SetAdminRoutesFiber(app *fiber.App, prefix string) {

	group := app.Group(prefix, p.adminFiberAuthMiddleware)
	group.Get("/projects", p.handleListProjectsFiber)
	group.Post("/projects", p.handleCreateProjectFiber)
	group.Get("/projects/:alias", p.handleGetProjectFiber)
	group.Put("/projects/:alias", p.handleUpdateProjectFiber)
	group.Delete("/projects/:alias", p.handleDeleteProjectFiber)
	group.Get("/projects/:alias/versions/:version/manifest", p.handleGetManifestAdminFiber)
	group.Get("/projects/:alias/versions", p.handleGetProjectVersionsAdminFiber)
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

func (p *Proxy) handleGetAggregateManifestFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()

	manifest, statusCode, err := p.handleGetAggregateManifest(c.Context())
	if err != nil {
		slog.Error("Failed to get aggregate manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
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

	slog.Info("Aggregate manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	c.Set("Content-Type", "application/x-yaml")
	return c.Status(statusCode).Send(manifest)
}

func (p *Proxy) handleGetAggregateManifestAtVersionFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()
	requestedVersion := c.Params("version")

	currentVersion, statusCode, err := p.handleGetCatalogVersion(c.Context())
	if err != nil {
		slog.Error("Failed to get catalog version",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
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
	if requestedVersion != currentVersion {
		return c.Status(fiber.StatusNotFound).SendString("Not found")
	}

	manifest, statusCode, err := p.handleGetAggregateManifest(c.Context())
	if err != nil {
		slog.Error("Failed to get aggregate manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
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

	slog.Info("Aggregate manifest request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetAggregateManifest),
		slog.String(helpers.LogKeyVersion, currentVersion),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("manifest_size", len(manifest)),
	)

	c.Set("Content-Type", "application/x-yaml")
	return c.Status(statusCode).Send(manifest)
}

func (p *Proxy) handleGetCatalogVersionFiber(c *fiber.Ctx) (err error) {

	startTime := time.Now()

	version, statusCode, err := p.handleGetCatalogVersion(c.Context())
	if err != nil {
		slog.Error("Failed to get catalog version",
			slog.String(helpers.LogKeyAction, helpers.ActionGetCatalogVersion),
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

	slog.Info("Catalog version request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetCatalogVersion),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	)

	return c.Status(statusCode).JSON([]string{version})
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

	statusCode, id, err := p.handleCreateProject(c.Context(), req)
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

	return c.Status(statusCode).JSON(fiber.Map{"id": id.String()})
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

func (p *Proxy) handleGetProjectVersionsAdminFiber(c *fiber.Ctx) (err error) {

	return p.handleGetVersionsFiber(c)
}

func (p *Proxy) handleGetManifestAdminFiber(c *fiber.Ctx) (err error) {

	alias := c.Params("alias")
	version := c.Params("version")
	if c.Query("aggregate") == "true" {
		return p.handleGetManifestAggregatedFiber(c, alias, version)
	}
	return p.handleGetManifestDataFiber(c, alias, version)
}

func (p *Proxy) handleGetManifestDataFiber(c *fiber.Ctx, alias string, version string) (err error) {

	startTime := time.Now()

	manifest, statusCode, err := p.handleGetManifestData(c.Context(), alias, version)
	if err != nil {
		slog.Error("Failed to get manifest data",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifestData),
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

	slog.Info("Manifest data request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifestData),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
	)

	return c.Status(statusCode).JSON(manifest)
}

func (p *Proxy) handleGetManifestAggregatedFiber(c *fiber.Ctx, alias string, version string) (err error) {

	startTime := time.Now()

	out, statusCode, err := p.handleGetManifestAggregated(c.Context(), alias, version)
	if err != nil {
		slog.Error("Failed to get manifest aggregated",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifestAggregated),
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

	slog.Info("Manifest aggregated request completed",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifestAggregated),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.Int(helpers.LogKeyStatusCode, statusCode),
		slog.String(helpers.LogKeyMethod, c.Method()),
		slog.String(helpers.LogKeyPath, c.Path()),
		slog.Duration(helpers.LogKeyDuration, time.Since(startTime)),
		slog.Int("packages_count", len(out.Packages)),
	)

	return c.Status(statusCode).JSON(out)
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
