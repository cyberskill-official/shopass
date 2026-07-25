package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	log := slog.Default()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // use 8081 for notifsvc to avoid conflict
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			UserID    int64          `json:"user_id"`
			ProductID int64          `json:"product_id"`
			Reason    string         `json:"reason"`
			Payload   map[string]any `json:"payload"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Info("Received notification request via HTTP",
			"user_id", payload.UserID,
			"product_id", payload.ProductID,
			"reason", payload.Reason)

		// TASK-NOTIF-003 fanout. Here we would push to Redis Streams or RabbitMQ.
		w.WriteHeader(http.StatusAccepted)
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("notifsvc started", "port", port)
	if err := server.ListenAndServe(); err != nil {
		log.Error("notifsvc failed", "err", err)
	}
}
