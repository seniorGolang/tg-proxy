package webui

import (
	"embed"
	"html/template"
	"io/fs"
	"strconv"
	"strings"

	"github.com/seniorGolang/tg-proxy/helpers"
)

type UI struct {
	provider     Provider
	uiPrefix     string
	baseURL      string
	publicPrefix string
	templates    *template.Template
	staticFS     fs.FS
}

func New(provider Provider) (ui *UI) {

	ui = &UI{
		provider:  provider,
		staticFS:  mustSubFS(FS, "static"),
		templates: mustParseTemplates(FS, "templates"),
	}
	return
}

func normalizePrefix(prefix string) (base string) {

	base = strings.TrimSuffix(prefix, "/")
	if base == "/" {
		base = ""
	}
	return
}

func (ui *UI) SetManifestBase(baseURL string, publicPrefix string) {

	ui.baseURL = strings.TrimSuffix(baseURL, "/")
	ui.publicPrefix = strings.TrimSuffix(publicPrefix, "/")
}

// ApplyManifestBaseFromProvider подставляет baseURL и publicPrefix из провайдера, если он реализует ManifestBaseProvider.
// Вызывать после New вместо SetManifestBase, когда провайдер — proxy или обёртка над ним (например cache).
func (ui *UI) ApplyManifestBaseFromProvider() {

	if p, ok := ui.provider.(ManifestBaseProvider); ok {
		ui.baseURL = strings.TrimSuffix(p.BaseURL(), "/")
		ui.publicPrefix = strings.TrimSuffix(p.PublicPrefix(), "/")
	}
}

func (ui *UI) manifestSourceURL(alias string) (url string) {

	if ui.baseURL == "" {
		return ""
	}
	pathParts := []string{strings.TrimPrefix(ui.publicPrefix, "/"), alias}
	return helpers.BuildURL(ui.baseURL, pathParts...)
}

func (ui *UI) manifestInstallCommand(alias string, version string) (cmd string) {

	u := ui.manifestSourceURL(alias)
	if u == "" {
		return "tg pkg add " + alias + "@" + version
	}
	return "tg pkg add " + u + "@" + version
}

func (ui *UI) packageSourceURL(sourceAlias string, sourceVersion string, fallbackAlias string, fallbackVersion string) (url string) {

	alias := fallbackAlias
	if sourceAlias != "" {
		alias = sourceAlias
	}
	return ui.manifestSourceURL(alias)
}

func (ui *UI) packageInstallCommand(sourceAlias string, sourceVersion string, fallbackAlias string, fallbackVersion string, packageName string) (cmd string) {

	alias := fallbackAlias
	version := fallbackVersion
	if sourceAlias != "" {
		alias = sourceAlias
	}
	if sourceVersion != "" {
		version = sourceVersion
	}
	u := ui.manifestSourceURL(alias)
	if u == "" {
		return "tg pkg add " + alias + ":" + packageName + "@" + version
	}
	return "tg pkg add " + u + ":" + packageName + "@" + version
}

func (ui *UI) pathForAlias(alias string) (p string) {

	if ui.uiPrefix == "" {
		return "/" + alias
	}
	return ui.uiPrefix + "/" + alias
}

func (ui *UI) pathForVersion(alias string, version string) (p string) {

	if ui.uiPrefix == "" {
		return "/" + alias + "/" + version
	}
	return ui.uiPrefix + "/" + alias + "/" + version
}

func (ui *UI) pathForPackage(alias string, version string, index int) (p string) {

	base := ui.pathForVersion(alias, version)
	return base + "?pkg=" + strconv.Itoa(index)
}

func mustSubFS(efs embed.FS, dir string) (f fs.FS) {

	var err error
	if f, err = fs.Sub(efs, dir); err != nil {
		panic("webui: static subfs: " + err.Error())
	}
	return
}

func safeID(s string) string {
	return strings.ReplaceAll(s, ".", "-")
}

func mustParseTemplates(efs embed.FS, dir string) (t *template.Template) {

	root, err := fs.Sub(efs, dir)
	if err != nil {
		panic("webui: templates subfs: " + err.Error())
	}
	t = template.Must(template.New("").Funcs(template.FuncMap{
		"safeID": safeID,
	}).ParseFS(root, "*.gohtml"))
	return
}
