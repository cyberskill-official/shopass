package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shopass/services/track/internal/track"
)

type WishlistHandler struct {
	repo track.WishlistRepo
	gate FeatureGate
}

func NewWishlistHandler(repo track.WishlistRepo) *WishlistHandler {
	return &WishlistHandler{repo: repo}
}

func (h *WishlistHandler) WithGate(gate FeatureGate) *WishlistHandler {
	h.gate = gate
	return h
}

func (h *WishlistHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/wishlists", h.HandleCreate)
	mux.HandleFunc("GET /v1/wishlists", h.HandleList)
	mux.HandleFunc("POST /v1/wishlists/{id}/items", h.HandleAddItem)
	mux.HandleFunc("DELETE /v1/wishlists/{id}/items/{product_id}", h.HandleRemoveItem)
	mux.HandleFunc("DELETE /v1/wishlists/{id}", h.HandleDelete)
	mux.HandleFunc("GET /v1/wishlists/{id}/items", h.HandleListItems) // Not explicitly in TASK but needed for testing/reading
}

func (h *WishlistHandler) HandleCreate(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.Name == "" {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	wl, err := h.repo.CreateWishlist(req.Context(), userID, body.Name)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   wl.ID,
		"name": wl.Name,
	})
}

func (h *WishlistHandler) HandleList(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	if userID == 0 {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	lists, err := h.repo.ListWishlists(req.Context(), userID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if lists == nil {
		lists = []track.Wishlist{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(lists)
}

func (h *WishlistHandler) HandleAddItem(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	wid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	owns, err := h.repo.OwnsWishlist(req.Context(), userID, wid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "wishlist not found") // 404, không 403 (DEC-TRACK-12)
		return
	}
	var body struct {
		ProductID   int64  `json:"product_id"`
		TargetPrice *int64 `json:"target_price"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	if h.gate != nil {
		exists, err := h.repo.HasItem(req.Context(), wid, body.ProductID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !exists {
			used, err := h.repo.CountUserItems(req.Context(), userID)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "internal error")
				return
			}
			allowed, limitReached, err := h.gate.Check(req.Context(), userID, "wishlist_items", &used)
			if err != nil {
				writeErr(w, http.StatusServiceUnavailable, "gating unavailable")
				return
			}
			if !allowed {
				status := http.StatusForbidden
				if limitReached {
					status = http.StatusPaymentRequired
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error":          "wishlist_limit_reached",
					"feature_key":    "wishlist_items",
					"suggested_tier": "premium_basic",
					"upgrade_path":   "/billing",
				})
				return
			}
		}
	}
	if err := h.repo.AddItem(req.Context(), wid, body.ProductID, body.TargetPrice); err != nil {
		if track.IsFKViolation(err) {
			writeErr(w, http.StatusBadRequest, "product not tracked")
			return
		}
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK) // idempotent
}

func (h *WishlistHandler) HandleRemoveItem(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	wid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	pid, err := strconv.ParseInt(req.PathValue("product_id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	owns, err := h.repo.OwnsWishlist(req.Context(), userID, wid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "wishlist not found")
		return
	}
	if err := h.repo.RemoveItem(req.Context(), wid, pid); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WishlistHandler) HandleDelete(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	wid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	owns, err := h.repo.OwnsWishlist(req.Context(), userID, wid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "wishlist not found")
		return
	}
	if err := h.repo.DeleteWishlist(req.Context(), wid); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *WishlistHandler) HandleListItems(w http.ResponseWriter, req *http.Request) {
	userID := UserID(req.Context())
	wid, err := strconv.ParseInt(req.PathValue("id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid wishlist id")
		return
	}
	owns, err := h.repo.OwnsWishlist(req.Context(), userID, wid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !owns {
		writeErr(w, http.StatusNotFound, "wishlist not found")
		return
	}
	items, err := h.repo.ListItems(req.Context(), wid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal error")
		return
	}
	if items == nil {
		items = []track.WishlistItem{}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(items)
}
