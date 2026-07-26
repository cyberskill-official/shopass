package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"shopass/services/notif/internal/notif"
)

// Store is the persistence surface the HTTP handlers need.
type Store interface {
	GetUserChannels(ctx context.Context, userID int64) (notif.UserChannels, error)
	InsertNotification(ctx context.Context, n notif.Notification) (int64, error)
	MarkQueued(ctx context.Context, id int64) error
	UpsertToken(ctx context.Context, userID int64, channel, platform, address string) error
	DeleteToken(ctx context.Context, userID int64, address string) error
}

// Server exposes notifsvc HTTP endpoints.
type Server struct {
	store Store
	log   *slog.Logger
}

func New(store Store, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{store: store, log: log}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /notify", s.handleNotify)
	mux.HandleFunc("POST /v1/devices", s.handleRegisterDevice)
	mux.HandleFunc("DELETE /v1/devices", s.handleDeleteDevice)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

type notifyRequest struct {
	UserID    int64          `json:"user_id"`
	ProductID int64          `json:"product_id"`
	Reason    string         `json:"reason"`
	Payload   map[string]any `json:"payload"`
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	var req notifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.UserID <= 0 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}
	template := req.Reason
	if template == "" {
		template = "bottom_predicted"
	}

	caps, err := s.store.GetUserChannels(r.Context(), req.UserID)
	if err != nil {
		s.log.Error("get user channels", "err", err, "user_id", req.UserID)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	// Phase A: push only (email/SMS land in later NOTIF tasks).
	channel, ok := notif.ResolveChannel([]string{"push"}, caps, false)
	if !ok {
		http.Error(w, "no verified push channel", http.StatusUnprocessableEntity)
		return
	}

	payload := req.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	if req.ProductID > 0 {
		payload["product_id"] = req.ProductID
	}

	rendered, err := notif.Render(template, payload)
	if err != nil {
		// Still enqueue with a generic body so deal path can proceed when
		// payload is sparse; attach render error for operators.
		s.log.Warn("template render fallback", "template", template, "err", err)
		rendered = notif.Rendered{
			Title: "Cảnh báo giá Shopass",
			Body:  "Có cập nhật giá cho sản phẩm bạn theo dõi.",
		}
	}
	payload["title"] = rendered.Title
	payload["body"] = rendered.Body
	if _, has := payload["deeplink"]; !has && req.ProductID > 0 {
		payload["deeplink"] = "/products/" + strconv.FormatInt(req.ProductID, 10) + "/chart"
	}

	id, err := s.store.InsertNotification(r.Context(), notif.Notification{
		UserID:   req.UserID,
		Channel:  channel,
		Template: template,
		Payload:  payload,
		Status:   "pending",
	})
	if err != nil {
		s.log.Error("insert notification", "err", err, "user_id", req.UserID)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	if err := s.store.MarkQueued(r.Context(), id); err != nil {
		s.log.Error("mark queued", "err", err, "id", id)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}

	s.log.Info("notification enqueued",
		"id", id,
		"user_id", req.UserID,
		"product_id", req.ProductID,
		"channel", channel,
		"template", template,
	)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "status": "queued"})
}

type deviceRequest struct {
	FCMToken string `json:"fcm_token"`
	Platform string `json:"platform"`
}

func (s *Server) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req deviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.FCMToken == "" {
		http.Error(w, "fcm_token required", http.StatusBadRequest)
		return
	}
	platform := req.Platform
	if platform == "" {
		platform = "web"
	}
	switch platform {
	case "web", "android", "ios":
	default:
		http.Error(w, "invalid platform", http.StatusBadRequest)
		return
	}
	if err := s.store.UpsertToken(r.Context(), userID, "push", platform, req.FCMToken); err != nil {
		s.log.Error("upsert token", "err", err, "user_id", userID)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "platform": platform})
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req deviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.FCMToken == "" {
		http.Error(w, "fcm_token required", http.StatusBadRequest)
		return
	}
	if err := s.store.DeleteToken(r.Context(), userID, req.FCMToken); err != nil {
		s.log.Error("delete token", "err", err, "user_id", userID)
		http.Error(w, "store error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func userIDFromRequest(r *http.Request) (int64, bool) {
	raw := r.Header.Get("X-User-Id")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
