package secrets

import (
	"context"
	"sync"
	"time"
)

type cacheItem struct {
	sec Secret
	at  time.Time
}

type cachedProvider struct {
	inner     SecretProvider
	ttl       time.Duration
	mu        sync.Mutex
	items     map[string]cacheItem
	rotations map[string]int
}

func NewCachedProvider(inner SecretProvider, ttl time.Duration) SecretProvider {
	return &cachedProvider{
		inner:     inner,
		ttl:       ttl,
		items:     make(map[string]cacheItem),
		rotations: make(map[string]int),
	}
}

func (c *cachedProvider) Get(ctx context.Context, path string) (Secret, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if it, ok := c.items[path]; ok && time.Since(it.at) < c.ttl {
		return it.sec, nil
	}

	sec, err := c.inner.Get(ctx, path)
	if err != nil {
		return Secret{}, err
	}

	if old, ok := c.items[path]; ok && old.sec.Version != sec.Version {
		c.rotations[path]++
	}
	c.items[path] = cacheItem{sec: sec, at: time.Now()}
	return sec, nil
}

func (c *cachedProvider) RotationCount(path string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rotations[path]
}
