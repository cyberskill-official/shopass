package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"shopass/services/bill/internal/referral"
)

type referralEventBus struct{ log *slog.Logger }

func (b referralEventBus) Publish(_ context.Context, event interface{}) {
	b.log.Info("referral.event", "payload", event)
}

type ReferralHandler struct {
	repo *referral.PGRepo
	svc  *referral.Service
	log  *slog.Logger
}

func NewReferralHandler(repo *referral.PGRepo, log *slog.Logger, opts ...referral.Option) *ReferralHandler {
	if log == nil {
		log = slog.Default()
	}
	opts = append(opts, referral.WithLogger(log))
	return &ReferralHandler{
		repo: repo,
		svc:  referral.NewService(repo, referralEventBus{log: log}, opts...),
		log:  log,
	}
}

func (h *ReferralHandler) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	code, err := h.repo.CreateCodeForUser(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	rc, found, err := h.repo.FindByUser(r.Context(), userID)
	if err != nil || !found {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	hasReferrer, err := h.repo.HasReferrer(r.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"code":         code,
		"uses":         rc.Uses,
		"has_referrer": hasReferrer,
		"reward_note":  "Cả hai nhận 1 tháng Premium sau khi vượt kiểm tra chống gian lận (đề xuất mặc định — chờ Stephen duyệt kinh tế).",
	})
}

type attributeBody struct {
	Code string `json:"code"`
}

func (h *ReferralHandler) HandleAttribute(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body attributeBody
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	code := strings.TrimSpace(strings.ToUpper(body.Code))
	if code == "" {
		writeErr(w, http.StatusBadRequest, "code required")
		return
	}
	err := h.svc.Attribute(r.Context(), userID, code)
	if err != nil {
		switch {
		case errors.Is(err, referral.ErrSelfReferral):
			writeErr(w, http.StatusBadRequest, "self_referral")
		case errors.Is(err, referral.ErrAlreadyAttributed):
			writeErr(w, http.StatusConflict, "already_attributed")
		case errors.Is(err, referral.ErrUnknownCode):
			writeErr(w, http.StatusNotFound, "unknown_code")
		default:
			writeErr(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
