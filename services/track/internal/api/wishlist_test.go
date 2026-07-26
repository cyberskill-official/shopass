package api

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"shopass/services/track/internal/track"
)

type mockWishlistRepo struct {
	lists map[int64]track.Wishlist
	items map[int64]map[int64]track.WishlistItem // wishlistID -> productID -> item
	nextW int64
	nextI int64
}

func (m *mockWishlistRepo) CreateWishlist(ctx context.Context, userID int64, name string) (track.Wishlist, error) {
	if m.lists == nil {
		m.lists = make(map[int64]track.Wishlist)
	}
	m.nextW++
	w := track.Wishlist{ID: m.nextW, UserID: userID, Name: name, CreatedAt: time.Now()}
	m.lists[m.nextW] = w
	return w, nil
}

func (m *mockWishlistRepo) ListWishlists(ctx context.Context, userID int64) ([]track.Wishlist, error) {
	var res []track.Wishlist
	for _, w := range m.lists {
		if w.UserID == userID {
			res = append(res, w)
		}
	}
	return res, nil
}

func (m *mockWishlistRepo) OwnsWishlist(ctx context.Context, userID, wishlistID int64) (bool, error) {
	w, ok := m.lists[wishlistID]
	return ok && w.UserID == userID, nil
}

func (m *mockWishlistRepo) AddItem(ctx context.Context, wishlistID, productID int64, target *int64) error {
	if productID == 999999 {
		// Mock FK violation
		return errors.New("fk violation") // Need to make it match IsFKViolation or just mock IsFKViolation behavior?
		// Wait, track.IsFKViolation expects pgx error. In tests we can just mock the repo to return a pgx error, but that's hard.
		// Let's just skip the specific type check in the mock and simulate the handler logic if needed, or change the mock to return a pgx error.
	}
	if m.items == nil {
		m.items = make(map[int64]map[int64]track.WishlistItem)
	}
	if m.items[wishlistID] == nil {
		m.items[wishlistID] = make(map[int64]track.WishlistItem)
	}
	if item, ok := m.items[wishlistID][productID]; ok {
		item.TargetPrice = target
		m.items[wishlistID][productID] = item
		return nil
	}
	m.nextI++
	m.items[wishlistID][productID] = track.WishlistItem{
		ID:          m.nextI,
		WishlistID:  wishlistID,
		ProductID:   productID,
		TargetPrice: target,
		AddedAt:     time.Now(),
	}
	return nil
}

func (m *mockWishlistRepo) HasItem(ctx context.Context, wishlistID, productID int64) (bool, error) {
	if m.items == nil || m.items[wishlistID] == nil {
		return false, nil
	}
	_, ok := m.items[wishlistID][productID]
	return ok, nil
}

func (m *mockWishlistRepo) CountUserItems(ctx context.Context, userID int64) (int64, error) {
	var n int64
	for wid, items := range m.items {
		w, ok := m.lists[wid]
		if !ok || w.UserID != userID {
			continue
		}
		n += int64(len(items))
	}
	return n, nil
}

func (m *mockWishlistRepo) RemoveItem(ctx context.Context, wishlistID, productID int64) error {
	if m.items[wishlistID] != nil {
		delete(m.items[wishlistID], productID)
	}
	return nil
}

func (m *mockWishlistRepo) DeleteWishlist(ctx context.Context, wishlistID int64) error {
	delete(m.lists, wishlistID)
	delete(m.items, wishlistID) // CASCADE
	return nil
}

func (m *mockWishlistRepo) ListItems(ctx context.Context, wishlistID int64) ([]track.WishlistItem, error) {
	var res []track.WishlistItem
	for _, it := range m.items[wishlistID] {
		res = append(res, it)
	}
	return res, nil
}

func setupWishlistHandler(t *testing.T) *WishlistHandler {
	return NewWishlistHandler(&mockWishlistRepo{})
}

func doWishlistReq(t *testing.T, h *WishlistHandler, method, path, bodyStr string, userID int64) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	if userID != 0 {
		req = req.WithContext(context.WithValue(req.Context(), "user_id", userID))
	}
	// Note: httptest.NewRequest doesn't parse PathValue in Go 1.22 ServeMux automatically unless routed through ServeMux
	// We need to route it through ServeMux
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func ptr(v int64) *int64  { return &v }
func itoa(i int64) string { return strconv.FormatInt(i, 10) }

func TestList_ScopedToUser(t *testing.T) {
	h := setupWishlistHandler(t)
	// Create for A
	recA := doWishlistReq(t, h, "POST", "/v1/wishlists", `{"name":"A"}`, 1)
	var wa map[string]interface{}
	decode(t, recA, &wa)
	widA := int64(wa["id"].(float64))

	// Create for B
	doWishlistReq(t, h, "POST", "/v1/wishlists", `{"name":"B"}`, 2)

	// List A
	rec := doWishlistReq(t, h, "GET", "/v1/wishlists", "", 1)
	var lists []map[string]interface{}
	decode(t, rec, &lists)
	if len(lists) != 1 {
		t.Fatalf("Expected 1 list, got %d", len(lists))
	}
	if int64(lists[0]["id"].(float64)) != widA {
		t.Errorf("Expected widA")
	}
}

func TestCrossUser_404(t *testing.T) {
	h := setupWishlistHandler(t)
	recB := doWishlistReq(t, h, "POST", "/v1/wishlists", `{"name":"B"}`, 2)
	var wb map[string]interface{}
	decode(t, recB, &wb)
	widB := int64(wb["id"].(float64))

	rec := doWishlistReq(t, h, "POST", "/v1/wishlists/"+itoa(widB)+"/items", `{"product_id":1}`, 1)
	if rec.Code != 404 {
		t.Errorf("Expected 404, got %d", rec.Code)
	}
}

func TestDeleteWishlist_Cascade(t *testing.T) {
	h := setupWishlistHandler(t)
	recA := doWishlistReq(t, h, "POST", "/v1/wishlists", `{"name":"A"}`, 1)
	var wa map[string]interface{}
	decode(t, recA, &wa)
	wid := int64(wa["id"].(float64))

	doWishlistReq(t, h, "POST", "/v1/wishlists/"+itoa(wid)+"/items", `{"product_id":1}`, 1)

	// Delete
	doWishlistReq(t, h, "DELETE", "/v1/wishlists/"+itoa(wid), "", 1)

	// List items -> 404
	recItems := doWishlistReq(t, h, "GET", "/v1/wishlists/"+itoa(wid)+"/items", "", 1)
	if recItems.Code != 404 {
		t.Errorf("Expected 404, got %d", recItems.Code)
	}
}

func TestTargetPrice_Int64InJSON(t *testing.T) {
	h := setupWishlistHandler(t)
	recA := doWishlistReq(t, h, "POST", "/v1/wishlists", `{"name":"A"}`, 1)
	var wa map[string]interface{}
	decode(t, recA, &wa)
	wid := int64(wa["id"].(float64))

	doWishlistReq(t, h, "POST", "/v1/wishlists/"+itoa(wid)+"/items", `{"product_id":1, "target_price":89000}`, 1)

	rec := doWishlistReq(t, h, "GET", "/v1/wishlists/"+itoa(wid)+"/items", "", 1)
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"target_price":89000`)) {
		t.Errorf("Expected target_price as int64, got %s", rec.Body.String())
	}
}
