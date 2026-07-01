package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shopass/services/track/internal/track"
)

type AlertRuleHandler struct {
	repo track.AlertRuleRepo
}

func NewAlertRuleHandler(repo track.AlertRuleRepo) *AlertRuleHandler {
	return &AlertRuleHandler{repo: repo}
}

func (h *AlertRuleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/alerts", h.HandleCreateRule)
	mux.HandleFunc("GET /v1/alerts", h.HandleListRules)
	mux.HandleFunc("PATCH /v1/alerts/{id}", h.HandleToggleActive)
	mux.HandleFunc("DELETE /v1/alerts/{id}", h.HandleDeleteRule)
	mux.HandleFunc("GET /v1/alerts/{id}/history", h.HandleListAlerts)
}

func (h *AlertRuleHandler) HandleCreateRule(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		ProductID int64    `json:"product_id"`
		RuleType  string   `json:"rule_type"`
		Threshold *int64   `json:"threshold"`
		Channel   []string `json:"channel"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if len(body.Channel) == 0 {
		body.Channel = []string{"push"} // mặc định kênh rẻ nhất
	}
	if err := track.ValidateRule(body.RuleType, body.Threshold, body.Channel); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	rule, err := h.repo.CreateRule(req.Context(), userID, body.ProductID,
		body.RuleType, body.Threshold, body.Channel)
	if err != nil {
		if track.IsFKViolation(err) { // product_id chưa track
			writeErr(w, http.StatusBadRequest, "product not tracked")
			return
		}
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rule)
}

func (h *AlertRuleHandler) HandleListRules(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rules, err := h.repo.ListRules(req.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if rules == nil {
		rules = []track.AlertRule{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(rules)
}

func (h *AlertRuleHandler) HandleToggleActive(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	rid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	owns, err := h.repo.OwnsRule(req.Context(), userID, rid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "rule not found") // DEC-TRACK-24
		return
	}
	var body struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.Active == nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.repo.ToggleActive(req.Context(), rid, *body.Active); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AlertRuleHandler) HandleDeleteRule(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	rid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	owns, err := h.repo.OwnsRule(req.Context(), userID, rid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "rule not found")
		return
	}
	if err := h.repo.DeleteRule(req.Context(), rid); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AlertRuleHandler) HandleListAlerts(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	rid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid rule id")
		return
	}
	owns, err := h.repo.OwnsRule(req.Context(), userID, rid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "rule not found")
		return
	}
	alerts, err := h.repo.ListAlerts(req.Context(), rid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if alerts == nil {
		alerts = []track.Alert{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(alerts)
}
