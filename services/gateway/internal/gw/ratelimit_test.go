package gw

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type mockRedis struct {
	counts map[string]int
}

func (m *mockRedis) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if m.counts == nil {
		m.counts = make(map[string]int)
	}
	key := keys[0]
	limit := args[0].(int)

	curr := m.counts[key]
	if curr >= limit {
		return redis.NewCmdResult([]interface{}{int64(0), int64(12)}, nil)
	}
	m.counts[key]++
	return redis.NewCmdResult([]interface{}{int64(1), int64(0)}, nil)
}

func TestRateLimit_PerIP_429(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}, Redis: &mockRedis{}}
	h := NewHandler(deps)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/v1/auth/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		require.Equal(t, 200, rr.Code, "request %d should pass", i)
	}

	req := httptest.NewRequest("POST", "/v1/auth/login", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, 429, rr.Code)
	require.Equal(t, "12", rr.Header().Get("Retry-After"))
}

func TestRateLimit_PerUser_Isolated(t *testing.T) {
	deps := Deps{JWKS: &mockJWKS{}, Redis: &mockRedis{}}
	h := NewHandler(deps)

	// User 1 uses all 100 limit on /v1/track
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/v1/track", nil)
		req.Header.Set("Authorization", "Bearer valid-1")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		require.Equal(t, 200, rr.Code)
	}

	// User 1 gets 429
	req1 := httptest.NewRequest("GET", "/v1/track", nil)
	req1.Header.Set("Authorization", "Bearer valid-1")
	rr1 := httptest.NewRecorder()
	h.ServeHTTP(rr1, req1)
	require.Equal(t, 429, rr1.Code)

	// User 2 still 200
	req2 := httptest.NewRequest("GET", "/v1/track", nil)
	req2.Header.Set("Authorization", "Bearer valid-2")
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	require.Equal(t, 200, rr2.Code)
}
