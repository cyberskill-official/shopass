package gw

import (
	"net/http"
	"strings"
)

type WAFConfig struct {
	MaxBodyBytes int64
}

var defaultWAF = WAFConfig{MaxBodyBytes: 1 << 20}

var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true,
	"PATCH": true, "DELETE": true, "OPTIONS": true,
}

func waf(cfg WAFConfig) Middleware {
	if cfg.MaxBodyBytes <= 0 {
		cfg = defaultWAF
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !allowedMethods[r.Method] {
				writeJSON(w, 405, errBody("method_not_allowed"))
				return
			}
			if strings.Contains(r.URL.Path, "..") || looksLikeSQLi(r.URL.RawQuery) {
				writeJSON(w, 400, errBody("bad_request"))
				return
			}
			if r.ContentLength > cfg.MaxBodyBytes {
				writeJSON(w, 413, errBody("body_too_large"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func looksLikeSQLi(s string) bool {
	q := strings.ToLower(s)
	return strings.Contains(q, " union ") ||
		strings.Contains(q, " or 1=1") ||
		strings.Contains(q, "drop table")
}
