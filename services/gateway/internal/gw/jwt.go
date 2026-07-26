package gw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Locale string `json:"locale"`
	Tier   string `json:"tier"`
	jwt.RegisteredClaims
}

type claimsKey struct{}

type JWKSCache interface {
	Verify(ctx context.Context, tokenString string) (*Claims, error)
}

func jwtVerify(jwks JWKSCache) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublic(r) {
				next.ServeHTTP(w, r)
				return
			}
			if jwks == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "authentication unavailable"})
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			raw := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwks.Verify(r.Context(), raw)
			if err != nil {
				if errors.Is(err, ErrJWKSUnavailable) {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "authentication unavailable"})
					return
				}
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			r = r.WithContext(ctx)
			r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10))
			r.Header.Set("X-User-Locale", claims.Locale)
			r.Header.Set("X-User-Tier", claims.Tier)

			next.ServeHTTP(w, r)
		})
	}
}

func isPublic(r *http.Request) bool {
	if r.URL.Path == "/healthz" || r.URL.Path == "/v1/health" {
		return true
	}
	// Payment gateway callbacks are signature-verified by billsvc, not JWT.
	if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/billing/ipn/") {
		return true
	}
	// Premium waitlist capture (R39) — public, rate-limited with other POSTs.
	if r.Method == http.MethodPost && r.URL.Path == "/v1/leads/waitlist" {
		return true
	}
	// Fake-sale checker lead magnet (R43) — public, rate-limited.
	if r.Method == http.MethodPost && r.URL.Path == "/v1/tools/fake-sale-check" {
		return true
	}
	if !strings.HasPrefix(r.URL.Path, "/v1/auth/") {
		return false
	}
	if r.Method == http.MethodPost {
		switch r.URL.Path {
		case "/v1/auth/login", "/v1/auth/register", "/v1/auth/refresh", "/v1/auth/logout":
			return true
		}
	}
	// OAuth starts at a user action and callbacks originate with the provider;
	// neither carries an application access token yet.
	return r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/auth/oauth/")
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
