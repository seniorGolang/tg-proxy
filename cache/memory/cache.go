package memory

import (
	"context"
	"sync"
	"time"

	"github.com/seniorGolang/tg-proxy/model/domain"
)

type cacheEntry struct {
	project  domain.Project
	expires  time.Time
	versions []string
}

type aggregateEntry struct {
	manifest []byte
	expires  time.Time
}

type Cache struct {
	mu        sync.RWMutex
	projects  map[string]*cacheEntry
	versions  map[string]*cacheEntry
	aggregate *aggregateEntry
}

func NewCache() (c *Cache) {
	return &Cache{
		projects: make(map[string]*cacheEntry),
		versions: make(map[string]*cacheEntry),
	}
}

func (c *Cache) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.projects[alias]
	if !exists {
		return
	}

	if time.Now().After(entry.expires) {
		return
	}

	return entry.project, true, nil
}

func (c *Cache) SetProject(ctx context.Context, alias string, project domain.Project, ttl time.Duration) (err error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.projects[alias] = &cacheEntry{
		project: project,
		expires: time.Now().Add(ttl),
	}

	return
}

func (c *Cache) DeleteProject(ctx context.Context, alias string) (err error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.projects, alias)
	delete(c.versions, alias)

	return
}

func (c *Cache) GetVersions(ctx context.Context, alias string) (versions []string, found bool, err error) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.versions[alias]
	if !exists {
		return
	}

	if time.Now().After(entry.expires) {
		return
	}

	versions = make([]string, len(entry.versions))
	copy(versions, entry.versions)
	found = true

	return
}

func (c *Cache) SetVersions(ctx context.Context, alias string, versions []string, ttl time.Duration) (err error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.versions[alias] = &cacheEntry{
		versions: versions,
		expires:  time.Now().Add(ttl),
	}

	return
}

func (c *Cache) GetAggregateManifest(ctx context.Context) (manifest []byte, found bool, err error) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.aggregate == nil {
		return
	}

	if time.Now().After(c.aggregate.expires) {
		return
	}

	manifest = make([]byte, len(c.aggregate.manifest))
	copy(manifest, c.aggregate.manifest)
	found = true

	return
}

func (c *Cache) SetAggregateManifest(ctx context.Context, manifest []byte, ttl time.Duration) (err error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	data := make([]byte, len(manifest))
	copy(data, manifest)
	c.aggregate = &aggregateEntry{
		manifest: data,
		expires:  time.Now().Add(ttl),
	}

	return
}

func (c *Cache) Clear(ctx context.Context) (err error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.projects = make(map[string]*cacheEntry)
	c.versions = make(map[string]*cacheEntry)
	c.aggregate = nil

	return
}
