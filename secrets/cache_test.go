package secrets

import (
	"context"
	"sync"
	"testing"
	"time"
)

type countingProvider struct {
	val   string
	ver   string
	calls int
}

func (c *countingProvider) Get(_ context.Context, _ string) (Secret, error) {
	c.calls++
	return Secret{value: c.val, Version: c.ver}, nil
}

func TestCache_HitWithinTTL(t *testing.T) {
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, 60*time.Second)
	ctx := context.Background()
	p.Get(ctx, "db/main")
	p.Get(ctx, "db/main")
	if backend.calls != 1 {
		t.Errorf("expected 1 backend call, got %d", backend.calls)
	}
}

func TestCache_RefreshAfterTTL(t *testing.T) {
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, 10*time.Millisecond)
	ctx := context.Background()
	p.Get(ctx, "db/main")
	time.Sleep(20 * time.Millisecond)
	p.Get(ctx, "db/main")
	if backend.calls != 2 {
		t.Errorf("expected 2 backend calls, got %d", backend.calls)
	}
}

func TestCache_RotationDetected(t *testing.T) {
	backend := &countingProvider{val: "old", ver: "v1"}
	p := NewCachedProvider(backend, 10*time.Millisecond)
	ctx := context.Background()
	p.Get(ctx, "auth/jwt-signing")
	backend.val = "new"
	backend.ver = "v2"
	time.Sleep(20 * time.Millisecond)
	s, err := p.Get(ctx, "auth/jwt-signing")
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != "v2" {
		t.Fatalf("expected v2, got %s", s.Version)
	}
	cached := p.(*cachedProvider)
	if cached.RotationCount("auth/jwt-signing") != 1 {
		t.Fatal("expected one rotation")
	}
}

type errProvider struct{}

func (e *errProvider) Get(_ context.Context, _ string) (Secret, error) {
	return Secret{}, errBackend
}

var errBackend = &backendError{}

type backendError struct{}

func (s *backendError) Error() string { return "backend error" }

func TestProvider_BackendError_NoFallback(t *testing.T) {
	p := NewCachedProvider(&errProvider{}, time.Minute)
	_, err := p.Get(context.Background(), "db/main")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCache_ConcurrentNoRace(t *testing.T) {
	backend := &countingProvider{val: "x", ver: "v1"}
	p := NewCachedProvider(backend, time.Minute)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Get(ctx, "db/main")
		}()
	}
	wg.Wait()
}
