package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

type engine struct {
	storage     storage
	encryptor   encryptor
	cache       cache
	sources     map[string]Source
	sourcesMu   sync.RWMutex
	resolver    *resolver
	transformer *transformer
}

// EngineOption - функция опции для настройки Engine
type EngineOption func(*engine)

// Storage устанавливает хранилище проектов
func Storage(stor storage) (opt EngineOption) {
	return func(e *engine) {
		e.storage = stor
	}
}

// Encryptor устанавливает шифратор для токенов
func Encryptor(enc encryptor) (opt EngineOption) {
	return func(e *engine) {
		e.encryptor = enc
	}
}

// Cache устанавливает кеш
func Cache(c cache) (opt EngineOption) {
	return func(e *engine) {
		e.cache = c
	}
}

// NewEngine создает новый Engine с использованием паттерна опций
// opts - опции для настройки Engine (Storage, Encryptor, Cache)
func NewEngine(opts ...EngineOption) (eng *engine) {

	e := &engine{
		sources: make(map[string]Source),
	}

	for _, opt := range opts {
		opt(e)
	}

	e.resolver = newResolver(e.storage, e.encryptor, e.cache)
	e.transformer = newTransformer(e.storage)

	return e
}

func (e *engine) RegisterSource(src Source) (err error) {

	e.sourcesMu.Lock()
	defer e.sourcesMu.Unlock()

	name := src.Name()

	if _, exists := e.sources[name]; exists {
		slog.Debug("Source already registered",
			slog.String(helpers.LogKeySource, name),
		)
		return fmt.Errorf("%w: %s", errs.ErrSourceAlreadyRegistered, name)
	}

	e.sources[name] = src

	args := []any{
		slog.String(helpers.LogKeySource, name),
	}

	if sourceInfo, ok := src.(SourceInfo); ok {
		if baseURL := sourceInfo.BaseURL(); baseURL != "" {
			args = append(args, slog.String(helpers.LogKeySourceURL, baseURL))
		}
	}

	slog.Info("Source registered", args...)

	return
}

func (e *engine) GetSource(name string) (src Source, err error) {

	e.sourcesMu.RLock()
	defer e.sourcesMu.RUnlock()

	var exists bool
	if src, exists = e.sources[name]; !exists {
		slog.Debug("Source not found",
			slog.String(helpers.LogKeySource, name),
		)
		return nil, fmt.Errorf("%w: %s", errs.ErrSourceNotFound, name)
	}

	return
}

func (e *engine) GetManifest(ctx context.Context, alias string, version string, baseURL string) (manifest []byte, err error) {

	var found bool
	var project domain.Project
	if project, found, err = e.resolver.ResolveProject(ctx, alias); err != nil {
		slog.Debug("Failed to resolve project",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	if !found {
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
		)
		err = errs.ErrProjectNotFound
		return
	}

	slog.Debug("Project resolved",
		slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeyVersion, version),
		slog.String(helpers.LogKeySource, project.SourceName),
		slog.String(helpers.LogKeyRepoURL, project.RepoURL),
	)

	var availableVersions []string
	if availableVersions, err = e.GetVersions(ctx, alias); err != nil {
		slog.Debug("Failed to get versions",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	var versionExists bool
	for _, v := range availableVersions {
		if v == version {
			versionExists = true
			break
		}
	}
	if !versionExists {
		slog.Debug("Version not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Int(helpers.LogKeyVersionsCount, len(availableVersions)),
		)
		err = errs.ErrVersionNotFound
		return
	}

	var src Source
	if src, err = e.GetSource(project.SourceName); err != nil {
		slog.Debug("Source not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	var domainManifest domain.Manifest
	if domainManifest, err = src.GetManifest(ctx, project, version); err != nil {
		// Если GitLab/GitHub API вернул 404, это означает, что манифест не найден
		// Преобразуем в ErrVersionNotFound, так как версия уже была проверена в списке версий
		if statusCode, found := helpers.ExtractStatusCode(err); found && statusCode == 404 {
			slog.Debug("Manifest not found (404), treating as version not found",
				slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
				slog.String(helpers.LogKeyAlias, alias),
				slog.String(helpers.LogKeyVersion, version),
				slog.String(helpers.LogKeySource, project.SourceName),
			)
			err = errs.ErrVersionNotFound
			return
		}
		slog.Debug("Failed to get manifest from source",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	var modelManifest model.Manifest
	modelManifest.FromDomain(domainManifest)

	// Извлекаем домен источника из RepoURL для проверки принадлежности URL
	var sourceDomain string
	if sourceDomain, err = ExtractSourceDomain(project.RepoURL); err != nil {
		slog.Debug("Failed to extract source domain, continuing without domain check",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyRepoURL, project.RepoURL),
			slog.Any(helpers.LogKeyError, err),
		)
		sourceDomain = ""
	}

	if manifest, err = e.transformer.Transform(ctx, &modelManifest, alias, version, baseURL, sourceDomain); err != nil {
		slog.Debug("Failed to transform manifest",
			slog.String(helpers.LogKeyAction, helpers.ActionGetManifest),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	return
}

func (e *engine) GetFile(ctx context.Context, alias string, version string, filename string) (stream io.ReadCloser, err error) {

	var found bool
	var project domain.Project
	if project, found, err = e.resolver.ResolveProject(ctx, alias); err != nil {
		slog.Debug("Failed to resolve project",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	if !found {
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
		)
		err = errs.ErrProjectNotFound
		return
	}

	var availableVersions []string
	if availableVersions, err = e.GetVersions(ctx, alias); err != nil {
		slog.Debug("Failed to get versions",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	var versionExists bool
	for _, v := range availableVersions {
		if v == version {
			versionExists = true
			break
		}
	}
	if !versionExists {
		slog.Debug("Version not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
		)
		err = errs.ErrVersionNotFound
		return
	}

	var src Source
	if src, err = e.GetSource(project.SourceName); err != nil {
		slog.Debug("Source not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	if stream, err = src.GetFileStream(ctx, project, version, filename); err != nil {
		// Если GitLab/GitHub API вернул 404, это означает, что файл не найден
		// Преобразуем в ErrVersionNotFound, так как версия уже была проверена в списке версий
		if statusCode, found := helpers.ExtractStatusCode(err); found && statusCode == 404 {
			slog.Debug("File not found (404), treating as version not found",
				slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
				slog.String(helpers.LogKeyAlias, alias),
				slog.String(helpers.LogKeyVersion, version),
				slog.String(helpers.LogKeyFilename, filename),
				slog.String(helpers.LogKeySource, project.SourceName),
			)
			err = errs.ErrVersionNotFound
			return
		}
		slog.Debug("Failed to get file stream from source",
			slog.String(helpers.LogKeyAction, helpers.ActionGetFile),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeyVersion, version),
			slog.String(helpers.LogKeyFilename, filename),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	return
}

func (e *engine) GetVersions(ctx context.Context, alias string) (versions []string, err error) {

	var project domain.Project
	var found bool
	if project, found, err = e.resolver.ResolveProject(ctx, alias); err != nil {
		slog.Debug("Failed to resolve project",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	if !found {
		slog.Debug("Project not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
		)
		err = errs.ErrProjectNotFound
		return
	}

	var cachedVersions []string
	var cachedFound bool
	if cachedVersions, cachedFound, err = e.cache.GetVersions(ctx, alias); err == nil && cachedFound {
		slog.Debug("Versions retrieved from cache",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Int(helpers.LogKeyVersionsCount, len(cachedVersions)),
		)
		versions = cachedVersions
		return
	}

	slog.Debug("Cache miss, fetching from source",
		slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
		slog.String(helpers.LogKeyAlias, alias),
		slog.String(helpers.LogKeySource, project.SourceName),
	)

	var src Source
	if src, err = e.GetSource(project.SourceName); err != nil {
		slog.Debug("Source not found",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	if versions, err = src.GetVersions(ctx, project); err != nil {
		// Если GitLab/GitHub API вернул 404, это означает, что пакеты не найдены
		// В этом случае возвращаем пустой список версий вместо ошибки
		if statusCode, found := helpers.ExtractStatusCode(err); found && statusCode == 404 {
			slog.Debug("No packages found (404), returning empty versions list",
				slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
				slog.String(helpers.LogKeyAlias, alias),
				slog.String(helpers.LogKeySource, project.SourceName),
			)
			versions = []string{}
			err = nil
			return
		}
		slog.Debug("Failed to get versions from source",
			slog.String(helpers.LogKeyAction, helpers.ActionGetVersions),
			slog.String(helpers.LogKeyAlias, alias),
			slog.String(helpers.LogKeySource, project.SourceName),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	sort.Sort(sort.Reverse(sort.StringSlice(versions)))

	_ = e.cache.SetVersions(ctx, alias, versions, 5*time.Minute)

	return
}

func (e *engine) CreateProject(ctx context.Context, project domain.Project) (err error) {

	project.RepoURL = helpers.NormalizeRepoURL(project.RepoURL)

	if project.Token != "" && e.encryptor != nil {
		var encryptedToken string
		if encryptedToken, err = e.encryptor.EncryptString(project.Token); err != nil {
			slog.Debug("Failed to encrypt token",
				slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
				slog.String(helpers.LogKeyAlias, project.Alias),
				slog.Any(helpers.LogKeyError, err),
			)
			return
		}
		project.EncryptedToken = encryptedToken
		project.Token = ""
	}

	if err = e.storage.CreateProject(ctx, project); err != nil {
		slog.Debug("Failed to create project in storage",
			slog.String(helpers.LogKeyAction, helpers.ActionCreateProject),
			slog.String(helpers.LogKeyAlias, project.Alias),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	_ = e.resolver.InvalidateCache(ctx, project.Alias)

	return
}

func (e *engine) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	return e.resolver.ResolveProject(ctx, alias)
}

func (e *engine) UpdateProject(ctx context.Context, alias string, project domain.Project) (err error) {

	project.Alias = alias
	project.RepoURL = helpers.NormalizeRepoURL(project.RepoURL)

	if project.Token != "" && e.encryptor != nil {
		var encryptedToken string
		if encryptedToken, err = e.encryptor.EncryptString(project.Token); err != nil {
			slog.Debug("Failed to encrypt token",
				slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
				slog.String(helpers.LogKeyAlias, alias),
				slog.Any(helpers.LogKeyError, err),
			)
			return
		}
		project.EncryptedToken = encryptedToken
		project.Token = ""
	}

	if err = e.storage.UpdateProject(ctx, alias, project); err != nil {
		slog.Debug("Failed to update project in storage",
			slog.String(helpers.LogKeyAction, helpers.ActionUpdateProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	_ = e.resolver.InvalidateCache(ctx, alias)

	return
}

func (e *engine) DeleteProject(ctx context.Context, alias string) (err error) {

	if err = e.storage.DeleteProject(ctx, alias); err != nil {
		slog.Debug("Failed to delete project from storage",
			slog.String(helpers.LogKeyAction, helpers.ActionDeleteProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	_ = e.cache.DeleteProject(ctx, alias)

	return
}

func (e *engine) ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error) {

	return e.storage.ListProjects(ctx, limit, offset)
}
