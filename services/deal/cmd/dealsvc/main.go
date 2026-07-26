package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"

	"shopass/services/deal/internal/api"
	"shopass/services/deal/internal/batch"
)

// pg_try_advisory_lock key for nightly score to prevent overlapping runs.
const nightlyScoreLockKey = 1001
const refreshPriorsLockKey = 1002

type httpNotif struct {
	url    string
	client *http.Client
}

func (h *httpNotif) Enqueue(ctx context.Context, item batch.NotifItem) error {
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("notifsvc returned status %d", resp.StatusCode)
	}
	return nil
}

func main() {
	log := slog.Default()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/shopass_deal?sslmode=disable"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	notifURL := os.Getenv("NOTIFSVC_URL")
	if notifURL == "" {
		notifURL = "http://localhost:8081/notify"
	}

	// Initialize batches
	b := batch.New(pool, log, &httpNotif{
		url:    notifURL,
		client: &http.Client{Timeout: 10 * time.Second},
	})

	// RUN_ONCE=1 triggers a single nightly-score pass and exits (manual/smoke trigger).
	if os.Getenv("RUN_ONCE") == "1" {
		var acquired bool
		if err := pool.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", nightlyScoreLockKey).Scan(&acquired); err != nil || !acquired {
			log.Error("run-once: nightly score lock not acquired")
			os.Exit(1)
		}
		defer pool.Exec(ctx, "SELECT pg_advisory_unlock($1)", nightlyScoreLockKey)
		if err := b.RunNightlyScore(ctx, time.Now()); err != nil {
			log.Error("run-once nightly score failed", "err", err)
			os.Exit(1)
		}
		log.Info("run-once nightly score completed")
		return
	}

	// Serve the chart feed over HTTP (TASK-DEAL-003), read by the web BFF. This
	// runs alongside the nightly cron below.
	dealAddr := os.Getenv("DEAL_ADDR")
	if dealAddr == "" {
		dealAddr = ":8082"
	}
	repo := &chartRepo{pool: pool}
	dealSvc := &dealService{pool: pool}
	chartHandler := api.NewHandler(repo, dealSvc)
	checkHandler := api.NewFakeSaleCheckHandler(repo, repo, dealSvc)
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, chartHandler, checkHandler)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	httpSrv := &http.Server{
		Addr:              dealAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		log.Info("dealsvc chart http listening", "addr", dealAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("chart http serve", "err", err)
		}
	}()

	// Setup Cron with Asia/Ho_Chi_Minh timezone
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Error("failed to load timezone", "err", err)
		os.Exit(1)
	}
	c := cron.New(cron.WithLocation(loc))

	// Register nightly score (TASK-DEAL-006): runs at 02:00
	_, err = c.AddFunc("0 2 * * *", func() {
		// Use advisory lock to prevent overlapping
		var acquired bool
		err := pool.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", nightlyScoreLockKey).Scan(&acquired)
		if err != nil || !acquired {
			log.Info("nightly score lock not acquired, skipping")
			return
		}
		defer pool.Exec(ctx, "SELECT pg_advisory_unlock($1)", nightlyScoreLockKey)

		log.Info("starting nightly score batch")
		if err := b.RunNightlyScore(ctx, time.Now()); err != nil {
			log.Error("nightly score failed", "err", err)
		} else {
			log.Info("nightly score completed")
		}
	})
	if err != nil {
		log.Error("failed to register cron for nightly score", "err", err)
		os.Exit(1)
	}

	c.Start()
	log.Info("dealsvc cron started")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Info("shutting down")
	c.Stop()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = httpSrv.Shutdown(shutdownCtx)
}
