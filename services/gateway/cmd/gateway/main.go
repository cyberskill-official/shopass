package main

import (
	"log"
	"net/http"

	"shopass/services/gateway/internal/gw"
)

func main() {
	// Initialize dependencies
	// Note: In a real application, Redis and JWKS would be instantiated with real clients.
	// For example, reading secrets via shopass/secrets package (FR-INFRA-003).
	deps := gw.Deps{
		WAFConfig: gw.WAFConfig{MaxBodySize: 1 << 20}, // 1 MiB limit
		Redis:     nil,                                // Real implementation injected here
		JWKS:      nil,                                // Real implementation injected here
	}

	handler := gw.NewHandler(deps)

	log.Println("API Gateway listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
