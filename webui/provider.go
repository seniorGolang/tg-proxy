package webui

import (
	"context"

	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/dto"
)

type Provider interface {
	ListProjects(ctx context.Context, limit int, offset int) (projects []dto.ProjectResponse, total int64, err error)
	GetVersions(ctx context.Context, alias string) (versions []string, err error)
	GetManifestAggregated(ctx context.Context, alias string, version string) (out *model.ManifestAggregatedResponse, err error)
}

// ManifestBaseProvider — опциональный интерфейс для baseURL и префикса public API (Proxy, webui/cache).
// Позволяет UI брать данные для команд установки из провайдера без явной передачи.
type ManifestBaseProvider interface {
	BaseURL() (baseURL string)
	PublicPrefix() (prefix string)
}
