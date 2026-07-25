package sink

import (
	"context"
	"testing"
	"time"
)

type spyRepo struct {
	insertCalls int
	returns     bool
}

func (s *spyRepo) InsertSnapshot(ctx context.Context, snap PriceSnapshot) (bool, error) {
	s.insertCalls++
	return s.returns, nil
}

type dummyMetrics struct{}

func (d *dummyMetrics) SnapshotWritten(productID int64) {}
func (d *dummyMetrics) SnapshotSkipped(productID int64) {}

func TestSink_DelegatesToInsertSnapshot(t *testing.T) {
	spy := &spyRepo{returns: true}
	s := NewSink(spy, &dummyMetrics{})
	ctx := context.Background()

	s.Write(ctx, PriceSnapshot{ProductID: 1, TS: time.Now(), Price: 89_000})
	if spy.insertCalls != 1 {
		t.Errorf("Expected 1 call to InsertSnapshot, got %d", spy.insertCalls)
	}
}

func TestSink_NoChangeSkips(t *testing.T) {
	spy := &spyRepo{returns: false} // delta-only báo skip
	s := NewSink(spy, &dummyMetrics{})
	ctx := context.Background()

	written, _, err := s.Write(ctx, PriceSnapshot{ProductID: 1, TS: time.Now(), Price: 89_000})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if written {
		t.Errorf("Expected written to be false")
	}
}

func TestSink_ReturnsFlashForRetier(t *testing.T) {
	spy := &spyRepo{returns: true}
	s := NewSink(spy, &dummyMetrics{})
	ctx := context.Background()

	_, flash, _ := s.Write(ctx, PriceSnapshot{ProductID: 1, TS: time.Now(), Price: 89_000, FlashSale: true})
	if !flash {
		t.Errorf("Expected flash to be true for re-tier signal")
	}
}
