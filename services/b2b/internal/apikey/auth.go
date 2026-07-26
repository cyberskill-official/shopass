package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidKey   = errors.New("apikey: invalid")
	ErrUnauthorized = errors.New("apikey: unauthorized")
)

type APIKey struct {
	ID           int64
	Prefix       string
	SecretHash   string
	OrgName      string
	Tier         string
	RatePerMin   int
	MonthlyQuota int
	Revoked      bool
}

type KeyStore interface {
	GetByPrefix(ctx context.Context, prefix string) (*APIKey, error)
	SetRevoked(ctx context.Context, prefix string, revoked bool) error
}

// NewKey returns prefix.secret and the hash to store (never store cleartext secret).
func NewKey() (prefix, secret, hash string, err error) {
	var b [24]byte
	if _, err = rand.Read(b[:]); err != nil {
		return "", "", "", err
	}
	prefix = hex.EncodeToString(b[:8])
	secret = hex.EncodeToString(b[8:])
	sum := sha256.Sum256([]byte(secret))
	hash = hex.EncodeToString(sum[:])
	return prefix, secret, hash, nil
}

func ParsePresented(raw string) (prefix, secret string, err error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidKey
	}
	return parts[0], parts[1], nil
}

func Verify(secret, storedHash string) bool {
	sum := sha256.Sum256([]byte(secret))
	got := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHash)) == 1
}

func Format(prefix, secret string) string {
	return fmt.Sprintf("%s.%s", prefix, secret)
}

// Auth authenticates X-API-Key with optional short-lived prefix cache (<=60s).
type Auth struct {
	Store    KeyStore
	CacheTTL time.Duration
	mu       sync.Mutex
	cache    map[string]cacheEntry
	now      func() time.Time
}

type cacheEntry struct {
	key       *APIKey
	expiresAt time.Time
}

func NewAuth(store KeyStore) *Auth {
	return &Auth{
		Store:    store,
		CacheTTL: 60 * time.Second,
		cache:    make(map[string]cacheEntry),
		now:      time.Now,
	}
}

func (a *Auth) Authenticate(ctx context.Context, raw string) (*APIKey, error) {
	prefix, secret, err := ParsePresented(raw)
	if err != nil {
		return nil, ErrUnauthorized
	}
	k, err := a.getCached(ctx, prefix)
	if err != nil || k == nil || k.Revoked {
		return nil, ErrUnauthorized
	}
	if !Verify(secret, k.SecretHash) {
		return nil, ErrUnauthorized
	}
	return k, nil
}

func (a *Auth) getCached(ctx context.Context, prefix string) (*APIKey, error) {
	now := a.now()
	a.mu.Lock()
	if e, ok := a.cache[prefix]; ok && now.Before(e.expiresAt) {
		cp := *e.key
		a.mu.Unlock()
		return &cp, nil
	}
	a.mu.Unlock()

	k, err := a.Store.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}
	ttl := a.CacheTTL
	if ttl <= 0 || ttl > 60*time.Second {
		ttl = 60 * time.Second
	}
	a.mu.Lock()
	a.cache[prefix] = cacheEntry{key: k, expiresAt: now.Add(ttl)}
	a.mu.Unlock()
	if k == nil {
		return nil, nil
	}
	cp := *k
	return &cp, nil
}

func (a *Auth) Revoke(ctx context.Context, prefix string) error {
	if err := a.Store.SetRevoked(ctx, prefix, true); err != nil {
		return err
	}
	a.mu.Lock()
	delete(a.cache, prefix)
	a.mu.Unlock()
	return nil
}

// AdvanceCache is a test hook to expire the in-memory prefix cache.
func (a *Auth) AdvanceCache(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	base := a.now()
	a.now = func() time.Time { return base.Add(d) }
}

// MemoryKeyStore for tests / noop.
type MemoryKeyStore struct {
	mu       sync.Mutex
	byPrefix map[string]*APIKey
}

func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{byPrefix: make(map[string]*APIKey)}
}

func (m *MemoryKeyStore) Put(k *APIKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *k
	m.byPrefix[k.Prefix] = &cp
}

func (m *MemoryKeyStore) GetByPrefix(_ context.Context, prefix string) (*APIKey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.byPrefix[prefix]
	if !ok {
		return nil, nil
	}
	cp := *k
	return &cp, nil
}

func (m *MemoryKeyStore) SetRevoked(_ context.Context, prefix string, revoked bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k, ok := m.byPrefix[prefix]
	if !ok {
		return ErrUnauthorized
	}
	k.Revoked = revoked
	return nil
}
