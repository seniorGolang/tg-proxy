package core

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

// Dependency представляет распарсенную зависимость
type Dependency struct {
	Source  string
	Package string
	Version string
}

type transformer struct {
	storage storage
}

func newTransformer(stor storage) (t *transformer) {
	return &transformer{
		storage: stor,
	}
}

// ExtractSourceDomain извлекает домен источника (схема + хост + порт) из RepoURL
func ExtractSourceDomain(repoURL string) (sourceDomain string, err error) {

	if repoURL == "" {
		return "", nil
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(repoURL); err != nil {
		return "", fmt.Errorf("failed to parse repo URL: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return "", nil
	}

	sourceDomain = parsedURL.Scheme + "://" + parsedURL.Host
	return
}

// isSameDomain проверяет, принадлежит ли URL тому же домену источника
// Сравнивает схему, хост и порт
func isSameDomain(urlStr string, sourceDomain string) (same bool) {

	if urlStr == "" || sourceDomain == "" {
		return false
	}

	var parsedURL *url.URL
	var err error
	if parsedURL, err = url.Parse(urlStr); err != nil {
		return false
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	var parsedSourceDomain *url.URL
	if parsedSourceDomain, err = url.Parse(sourceDomain); err != nil {
		return false
	}

	// Сравниваем схему, хост и порт
	same = parsedURL.Scheme == parsedSourceDomain.Scheme &&
		parsedURL.Host == parsedSourceDomain.Host

	return
}

func (t *transformer) Transform(ctx context.Context, manifest *model.Manifest, alias string, version string, baseURL string, sourceDomain string) (transformed []byte, err error) {

	if err = t.replaceManifestURLs(ctx, manifest, alias, version, baseURL, sourceDomain); err != nil {
		return
	}

	if transformed, err = yaml.Marshal(manifest); err != nil {
		transformed = nil
		err = fmt.Errorf("%w: %w", errs.ErrManifestMarshalError, err)
		return
	}

	return
}

func (t *transformer) replaceManifestURLs(ctx context.Context, manifest *model.Manifest, alias string, version string, baseURL string, sourceDomain string) (err error) {

	for i := range manifest.Manifests {
		if manifest.Manifests[i].URL, err = t.replaceURL(manifest.Manifests[i].URL, alias, version, baseURL, sourceDomain); err != nil {
			return
		}
	}

	for i := range manifest.Packages {
		for j := range manifest.Packages[i].Downloads {
			if manifest.Packages[i].Downloads[j].URL, err = t.replaceURL(
				manifest.Packages[i].Downloads[j].URL,
				alias,
				version,
				baseURL,
				sourceDomain,
			); err != nil {
				return
			}
		}

		if err = t.replaceScriptURLs(manifest.Packages[i].Scripts, alias, version, baseURL, sourceDomain); err != nil {
			return
		}

		if err = t.replaceDependencies(ctx, &manifest.Packages[i], baseURL); err != nil {
			return
		}
	}

	return
}

func (t *transformer) replaceScriptURLs(scripts *model.Scripts, alias string, version string, baseURL string, sourceDomain string) (err error) {

	if scripts == nil {
		return
	}

	scriptsToReplace := []*model.ScriptAction{
		scripts.PreInstall,
		scripts.PostInstall,
		scripts.PreUninstall,
		scripts.PostUninstall,
	}

	for _, script := range scriptsToReplace {
		if script != nil {
			if script.Source, err = t.replaceURL(script.Source, alias, version, baseURL, sourceDomain); err != nil {
				return
			}
		}
	}

	return
}

// parseDependencyString парсит строку зависимости в структуру Dependency.
// Поддерживаемые форматы:
// - package - зависимость без версии
// - package@version - зависимость с версией
// - source:package - зависимость из другого репозитория без версии
// - source:package@version - зависимость из другого репозитория с версией
// - URL:package@version - зависимость из URL репозитория
func parseDependencyString(depStr string) (dep Dependency) {

	depStr = strings.TrimSpace(depStr)
	if depStr == "" {
		return
	}

	// Шаг 1: Разделяем по "@" для извлечения версии
	parts := strings.Split(depStr, "@")
	specWithoutVersion := parts[0]
	if len(parts) == 2 {
		dep.Version = parts[1]
	}

	// Шаг 2: Проверяем наличие ":" для разделения source и package
	colonIndex := strings.LastIndex(specWithoutVersion, ":")
	if colonIndex > 0 {
		// Проверяем, не является ли ":" частью схемы URL (://)
		schemeEndIndex := strings.Index(specWithoutVersion, "://")
		if schemeEndIndex < 0 || colonIndex > schemeEndIndex+2 {
			// Это не часть схемы, значит ":" разделяет source и package
			beforeColon := specWithoutVersion[:colonIndex]
			afterColon := specWithoutVersion[colonIndex+1:]

			// Проверяем, является ли часть до ":" валидным URL
			testURL := beforeColon
			if !strings.Contains(testURL, "://") {
				// Пробуем добавить схему для проверки
				testURL = "https://" + testURL
			}
			testParsedURL, testErr := url.Parse(testURL)
			if testErr == nil && testParsedURL.Scheme != "" {
				// Это валидный URL, значит это source
				// Восстанавливаем оригинальный URL без добавленной схемы
				if !strings.Contains(beforeColon, "://") {
					// Если не было схемы, используем https://
					dep.Source = "https://" + beforeColon
				} else {
					dep.Source = beforeColon
				}
				dep.Package = afterColon
				return
			}
		}
	}

	// Если не нашли source через ":", значит это просто package
	dep.Package = specWithoutVersion
	return
}

func (t *transformer) replaceDependencies(ctx context.Context, pkg *model.Package, baseURL string) (err error) {

	if len(pkg.Dependencies) == 0 {
		return
	}

	for i := range pkg.Dependencies {
		if pkg.Dependencies[i], err = t.replaceDependencyURL(ctx, pkg.Dependencies[i], baseURL); err != nil {
			return
		}
	}

	return
}

func (t *transformer) replaceDependencyURL(ctx context.Context, depStr string, baseURL string) (replaced string, err error) {

	if depStr == "" || baseURL == "" {
		replaced = depStr
		return
	}

	dep := parseDependencyString(depStr)

	// Если Source пустой или не является валидным URL, оставляем как есть
	if dep.Source == "" || !strings.Contains(dep.Source, "://") {
		replaced = depStr
		return
	}

	// Нормализуем URL источника (удаляем .git если есть)
	normalizedSource := helpers.NormalizeRepoURL(dep.Source)

	// Ищем проект по нормализованному URL
	var project domain.Project
	var found bool
	if project, found, err = t.storage.GetProjectByRepoURL(ctx, normalizedSource); err != nil {
		return
	}

	// Если проект не найден, оставляем зависимость без изменений
	if !found {
		replaced = depStr
		return
	}

	// Формируем новый URL зависимости: baseURL/alias:package@version
	var baseParsedURL *url.URL
	if baseParsedURL, err = url.Parse(baseURL); err != nil {
		replaced = depStr
		return
	}

	var depSpec string
	if dep.Version != "" {
		depSpec = fmt.Sprintf("%s:%s@%s", project.Alias, dep.Package, dep.Version)
	} else {
		depSpec = fmt.Sprintf("%s:%s", project.Alias, dep.Package)
	}

	// Объединяем baseURL с зависимостью, правильно обрабатывая слэши
	basePath := strings.TrimSuffix(baseParsedURL.Path, "/")
	newPath := path.Join("/", basePath, depSpec)
	baseParsedURL.Path = newPath

	replaced = baseParsedURL.String()
	return
}

func (t *transformer) replaceURL(originalURL string, alias string, version string, baseURL string, sourceDomain string) (replaced string, err error) {

	if baseURL == "" || originalURL == "" {
		replaced = originalURL
		return
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(originalURL); err != nil {
		replaced = originalURL
		return
	}

	// Используем strings.Index для эффективного поиска подстроки
	genericPrefix := "/packages/generic/"
	genericIndex := strings.Index(parsedURL.Path, genericPrefix)

	// Если sourceDomain задан, проверяем принадлежность URL домену источника
	if sourceDomain != "" {
		if !isSameDomain(originalURL, sourceDomain) {
			// URL не принадлежит домену источника - оставляем без изменений
			// (файлы будут скачиваться напрямую, мимо прокси)
			replaced = originalURL
			return
		}

		// URL принадлежит домену источника
		if genericIndex == -1 {
			// Если URL принадлежит домену источника, но не содержит /packages/generic/,
			// заменяем только схему и хост, сохраняя путь
			var newURL *url.URL
			if newURL, err = url.Parse(baseURL); err != nil {
				replaced = originalURL
				return
			}
			newURL.Path = parsedURL.Path
			newURL.RawPath = parsedURL.RawPath
			newURL.RawQuery = parsedURL.RawQuery
			newURL.Fragment = parsedURL.Fragment
			replaced = newURL.String()
			return
		}
		// Продолжаем обработку URL с /packages/generic/
	} else {
		// Если sourceDomain не задан, используем старую логику
		// (для обратной совместимости)
		if genericIndex == -1 {
			replaced = originalURL
			return
		}
	}

	// Извлекаем путь после "/packages/generic/"
	pathAfterGeneric := parsedURL.Path[genericIndex+len(genericPrefix):]
	if pathAfterGeneric == "" {
		replaced = originalURL
		return
	}

	// Разбиваем путь на части для проверки версии и получения имени файла
	pathParts := strings.Split(strings.Trim(pathAfterGeneric, "/"), "/")
	if len(pathParts) < 2 {
		replaced = originalURL
		return
	}

	// Определяем формат URL и проверяем версию
	// GitLab формат: /packages/generic/release/version/filename
	// GitHub формат: /packages/generic/version/filename
	if pathParts[0] != version {
		// Если первая часть не совпадает с версией, это может быть package_name (GitLab формат)
		// Проверяем вторую часть
		if len(pathParts) < 3 || pathParts[1] != version {
			err = fmt.Errorf("%w: expected %s, got %s in URL %s", errs.ErrVersionMismatch, version, pathParts[0], originalURL)
			return
		}
		// Версия найдена на позиции 1 (GitLab формат)
	}

	// Получаем имя файла (последняя часть пути)
	filename := pathParts[len(pathParts)-1]
	if filename == "" {
		replaced = originalURL
		return
	}

	// Используем path.Join для безопасной сборки нового пути
	newPath := path.Join("/", alias, version, filename)

	// Используем url.URL для правильной сборки нового URL с query параметрами
	var newURL *url.URL
	if newURL, err = url.Parse(baseURL); err != nil {
		replaced = originalURL
		return
	}

	newURL.Path = newPath
	newURL.RawQuery = parsedURL.RawQuery
	replaced = newURL.String()

	return
}
