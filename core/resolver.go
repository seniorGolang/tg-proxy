package core

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/seniorGolang/tg-proxy/helpers"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

const projectTTL = 1 * time.Hour

type resolver struct {
	storage   storage
	encryptor encryptor
	cache     cache
}

func newResolver(stor storage, enc encryptor, c cache) (res *resolver) {

	return &resolver{
		storage:   stor,
		encryptor: enc,
		cache:     c,
	}
}

func (r *resolver) ResolveProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	if r.cache != nil {
		if project, found, _ = r.cache.GetProject(ctx, alias); found {
			slog.Debug("Project found in cache",
				slog.String(helpers.LogKeyAction, helpers.ActionResolveProject),
				slog.String(helpers.LogKeyAlias, alias),
			)
			return
		}
	}

	if r.storage == nil {
		return
	}

	if project, found, err = r.storage.GetProject(ctx, alias); err != nil {
		slog.Debug("Failed to get project from storage",
			slog.String(helpers.LogKeyAction, helpers.ActionResolveProject),
			slog.String(helpers.LogKeyAlias, alias),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	if !found {
		return
	}

	if project.EncryptedToken != "" && r.encryptor != nil {
		var token string
		if token, err = r.encryptor.DecryptString(project.EncryptedToken); err != nil {
			slog.Debug("Failed to decrypt token",
				slog.String(helpers.LogKeyAction, helpers.ActionResolveProject),
				slog.String(helpers.LogKeyAlias, alias),
				slog.Any(helpers.LogKeyError, err),
			)
			project = domain.Project{}
			err = fmt.Errorf("failed to decrypt token: %w", err)
			return
		}
		project.Token = token
		project.EncryptedToken = ""
	}

	if r.cache != nil {
		_ = r.cache.SetProject(ctx, alias, project, projectTTL)
	}

	return
}

func (r *resolver) InvalidateCache(ctx context.Context, alias string) (err error) {

	if r.cache != nil {
		_ = r.cache.DeleteProject(ctx, alias)
	}
	return
}

func (r *resolver) ClearCache(ctx context.Context) (err error) {

	if r.cache != nil {
		_ = r.cache.Clear(ctx)
	}
	return
}
