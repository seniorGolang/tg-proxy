package internal

// Package представляет пакет GitLab для парсинга JSON ответа
type Package struct {
	Version string `json:"version"`
}
