// Command scrapesvc wires the real Shopee adapter to the price ingest endpoint.
//
// With DATABASE_URL set it uses the durable Postgres queue (TASK-SCRAPE-001):
// SCRAPE_SEED jobs are enqueued, then all due jobs for the platform are drained.
// Without a database it falls back to an in-memory one-shot over SCRAPE_SEED.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/scrape/internal/adapters/shopee"
	"shopass/services/scrape/internal/feeder"
	"shopass/services/scrape/internal/memqueue"
	"shopass/services/scrape/internal/orchestrator"
	"shopass/services/scrape/internal/pgqueue"
	"shopass/services/scrape/internal/priceclient"
)

const shopeePlatformID int16 = 1

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	shopeeBase := env("SHOPEE_BASE_URL", "https://shopee.vn")
	priceBase := env("PRICE_BASE_URL", "http://localhost:8081")
	dbURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	// http.Transport uses ProxyFromEnvironment, so HTTPS_PROXY wires the
	// residential proxy required by TASK-SCRAPE-002.
	httpClient := &http.Client{Timeout: 15 * time.Second}
	adapter := shopee.NewShopeeAdapter(shopeeBase, httpClient, nil)
	pc := priceclient.New(priceBase, 10*time.Second)
	cfg := orchestrator.Config{MaxConcurrency: map[int16]int{shopeePlatformID: 4}, MaxAttempts: 3, BackoffBaseMs: 200}
	jobs := parseSeed(env("SCRAPE_SEED", ""))

	if dbURL != "" {
		pool, err := pgxpool.New(ctx, dbURL)
		if err != nil {
			log.Error("db connect", "err", err.Error())
			os.Exit(1)
		}
		defer pool.Close()
		q := pgqueue.New(pool, 2*time.Minute)
		orch := orchestrator.NewPool(cfg, pc, q)
		orch.RegisterAdapter(adapter)
		orch.SetRescheduler(q) // persist tier-based next_run after each scrape

		if n, err := feeder.SyncJobs(ctx, pool); err != nil {
			log.Error("feeder sync", "err", err.Error())
		} else if n > 0 {
			log.Info("registered new products into scrape_job", "count", n)
		}

		for _, j := range jobs {
			if err := q.Enqueue(ctx, j); err != nil {
				log.Error("enqueue", "product_id", j.ProductID, "err", err.Error())
			}
		}
		var succeeded, deferred, failed int
		for {
			j, ok, err := q.Claim(ctx, shopeePlatformID)
			if err != nil {
				log.Error("claim", "err", err.Error())
				break
			}
			if !ok {
				break
			}
			result, err := orch.ProcessJob(ctx, j)
			if err != nil {
				log.Error("scrape job outcome not persisted", "product_id", j.ProductID, "err", err.Error())
				continue
			}
			switch result.Outcome {
			case orchestrator.JobSucceeded:
				succeeded++
				log.Info("scraped and posted (durable queue)", "product_id", j.ProductID)
			case orchestrator.JobDeferred:
				deferred++
				log.Warn("scrape deferred for retry", "product_id", j.ProductID, "attempt", result.Attempts, "retry_at", result.RetryAt, "err", result.Cause)
			case orchestrator.JobFailed:
				failed++
				log.Error("scrape failed permanently", "product_id", j.ProductID, "attempt", result.Attempts, "err", result.Cause)
			default:
				log.Error("scrape returned unknown outcome", "product_id", j.ProductID, "outcome", result.Outcome)
			}
		}
		log.Info("drain complete", "succeeded", succeeded, "deferred", deferred, "failed", failed)
		return
	}

	// in-memory one-shot fallback
	if len(jobs) == 0 {
		log.Info("no jobs; set SCRAPE_SEED=productID:itemID:shopID,... (and DATABASE_URL for the durable queue)")
		return
	}
	orch := orchestrator.NewPool(cfg, pc, memqueue.New())
	orch.RegisterAdapter(adapter)
	for _, j := range jobs {
		result, err := orch.ProcessJob(ctx, j)
		if err != nil {
			log.Error("scrape job outcome not persisted", "product_id", j.ProductID, "err", err.Error())
			continue
		}
		switch result.Outcome {
		case orchestrator.JobSucceeded:
			log.Info("scraped and posted to price", "product_id", j.ProductID)
		case orchestrator.JobDeferred:
			log.Warn("scrape deferred for retry", "product_id", j.ProductID, "attempt", result.Attempts, "retry_at", result.RetryAt, "err", result.Cause)
		case orchestrator.JobFailed:
			log.Error("scrape failed permanently", "product_id", j.ProductID, "attempt", result.Attempts, "err", result.Cause)
		default:
			log.Error("scrape returned unknown outcome", "product_id", j.ProductID, "outcome", result.Outcome)
		}
	}
}

func parseSeed(seed string) []orchestrator.ScrapeJob {
	var jobs []orchestrator.ScrapeJob
	for _, part := range strings.Split(seed, ",") {
		f := strings.Split(strings.TrimSpace(part), ":")
		if len(f) != 3 {
			continue
		}
		pid, err := strconv.ParseInt(f[0], 10, 64)
		if err != nil {
			continue
		}
		jobs = append(jobs, orchestrator.ScrapeJob{
			ProductID: pid, PlatformID: shopeePlatformID, PlatformItemID: f[1] + ":" + f[2], Tier: orchestrator.TierHot,
		})
	}
	return jobs
}
