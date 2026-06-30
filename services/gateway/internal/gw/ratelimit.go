package gw

import (
	"context"
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
				// Fail-open strategy if Redis fails
				next.ServeHTTP(w, r)
				return
			}

			vals := res.([]interface{})
			allowed := vals[0].(int64) == 1
			retryAfter := vals[1].(int64)

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

	if r.URL.Path == "/v1/auth/login" {
		limit = 5 // stricter for login
	}

	if claims, ok := r.Context().Value(claimsKey{}).(*Claims); ok {
		return "rl:user:" + strconv.FormatInt(claims.UserID, 10), limit
	}

	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return "rl:ip:" + ip, limit
}
