package gw

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Locale string `json:"locale"`
	Tier   string `json:"tier"`
	jwt.RegisteredClaims
}

type JWKSCache struct {
	mu        sync.Mutex
	keys      map[string]any
	fetch     func() (map[string]any, error)
	ttl       time.Duration
	fetchedAt time.Time
	issuer    string
	audience  string
}

func NewJWKSCache(fetch func() (map[string]any, error), ttl time.Duration) *JWKSCache {
	return &JWKSCache{
		fetch:    fetch,
		ttl:      ttl,
		keys:     make(map[string]any),
		issuer:   "sandeal-auth",
		audience: "sandeal-api",
	}
}

func (c *JWKSCache) Verify(ctx context.Context, raw string) (*Claims, error) {
	if c == nil {
		return nil, errors.New("jwks not configured")
	}
	token, err := jwt.ParseWithClaims(
		raw,
		&Claims{},
		c.keyFunc(ctx),
		jwt.WithIssuer(c.issuer),
		jwt.WithAudience(c.audience),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

func (c *JWKSCache) keyFunc(_ context.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, jwt.ErrTokenUnverifiable
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		if key, ok := c.keys[kid]; ok && !c.expired() {
			return key, nil
		}
		keys, err := c.fetch()
		if err != nil {
			return nil, err
		}
		c.keys = keys
		c.fetchedAt = time.Now()
		if key, ok := c.keys[kid]; ok {
			return key, nil
		}
		return nil, jwt.ErrSignatureInvalid
	}
}

func (c *JWKSCache) expired() bool {
	return c.ttl > 0 && !c.fetchedAt.IsZero() && time.Since(c.fetchedAt) >= c.ttl
}

func jwtVerify(jwks *JWKSCache) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			raw := bearer(r)
			if raw == "" {
				writeJSON(w, 401, errBody("unauthorized"))
				return
			}
			claims, err := jwks.Verify(r.Context(), raw)
			if err != nil {
				writeJSON(w, 401, errBody("unauthorized"))
				return
			}
			r = r.WithContext(withClaims(r.Context(), claims))
			r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10))
			next.ServeHTTP(w, r)
		})
	}
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return h[7:]
}

type contextKey struct{ name string }

var claimsKey = &contextKey{"claims"}

func withClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

func GetClaims(ctx context.Context) *Claims {
	c, _ := ctx.Value(claimsKey).(*Claims)
	return c
}

func isPublic(path string) bool {
	return path == "/v1/health" || path == "/v1/auth/login"
}
