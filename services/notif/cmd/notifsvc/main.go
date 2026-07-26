package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/notif/internal/fcm"
	"shopass/services/notif/internal/notif"
	"shopass/services/notif/internal/server"
)

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
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
	port := env("PORT", "8082")

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

	repo := notif.NewRepo(pool)
	srv := server.New(repo, log)

	httpSrv := &http.Server{
		Addr:              ":" + port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go startFCMLoop(ctx, log, repo)

	go func() {
		log.Info("notifsvc started", "port", port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("notifsvc failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func startFCMLoop(ctx context.Context, log *slog.Logger, repo *notif.Repo) {
	projectID := os.Getenv("FCM_PROJECT_ID")
	if projectID == "" {
		log.Warn("FCM_PROJECT_ID unset; notifications stay queued until FCM is configured")
		return
	}
	oauth, err := fcm.NewGoogleOAuthFromEnv(ctx)
	if err != nil {
		log.Error("fcm oauth init", "err", err)
		return
	}
	if oauth == nil {
		log.Warn("FCM credentials unset (FCM_SERVICE_ACCOUNT_JSON or GOOGLE_APPLICATION_CREDENTIALS); dispatcher idle")
		return
	}

	client := fcm.NewClient(projectID, oauth, &http.Client{Timeout: 15 * time.Second})
	dispatcher := fcm.NewDispatcher(client, fcm.RepoAdapter{Repo: repo}, fcm.NewBucket(), 50, 3)
	log.Info("fcm dispatcher started", "project_id", projectID)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := dispatcher.RunOnce(ctx); err != nil {
				log.Error("fcm dispatch", "err", err)
			}
		}
	}
}
