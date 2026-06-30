package gw

import (
	"net/http"
	"strings"
)

type WAFConfig struct {
	MaxBodySize int64
}

func waf(cfg WAFConfig) func(http.Handler) http.Handler {
	if cfg.MaxBodySize == 0 {
		cfg.MaxBodySize = 1 << 20 // 1 MiB default
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Method allowlist
			switch r.Method {
			case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
				// Allowed
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			// 2. Body-size cap
			if r.ContentLength > cfg.MaxBodySize {
				http.Error(w, "Payload Too Large", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxBodySize)

			// 3. Path traversal & SQLi thô (very basic pattern check)
			if strings.Contains(r.URL.Path, "../") || strings.Contains(r.URL.RawQuery, "../") {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			// A very naive SQLi check for demo/compliance with spec "pattern SQLi thô"
			// Just checking some obvious bad patterns in query
			upperQuery := strings.ToUpper(r.URL.RawQuery)
			if strings.Contains(upperQuery, "UNION SELECT") || strings.Contains(upperQuery, "OR 1=1") {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
