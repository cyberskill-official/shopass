package gw

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

const tokenBucketScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local current = redis.call("GET", key)
if current and tonumber(current) >= limit then
    return {0, redis.call("TTL", key)}
end
redis.call("INCR", key)
if not current then
    redis.call("EXPIRE", key, ARGV[2])
end
return {1, 0}
`

func rateLimit(rdb RedisClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				next.ServeHTTP(w, r)
				return
			}

			key, limit := bucketKeyAndLimit(r)
			window := 60 // 60 seconds TTL

			cmd := rdb.Eval(r.Context(), tokenBucketScript, []string{key}, limit, window)
			res, err := cmd.Result()
			if err != nil {
				// Login throttling is a security boundary. A broken limiter must not
				// silently turn into unlimited credential guessing.
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "rate limiter unavailable"})
				return
			}

			vals, ok := res.([]interface{})
			if !ok || len(vals) != 2 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "rate limiter unavailable"})
				return
			}
			allowedValue, allowedOK := vals[0].(int64)
			retryAfter, retryOK := vals[1].(int64)
			if !allowedOK || !retryOK {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "rate limiter unavailable"})
				return
			}
			allowed := allowedValue == 1

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limited"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bucketKeyAndLimit(r *http.Request) (string, int) {
	limit := 100 // default

	switch r.URL.Path {
	case "/v1/auth/login":
		limit = 5 // stricter for credential guessing
	case "/v1/auth/refresh":
		limit = 10 // session restore / refresh storms
	}

	if claims, ok := r.Context().Value(claimsKey{}).(*Claims); ok {
		return "rl:user:" + strconv.FormatInt(claims.UserID, 10), limit
	}

	return "rl:ip:" + clientIP(r), limit
}

// clientIP accepts X-Real-IP only after the request has crossed the private
// Caddy -> web/gateway boundary. Caddy overwrites this header with
// {remote_host}; the gateway itself has no public port. A syntactically invalid
// header is ignored so a malformed value cannot splinter rate-limit buckets.
func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Real-IP")); net.ParseIP(forwarded) != nil {
		return forwarded
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	if net.ParseIP(r.RemoteAddr) != nil {
		return r.RemoteAddr
	}
	return "unknown"
}
