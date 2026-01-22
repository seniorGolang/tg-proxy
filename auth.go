package tgproxy

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// AuthProvider интерфейс для авторизации в net/http
type AuthProvider interface {
	// Authorize проверяет авторизацию HTTP запроса
	Authorize(r *http.Request) (err error)
}

// FiberAuthProvider интерфейс для авторизации в go-fiber
type FiberAuthProvider interface {
	// AuthorizeFiber проверяет авторизацию Fiber контекста
	AuthorizeFiber(c *fiber.Ctx) (err error)
}

// UnifiedAuthProvider поддерживает оба фреймворка
type UnifiedAuthProvider interface {
	AuthProvider
	FiberAuthProvider
}
