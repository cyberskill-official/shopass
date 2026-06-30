package secrets

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type countingProvider struct {
	val   string
	ver   string
	calls int
	mu    sync.Mutex
}

func (p *countingProvider) Get(ctx context.Context, path string) (Secret, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls++
	return Secret{value: p.val, Version: p.ver}, nil
}

type errProvider struct{}

func (p *errProvider) Get(ctx context.Context, path string) (Secret, error) {
	return Secret{}, errors.New("backend error")
}

func TestCache_HitWithinTTL(t *testing.T) {
	ctx := context.Background()
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, 60*time.Second)

	_, _ = p.Get(ctx, "db/main")
	_, _ = p.Get(ctx, "db/main")

	require.Equal(t, 1, backend.calls)
}

func TestCache_RefreshAfterTTL(t *testing.T) {
	ctx := context.Background()
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, 10*time.Millisecond)

	_, _ = p.Get(ctx, "db/main")
	time.Sleep(20 * time.Millisecond)
	_, _ = p.Get(ctx, "db/main")

	require.Equal(t, 2, backend.calls)
}

func TestCache_RotationDetected(t *testing.T) {
	ctx := context.Background()
	backend := &countingProvider{val: "old", ver: "v1"}
	p := NewCachedProvider(backend, 10*time.Millisecond)

	_, _ = p.Get(ctx, "auth/jwt-signing")

	backend.mu.Lock()
	backend.val = "new"
	backend.ver = "v2"
	backend.mu.Unlock()

	time.Sleep(20 * time.Millisecond)
	s, _ := p.Get(ctx, "auth/jwt-signing")

	require.Equal(t, "v2", s.Version)
}

func TestProvider_BackendError_NoFallback(t *testing.T) {
	ctx := context.Background()
	p := NewCachedProvider(&errProvider{}, time.Minute)

	_, err := p.Get(ctx, "db/main")
	require.Error(t, err) // NO cleartext fallback
}

func TestCache_ConcurrentNoRace(t *testing.T) {
	ctx := context.Background()
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = p.Get(ctx, "db/main")
		}()
	}
	wg.Wait()
	// Race detector will catch issues if any
	require.GreaterOrEqual(t, backend.calls, 1)
}
