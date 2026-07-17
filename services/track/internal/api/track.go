package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"shopass/services/track/internal/priceclient"
	"shopass/services/track/internal/track"
)

// auth.UserID placeholder. Giả lập logic gateway
func UserID(ctx context.Context) int64 {
	v := ctx.Value("user_id")
	id, ok := v.(int64)
	if !ok || id <= 0 {
		return 0
	}
	return id
}

// Interfaces expected by Handler
type PlatformMap interface {
	IDByCode(code string) (int16, bool)
}

// TrackedProduct is the registry payload shared with the private pricesvc
// client. The alias lets that client satisfy PriceService without exposing the
// price service's internal package across a service boundary.
type TrackedProduct = priceclient.TrackedProduct

type PriceService interface {
	Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error)
}

type TrackRepo interface {
	LinkUserProduct(ctx context.Context, userID, productID int64) (bool, error)
	ListUserTrackedProducts(ctx context.Context, userID int64) ([]track.UserTrackedProduct, error)
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
	decoder := json.NewDecoder(http.MaxBytesReader(w, req.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	platform := strings.ToLower(strings.TrimSpace(body.Platform))
	itemURL := strings.TrimSpace(body.ItemURL)
	if platform == "" || itemURL == "" {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	platformID, ok := h.platforms.IDByCode(platform)
	if !ok {
		writeErr(w, http.StatusBadRequest, "unsupported platform")
		return
	}
	item, ok := track.ParseItemURL(platform, itemURL)
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
	if h.scrapeQueue == nil {
		slog.Default().Warn("priming scrape queue is not configured", "product_id", tp.ID)
	} else if err := h.scrapeQueue.EnqueuePriming(req.Context(), tp.ID); err != nil {
		// The product link is durable already; feeder/retry work can pick this
		// up later, so never roll the user action back for a queue failure.
		slog.Default().Warn("priming scrape enqueue failed", "product_id", tp.ID, "err", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(TrackResponse{
		ProductID: tp.ID, Platform: platform, AlreadyTracked: !linkedNew,
	})
}

// HandleListTrackedProducts exposes the caller's own tracked products for the
// closed-beta dashboard. There is deliberately no user ID path/query input;
// identity comes only from the gateway-populated request context.
func (h *Handler) HandleListTrackedProducts(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	products, err := h.repo.ListUserTrackedProducts(req.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if products == nil {
		products = []track.UserTrackedProduct{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(products)
}
