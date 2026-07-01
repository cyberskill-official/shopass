package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sandeal/gateway/internal/gw"
)

func main() {
	addr := os.Getenv("GATEWAY_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	handler := gw.NewHandler(gw.Deps{})
	log.Printf("gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
