// Command pricesvc serves the price read (history) and internal write (ingest)
// HTTP endpoints backed by Postgres/TimescaleDB.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/price/internal/api"
	"shopass/services/price/internal/price"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.Default()
	dbURL := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shopass_price?sslmode=disable")
	addr := env("PRICE_ADDR", ":8081")
	serviceToken := os.Getenv("PRICE_INTERNAL_SERVICE_TOKEN")
	if strings.TrimSpace(serviceToken) == "" {
		log.Error("PRICE_INTERNAL_SERVICE_TOKEN is required for private product upserts")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	productRepo := price.NewRepo(pool)
	mux := http.NewServeMux()
	api.NewHandler(productRepo).RegisterRoutes(mux)                                     // GET price-history
	api.NewIngestHandler(price.NewSnapshotRepo(pool), serviceToken).RegisterRoutes(mux) // POST snapshots
	api.NewProductUpsertHandler(productRepo, serviceToken).RegisterRoutes(mux)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("pricesvc listening", "addr", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}
