// Command tracksvc serves the track HTTP endpoints (wishlists and alert rules).
// It sits behind the API gateway: the gateway verifies the JWT and forwards the
// caller id as X-User-Id, which this service trusts (it does not re-verify).
package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"shopass/services/track/internal/api"
	"shopass/services/track/internal/track"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// userMiddleware injects the gateway-provided X-User-Id into the request context
// under the key the handlers read (UserID -> ctx.Value("user_id")).
func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw := r.Header.Get("X-User-Id"); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
				r = r.WithContext(context.WithValue(r.Context(), "user_id", id))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	log := slog.Default()
	dbURL := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shopass?sslmode=disable")
	addr := env("TRACK_ADDR", ":8083")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	wh := api.NewWishlistHandler(track.NewWishlistRepo(db))

	mux := http.NewServeMux()
	wh.RegisterRoutes(mux) // /v1/wishlists (list/create) + items
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	log.Info("tracksvc listening", "addr", addr)
	if err := http.ListenAndServe(addr, userMiddleware(mux)); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
