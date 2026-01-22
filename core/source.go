package core

import (
	"context"
	"io"
	"net/http"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

// Source интерфейс для работы с источниками пакетов
type Source interface {
	Name() (name string)
	GetManifest(ctx context.Context, project domain.Project, version string) (manifest domain.Manifest, err error)
	GetFileStream(ctx context.Context, project domain.Project, version string, filename string) (stream io.ReadCloser, err error)
	GetFileResponse(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error)
	GetVersions(ctx context.Context, project domain.Project) (versions []string, err error)
}

// SourceInfo опциональный интерфейс для получения метаданных источника
type SourceInfo interface {
	BaseURL() (url string)
}
