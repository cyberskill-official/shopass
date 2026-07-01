package shopee

import (
	"errors"
	"time"

	"shopass/services/scrape/internal/orchestrator"
)

const shopeePriceUnit = 100_000 // Shopee trả giá theo micro-đơn-vị

var ErrItemGone = errors.New("shopee: item removed or unavailable")

type pdpResponse struct {
	Error int `json:"error"`
	Data  struct {
		Item struct {
			Price           int64 `json:"price"` // micro-VND
			PriceBeforeDisc int64 `json:"price_before_discount"`
			Stock           int32 `json:"stock"`
			HistoricalSold  int32 `json:"historical_sold"`
			FlashSale       *struct {
				Status int `json:"status"`
			} `json:"flash_sale"`
		} `json:"item"`
	} `json:"data"`
}

// ptr returns a pointer to v.
func ptr[T any](v T) *T {
	return &v
}

func (r pdpResponse) toSnapshot(productID int64, ts time.Time) (orchestrator.PriceSnapshot, error) {
	if r.Error != 0 {
		return orchestrator.PriceSnapshot{}, ErrItemGone
	}
	it := r.Data.Item
	snap := orchestrator.PriceSnapshot{
		ProductID: productID,
		TS:        ts,
		Price:     it.Price / shopeePriceUnit,
		Stock:     ptr(it.Stock),
		Sold:      ptr(it.HistoricalSold),
		FlashSale: it.FlashSale != nil && it.FlashSale.Status == 1,
	}
	if it.PriceBeforeDisc > 0 {
		snap.ListPrice = ptr(it.PriceBeforeDisc / shopeePriceUnit)
	}
	return snap, nil
}
