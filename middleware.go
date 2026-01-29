package tgproxy

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/seniorGolang/tg-proxy/helpers"
)

func (p *Proxy) publicAuthMiddleware(next http.HandlerFunc) (handler http.HandlerFunc) {

	return func(w http.ResponseWriter, r *http.Request) {
		if p.publicAuth != nil {
			var err error
			if err = p.publicAuth.Authorize(r); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "public"),
					slog.String(helpers.LogKeyMethod, r.Method),
					slog.String(helpers.LogKeyPath, r.URL.Path),
					slog.Any(helpers.LogKeyError, err),
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

func (p *Proxy) adminAuthMiddleware(next http.HandlerFunc) (handler http.HandlerFunc) {

	return func(w http.ResponseWriter, r *http.Request) {
		if p.adminAuth != nil {
			var err error
			if err = p.adminAuth.Authorize(r); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "admin"),
					slog.String(helpers.LogKeyMethod, r.Method),
					slog.String(helpers.LogKeyPath, r.URL.Path),
					slog.Any(helpers.LogKeyError, err),
				)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

func (p *Proxy) publicFiberAuthMiddleware(c *fiber.Ctx) (err error) {

	if p.publicAuth != nil {
		var fiberAuth FiberAuthProvider
		var ok bool
		if fiberAuth, ok = p.publicAuth.(FiberAuthProvider); ok {
			if err = fiberAuth.AuthorizeFiber(c); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "public"),
					slog.String(helpers.LogKeyMethod, c.Method()),
					slog.String(helpers.LogKeyPath, c.Path()),
					slog.Any(helpers.LogKeyError, err),
				)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized",
				})
			}
		} else {
			var httpReq *http.Request
			if httpReq, err = http.NewRequest(c.Method(), string(c.Request().URI().FullURI()), nil); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to process request",
				})
			}
			for key, values := range c.GetReqHeaders() {
				if len(values) > 0 {
					httpReq.Header.Set(key, values[0])
					for i := 1; i < len(values); i++ {
						httpReq.Header.Add(key, values[i])
					}
				}
			}
			if err = p.publicAuth.Authorize(httpReq); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "public"),
					slog.String(helpers.LogKeyMethod, c.Method()),
					slog.String(helpers.LogKeyPath, c.Path()),
					slog.Any(helpers.LogKeyError, err),
				)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized",
				})
			}
		}
	}
	return c.Next()
}

func (p *Proxy) adminFiberAuthMiddleware(c *fiber.Ctx) (err error) {

	if p.adminAuth != nil {
		var fiberAuth FiberAuthProvider
		var ok bool
		if fiberAuth, ok = p.adminAuth.(FiberAuthProvider); ok {
			if err = fiberAuth.AuthorizeFiber(c); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "admin"),
					slog.String(helpers.LogKeyMethod, c.Method()),
					slog.String(helpers.LogKeyPath, c.Path()),
					slog.Any(helpers.LogKeyError, err),
				)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized",
				})
			}
		} else {
			var httpReq *http.Request
			if httpReq, err = http.NewRequest(c.Method(), string(c.Request().URI().FullURI()), nil); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to process request",
				})
			}
			for key, values := range c.GetReqHeaders() {
				if len(values) > 0 {
					httpReq.Header.Set(key, values[0])
					for i := 1; i < len(values); i++ {
						httpReq.Header.Add(key, values[i])
					}
				}
			}
			if err = p.adminAuth.Authorize(httpReq); err != nil {
				slog.Info("Authorization failed",
					slog.String(helpers.LogKeyAuthProvider, "admin"),
					slog.String(helpers.LogKeyMethod, c.Method()),
					slog.String(helpers.LogKeyPath, c.Path()),
					slog.Any(helpers.LogKeyError, err),
				)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Unauthorized",
				})
			}
		}
	}
	return c.Next()
}
