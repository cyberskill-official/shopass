package engine

import (
	"context"
	"testing"

	"shopass/services/track/internal/track"
)

type mockRuleRepo struct {
	rules []track.AlertRule
}

func (m *mockRuleRepo) ActiveByProduct(ctx context.Context, productID int64) ([]track.AlertRule, error) {
	var res []track.AlertRule
	for _, r := range m.rules {
		if r.ProductID == productID && r.Active {
			res = append(res, r)
		}
	}
	return res, nil
}

type mockPriceService struct {
	median int64
}

func (m *mockPriceService) Median7d(ctx context.Context, productID int64) (int64, error) {
	return m.median, nil
}

type mockDealService struct {
	verdict DealVerdict
}

func (m *mockDealService) DetectFakeSale(ctx context.Context, productID int64, price int64, listPrice *int64) (DealVerdict, error) {
	return m.verdict, nil
}

type mockStateRepo struct {
	state map[int64]bool
}

func (m *mockStateRepo) LastConditionMet(ctx context.Context, ruleID int64) (bool, error) {
	return m.state[ruleID], nil
}

func (m *mockStateRepo) Set(ctx context.Context, ruleID int64, met bool) error {
	if m.state == nil {
		m.state = make(map[int64]bool)
	}
	m.state[ruleID] = met
	return nil
}

type mockHandoff struct {
	alerts   int
	enqueues int
}

func (m *mockHandoff) CreateAndEnqueue(ctx context.Context, r track.AlertRule, payload map[string]any) (int64, error) {
	m.alerts++
	m.enqueues++
	return int64(m.alerts), nil
}

func (m *mockHandoff) AlertCount(ruleID int64) int {
	return m.alerts // simplified
}

func (m *mockHandoff) EnqueueCount() int {
	return m.enqueues
}

func (m *mockHandoff) DirectSendCount() int {
	return 0 // By design, Handoff just enqueues, it doesn't send directly.
}

func setupEngine(t *testing.T) (*Engine, *mockHandoff) {
	e, h := setupEngineWithMedian(t, 0)
	return e, h
}

func setupEngineWithMedian(t *testing.T, median int64) (*Engine, *mockHandoff) {
	h := &mockHandoff{}
	e := NewEngine(
		&mockRuleRepo{},
		&mockPriceService{median: median},
		&mockDealService{verdict: SaleXin},
		&mockStateRepo{},
		h,
	)
	return e, h
}

func seedRule(t *testing.T, e *Engine, ruleType string, threshold *int64) int64 {
	repo := e.rules.(*mockRuleRepo)
	id := int64(len(repo.rules) + 1)
	repo.rules = append(repo.rules, track.AlertRule{
		ID:        id,
		ProductID: 1,
		RuleType:  ruleType,
		Threshold: threshold,
		Active:    true,
	})
	return id
}

func ptr(v int64) *int64 { return &v }

var ctx = context.Background()
var pid = int64(1)

func TestPriceBelow_Met(t *testing.T) {
	e, h := setupEngine(t)
	rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
	if err := e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}); err != nil {
		t.Fatal(err)
	}
	if h.AlertCount(rid) != 1 {
		t.Errorf("Expected 1 alert")
	}
	if h.EnqueueCount() != 1 {
		t.Errorf("Expected 1 enqueue")
	}
	if h.DirectSendCount() != 0 {
		t.Errorf("Expected 0 direct sends")
	}
}

func (e *Engine) evalOnce(t *testing.T, snap Snapshot) bool {
	rules, _ := e.rules.ActiveByProduct(ctx, snap.ProductID)
	met, _, _ := e.conditionMet(ctx, rules[0], snap)
	return met
}

func TestDropPct_Median7Reference(t *testing.T) {
	e, _ := setupEngineWithMedian(t, 100_000)
	seedRule(t, e, "drop_pct", ptr(int64(20)))
	metLow := e.evalOnce(t, Snapshot{ProductID: pid, Price: 75_000})
	metHigh := e.evalOnce(t, Snapshot{ProductID: pid, Price: 85_000})
	if !metLow {
		t.Errorf("Expected metLow=true")
	}
	if metHigh {
		t.Errorf("Expected metHigh=false")
	}
}

func TestRealSale_OnlyOnSaleXin(t *testing.T) {
	e, _ := setupEngine(t)
	seedRule(t, e, "real_sale", nil)
	
	e.deal.(*mockDealService).verdict = SaleXin
	if !e.evalOnce(t, Snapshot{ProductID: pid, Price: 79_000}) {
		t.Errorf("Expected true for SaleXin")
	}
	
	e.deal.(*mockDealService).verdict = SaleAo
	if e.evalOnce(t, Snapshot{ProductID: pid, Price: 79_000}) {
		t.Errorf("Expected false for SaleAo")
	}
}

func TestBottomPredicted_Skipped(t *testing.T) {
	e, h := setupEngine(t)
	rid := seedRule(t, e, "bottom_predicted", nil)
	e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 1})
	if h.AlertCount(rid) != 0 {
		t.Errorf("Expected 0 alert count")
	}
}
