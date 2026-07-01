package gw

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

type RedisClient interface {
	AllowN(ctx context.Context, key string, limit int) (bool, int, error)
}

type RateLimiter struct {
	rdb RedisClient
}

func rateLimit(rdb RedisClient) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				next.ServeHTTP(w, r)
				return
			}
			key, limit := bucketKeyAndLimit(r)
			ok, retryAfter, err := rdb.AllowN(r.Context(), key, limit)
			if err == nil && !ok {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				writeJSON(w, 429, errBody("rate_limited"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bucketKeyAndLimit(r *http.Request) (string, int) {
	if claims := GetClaims(r.Context()); claims != nil {
		return "rl:user:" + strconv.FormatInt(claims.UserID, 10), 100
	}
	if r.URL.Path == "/v1/auth/login" {
		return "rl:ip:" + clientIP(r), 5
	}
	return "rl:ip:" + clientIP(r), 20
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	ip := r.RemoteAddr
	if host, _, ok := strings.Cut(ip, ":"); ok {
		return host
	}
	return ip
}

type mockRedis struct{}

func (m *mockRedis) AllowN(_ context.Context, _ string, _ int) (bool, int, error) {
	return true, 0, nil
}
