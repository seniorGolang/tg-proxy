package core

import (
	"context"

	"github.com/google/uuid"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type storage interface {
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (project domain.Project, found bool, err error)
	GetProjectByRepoURL(ctx context.Context, repoURL string) (project domain.Project, found bool, err error)
	CreateProject(ctx context.Context, project domain.Project) (id uuid.UUID, err error)
	UpdateProject(ctx context.Context, alias string, project domain.Project) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	ListProjects(ctx context.Context, limit int, offset int) (projects []domain.Project, total int64, err error)
	GetCatalogVersion(ctx context.Context) (version string, err error)
}
