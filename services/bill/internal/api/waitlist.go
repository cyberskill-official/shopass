package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var emailOK = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type WaitlistHandler struct {
	pool *pgxpool.Pool
}

func NewWaitlistHandler(pool *pgxpool.Pool) *WaitlistHandler {
	return &WaitlistHandler{pool: pool}
}

type waitlistBody struct {
	Email        string `json:"email"`
	Zalo         string `json:"zalo"`
	Source       string `json:"source"`
	TierInterest string `json:"tier_interest"`
}

func (h *WaitlistHandler) HandleWaitlist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body waitlistBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	email := strings.TrimSpace(strings.ToLower(body.Email))
	if !emailOK.MatchString(email) || len(email) > 254 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid email"})
		return
	}
	source := strings.TrimSpace(body.Source)
	if source == "" {
		source = "pricing"
	}
	if len(source) > 64 {
		source = source[:64]
	}
	tier := strings.TrimSpace(body.TierInterest)
	switch tier {
	case "premium_basic", "premium_plus", "premium_pro":
	default:
		tier = "premium_basic"
	}
	zalo := strings.TrimSpace(body.Zalo)
	if len(zalo) > 32 {
		zalo = zalo[:32]
	}
	var zaloArg any
	if zalo != "" {
		zaloArg = zalo
	}

	var id int64
	err := h.pool.QueryRow(r.Context(), `
		INSERT INTO marketing_lead (email, zalo, source, tier_interest)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email, source) DO UPDATE SET
		  zalo = COALESCE(EXCLUDED.zalo, marketing_lead.zalo),
		  tier_interest = EXCLUDED.tier_interest
		RETURNING id
	`, email, zaloArg, source, tier).Scan(&id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "persist failed"})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "id": id})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
