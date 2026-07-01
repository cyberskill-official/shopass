package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
)

type mockPlatforms struct{}

func (m mockPlatforms) IDByCode(code string) (int16, bool) {
	switch code {
	case "shopee":
		return 1, true
	case "lazada":
		return 2, true
	case "tiktok":
		return 3, true
	default:
		return 0, false
	}
}

type mockPrice struct {
	products map[string]int64
	nextID   int64
}

func (m *mockPrice) Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error) {
	if m.products == nil {
		m.products = make(map[string]int64)
	}
	key := p.PlatformItemID
	if id, ok := m.products[key]; ok {
		p.ID = id
		return p, nil
	}
	m.nextID++
	p.ID = m.nextID
	m.products[key] = p.ID
	return p, nil
}

type mockRepo struct {
	links map[int64]map[int64]bool
}

func (m *mockRepo) LinkUserProduct(ctx context.Context, userID, productID int64) (bool, error) {
	if m.links == nil {
		m.links = make(map[int64]map[int64]bool)
	}
	if _, ok := m.links[userID]; !ok {
		m.links[userID] = make(map[int64]bool)
	}
	if m.links[userID][productID] {
		return false, nil // already linked
	}
	m.links[userID][productID] = true
	return true, nil
}

type mockQueue struct {
	count    int
	failNext bool
}

func (m *mockQueue) EnqueuePriming(ctx context.Context, productID int64) error {
	if m.failNext {
		m.failNext = false
		return errors.New("mock error")
	}
	m.count++
	return nil
}

func (m *mockQueue) PrimingCount() int {
	return m.count
}

func (m *mockQueue) FailNext() {
	m.failNext = true
}

func setupHandler(t *testing.T) (*Handler, *mockQueue) {
	q := &mockQueue{}
	return NewHandler(&mockPlatforms{}, &mockPrice{}, &mockRepo{}, q), q
}

func doPOST(t *testing.T, h *Handler, path string, bodyStr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), "user_id", int64(123))
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	h.HandleTrack(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

func countLinks(t *testing.T, h *Handler) int {
	repo := h.repo.(*mockRepo)
	count := 0
	for _, prods := range repo.links {
		count += len(prods)
	}
	return count
}

func TestTrack_NewProduct_201(t *testing.T) {
	h, q := setupHandler(t)
	rec := doPOST(t, h, "/v1/track",
		`{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`)
	if rec.Code != 201 {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
	if q.PrimingCount() != 1 {
		t.Errorf("Expected 1, got %d", q.PrimingCount())
	}
}

func TestTrack_UnsupportedPlatform_400(t *testing.T) {
	h, q := setupHandler(t)
	rec := doPOST(t, h, "/v1/track", `{"platform":"amazon","item_url":"https://x"}`)
	if rec.Code != 400 {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
	if q.PrimingCount() != 0 {
		t.Errorf("Expected 0, got %d", q.PrimingCount())
	}
}

func TestTrack_Idempotent_SameUser(t *testing.T) {
	h, _ := setupHandler(t)
	body := `{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`
	doPOST(t, h, "/v1/track", body)
	rec := doPOST(t, h, "/v1/track", body)
	var resp TrackResponse
	decode(t, rec, &resp)
	if !resp.AlreadyTracked {
		t.Errorf("Expected true for AlreadyTracked")
	}
	if countLinks(t, h) != 1 {
		t.Errorf("Expected 1 link, got %d", countLinks(t, h))
	}
}

func TestTrack_EnqueueFails_Still201(t *testing.T) {
	h, q := setupHandler(t)
	q.FailNext() // queue từ chối job kế
	rec := doPOST(t, h, "/v1/track",
		`{"platform":"shopee","item_url":"https://shopee.vn/x-i.88123.20114455667"}`)
	if rec.Code != 201 {
		t.Errorf("Expected 201, got %d", rec.Code)
	}
	if countLinks(t, h) != 1 {
		t.Errorf("Expected 1 link, got %d", countLinks(t, h))
	}
}
