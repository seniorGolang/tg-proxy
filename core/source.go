package core

import (
	"context"
	"io"
	"net/http"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type Source interface {
	Info() (name, url string)
	ParseFileURL(fileURL string) (version string, filename string, ok bool)
	GetManifest(ctx context.Context, project domain.Project, version string) (manifest domain.Manifest, err error)
	GetFileStream(ctx context.Context, project domain.Project, version string, filename string) (stream io.ReadCloser, err error)
	GetFileResponse(ctx context.Context, project domain.Project, version string, filename string) (resp *http.Response, err error)
	GetVersions(ctx context.Context, project domain.Project) (versions []string, err error)
}
