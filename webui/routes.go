package webui

import (
	"net/http"
	"path"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// SetRoutes регистрирует маршруты Web UI в net/http ServeMux. prefix — базовый путь ("" или "/" — в корне, "/ui" — по пути /ui).
func (ui *UI) SetRoutes(mux *http.ServeMux, prefix string) {

	ui.uiPrefix = normalizePrefix(prefix)
	base := ui.uiPrefix
	p := func(parts ...string) string {
		if base == "" {
			return "/" + path.Join(parts...)
		}
		return path.Join(append([]string{base}, parts...)...)
	}

	if base == "" {
		mux.HandleFunc("GET /", ui.handleIndex)
		mux.HandleFunc("GET /{alias}", ui.handleIndex)
		mux.HandleFunc("GET /{alias}/{version}", ui.handleIndex)
	} else {
		mux.HandleFunc("GET "+base, ui.handleIndex)
		mux.HandleFunc("GET "+base+"/", ui.handleIndex)
		mux.HandleFunc("GET "+base+"/{alias}", ui.handleIndex)
		mux.HandleFunc("GET "+base+"/{alias}/{version}", ui.handleIndex)
	}
	mux.HandleFunc("GET "+p("fragments", "projects"), ui.handleFragmentsProjects)
	mux.HandleFunc("GET "+p("fragments", "projects", "{alias}", "versions"), ui.handleFragmentsVersions)
	mux.HandleFunc("GET "+p("fragments", "projects", "{alias}", "collapse"), ui.handleFragmentsProjectCollapse)
	mux.HandleFunc("GET "+p("fragments", "projects", "{alias}", "versions", "{version}", "packages"), ui.handleFragmentsPackages)
	mux.HandleFunc("GET "+p("fragments", "package"), ui.handleFragmentsPackage)
	mux.HandleFunc("GET "+p("static", "{file}"), ui.handleStatic)
	mux.HandleFunc("GET /favicon.ico", ui.handleFavicon)
	if base != "" {
		mux.HandleFunc("GET "+base+"/favicon.ico", ui.handleFavicon)
	}
}

// SetRoutesFiber регистрирует маршруты Web UI в Fiber. prefix — базовый путь ("" или "/" — в корне, "/ui" — по пути /ui).
func (ui *UI) SetRoutesFiber(app *fiber.App, prefix string) {

	ui.uiPrefix = normalizePrefix(prefix)
	base := ui.uiPrefix
	if base == "" {
		base = "/"
	}
	group := app.Group(base)

	group.Get("/", ui.fiberIndex)
	group.Get("/:alias/:version", ui.fiberIndexWithPath)
	group.Get("/:alias", ui.fiberIndexWithPath)
	group.Get("/fragments/projects", ui.fiberFragmentsProjects)
	group.Get("/fragments/projects/:alias/versions", ui.fiberFragmentsVersions)
	group.Get("/fragments/projects/:alias/collapse", ui.fiberFragmentsProjectCollapse)
	group.Get("/fragments/projects/:alias/versions/:version/packages", ui.fiberFragmentsPackages)
	group.Get("/fragments/package", ui.fiberFragmentsPackage)
	group.Get("/static/*", ui.fiberStatic)
	group.Get("/favicon.ico", ui.fiberFavicon)
}

func (ui *UI) fiberIndex(c *fiber.Ctx) (err error) {

	return ui.fiberIndexWithPath(c)
}

func (ui *UI) fiberIndexWithPath(c *fiber.Ctx) (err error) {

	initialPath := ""
	if alias := c.Params("alias"); alias != "" {
		initialPath = alias
		if version := c.Params("version"); version != "" {
			initialPath = alias + "/" + version
		}
	}
	theme := "system"
	if cookie := c.Cookies("webui-theme"); cookie != "" {
		theme = cookie
	}
	html, err := ui.renderIndex(theme, initialPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(html)
}

func (ui *UI) fiberFragmentsProjects(c *fiber.Ctx) (err error) {

	html, err := ui.serveProjects(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.Send(html)
}

func (ui *UI) fiberFragmentsVersions(c *fiber.Ctx) (err error) {

	alias := c.Params("alias")
	if alias == "" {
		return c.Status(fiber.StatusBadRequest).SendString("alias required")
	}
	html, err := ui.serveVersions(c.Context(), alias)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("HX-Push-Url", ui.pathForAlias(alias))
	return c.Send(html)
}

func (ui *UI) fiberFragmentsProjectCollapse(c *fiber.Ctx) (err error) {

	alias := c.Params("alias")
	if alias == "" {
		return c.Status(fiber.StatusBadRequest).SendString("alias required")
	}
	html, err := ui.serveProjectCollapse(alias)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	pushURL := ui.uiPrefix
	if pushURL == "" {
		pushURL = "/"
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("HX-Push-Url", pushURL)
	return c.Send(html)
}

func (ui *UI) fiberFragmentsPackages(c *fiber.Ctx) (err error) {

	alias := c.Params("alias")
	version := c.Params("version")
	mainHtml, _, err := ui.servePackagesWithTree(c.Context(), alias, version)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("HX-Push-Url", ui.pathForVersion(alias, version))
	return c.Send(mainHtml)
}

func (ui *UI) fiberFragmentsPackage(c *fiber.Ctx) (err error) {

	alias := c.Query("alias")
	version := c.Query("version")
	indexStr := c.Query("index")
	html, statusCode, err := ui.servePackageCard(c.Context(), alias, version, indexStr)
	if err != nil {
		return c.Status(statusCode).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html; charset=utf-8")
	if c.Query("modal") != "1" {
		index, _ := strconv.Atoi(indexStr)
		c.Set("HX-Push-Url", ui.pathForPackage(alias, version, index))
	}
	return c.Send(html)
}

func (ui *UI) fiberStatic(c *fiber.Ctx) (err error) {

	file := c.Params("*")
	if file == "" {
		return c.Status(fiber.StatusNotFound).SendString("not found")
	}
	data, contentType, err := ui.serveStatic(file)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("not found")
	}
	if contentType != "" {
		c.Set("Content-Type", contentType)
	}
	return c.Send(data)
}

func (ui *UI) fiberFavicon(c *fiber.Ctx) (err error) {

	if len(faviconSVG) == 0 {
		return c.SendStatus(fiber.StatusNotFound)
	}
	c.Set("Content-Type", "image/svg+xml; charset=utf-8")
	return c.Send(faviconSVG)
}
