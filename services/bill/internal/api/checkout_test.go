package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"shopass/services/bill/internal/pay"
)

type mockPlanCatalog struct{}

func (m mockPlanCatalog) ByTier(ctx context.Context, tier string) (Plan, bool) {
	if tier == "premium_basic" {
		return Plan{Tier: "premium_basic", Price: 29000}, true
	}
	return Plan{}, false
}

type mockPaymentRepo struct {
	payments map[string]pay.PaymentResult
}

func (m *mockPaymentRepo) ByOrderRef(ctx context.Context, orderRef string) (pay.PaymentResult, bool) {
	res, ok := m.payments[orderRef]
	return res, ok
}

func (m *mockPaymentRepo) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {
	m.payments[orderRef] = pay.PaymentResult{
		OrderRef: orderRef,
		Gateway:  gateway,
		Amount:   amount,
	}
}

type mockGateway struct {
	calls int
	err   error
}

func (m *mockGateway) Code() string { return "vietqr" }
func (m *mockGateway) CreatePayment(ctx context.Context, r pay.PaymentRequest) (pay.PaymentResult, error) {
	m.calls++
	if m.err != nil {
		return pay.PaymentResult{}, m.err
	}
	return pay.PaymentResult{
		OrderRef:  r.OrderRef,
		Gateway:   "vietqr",
		Amount:    r.Amount,
		QRPayload: "qr_payload",
	}, nil
}

type testMetrics struct {
	calls int
	errs  int
}

func (t *testMetrics) GatewayError(gw string)  { t.errs++ }
func (t *testMetrics) IntentCreated(gw string) { t.calls++ }

func setupCheckout(t *testing.T) (*Handler, *mockGateway, *mockPaymentRepo) {
	plans := mockPlanCatalog{}
	gw := &mockGateway{}
	reg := pay.NewRegistry()
	reg.Register(gw)
	repo := &mockPaymentRepo{payments: make(map[string]pay.PaymentResult)}
	
	h := NewHandler(plans, reg, repo)
	h.metrics = &testMetrics{}
	return h, gw, repo
}

func doPOST(t *testing.T, h *Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.HandleCheckout(rec, req)
	return rec
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}

func TestCheckout_HappyPath(t *testing.T) {
	h, _, _ := setupCheckout(t)
	rec := doPOST(t, h, "/v1/billing/checkout", `{"plan_tier":"premium_basic","gateway":"vietqr"}`)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var res pay.PaymentResult
	decode(t, rec, &res)
	if res.Amount != 29000 {
		t.Fatalf("expected 29000, got %d", res.Amount)
	}
}

func TestCheckout_AmountFromCatalog_NotBody(t *testing.T) {
	h, _, _ := setupCheckout(t)
	rec := doPOST(t, h, "/v1/billing/checkout", `{"plan_tier":"premium_basic","gateway":"vietqr","amount":1}`)
	var res pay.PaymentResult
	decode(t, rec, &res)
	if res.Amount != 29000 {
		t.Fatalf("expected amount from catalog 29000, got %d", res.Amount) // bỏ qua amount client
	}
}

func TestCheckout_Idempotent_NoDoubleCharge(t *testing.T) {
	h, gw, _ := setupCheckout(t)
	body := `{"plan_tier":"premium_basic","gateway":"vietqr"}`
	doPOST(t, h, "/v1/billing/checkout", body)
	doPOST(t, h, "/v1/billing/checkout", body) // bấm lại
	if gw.calls != 1 {
		t.Fatalf("expected exactly 1 gateway call, got %d", gw.calls)
	}
}

func TestCheckout_GatewayError_502_StaysPending(t *testing.T) {
	h, gw, repo := setupCheckout(t)
	gw.err = context.DeadlineExceeded // mock fail
	rec := doPOST(t, h, "/v1/billing/checkout", `{"plan_tier":"premium_basic","gateway":"vietqr"}`)
	if rec.Code != 502 {
		t.Fatalf("expected 502, got %d", rec.Code)
	}
	
	// ensure not inserted to repo when fails
	if len(repo.payments) > 0 {
		t.Fatalf("expected no payment inserted on failure, got %d", len(repo.payments))
	}
}
