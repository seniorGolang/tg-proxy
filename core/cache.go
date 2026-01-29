package core

import (
	"context"
	"time"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type cache interface {
	GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error)
	SetProject(ctx context.Context, alias string, project domain.Project, ttl time.Duration) (err error)
	DeleteProject(ctx context.Context, alias string) (err error)
	GetVersions(ctx context.Context, alias string) (versions []string, found bool, err error)
	SetVersions(ctx context.Context, alias string, versions []string, ttl time.Duration) (err error)
	GetAggregateManifest(ctx context.Context) (manifest []byte, found bool, err error)
	SetAggregateManifest(ctx context.Context, manifest []byte, ttl time.Duration) (err error)
	Clear(ctx context.Context) (err error)
}
