package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"shopass/services/auth/internal/auth"
)

const resetAckMessage = "Nếu tài khoản tồn tại, hướng dẫn đặt lại mật khẩu đã được gửi."

func (h *handlers) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if h.lifecycle == nil {
		writeErr(w, http.StatusServiceUnavailable, "account lifecycle unavailable")
		return
	}
	var body struct {
		Identifier string `json:"identifier"`
		Email      string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	identifier := body.Identifier
	if identifier == "" {
		identifier = body.Email
	}
	if identifier == "" {
		writeErr(w, http.StatusBadRequest, "identifier required")
		return
	}
	// Always identical success (§1 #3) — never leak existence.
	_ = h.lifecycle.RequestReset(r.Context(), identifier)
	writeJSON(w, http.StatusOK, map[string]string{"message": resetAckMessage})
}

func (h *handlers) confirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if h.lifecycle == nil {
		writeErr(w, http.StatusServiceUnavailable, "account lifecycle unavailable")
		return
	}
	var body struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if body.Token == "" || body.NewPassword == "" {
		writeErr(w, http.StatusBadRequest, "token and new_password required")
		return
	}
	err := h.lifecycle.ConfirmReset(r.Context(), body.Token, body.NewPassword)
	switch {
	case errors.Is(err, auth.ErrInvalidResetToken):
		writeErr(w, http.StatusBadRequest, "invalid or expired reset token")
	case errors.Is(err, auth.ErrWeakPassword):
		writeErr(w, http.StatusBadRequest, err.Error())
	case err != nil:
		h.log.Error("confirm reset", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, map[string]string{"status": "password_updated"})
	}
}

func (h *handlers) deleteAccount(w http.ResponseWriter, r *http.Request) {
	if h.lifecycle == nil {
		writeErr(w, http.StatusServiceUnavailable, "account lifecycle unavailable")
		return
	}
	raw := r.Header.Get("X-User-Id")
	if raw == "" {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || userID <= 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.lifecycle.DeleteAccount(r.Context(), userID); err != nil {
		h.log.Error("delete account", "err", err)
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	grace := auth.GraceUntil(time.Now())
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "deleted",
		"grace_until": grace.Format(time.RFC3339),
	})
}
