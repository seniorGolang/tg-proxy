package github

import (
	"net/http"
)

const sourceName = "github"

type Source struct {
	baseURL string
	token   string
	http    *http.Client
}

// ClientOption - функция опции для настройки GitHub клиента
type ClientOption func(*Source)

// DefaultToken устанавливает fallback токен доступа
// Этот токен используется только если токен проекта не задан
// В норме токен задан для каждого проекта через project.Token
func DefaultToken(token string) (opt ClientOption) {
	return func(s *Source) {
		s.token = token
	}
}

// NewClient создает новый GitHub клиент с использованием паттерна опций
// baseURL - базовый URL GitHub (например, https://api.github.com)
// opts - опции для настройки клиента (например, DefaultToken для fallback токена)
func NewClient(baseURL string, opts ...ClientOption) (src *Source) {

	s := &Source{
		baseURL: baseURL,
		http:    &http.Client{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Source) Name() (name string) {
	return sourceName
}

func (s *Source) BaseURL() (url string) {
	return s.baseURL
}
