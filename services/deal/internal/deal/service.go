package deal

import (
	"context"

	"shopass/services/deal/internal/fakesale"
)

type PriceRepo interface {
	// QueryRange trả về danh sách giá trong 90 ngày qua (ví dụ: mảng int64)
	QueryRange(ctx context.Context, productID int64, fromUnix, toUnix int64) ([]int64, error)
}

type Service struct {
	price PriceRepo
}

func NewService(price PriceRepo) *Service {
	return &Service{
		price: price,
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
	// TODO: Dùng thư viện time.Now() để lấy range thực tế
	// hist, err := s.price.QueryRange(ctx, productID, time.Now().Unix()-7776000, time.Now().Unix())
	
	// Mock history cho build thành công
	hist := []int64{}
	
	verdict := fakesale.DetectFakeSale(hist, currentPrice, lp)
	
	// TODO: Phát OTel counter fake_sale_verdict_total{verdict}
	
	return verdict, nil
}
