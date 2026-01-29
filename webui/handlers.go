package webui

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/seniorGolang/tg-proxy/model"
)

type pageData struct {
	UIPrefix       string
	Theme          string
	InitialPath    string
	FaviconDataURI string
}

type projectsData struct {
	UIPrefix string
	Projects []projectItem
}

type projectItem struct {
	Alias string
}

type versionsData struct {
	UIPrefix string
	Alias    string
	Versions []string
}

type projectCollapseData struct {
	UIPrefix string
	Alias    string
}

type packagesData struct {
	UIPrefix           string
	Alias              string
	Version            string
	ManifestSourceURL  string
	ManifestInstallCmd string
	Packages           []packageItem
}

type packageItem struct {
	model.PackageWithSource
	Index          int
	SourceURL      string
	InstallCommand string
}

type packageCardData struct {
	model.PackageWithSource
	Alias          string
	Version        string
	SourceURL      string
	InstallCommand string
}

func themeFromRequest(r *http.Request) (theme string) {

	if c, err := r.Cookie("webui-theme"); err == nil && c.Value != "" {
		return c.Value
	}
	return "system"
}

func (ui *UI) handleIndex(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	base := ui.uiPrefix
	if base == "" {
		base = "/"
	}
	trimmed := strings.TrimSuffix(base, "/")
	ok := path == base || path == trimmed || path == trimmed+"/"
	if !ok && strings.HasPrefix(path, trimmed+"/") {
		rest := strings.TrimPrefix(path, trimmed+"/")
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) >= 1 && parts[0] != "" {
			ok = true
		}
		if len(parts) == 2 && parts[1] != "" {
			ok = true
		}
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	initialPath := ""
	if a := r.PathValue("alias"); a != "" {
		initialPath = a
		if v := r.PathValue("version"); v != "" {
			initialPath = a + "/" + v
		}
	} else if path != base && path != trimmed && path != trimmed+"/" {
		rest := strings.TrimPrefix(path, trimmed+"/")
		if rest == "" {
			rest = strings.TrimPrefix(path, trimmed)
		}
		initialPath = strings.TrimPrefix(rest, "/")
	}
	html, err := ui.renderIndex(themeFromRequest(r), initialPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

func (ui *UI) renderIndex(theme string, initialPath string) (html []byte, err error) {

	data := pageData{
		UIPrefix:       ui.uiPrefix,
		Theme:          theme,
		InitialPath:    initialPath,
		FaviconDataURI: faviconDataURI,
	}
	return ui.renderTemplate("layout", data)
}

func (ui *UI) renderTemplate(name string, data any) (html []byte, err error) {

	var buf strings.Builder
	if err = ui.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func (ui *UI) handleFragmentsProjects(w http.ResponseWriter, r *http.Request) {

	html, err := ui.serveProjects(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(html)
}

func (ui *UI) serveProjects(ctx context.Context) (html []byte, err error) {

	projects, _, err := ui.provider.ListProjects(ctx, 100, 0)
	if err != nil {
		return nil, err
	}
	items := make([]projectItem, len(projects))
	for i := range projects {
		items[i] = projectItem{Alias: projects[i].Alias}
	}
	data := projectsData{UIPrefix: ui.uiPrefix, Projects: items}
	return ui.renderTemplate("projects", data)
}

func (ui *UI) handleFragmentsVersions(w http.ResponseWriter, r *http.Request) {

	alias := r.PathValue("alias")
	if alias == "" {
		http.Error(w, "alias required", http.StatusBadRequest)
		return
	}
	html, err := ui.serveVersions(r.Context(), alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Push-Url", ui.pathForAlias(alias))
	_, _ = w.Write(html)
}

func (ui *UI) handleFragmentsProjectCollapse(w http.ResponseWriter, r *http.Request) {

	alias := r.PathValue("alias")
	if alias == "" {
		http.Error(w, "alias required", http.StatusBadRequest)
		return
	}
	html, err := ui.serveProjectCollapse(alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pushURL := ui.uiPrefix
	if pushURL == "" {
		pushURL = "/"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Push-Url", pushURL)
	_, _ = w.Write(html)
}

func (ui *UI) serveProjectCollapse(alias string) (html []byte, err error) {

	data := projectCollapseData{UIPrefix: ui.uiPrefix, Alias: alias}
	return ui.renderTemplate("project-collapsed", data)
}

func (ui *UI) serveVersions(ctx context.Context, alias string) (html []byte, err error) {

	versions, err := ui.provider.GetVersions(ctx, alias)
	if err != nil {
		return nil, err
	}
	data := versionsData{UIPrefix: ui.uiPrefix, Alias: alias, Versions: versions}
	return ui.renderTemplate("versions", data)
}

func (ui *UI) handleFragmentsPackages(w http.ResponseWriter, r *http.Request) {

	alias := r.PathValue("alias")
	version := r.PathValue("version")
	if alias == "" || version == "" {
		http.Error(w, "alias and version required", http.StatusBadRequest)
		return
	}
	mainHtml, _, err := ui.servePackagesWithTree(r.Context(), alias, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Push-Url", ui.pathForVersion(alias, version))
	_, _ = w.Write(mainHtml)
}

func (ui *UI) servePackagesWithTree(ctx context.Context, alias string, version string) (mainHtml []byte, oobHtml []byte, err error) {

	out, err := ui.provider.GetManifestAggregated(ctx, alias, version)
	if err != nil {
		return nil, nil, err
	}
	items := ui.buildPackageItems(out.Packages, alias, version)
	data := packagesData{
		UIPrefix:           ui.uiPrefix,
		Alias:              alias,
		Version:            version,
		ManifestSourceURL:  ui.manifestSourceURL(alias),
		ManifestInstallCmd: ui.manifestInstallCommand(alias, version),
		Packages:           items,
	}
	mainHtml, err = ui.renderTemplate("packages", data)
	if err != nil {
		return nil, nil, err
	}
	return mainHtml, nil, nil
}

func (ui *UI) buildPackageItems(packages []model.PackageWithSource, manifestAlias string, manifestVersion string) (items []packageItem) {

	items = make([]packageItem, len(packages))
	for i := range packages {
		p := &packages[i]
		items[i] = packageItem{
			PackageWithSource: *p,
			Index:             i,
			SourceURL:         ui.packageSourceURL(p.SourceAlias, p.SourceVersion, manifestAlias, manifestVersion),
			InstallCommand:    ui.packageInstallCommand(p.SourceAlias, p.SourceVersion, manifestAlias, manifestVersion, p.Name),
		}
	}
	return items
}

func (ui *UI) handleFragmentsPackage(w http.ResponseWriter, r *http.Request) {

	alias := r.URL.Query().Get("alias")
	version := r.URL.Query().Get("version")
	indexStr := r.URL.Query().Get("index")
	html, statusCode, err := ui.servePackageCard(r.Context(), alias, version, indexStr)
	if err != nil {
		http.Error(w, err.Error(), statusCode)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.URL.Query().Get("modal") != "1" {
		index, _ := strconv.Atoi(indexStr)
		w.Header().Set("HX-Push-Url", ui.pathForPackage(alias, version, index))
	}
	_, _ = w.Write(html)
}

func (ui *UI) servePackageCard(ctx context.Context, alias string, version string, indexStr string) (html []byte, statusCode int, err error) {

	if alias == "" || version == "" || indexStr == "" {
		return nil, http.StatusBadRequest, errors.New("alias, version, index required")
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		return nil, http.StatusBadRequest, errors.New("invalid index")
	}
	out, err := ui.provider.GetManifestAggregated(ctx, alias, version)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if index >= len(out.Packages) {
		return nil, http.StatusBadRequest, errors.New("index out of range")
	}
	pkg := &out.Packages[index]
	data := packageCardData{
		PackageWithSource: *pkg,
		Alias:             alias,
		Version:           version,
		SourceURL:         ui.packageSourceURL(pkg.SourceAlias, pkg.SourceVersion, alias, version),
		InstallCommand:    ui.packageInstallCommand(pkg.SourceAlias, pkg.SourceVersion, alias, version, pkg.Name),
	}
	html, err = ui.renderTemplate("package-card", data)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return html, http.StatusOK, nil
}

func (ui *UI) handleStatic(w http.ResponseWriter, r *http.Request) {

	name := r.PathValue("file")
	if name == "" {
		prefix := ui.uiPrefix + "/static/"
		if ui.uiPrefix == "" {
			prefix = "/static/"
		}
		name = strings.TrimPrefix(r.URL.Path, prefix)
	}
	name = strings.TrimPrefix(name, "/")
	data, contentType, err := ui.serveStatic(name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	_, _ = w.Write(data)
}

func (ui *UI) handleFavicon(w http.ResponseWriter, r *http.Request) {

	if len(faviconSVG) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(faviconSVG)))
	_, _ = w.Write(faviconSVG)
}

func (ui *UI) serveStatic(name string) (data []byte, contentType string, err error) {

	if name == "" || strings.Contains(name, "..") {
		return nil, "", fs.ErrNotExist
	}
	data, err = fs.ReadFile(ui.staticFS, name)
	if err != nil {
		return nil, "", err
	}
	switch {
	case strings.HasSuffix(name, ".css"):
		contentType = "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		contentType = "application/javascript; charset=utf-8"
	case strings.HasSuffix(name, ".svg"):
		contentType = "image/svg+xml; charset=utf-8"
	}
	return data, contentType, nil
}
