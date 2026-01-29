package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/seniorGolang/tg-proxy/cache/redis/internal"
	"github.com/seniorGolang/tg-proxy/model/domain"
)

const (
	projectKeyPrefix     = "project:"
	versionsKeyPrefix    = "versions:"
	aggregateManifestKey = "aggregate_manifest"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) (c *Cache) {
	return &Cache{
		client: client,
	}
}

func (c *Cache) GetProject(ctx context.Context, alias string) (project domain.Project, found bool, err error) {

	key := projectKeyPrefix + alias
	var data string
	if data, err = c.client.Get(ctx, key).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
			return
		}
		return
	}

	var doc internal.Project
	if err = json.Unmarshal([]byte(data), &doc); err != nil {
		err = fmt.Errorf("failed to unmarshal project: %w", err)
		return
	}

	project = doc.ToDomain()
	found = true
	return
}

func (c *Cache) SetProject(ctx context.Context, alias string, project domain.Project, ttl time.Duration) (err error) {

	key := projectKeyPrefix + alias
	doc := internal.FromDomain(project)
	var data []byte
	if data, err = json.Marshal(doc); err != nil {
		err = fmt.Errorf("failed to marshal project: %w", err)
		return
	}

	if err = c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		err = fmt.Errorf("failed to set project in cache: %w", err)
		return
	}

	return
}

func (c *Cache) DeleteProject(ctx context.Context, alias string) (err error) {

	projectKey := projectKeyPrefix + alias
	versionsKey := versionsKeyPrefix + alias

	pipe := c.client.Pipeline()
	pipe.Del(ctx, projectKey)
	pipe.Del(ctx, versionsKey)

	if _, err = pipe.Exec(ctx); err != nil {
		err = fmt.Errorf("failed to delete project from cache: %w", err)
		return
	}

	return
}

func (c *Cache) GetVersions(ctx context.Context, alias string) (versions []string, found bool, err error) {

	key := versionsKeyPrefix + alias
	var data string
	if data, err = c.client.Get(ctx, key).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
			return
		}
		return
	}

	if err = json.Unmarshal([]byte(data), &versions); err != nil {
		err = fmt.Errorf("failed to unmarshal versions: %w", err)
		return
	}

	found = true
	return
}

func (c *Cache) SetVersions(ctx context.Context, alias string, versions []string, ttl time.Duration) (err error) {

	key := versionsKeyPrefix + alias
	var data []byte
	if data, err = json.Marshal(versions); err != nil {
		err = fmt.Errorf("failed to marshal versions: %w", err)
		return
	}

	if err = c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		err = fmt.Errorf("failed to set versions in cache: %w", err)
		return
	}

	return
}

func (c *Cache) GetAggregateManifest(ctx context.Context) (manifest []byte, found bool, err error) {

	var data string
	if data, err = c.client.Get(ctx, aggregateManifestKey).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil
			return
		}
		err = fmt.Errorf("failed to get aggregate manifest from cache: %w", err)
		return
	}

	manifest = []byte(data)
	found = true
	return
}

func (c *Cache) SetAggregateManifest(ctx context.Context, manifest []byte, ttl time.Duration) (err error) {

	if err = c.client.Set(ctx, aggregateManifestKey, manifest, ttl).Err(); err != nil {
		err = fmt.Errorf("failed to set aggregate manifest in cache: %w", err)
		return
	}

	return
}

func (c *Cache) Clear(ctx context.Context) (err error) {

	var keys []string

	projectPattern := projectKeyPrefix + "*"
	iter := c.client.Scan(ctx, 0, projectPattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err = iter.Err(); err != nil {
		err = fmt.Errorf("failed to scan project keys: %w", err)
		return
	}

	versionsPattern := versionsKeyPrefix + "*"
	iter = c.client.Scan(ctx, 0, versionsPattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err = iter.Err(); err != nil {
		err = fmt.Errorf("failed to scan version keys: %w", err)
		return
	}

	keys = append(keys, aggregateManifestKey)

	if len(keys) > 0 {
		if err = c.client.Del(ctx, keys...).Err(); err != nil {
			err = fmt.Errorf("failed to delete keys: %w", err)
			return
		}
	}

	return
}
