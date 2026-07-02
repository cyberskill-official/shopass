package auth

import (
	"context"
	"sync"
	"time"
)

// oauthTmp is the short-lived per-flow state stored between StartOAuth and the
// callback: the PKCE verifier, the target provider, and the OIDC nonce.
type oauthTmp struct {
	Verifier string
	Provider string
	Nonce    string
}

// TmpStore holds oauthTmp values keyed by the flow state, with a short TTL and
// one-time reads (§1 #11). Take both returns and removes the entry so a delayed
// or replayed callback cannot reuse it. A production deployment backs this with
// Redis; MemTmpStore is the in-process implementation.
type TmpStore interface {
	Put(ctx context.Context, key string, v oauthTmp, ttl time.Duration)
	Take(ctx context.Context, key string) (oauthTmp, bool)
}

type tmpEntry struct {
	v      oauthTmp
	expiry time.Time
}

// MemTmpStore is an in-memory, mutex-guarded TmpStore.
type MemTmpStore struct {
	mu  sync.Mutex
	m   map[string]tmpEntry
	now func() time.Time
}

func NewMemTmpStore() *MemTmpStore {
	return &MemTmpStore{m: make(map[string]tmpEntry), now: time.Now}
}

func (s *MemTmpStore) Put(_ context.Context, key string, v oauthTmp, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = tmpEntry{v: v, expiry: s.now().Add(ttl)}
}

// Take returns the value once and deletes it. An expired entry is treated as
// absent (and removed).
func (s *MemTmpStore) Take(_ context.Context, key string) (oauthTmp, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[key]
	if !ok {
		return oauthTmp{}, false
	}
	delete(s.m, key) // one-time use
	if s.now().After(e.expiry) {
		return oauthTmp{}, false
	}
	return e.v, true
}
