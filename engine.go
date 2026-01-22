package tgproxy

import (
	"context"
	"io"

	"github.com/seniorGolang/tg-proxy/core"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

// engine интерфейс для Core Engine (приватный, используется только в этом пакете)
type engine interface {
	// Получение манифеста для проекта и версии
	GetManifest(ctx context.Context, alias string, version string, baseURL string) (manifest []byte, err error)

	// Получение файла как потока (потоковая передача)
	GetFile(ctx context.Context, alias string, version string, filename string) (stream io.ReadCloser, err error)

	// Получение списка версий для проекта
	GetVersions(ctx context.Context, alias string) (versions []string, err error)

	// Получение источника по имени
	GetSource(name string) (src core.Source, err error)

	// Управление проектами
	CreateProject(ctx context.Context, project domain.Project) (err error)
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	UpdateProject(ctx context.Context, alias string, project domain.Project) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error)
}
