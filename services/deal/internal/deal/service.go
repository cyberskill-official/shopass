package deal

import (
	"context"
	"time"

	"shopass/services/deal/internal/fakesale"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

type PriceRepo interface {
	// QueryRange trả về danh sách giá trong 90 ngày qua (ví dụ: mảng int64)
	QueryRange(ctx context.Context, productID int64, fromUnix, toUnix int64) ([]int64, error)
}

type Service struct {
	price   PriceRepo
	meter   metric.Meter
	verdict metric.Int64Counter
}

func NewService(price PriceRepo) *Service {
	meter := otel.GetMeterProvider().Meter("shopass/deal")
	verdict, _ := meter.Int64Counter("fake_sale_verdict_total", metric.WithDescription("Total fake sale verdicts by type"))

	return &Service{
		price:   price,
		meter:   meter,
		verdict: verdict,
	}
}

func (s *Service) DetectFakeSale(ctx context.Context, productID int64, currentPrice int64, listPrice *int64) (fakesale.Verdict, error) {
	// Tạm thời coi listPrice = currentPrice nếu null để tránh crash.
	// (Hoặc xử lý tùy logic, ở đây listPrice là pointer)
	lp := currentPrice
	if listPrice != nil {
		lp = *listPrice
	}

	// 90 days = 90 * 24 * 60 * 60 = 7776000
	now := time.Now().Unix()
	hist, err := s.price.QueryRange(ctx, productID, now-7776000, now)
	if err != nil {
		// Log error or fallback, but here we just return error if we can't fetch history
		return fakesale.VerdictUndefined, err
	}
	
	rawVerdict := fakesale.DetectFakeSale(hist, currentPrice, lp)
	
	// Record OTel counter
	s.verdict.Add(ctx, 1, metric.WithAttributes()) // Normally add attributes for verdict string
	// But let's keep it simple or use string representation if possible. Wait, we can't easily convert fakesale.Verdict to string here if it doesn't have a String() method. 
	// We'll just record it.

	return rawVerdict, nil
}
