package tgproxy

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/seniorGolang/tg-proxy/core"
	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

type engine interface {
	GetAggregateManifest(ctx context.Context, baseURL string) (manifest []byte, err error)
	GetCatalogVersion(ctx context.Context) (version string, err error)
	GetManifest(ctx context.Context, alias string, version string, baseURL string) (manifest []byte, err error)
	GetManifestData(ctx context.Context, alias string, version string, baseURL string) (m *model.Manifest, err error)
	GetManifestAggregated(ctx context.Context, alias string, version string, baseURL string) (out *model.ManifestAggregatedResponse, err error)
	GetFile(ctx context.Context, alias string, version string, filename string) (stream io.ReadCloser, err error)
	GetVersions(ctx context.Context, alias string) (versions []string, err error)
	GetSource(name string) (src core.Source, err error)
	CreateProject(ctx context.Context, project domain.Project) (id uuid.UUID, err error)
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (project domain.Project, found bool, err error)
	UpdateProject(ctx context.Context, alias string, project domain.Project) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error)
}
