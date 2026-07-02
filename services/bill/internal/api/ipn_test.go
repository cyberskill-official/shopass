package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shopass/services/bill/internal/bill"
)

type mockPaymentRepoIPN struct {
	payments map[int64]*bill.PaymentRecord
	byOrderRef map[string]int64
}

func (m *mockPaymentRepoIPN) ByOrderRef(ctx context.Context, orderRef string) (bill.PaymentRecord, bool) {
	id, ok := m.byOrderRef[orderRef]
	if !ok {
		return bill.PaymentRecord{}, false
	}
	return *m.payments[id], true
}
func (m *mockPaymentRepoIPN) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {}
func (m *mockPaymentRepoIPN) MarkPaid(ctx context.Context, id int64, txID string) {
	m.payments[id].Status = "paid"
	m.payments[id].TransactionID = &txID
	now := time.Now()
	m.payments[id].PaidAt = &now
}
func (m *mockPaymentRepoIPN) MarkFailed(ctx context.Context, id int64) {
	m.payments[id].Status = "failed"
}
func (m *mockPaymentRepoIPN) MarkMismatch(ctx context.Context, id int64, gwAmt int64) {
	m.payments[id].Status = "mismatch"
}
func (m *mockPaymentRepoIPN) GetPendingOlderThan(ctx context.Context, d time.Duration) []bill.PaymentRecord {
	return nil
}

type fakeSecrets map[string]string
func (f fakeSecrets) Get(ctx context.Context, path string) (string, error) {
	return f[path], nil
}

type mockSubActivatorIPN struct{}

func (mockSubActivatorIPN) ActivateSubscription(ctx context.Context, subID int64, duration time.Duration) error {
	return nil
}

func setupIPN(t *testing.T) (*IPNHandler, *mockPaymentRepoIPN) {
	repo := &mockPaymentRepoIPN{
		payments: make(map[int64]*bill.PaymentRecord),
		byOrderRef: make(map[string]int64),
	}
	h := NewIPNHandler(repo, mockSubActivatorIPN{}, fakeSecrets{})
	return h, repo
}

func doIPN(t *testing.T, h *IPNHandler, gw, sig, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/billing/ipn/"+gw, bytes.NewBufferString(body))
	req.Header.Set("X-Signature", sig)
	req.SetPathValue("gateway", gw)
	rec := httptest.NewRecorder()
	h.HandleIPN(rec, req)
	return rec
}

func TestIPN_BadSignature(t *testing.T) {
	h, _ := setupIPN(t)
	rec := doIPN(t, h, "momo", "bad-sig", `{"order_ref":"123","amount":100,"status":"paid"}`)
	if rec.Code != 400 {
		t.Fatalf("expected 400 bad signature, got %d", rec.Code)
	}
}

func TestIPN_MismatchAmount(t *testing.T) {
	h, repo := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "pending"}
	repo.byOrderRef["o1"] = 1

	rec := doIPN(t, h, "momo", "good-sig", `{"order_ref":"o1","amount":1000,"status":"paid"}`)
	if rec.Code != 200 {
		t.Fatalf("expected 200 (to stop retry), got %d", rec.Code)
	}
	if repo.payments[1].Status != "mismatch" {
		t.Fatalf("expected status mismatch, got %s", repo.payments[1].Status)
	}
}

func TestIPN_SuccessPaid(t *testing.T) {
	h, repo := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "pending"}
	repo.byOrderRef["o1"] = 1

	rec := doIPN(t, h, "momo", "good-sig", `{"order_ref":"o1","amount":29000,"status":"paid","transaction_id":"tx123"}`)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if repo.payments[1].Status != "paid" {
		t.Fatalf("expected status paid, got %s", repo.payments[1].Status)
	}
	if *repo.payments[1].TransactionID != "tx123" {
		t.Fatalf("expected tx tx123")
	}
}

func TestIPN_Idempotent(t *testing.T) {
	h, repo := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "paid"}
	repo.byOrderRef["o1"] = 1

	rec := doIPN(t, h, "momo", "good-sig", `{"order_ref":"o1","amount":29000,"status":"paid"}`)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	// duplicate should just return 200 and not crash or fail
}
