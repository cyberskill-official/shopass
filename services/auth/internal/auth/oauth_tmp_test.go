package auth

import (
	"context"
	"testing"
	"time"
)

func TestMemTmpStore_OneTimeTake(t *testing.T) {
	s := NewMemTmpStore()
	ctx := context.Background()
	s.Put(ctx, "oauth:st1", oauthTmp{Verifier: "v", Provider: "google", Nonce: "n"}, time.Minute)

	got, ok := s.Take(ctx, "oauth:st1")
	if !ok || got.Verifier != "v" || got.Nonce != "n" {
		t.Fatalf("first Take failed: %+v ok=%v", got, ok)
	}
	// Second take must fail: entries are single-use (§1 #11).
	if _, ok := s.Take(ctx, "oauth:st1"); ok {
		t.Fatal("second Take must miss (one-time use)")
	}
}

func TestMemTmpStore_Expiry(t *testing.T) {
	s := NewMemTmpStore()
	now := time.Unix(1_000_000, 0)
	s.now = func() time.Time { return now }
	ctx := context.Background()
	s.Put(ctx, "k", oauthTmp{Verifier: "v"}, 5*time.Minute)

	now = now.Add(6 * time.Minute) // past TTL
	if _, ok := s.Take(ctx, "k"); ok {
		t.Fatal("expired entry must be treated as absent")
	}
}

func TestMemTmpStore_MissingKey(t *testing.T) {
	s := NewMemTmpStore()
	if _, ok := s.Take(context.Background(), "nope"); ok {
		t.Fatal("missing key must return ok=false")
	}
}
