package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, message string, extra map[string]any) {
	body := map[string]any{"error": code, "message": message}
	for k, v := range extra {
		body[k] = v
	}
	writeJSON(w, status, body)
}
