// Command affilsvc serves affiliate link + postback + cashback summary endpoints.
// Gateway verifies JWT and forwards X-User-Id (trusted here; not re-verified).
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/affil/internal/affil"
	"shopass/services/affil/internal/api"
	"shopass/services/affil/internal/auth"
	"shopass/services/affil/internal/cashback"
	trustpayout "shopass/services/trust/payout"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if raw := r.Header.Get("X-User-Id"); raw != "" {
			if id, err := strconv.ParseInt(raw, 10, 64); err == nil {
				r = r.WithContext(auth.WithUserID(r.Context(), id))
			}
		}
		next.ServeHTTP(w, r)
	})
}

type noopMetrics struct{}

func (noopMetrics) LinkRejected(string)       {}
func (noopMetrics) LinkCreated(int16, string) {}

func main() {
	log := slog.Default()
	dbURL := env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shopass?sslmode=disable")
	addr := env("AFFIL_ADDR", ":8086")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Error("db open", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	repo := affil.NewRepo(pool)
	cbStore := cashback.NewPGStore(pool)
	trustSvc := &trustpayout.Service{
		Cfg:   trustpayout.DefaultConfig(),
		Store: trustpayout.NewPGStore(pool),
	}
	ledger := &cashback.Ledger{
		Cfg:   cashback.DefaultConfig(),
		Store: cbStore,
		Hold:  trustpayout.HoldChecker{Svc: trustSvc},
		Payer: cashback.NewVietQRStub(log),
	}

	// Minimal product/network stubs for link endpoint when not fully configured.
	linkHandler := api.NewHandler(nilProducts{}, nilNetworks{}, repo, noopMetrics{})
	postback := api.NewPostbackHandler(repo, envSecrets{}).
		WithPayoutHolds(trustpayout.CreateHoldAdapter{Svc: trustSvc}).
		WithCashback(ledger)
	cashbackHandler := api.NewCashbackHandler(ledger)

	mux := http.NewServeMux()
	api.RegisterRoutesWithCashback(mux, linkHandler, postback, cashbackHandler)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	// Periodic release + payout sweep (TRUST-005 window + threshold).
	go func() {
		t := time.NewTicker(1 * time.Hour)
		defer t.Stop()
		for range t.C {
			n, err := cashback.NewReleaser(ledger).ReleaseDue(context.Background(), time.Now().UTC())
			if err != nil {
				log.Error("cashback release", "err", err)
				continue
			}
			if n > 0 {
				log.Info("cashback released", "count", n)
			}
		}
	}()

	server := &http.Server{
		Addr:              addr,
		Handler:           userMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("affilsvc listening", "addr", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Error("serve", "err", err)
		os.Exit(1)
	}
}

type envSecrets struct{}

func (envSecrets) Get(ctx context.Context, keyPath string) (string, error) {
	return os.Getenv(keyPath), nil
}

type nilProducts struct{}

func (nilProducts) Get(ctx context.Context, productID int64) (api.Product, bool) {
	return nil, false
}

type nilNetworks struct{}

func (nilNetworks) TemplateFor(platformID int16) (string, affil.NetworkTemplate, bool) {
	return "", affil.NetworkTemplate{}, false
}
