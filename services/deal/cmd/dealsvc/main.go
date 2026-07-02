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

	"shopass/services/deal/internal/batch"
)

// pg_try_advisory_lock key for nightly score to prevent overlapping runs.
const nightlyScoreLockKey = 1001
const refreshPriorsLockKey = 1002

type httpNotif struct {
	url string
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
	
	resp, err := http.DefaultClient.Do(req)
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
	b := batch.New(pool, log, &httpNotif{url: notifURL})

	// Setup Cron with Asia/Ho_Chi_Minh timezone
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		log.Error("failed to load timezone", "err", err)
		os.Exit(1)
	}
	c := cron.New(cron.WithLocation(loc))

	// Register nightly score (FR-DEAL-006): runs at 02:00
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
}
