package engine

import (
	"testing"
)

func TestRisingEdge_NoSpam(t *testing.T) {
	e, h := setupEngine(t)
	rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
	for i := 0; i < 3; i++ {
		if err := e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}); err != nil {
			t.Fatal(err)
		}
	}
	if h.AlertCount(rid) != 1 {
		t.Errorf("Expected 1 alert, not three")
	}
}

func TestRisingEdge_ResetAllowsRefire(t *testing.T) {
	e, h := setupEngine(t)
	rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
	
	e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000})
	e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 99_000}) // reset
	e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}) // refire
	
	if h.AlertCount(rid) != 2 {
		t.Errorf("Expected 2 alerts, got %d", h.AlertCount(rid))
	}
}
