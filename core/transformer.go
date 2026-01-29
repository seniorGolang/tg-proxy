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

	same = parsedURL.Scheme == parsedSourceDomain.Scheme &&
		parsedURL.Host == parsedSourceDomain.Host

	return
}

func (t *transformer) Transform(ctx context.Context, manifest *model.Manifest, alias string, version string, baseURL string, sourceDomain string, resolver Source) (transformed []byte, err error) {

	if err = t.ReplaceManifestURLs(ctx, manifest, alias, version, baseURL, sourceDomain, resolver); err != nil {
		return
	}

	if transformed, err = yaml.Marshal(manifest); err != nil {
		transformed = nil
		err = fmt.Errorf("%w: %w", errs.ErrManifestMarshalError, err)
		return
	}

	return
}

// ReplaceManifestURLs заменяет URL в манифесте на прокси (модифицирует manifest на месте).
func (t *transformer) ReplaceManifestURLs(ctx context.Context, manifest *model.Manifest, alias string, version string, baseURL string, sourceDomain string, resolver Source) (err error) {

	return t.replaceManifestURLs(ctx, manifest, alias, version, baseURL, sourceDomain, resolver)
}

func (t *transformer) replaceManifestURLs(ctx context.Context, manifest *model.Manifest, alias string, version string, baseURL string, sourceDomain string, resolver Source) (err error) {

	for i := range manifest.Manifests {
		if manifest.Manifests[i].URL, err = t.replaceURL(manifest.Manifests[i].URL, alias, version, baseURL, sourceDomain, resolver); err != nil {
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
				resolver,
			); err != nil {
				return
			}
		}

		if err = t.replaceScriptURLs(manifest.Packages[i].Scripts, alias, version, baseURL, sourceDomain, resolver); err != nil {
			return
		}

		if err = t.replaceDependencies(ctx, &manifest.Packages[i], baseURL); err != nil {
			return
		}
	}

	return
}

func (t *transformer) replaceScriptURLs(scripts *model.Scripts, alias string, version string, baseURL string, sourceDomain string, resolver Source) (err error) {

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
			if script.Source, err = t.replaceURL(script.Source, alias, version, baseURL, sourceDomain, resolver); err != nil {
				return
			}
		}
	}

	return
}

// parseDependencyString: package | package@version | source:package | source:package@version | URL:package@version
func parseDependencyString(depStr string) (dep Dependency) {

	depStr = strings.TrimSpace(depStr)
	if depStr == "" {
		return
	}

	parts := strings.Split(depStr, "@")
	specWithoutVersion := parts[0]
	if len(parts) == 2 {
		dep.Version = parts[1]
	}

	colonIndex := strings.LastIndex(specWithoutVersion, ":")
	if colonIndex > 0 {
		schemeEndIndex := strings.Index(specWithoutVersion, "://")
		if schemeEndIndex < 0 || colonIndex > schemeEndIndex+2 {
			beforeColon := specWithoutVersion[:colonIndex]
			afterColon := specWithoutVersion[colonIndex+1:]

			testURL := beforeColon
			if !strings.Contains(testURL, "://") {
				testURL = "https://" + testURL
			}
			testParsedURL, testErr := url.Parse(testURL)
			if testErr == nil && testParsedURL.Scheme != "" {
				if !strings.Contains(beforeColon, "://") {
					dep.Source = "https://" + beforeColon
				} else {
					dep.Source = beforeColon
				}
				dep.Package = afterColon
				return
			}
		}
	}

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

	if dep.Source == "" || !strings.Contains(dep.Source, "://") {
		replaced = depStr
		return
	}

	normalizedSource := helpers.NormalizeRepoURL(dep.Source)

	var project domain.Project
	var found bool
	if project, found, err = t.storage.GetProjectByRepoURL(ctx, normalizedSource); err != nil {
		return
	}

	if !found {
		replaced = depStr
		return
	}

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

	basePath := strings.TrimSuffix(baseParsedURL.Path, "/")
	newPath := path.Join("/", basePath, depSpec)
	baseParsedURL.Path = newPath

	replaced = baseParsedURL.String()
	return
}

func (t *transformer) replaceURL(originalURL string, alias string, version string, baseURL string, sourceDomain string, resolver Source) (replaced string, err error) {

	if baseURL == "" || originalURL == "" {
		replaced = originalURL
		return
	}

	if sourceDomain == "" || !isSameDomain(originalURL, sourceDomain) {
		replaced = originalURL
		return
	}

	if resolver == nil {
		replaced = originalURL
		return
	}

	parsedVersion, filename, ok := resolver.ParseFileURL(originalURL)
	if !ok {
		replaced = originalURL
		return
	}

	if parsedVersion != version {
		err = fmt.Errorf("%w: expected %s, got %s in URL %s", errs.ErrVersionMismatch, version, parsedVersion, originalURL)
		return
	}

	var parsedURL *url.URL
	if parsedURL, err = url.Parse(originalURL); err != nil {
		replaced = originalURL
		return
	}

	replaced = helpers.BuildURL(baseURL, alias, version, filename)
	if parsedURL.RawQuery != "" {
		replaced = replaced + "?" + parsedURL.RawQuery
	}

	return
}
