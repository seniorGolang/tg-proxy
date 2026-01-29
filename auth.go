package tgproxy

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type AuthProvider interface {
	Authorize(r *http.Request) (err error)
}

type FiberAuthProvider interface {
	AuthorizeFiber(c *fiber.Ctx) (err error)
}

type UnifiedAuthProvider interface {
	AuthProvider
	FiberAuthProvider
}
