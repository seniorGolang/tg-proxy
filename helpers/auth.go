package helpers

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	basicAuthPrefix     = "Basic "
	bearerAuthPrefix    = "Bearer "
	defaultAPIKeyHeader = "X-Tg-Proxy-Key" //nolint:gosec
)

// StaticKeyAuth реализация авторизации через статический ключ (API Key)
// Поддерживает оба фреймворка (net/http и go-fiber)
type StaticKeyAuth struct {
	key    string
	header string
}

func (a *StaticKeyAuth) Authorize(r *http.Request) (err error) {

	var headerName string
	if a.header != "" {
		headerName = a.header
	} else {
		headerName = defaultAPIKeyHeader
	}

	key := r.Header.Get(headerName)
	if key == "" {
		return errors.New("missing API key")
	}

	if key != a.key {
		return errors.New("invalid API key")
	}

	return
}

func (a *StaticKeyAuth) AuthorizeFiber(c *fiber.Ctx) (err error) {

	var headerName string
	if a.header != "" {
		headerName = a.header
	} else {
		headerName = defaultAPIKeyHeader
	}

	key := c.Get(headerName)
	if key == "" {
		return errors.New("missing API key")
	}

	if key != a.key {
		return errors.New("invalid API key")
	}

	return
}

// NewStaticKeyAuth создает новый StaticKeyAuth
func NewStaticKeyAuth(key string, headerName string) (auth *StaticKeyAuth) {
	return &StaticKeyAuth{key: key, header: headerName}
}

// BasicAuth реализация авторизации через HTTP Basic Authentication
// Поддерживает оба фреймворка (net/http и go-fiber)
type BasicAuth struct {
	username string
	password string
}

func (a *BasicAuth) Authorize(r *http.Request) (err error) {

	var username, password string
	var ok bool
	if username, password, ok = r.BasicAuth(); !ok {
		return errors.New("missing basic auth")
	}

	if username != a.username || password != a.password {
		return errors.New("invalid credentials")
	}

	return
}

func (a *BasicAuth) AuthorizeFiber(c *fiber.Ctx) (err error) {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, basicAuthPrefix) {
		return errors.New("invalid authorization header")
	}

	encoded := strings.TrimPrefix(authHeader, basicAuthPrefix)
	var decoded []byte
	if decoded, err = base64.StdEncoding.DecodeString(encoded); err != nil {
		return errors.New("invalid basic auth encoding")
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return errors.New("invalid basic auth format")
	}

	if parts[0] != a.username || parts[1] != a.password {
		return errors.New("invalid credentials")
	}

	return
}

// NewBasicAuth создает новый BasicAuth
func NewBasicAuth(username string, password string) (auth *BasicAuth) {
	return &BasicAuth{username: username, password: password}
}

// JWTAuth реализация авторизации через JWT с асинхронной подписью (RSA)
// Поддерживает оба фреймворка (net/http и go-fiber)
type JWTAuth struct {
	publicKey *rsa.PublicKey
}

func (a *JWTAuth) Authorize(r *http.Request) (err error) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, bearerAuthPrefix) {
		return errors.New("invalid authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, bearerAuthPrefix)

	var token *jwt.Token
	if token, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.publicKey, nil
	}); err != nil {
		return
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return
}

func (a *JWTAuth) AuthorizeFiber(c *fiber.Ctx) (err error) {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing authorization header")
	}

	if !strings.HasPrefix(authHeader, bearerAuthPrefix) {
		return errors.New("invalid authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, bearerAuthPrefix)

	var token *jwt.Token
	if token, err = jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.publicKey, nil
	}); err != nil {
		return
	}

	if !token.Valid {
		return errors.New("invalid token")
	}

	return
}

// NewJWTAuth создает новый JWTAuth из PEM-encoded публичного ключа
func NewJWTAuth(publicKeyPEM string) (auth *JWTAuth, err error) {

	var publicKey *rsa.PublicKey
	if publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM)); err != nil {
		return
	}

	auth = &JWTAuth{publicKey: publicKey}
	return
}
