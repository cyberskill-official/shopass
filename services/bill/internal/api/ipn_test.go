package api

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shopass/services/bill/internal/bill"
	"shopass/services/bill/internal/pay"
)

type mockPaymentRepoIPN struct {
	payments   map[int64]*bill.PaymentRecord
	byOrderRef map[string]int64
}

func (m *mockPaymentRepoIPN) ByOrderRef(ctx context.Context, orderRef string) (bill.PaymentRecord, bool) {
	id, ok := m.byOrderRef[orderRef]
	if !ok {
		return bill.PaymentRecord{}, false
	}
	return *m.payments[id], true
}
func (m *mockPaymentRepoIPN) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {
}
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

func testSecrets() fakeSecrets {
	return fakeSecrets{
		"bill/momo/secret_key":   "momo-test-key",
		"bill/zalopay/mac_key":   "zalopay-test-key",
		"bill/vnpay/hash_secret": "vnpay-test-key",
	}
}

func signBody(t *testing.T, secrets pay.SecretReader, gateway, body string) string {
	t.Helper()
	var (
		sig string
		err error
	)
	switch gateway {
	case "momo":
		sig, err = pay.SignMoMo(context.Background(), secrets, body)
	case "zalopay":
		sig, err = pay.SignZaloPay(context.Background(), secrets, body)
	case "vnpay":
		sig, err = pay.SignVNPay(context.Background(), secrets, body)
	default:
		t.Fatalf("unknown gateway %s", gateway)
	}
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return sig
}

func hmacHex(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func setupIPN(t *testing.T) (*IPNHandler, *mockPaymentRepoIPN, fakeSecrets) {
	t.Helper()
	repo := &mockPaymentRepoIPN{
		payments:   make(map[int64]*bill.PaymentRecord),
		byOrderRef: make(map[string]int64),
	}
	secrets := testSecrets()
	h := NewIPNHandler(repo, mockSubActivatorIPN{}, secrets)
	return h, repo, secrets
}

func doIPN(t *testing.T, h *IPNHandler, gw, sig, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/billing/ipn/"+gw, bytes.NewBufferString(body))
	req.Header.Set("X-Signature", sig)
	req.SetPathValue("gateway", gw)
	rec := httptest.NewRecorder()
	h.HandleIPN(rec, req)
	return rec
}

func TestIPN_BadSignature(t *testing.T) {
	h, _, _ := setupIPN(t)
	body := `{"order_ref":"123","amount":100,"status":"paid"}`
	rec := doIPN(t, h, "momo", "not-a-valid-hmac", body)
	if rec.Code != 400 {
		t.Fatalf("expected 400 bad signature, got %d", rec.Code)
	}
}

func TestIPN_WrongKeySignature(t *testing.T) {
	h, repo, _ := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "pending"}
	repo.byOrderRef["o1"] = 1

	body := `{"order_ref":"o1","amount":29000,"status":"paid","transaction_id":"tx123"}`
	wrongSig := hmacHex("attacker-key", body)
	rec := doIPN(t, h, "momo", wrongSig, body)
	if rec.Code != 400 {
		t.Fatalf("expected 400 for wrong-key signature, got %d", rec.Code)
	}
	if repo.payments[1].Status != "pending" {
		t.Fatalf("payment must stay pending on bad sig, got %s", repo.payments[1].Status)
	}
}

func TestVerifyIPN_WrongKey(t *testing.T) {
	secrets := testSecrets()
	body := []byte(`{"order_ref":"o1","amount":29000,"status":"paid"}`)
	wrong := hmacHex("other-key", string(body))
	ok, err := VerifyIPN(context.Background(), secrets, "momo", body, wrong)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected VerifyIPN to reject wrong-key signature")
	}
}

func TestIPN_MismatchAmount(t *testing.T) {
	h, repo, secrets := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "pending"}
	repo.byOrderRef["o1"] = 1

	body := `{"order_ref":"o1","amount":1000,"status":"paid"}`
	sig := signBody(t, secrets, "momo", body)
	rec := doIPN(t, h, "momo", sig, body)
	if rec.Code != 200 {
		t.Fatalf("expected 200 (to stop retry), got %d", rec.Code)
	}
	if repo.payments[1].Status != "mismatch" {
		t.Fatalf("expected status mismatch, got %s", repo.payments[1].Status)
	}
}

func TestIPN_SuccessPaid(t *testing.T) {
	h, repo, secrets := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "pending"}
	repo.byOrderRef["o1"] = 1

	body := `{"order_ref":"o1","amount":29000,"status":"paid","transaction_id":"tx123"}`
	sig := signBody(t, secrets, "momo", body)
	rec := doIPN(t, h, "momo", sig, body)
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
	h, repo, secrets := setupIPN(t)
	repo.payments[1] = &bill.PaymentRecord{ID: 1, OrderRef: "o1", Amount: 29000, Status: "paid"}
	repo.byOrderRef["o1"] = 1

	body := `{"order_ref":"o1","amount":29000,"status":"paid"}`
	sig := signBody(t, secrets, "momo", body)
	rec := doIPN(t, h, "momo", sig, body)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
