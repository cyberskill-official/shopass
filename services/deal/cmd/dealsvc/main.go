package main

import (
	"context"
	"log/slog"
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

type dummyNotif struct{}

func (d *dummyNotif) Enqueue(ctx context.Context, item batch.NotifItem) error {
	slog.Info("dummy enqueue", "item", item)
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

	// Initialize batches
	b := batch.New(pool, log, &dummyNotif{})

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
