package gw

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var testKey *rsa.PrivateKey

func init() {
	var err error
	testKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
}

type memoryRedis struct {
	mu     sync.Mutex
	counts map[string]int
	keys   []string
}

func newMemoryRedis() *memoryRedis {
	return &memoryRedis{counts: make(map[string]int)}
}

func (m *memoryRedis) AllowN(_ context.Context, key string, limit int) (bool, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counts[key]++
	m.keys = append(m.keys, key)
	if m.counts[key] > limit {
		return false, 30, nil
	}
	return true, 0, nil
}

func (m *memoryRedis) saw(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range m.keys {
		if k == key {
			return true
		}
	}
	return false
}

type captureHandler struct {
	last http.Header
}

func (c *captureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.last = r.Header.Clone()
	w.WriteHeader(http.StatusOK)
}

func signToken(t *testing.T, claims jwt.Claims, kid string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if kid != "" {
		token.Header["kid"] = kid
	}
	s, err := token.SignedString(testKey)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func validClaims(userID int64) Claims {
	return Claims{
		UserID: userID,
		Locale: "vi-VN",
		Tier:   "free",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sandeal-auth",
			Audience:  jwt.ClaimStrings{"sandeal-api"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
}

func newTestHandler(rest http.Handler, rdb RedisClient) http.Handler {
	return NewHandler(Deps{
		REST:  rest,
		Redis: rdb,
		JWKS: NewJWKSCache(func() (map[string]any, error) {
			return map[string]any{"test-key": &testKey.PublicKey}, nil
		}, time.Minute),
	})
}

func do(h http.Handler, method, path, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func doIP(h http.Handler, method, path, ip string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	r.RemoteAddr = ip
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestPublicRoute_NoJWTRequired(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	rr := do(h, "GET", "/v1/health", "")
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestProtectedRoute_NoToken_401(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	rr := do(h, "GET", "/v1/track", "")
	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestProtectedRoute_MissingKID_401NoPanic(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	tok := signToken(t, validClaims(1), "")
	rr := do(h, "GET", "/v1/track", tok)
	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWT_Expired_401(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	claims := validClaims(1)
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))
	tok := signToken(t, claims, "test-key")
	rr := do(h, "GET", "/v1/track", tok)
	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWT_BadAudience_401(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	claims := validClaims(1)
	claims.Audience = jwt.ClaimStrings{"other-service"}
	tok := signToken(t, claims, "test-key")
	rr := do(h, "GET", "/v1/track", tok)
	if rr.Code != 401 {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestJWT_Valid_PropagatesUserID(t *testing.T) {
	upstream := &captureHandler{}
	h := newTestHandler(upstream, newMemoryRedis())
	tok := signToken(t, validClaims(90112), "test-key")
	rr := do(h, "GET", "/v1/track", tok)
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if got := upstream.last.Get("X-User-Id"); got != "90112" {
		t.Fatalf("expected X-User-Id 90112, got %q", got)
	}
}

func TestRateLimit_PerIP_429(t *testing.T) {
	rdb := newMemoryRedis()
	h := newTestHandler(&captureHandler{}, rdb)
	for i := 0; i < 5; i++ {
		if got := doIP(h, "POST", "/v1/auth/login", "1.2.3.4:5678").Code; got != 200 {
			t.Fatalf("request %d expected 200, got %d", i, got)
		}
	}
	over := doIP(h, "POST", "/v1/auth/login", "1.2.3.4:5678")
	if over.Code != 429 {
		t.Fatalf("expected 429, got %d", over.Code)
	}
	if over.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After")
	}
}

func TestRateLimit_PerUser_Isolated(t *testing.T) {
	rdb := newMemoryRedis()
	h := newTestHandler(&captureHandler{}, rdb)
	tokA := signToken(t, validClaims(1), "test-key")
	tokB := signToken(t, validClaims(2), "test-key")
	for i := 0; i < 100; i++ {
		do(h, "GET", "/v1/track", tokA)
	}
	if got := do(h, "GET", "/v1/track", tokB).Code; got != 200 {
		t.Fatalf("expected user B 200, got %d", got)
	}
	if !rdb.saw("rl:user:1") || !rdb.saw("rl:user:2") {
		t.Fatalf("expected per-user rate-limit keys, saw %#v", rdb.keys)
	}
}

func TestWAF_PathTraversal_400(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	rr := do(h, "GET", "/v1/../etc/passwd", "")
	if rr.Code != 400 {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	h := newTestHandler(&captureHandler{}, newMemoryRedis())
	rr := do(h, "TRACE", "/v1/track", "")
	if rr.Code != 405 {
		t.Fatalf("expected 405, got %d", rr.Code)
	}
}

func TestRequestID_Generated(t *testing.T) {
	upstream := &captureHandler{}
	h := newTestHandler(upstream, newMemoryRedis())
	rr := do(h, "GET", "/v1/health", "")
	if rr.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header in response")
	}
	if upstream.last.Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id propagated to upstream")
	}
}

func TestWSHandshakeRequiresJWT(t *testing.T) {
	h := NewHandler(Deps{
		JWKS: NewJWKSCache(func() (map[string]any, error) {
			return map[string]any{"test-key": &testKey.PublicKey}, nil
		}, time.Minute),
	})
	if got := do(h, "GET", "/ws", "").Code; got != 401 {
		t.Fatalf("expected 401, got %d", got)
	}
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, validClaims(1), "test-key"))
	req.Header.Set("Upgrade", "websocket")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusSwitchingProtocols {
		t.Fatalf("expected 101, got %d", rr.Code)
	}
}
