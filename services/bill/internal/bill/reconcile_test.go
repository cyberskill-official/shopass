package bill

import (
	"context"
	"testing"
	"time"
)

type mockPaymentRepoRec struct {
	stale []PaymentRecord
	paid map[int64]string
}

func (m *mockPaymentRepoRec) ByOrderRef(ctx context.Context, orderRef string) (PaymentRecord, bool) { return PaymentRecord{}, false }
func (m *mockPaymentRepoRec) InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string) {}
func (m *mockPaymentRepoRec) MarkPaid(ctx context.Context, id int64, txID string) {
	m.paid[id] = txID
}
func (m *mockPaymentRepoRec) MarkFailed(ctx context.Context, id int64) {}
func (m *mockPaymentRepoRec) MarkMismatch(ctx context.Context, id int64, gwAmt int64) {}
func (m *mockPaymentRepoRec) GetPendingOlderThan(ctx context.Context, d time.Duration) []PaymentRecord {
	return m.stale
}

type mockGatewayClient struct{}

func (m mockGatewayClient) CheckStatus(ctx context.Context, orderRef string) (string, string, error) {
	if orderRef == "o1" {
		return "paid", "tx1", nil
	}
	return "failed", "", nil
}

func TestReconcile(t *testing.T) {
	repo := &mockPaymentRepoRec{
		stale: []PaymentRecord{
			{ID: 1, OrderRef: "o1", Status: "pending"},
			{ID: 2, OrderRef: "o2", Status: "pending"},
		},
		paid: make(map[int64]string),
	}
	job := NewReconcileJob(repo, mockGatewayClient{})
	job.Run(context.Background())

	if repo.paid[1] != "tx1" {
		t.Fatalf("expected o1 to be paid with tx1")
	}
	if _, ok := repo.paid[2]; ok {
		t.Fatalf("expected o2 not to be paid")
	}
}
