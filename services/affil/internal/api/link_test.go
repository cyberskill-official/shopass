package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"shopass/services/affil/internal/affil"
)

type mockProductImpl struct {
	id         int64
	platformID int16
	targetURL  string
}

func (m mockProductImpl) ID() int64 { return m.id }
func (m mockProductImpl) PlatformID() int16 { return m.platformID }
func (m mockProductImpl) TargetURL() string { return m.targetURL }

type mockProductService struct {
	products map[int64]Product
}

func (m *mockProductService) Get(ctx context.Context, productID int64) (Product, bool) {
	p, ok := m.products[productID]
	return p, ok
}

type mockNetworkService struct {
	hasTemplate bool
}

func (m *mockNetworkService) TemplateFor(platformID int16) (string, affil.NetworkTemplate, bool) {
	if !m.hasTemplate {
		return "", affil.NetworkTemplate{}, false
	}
	return "involve_asia", affil.NetworkTemplate{
		BaseURL:     "https://go.involve.asia/aff",
		TargetParam: "url",
		SubIDParam:  "sub_id",
	}, true
}

type mockAffiliateRepo struct {
	clicks []affil.AffiliateClick
	fail   bool
}

func (m *mockAffiliateRepo) RecordClick(ctx context.Context, c affil.AffiliateClick) (int64, error) {
	if m.fail {
		return 0, context.DeadlineExceeded // mock error
	}
	m.clicks = append(m.clicks, c)
	return int64(len(m.clicks)), nil
}
func (m *mockAffiliateRepo) ClickCount() int {
	return len(m.clicks)
}
func (m *mockAffiliateRepo) FailNextClick() {
	m.fail = true
}

type mockMetrics struct{}

func (m *mockMetrics) LinkRejected(reason string)               {}
func (m *mockMetrics) LinkCreated(platformID int16, network string) {}

func setupHandler(t *testing.T) (*Handler, *mockAffiliateRepo) {
	ps := &mockProductService{
		products: map[int64]Product{
			90112: mockProductImpl{id: 90112, platformID: 1, targetURL: "https://shopee.vn/product/88123/20114455667"},
		},
	}
	ns := &mockNetworkService{hasTemplate: true}
	repo := &mockAffiliateRepo{}
	h := NewHandler(ps, ns, repo, &mockMetrics{})
	return h, repo
}

func setupHandlerNoNetwork(t *testing.T) (*Handler, *mockAffiliateRepo) {
	h, repo := setupHandler(t)
	h.networks = &mockNetworkService{hasTemplate: false}
	return h, repo
}

func doPOST(t *testing.T, h *Handler, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleCreateLink(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestLink_NotUserInitiated_Rejected_NoClick(t *testing.T) {
	h, repo := setupHandler(t)
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":false}`)
	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if repo.ClickCount() != 0 {
		t.Errorf("expected 0 clicks, got %d", repo.ClickCount())
	}
}

func TestLink_MissingFlag_Rejected(t *testing.T) {
	h, repo := setupHandler(t)
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112}`) // thiếu user_initiated
	if rec.Code != 400 {
		t.Errorf("expected 400, got %d", rec.Code)
	}
	if repo.ClickCount() != 0 {
		t.Errorf("expected 0 clicks, got %d", repo.ClickCount())
	}
}

func TestLink_HappyPath_HasDisclosureAndTarget(t *testing.T) {
	h, repo := setupHandler(t)
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var resp LinkResponse
	decode(t, rec, &resp)
	if resp.DeepLink == "" {
		t.Errorf("expected non-empty DeepLink")
	}
	if resp.TargetURL == "" || len(resp.TargetURL) < 9 || resp.TargetURL[:10] != "https://sh" {
		t.Errorf("expected target url to be shopee.vn")
	}
	if resp.Disclosure == "" {
		t.Errorf("expected non-empty Disclosure")
	}
	if repo.ClickCount() != 1 {
		t.Errorf("expected 1 click, got %d", repo.ClickCount())
	}
}

func TestLink_ProductNotFound_404(t *testing.T) {
	h, repo := setupHandler(t)
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":999999,"user_initiated":true}`)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	if repo.ClickCount() != 0 {
		t.Errorf("expected 0 clicks, got %d", repo.ClickCount())
	}
}

func TestLink_NoNetwork_503_NoClick(t *testing.T) {
	h, repo := setupHandlerNoNetwork(t) // không cấu hình template
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
	if rec.Code != 503 {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	if repo.ClickCount() != 0 {
		t.Errorf("expected 0 clicks, got %d", repo.ClickCount())
	}
}

func TestLink_RecordClickFails_NoLink(t *testing.T) {
	h, repo := setupHandler(t)
	repo.FailNextClick()
	rec := doPOST(t, h, "/v1/affiliate/link", `{"product_id":90112,"user_initiated":true}`)
	if rec.Code != 500 {
		t.Errorf("expected 500, got %d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("deep_link")) {
		t.Errorf("expected no deep_link in response")
	}
}
