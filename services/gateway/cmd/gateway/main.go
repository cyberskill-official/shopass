package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"shopass/obs"
	"shopass/services/gateway/internal/gw"
)

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

	// Initialize dependencies
	// Note: In a real application, Redis and JWKS would be instantiated with real clients.
	// For example, reading secrets via shopass/secrets package (FR-INFRA-003).
	deps := gw.Deps{
		WAFConfig: gw.WAFConfig{MaxBodySize: 1 << 20}, // 1 MiB limit
		Redis:     nil,                                // Real implementation injected here
		JWKS:      nil,                                // Real implementation injected here
	}

	handler := gw.NewHandler(deps)
	obsHandler := obs.HTTP("gateway")(handler)

	log.Println("API Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", obsHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

