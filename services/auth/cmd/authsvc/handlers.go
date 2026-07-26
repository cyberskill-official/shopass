package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"shopass/obs"
	"shopass/services/auth/internal/auth"
)

type handlers struct {
	log           *slog.Logger
	tokens        *auth.TokenService
	reg           *auth.Service
	oauth         *auth.OAuthService
	socialEnabled bool
}

func (h *handlers) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.health)
	mux.Handle("GET /metrics", obs.MetricsHandler())
	mux.HandleFunc("GET /.well-known/jwks.json", h.jwks)
	mux.HandleFunc("POST /v1/auth/register", h.registerUser)
	mux.HandleFunc("POST /v1/auth/login", h.login)
	mux.HandleFunc("POST /v1/auth/refresh", h.refresh)
	mux.HandleFunc("POST /v1/auth/logout", h.logout)
	mux.HandleFunc("GET /v1/auth/oauth/{provider}/start", h.oauthStart)
	mux.HandleFunc("GET /v1/auth/oauth/{provider}/callback", h.oauthCallback)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *handlers) jwks(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.tokens.GetJWKS())
}

func (h *handlers) registerUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	id, err := h.reg.Register(r.Context(), auth.RegisterInput{Email: body.Email, Phone: body.Phone, Password: body.Password})
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		writeErr(w, http.StatusConflict, "email already taken")
	case errors.Is(err, auth.ErrWeakPassword), errors.Is(err, auth.ErrNoIdentifier):
		writeErr(w, http.StatusBadRequest, err.Error())
	case err != nil:
		writeErr(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, map[string]int64{"user_id": id})
	}
}

func (h *handlers) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	pair, err := h.tokens.Login(r.Context(), body.Email, body.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrAccountNotActive):
		writeErr(w, http.StatusUnauthorized, "invalid email or password")
	case err != nil:
		writeErr(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, pair)
	}
}

func (h *handlers) refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	pair, err := h.tokens.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (h *handlers) logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	// Idempotent at the web boundary: no token detail is disclosed to an
	// unauthenticated caller, and the browser always clears its cookie.
	_ = h.tokens.Logout(r.Context(), body.RefreshToken)
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) oauthStart(w http.ResponseWriter, r *http.Request) {
	if !h.socialEnabled {
		writeErr(w, http.StatusServiceUnavailable, "social login not configured")
		return
	}
	url, err := h.oauth.StartOAuth(r.Context(), r.PathValue("provider"))
	switch {
	case errors.Is(err, auth.ErrUnknownProvider):
		writeErr(w, http.StatusNotFound, "unknown provider")
	case err != nil:
		writeErr(w, http.StatusInternalServerError, "internal error")
	default:
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (h *handlers) oauthCallback(w http.ResponseWriter, r *http.Request) {
	if !h.socialEnabled {
		writeErr(w, http.StatusServiceUnavailable, "social login not configured")
		return
	}
	q := r.URL.Query()
	pair, err := h.oauth.OAuthCallback(r.Context(), r.PathValue("provider"), q.Get("code"), q.Get("state"))
	switch {
	case errors.Is(err, auth.ErrBadState):
		writeErr(w, http.StatusBadRequest, "invalid oauth state")
	case errors.Is(err, auth.ErrUnknownProvider):
		writeErr(w, http.StatusNotFound, "unknown provider")
	case err != nil:
		writeErr(w, http.StatusBadGateway, "oauth exchange failed")
	default:
		writeJSON(w, http.StatusOK, pair)
	}
}
