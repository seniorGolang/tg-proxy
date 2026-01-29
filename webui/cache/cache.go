package cache

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/seniorGolang/tg-proxy/model"
	"github.com/seniorGolang/tg-proxy/model/dto"
	"github.com/seniorGolang/tg-proxy/webui"
)

type entry struct {
	expiresAt time.Time
	value     interface{}
}

// Cache — in-memory обёртка над webui.Provider с настраиваемым TTL.
type Cache struct {
	provider webui.Provider
	ttl      time.Duration
	mu       sync.RWMutex
	items    map[string]*entry
}

func New(provider webui.Provider, ttl time.Duration) *Cache {

	return &Cache{
		provider: provider,
		ttl:      ttl,
		items:    make(map[string]*entry),
	}
}

func (c *Cache) get(key string) (interface{}, bool) {

	c.mu.RLock()
	e := c.items[key]
	c.mu.RUnlock()
	if e == nil || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) set(key string, value interface{}) {

	expiresAt := time.Now().Add(c.ttl)
	c.mu.Lock()
	c.items[key] = &entry{expiresAt: expiresAt, value: value}
	c.mu.Unlock()
}

func (c *Cache) ListProjects(ctx context.Context, limit int, offset int) (projects []dto.ProjectResponse, total int64, err error) {

	key := "projects:" + strconv.Itoa(limit) + ":" + strconv.Itoa(offset)
	if v, ok := c.get(key); ok {
		val := v.(*projectsVal)
		return val.Projects, val.Total, nil
	}
	projects, total, err = c.provider.ListProjects(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	c.set(key, &projectsVal{Projects: projects, Total: total})
	return projects, total, nil
}

func (c *Cache) GetVersions(ctx context.Context, alias string) (versions []string, err error) {

	key := "versions:" + alias
	if v, ok := c.get(key); ok {
		return v.(*versionsVal).Versions, nil
	}
	versions, err = c.provider.GetVersions(ctx, alias)
	if err != nil {
		return nil, err
	}
	c.set(key, &versionsVal{Versions: versions})
	return versions, nil
}

func (c *Cache) GetManifestAggregated(ctx context.Context, alias string, version string) (out *model.ManifestAggregatedResponse, err error) {

	key := "manifest:" + alias + ":" + version
	if v, ok := c.get(key); ok {
		return v.(*model.ManifestAggregatedResponse), nil
	}
	out, err = c.provider.GetManifestAggregated(ctx, alias, version)
	if err != nil {
		return nil, err
	}
	c.set(key, out)
	return out, nil
}

func (c *Cache) BaseURL() (baseURL string) {

	if p, ok := c.provider.(webui.ManifestBaseProvider); ok {
		return p.BaseURL()
	}
	return ""
}

func (c *Cache) PublicPrefix() (prefix string) {

	if p, ok := c.provider.(webui.ManifestBaseProvider); ok {
		return p.PublicPrefix()
	}
	return ""
}

type projectsVal struct {
	Projects []dto.ProjectResponse
	Total    int64
}

type versionsVal struct {
	Versions []string
}
