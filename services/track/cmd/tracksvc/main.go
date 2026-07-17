// Command tracksvc serves product tracking, wishlist, and alert-rule endpoints.
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
	"time"

	_ "github.com/lib/pq"

	"shopass/services/track/internal/api"
	"shopass/services/track/internal/priceclient"
	"shopass/services/track/internal/priming"
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
	priceClient, err := priceclient.New(
		env("PRICE_INTERNAL_URL", "http://pricesvc:8081"),
		os.Getenv("PRICE_INTERNAL_SERVICE_TOKEN"),
		5*time.Second,
	)
	if err != nil {
		log.Error("price client configuration", "err", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	wh := api.NewWishlistHandler(track.NewWishlistRepo(db))
	ah := api.NewAlertRuleHandler(track.NewAlertRuleRepo(db))
	th := api.NewHandler(
		track.NewShopeeVNPlatformMap(),
		priceClient,
		track.NewRepo(db),
		priming.NewNoopQueue(),
	)

	mux := http.NewServeMux()
	// Register the new track endpoint together with the existing wishlist and
	// alert routes. This preserves their public surface while wiring the beta
	// flow through pricesvc rather than a cross-service database write.
	api.RegisterRoutes(mux, th, wh, ah)
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
