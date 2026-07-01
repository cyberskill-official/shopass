package sink

import (
	"context"
	"time"
)

// Giả định định nghĩa PriceSnapshot từ package price.
type PriceSnapshot struct {
	ProductID int64
	TS        time.Time
	Price     int64
	FlashSale bool
}

type PriceRepo interface {
	InsertSnapshot(ctx context.Context, snap PriceSnapshot) (bool, error)
}

type Metrics interface {
	SnapshotWritten(productID int64)
	SnapshotSkipped(productID int64)
}

type Sink struct {
	price   PriceRepo
	metrics Metrics
}

func NewSink(price PriceRepo, metrics Metrics) *Sink {
	return &Sink{price: price, metrics: metrics}
}

// Write đẩy snapshot vào PRICE qua delta-only; trả tín hiệu re-tier cho orchestrator.
func (s *Sink) Write(ctx context.Context, snap PriceSnapshot) (written, flashSale bool, err error) {
	written, err = s.price.InsertSnapshot(ctx, snap) // FR-PRICE-002 delta-only, ON CONFLICT DO NOTHING
	if err != nil {
		return false, snap.FlashSale, err
	}
	if written {
		s.metrics.SnapshotWritten(snap.ProductID)
	} else {
		s.metrics.SnapshotSkipped(snap.ProductID) // giá không đổi -> bỏ qua, không phải lỗi
	}
	return written, snap.FlashSale, nil
}
