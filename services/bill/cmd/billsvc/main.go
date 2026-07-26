package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/bill/internal/api"
	"shopass/services/bill/internal/bill"
	"shopass/services/bill/internal/gating"
	"shopass/services/bill/internal/pay"
	"shopass/services/bill/internal/referral"
	trustfraud "shopass/services/trust/fraud"
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
				r = r.WithContext(context.WithValue(r.Context(), "user_id", id))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	log := slog.Default()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	addr := env("BILL_ADDR", ":8086")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Error("db ping", "err", err)
		os.Exit(1)
	}

	repo := bill.NewRepo(pool, nil)
	secrets := pay.NewEnvSecrets()
	reg := pay.NewRegistry()
	reg.Register(pay.NewMoMoSandbox(secrets))
	reg.Register(pay.NewZaloPaySandbox(secrets))
	reg.Register(pay.NewVNPaySandbox(secrets))
	reg.Register(pay.NewVietQR())

	checkout := api.NewHandler(api.NewSQLPlanCatalog(repo), reg, api.NewCheckoutPayments(repo))
	ipn := api.NewIPNHandler(repo, repo, secrets)
	waitlist := api.NewWaitlistHandler(pool)
	referralRepo := referral.NewPGRepo(pool)
	fraudEngine := trustfraud.NewEngine(
		trustfraud.DefaultConfig(),
		trustfraud.NewPGEventCounter(pool),
		trustfraud.NewPGClusterSizer(pool),
		trustfraud.NewPGSignalStore(pool),
		trustfraud.NewPGRewardHolder(pool, log),
	)
	fraudAssessor := referral.AssessorFunc(func(ctx context.Context, userID int64, extras map[string]any) error {
		_, err := fraudEngine.Assess(ctx, userID, extras)
		return err
	})
	referralH := api.NewReferralHandler(
		referralRepo,
		log,
		referral.WithReferralFraud(trustfraud.NewPGAccountLinkStore(pool), fraudAssessor),
	)
	gate := gating.NewGate(gating.NewSQLRepo(pool), repo, gating.NewSQLPlanCatalog(pool))
	gatingH := api.NewGatingHandler(gate, os.Getenv("BILL_INTERNAL_SERVICE_TOKEN"))

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, checkout, ipn, waitlist, referralH)
	gatingH.RegisterRoutes(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           userMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("billsvc listening", "addr", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("serve", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}
