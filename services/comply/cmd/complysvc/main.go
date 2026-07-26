package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"shopass/services/comply/internal/adapters"
	"shopass/services/comply/internal/api"
	"shopass/services/comply/internal/breach"
	"shopass/services/comply/internal/consent"
	"shopass/services/comply/internal/dsar"
	"shopass/services/comply/internal/gating"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
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

	sqlDB, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Error("sql db open", "err", err)
		os.Exit(1)
	}
	defer sqlDB.Close()
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error("sql db ping", "err", err)
		os.Exit(1)
	}

	gatingRegistry := gating.NewRegistry(gating.NewPgRuleSource(sqlDB))
	if err := gatingRegistry.Reload(ctx); err != nil {
		log.Error("gating reload", "err", err)
		os.Exit(1)
	}
	log.Info("gating registry loaded", "vn_voucher_stacking", gatingRegistry.Allow("VN", gating.GateVoucherStacking))

	consentSvc := consent.NewService(consent.NewRepo(sqlDB))
	dsarSvc := dsar.NewService(
		dsar.NewPgRepo(pool),
		adapters.NewUsers(pool),
		adapters.NewTrack(pool),
		adapters.NewBill(pool),
		consentSvc,
	)
	breachSvc := breach.NewService(breach.NewPgRepo(pool))

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, consentSvc, dsarSvc, breachSvc)

	server := &http.Server{
		Addr:              env("COMPLY_ADDR", ":8087"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("complysvc listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("serve", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}
