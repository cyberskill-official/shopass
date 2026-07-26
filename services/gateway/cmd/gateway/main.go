package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"shopass/obs"
	"shopass/services/gateway/internal/gw"
)

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Initialize tracing
	ctx := context.Background()
	shutdown, err := obs.InitTracer(ctx, "gateway")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracer: %v", err)
	} else {
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			shutdown(ctx)
		}()
	}

	// Expose metrics on a separate multiplexer or route
	go func() {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", obs.MetricsHandler())
		log.Println("Metrics listening on :9090")
		if err := http.ListenAndServe(":9090", metricsMux); err != nil {
			log.Printf("Metrics server failed: %v", err)
		}
	}()

	redisOptions, err := redis.ParseURL(env("REDIS_URL", "redis://redis:6379/0"))
	if err != nil {
		log.Fatalf("invalid REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(redisOptions)
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("gateway Redis unavailable: %v", err)
	}

	// The auth service owns signing keys. The gateway receives only its public
	// JWKS over the private network and fails closed when it cannot verify a JWT.
	verifier := gw.NewHTTPJWKS(
		env("AUTH_JWKS_URL", "http://authsvc:8084/.well-known/jwks.json"),
		env("AUTH_ISS", "shopass-auth"),
		env("AUTH_AUD", "shopass-gateway"),
		5*time.Minute,
		nil,
	)

	deps := gw.Deps{
		WAFConfig: gw.WAFConfig{MaxBodySize: 1 << 20}, // 1 MiB limit
		Redis:     rdb,
		JWKS:      verifier,
		Upstreams: gw.Upstreams{
			Auth:  env("AUTH_UPSTREAM_URL", "http://authsvc:8084"),
			Track: env("TRACK_UPSTREAM_URL", "http://tracksvc:8083"),
			Price: env("PRICE_UPSTREAM_URL", "http://pricesvc:8081"),
			Deal:  env("DEAL_UPSTREAM_URL", "http://dealsvc:8082"),
			Notif: env("NOTIF_UPSTREAM_URL", "http://notifsvc:8082"),
			BFF:   env("BFF_UPSTREAM_URL", "http://bff:8085"),
		},
	}

	handler := gw.NewHandler(deps)
	obsHandler := obs.HTTP("gateway")(handler)

	server := &http.Server{
		Addr:              env("GATEWAY_ADDR", ":8080"),
		Handler:           obsHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		log.Printf("API Gateway listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("gateway shutdown: %v", err)
	}
}
