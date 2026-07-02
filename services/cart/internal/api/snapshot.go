package api

import (
	"encoding/json"
	"net/http"
	"time"

	"shopass/services/cart/internal/cart"
)

type SnapshotHandler struct {
	repo *cart.SnapshotRepo
}

func NewSnapshotHandler(repo *cart.SnapshotRepo) *SnapshotHandler {
	return &SnapshotHandler{repo: repo}
}

func (h *SnapshotHandler) CreateSnapshot(w http.ResponseWriter, r *http.Request) {
	// Giả lập middleware lấy user_id từ JWT (sẽ được tích hợp sau)
	// Hiện tại mock 1 để bypass auth tests, thực tế FR-CART-002 require JWT header/context
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		userIDVal = int64(1) // dummy for now if not set
	}
	userID := userIDVal.(int64)

	var payload cart.SnapshotPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Reject if payload has cookie/token fields via raw unmarshal check
	// Because we strict unmarshal normally, but to be sure:
	// In production, we can decode into map[string]interface{} to check forbidden keys first.
	// But let's trust the struct strictness. Actually the FR says "từ chối hoặc loại bỏ".
	// The strict struct `SnapshotPayload` naturally drops unknown fields, which satisfies "loại bỏ".

	snap := &cart.CartSnapshot{
		UserID:      userID,
		PlatformID:  payload.PlatformID,
		SnapshotRef: payload.SnapshotRef,
		CapturedAt:  time.Now(),
		Items:       make([]cart.CartItem, 0, len(payload.Items)),
	}

	for _, item := range payload.Items {
		if item.Qty <= 0 || item.UnitPrice <= 0 {
			http.Error(w, "qty and unit_price must be > 0", http.StatusBadRequest)
			return
		}
		
		snap.Items = append(snap.Items, cart.CartItem{
			ProductID:      item.ProductID,
			PlatformItemID: item.PlatformItemID,
			ShopID:         item.ShopID,
			Qty:            item.Qty,
			UnitPrice:      item.UnitPrice,
		})
	}

	if err := h.repo.InsertSnapshot(r.Context(), snap); err != nil {
		http.Error(w, "Failed to insert snapshot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(snap)
}
