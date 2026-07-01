package api

import (
	"context"
	"encoding/json"
	"net/http"

	"shopass/services/track/internal/track"
)

// auth.UserID placeholder. Giả lập logic gateway
func UserID(ctx context.Context) int64 {
	v := ctx.Value("user_id")
	if v == nil {
		return 0
	}
	return v.(int64)
}

// Interfaces expected by Handler
type PlatformMap interface {
	IDByCode(code string) (int16, bool)
}

// TrackedProduct input to PriceService
type TrackedProduct struct {
	ID             int64
	PlatformID     int16
	PlatformItemID string
	ShopID         *string
}

type PriceService interface {
	Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error)
}

type TrackRepo interface {
	LinkUserProduct(ctx context.Context, userID, productID int64) (bool, error)
}

type ScrapeQueue interface {
	EnqueuePriming(ctx context.Context, productID int64) error
}

type Handler struct {
	platforms   PlatformMap
	price       PriceService
	repo        TrackRepo
	scrapeQueue ScrapeQueue
}

func NewHandler(platforms PlatformMap, p PriceService, r TrackRepo, q ScrapeQueue) *Handler {
	return &Handler{
		platforms:   platforms,
		price:       p,
		repo:        r,
		scrapeQueue: q,
	}
}

type TrackRequest struct {
	Platform string `json:"platform"`
	ItemURL  string `json:"item_url"`
}

type TrackResponse struct {
	ProductID      int64  `json:"product_id"`
	Platform       string `json:"platform"`
	AlreadyTracked bool   `json:"already_tracked"`
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handler) HandleTrack(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body TrackRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	platformID, ok := h.platforms.IDByCode(body.Platform)
	if !ok {
		writeErr(w, http.StatusBadRequest, "unsupported platform")
		return
	}
	item, ok := track.ParseItemURL(body.Platform, body.ItemURL)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid item_url")
		return
	}
	tp, err := h.price.Upsert(req.Context(), TrackedProduct{
		PlatformID:     platformID,
		PlatformItemID: item.PlatformItemID,
		ShopID:         nilIfEmpty(item.ShopID),
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	linkedNew, err := h.repo.LinkUserProduct(req.Context(), userID, tp.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if err := h.scrapeQueue.EnqueuePriming(req.Context(), tp.ID); err != nil {
		// Log warning only
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(TrackResponse{
		ProductID: tp.ID, Platform: body.Platform, AlreadyTracked: !linkedNew,
	})
}
