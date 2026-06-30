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
	inner SecretProvider
	ttl   time.Duration
	mu    sync.Mutex
	items map[string]cacheItem
}

// NewCachedProvider wraps a SecretProvider with an in-memory TTL cache.
func NewCachedProvider(inner SecretProvider, ttl time.Duration) SecretProvider {
	return &cachedProvider{
		inner: inner,
		ttl:   ttl,
		items: make(map[string]cacheItem),
	}
}

func (c *cachedProvider) Get(ctx context.Context, path string) (Secret, error) {
	c.mu.Lock()
	it, ok := c.items[path]
	if ok && time.Since(it.at) < c.ttl {
		c.mu.Unlock()
		return it.sec, nil
	}
	c.mu.Unlock()

	sec, err := c.inner.Get(ctx, path)
	if err != nil {
		return Secret{}, err // No fallback to cleartext
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	// Could emit metric here: if old.sec.Version != sec.Version { metrics.RotationDetected(path) }
	c.items[path] = cacheItem{sec: sec, at: time.Now()}
	return sec, nil
}
