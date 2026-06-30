package gw

import (
	"context"
	"encoding/json"
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
			if isPublic(r.URL.Path) {
				next.ServeHTTP(w, r)
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
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			r = r.WithContext(ctx)
			r.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10))

			next.ServeHTTP(w, r)
		})
	}
}

func isPublic(path string) bool {
	return path == "/v1/health" || path == "/v1/auth/login"
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
