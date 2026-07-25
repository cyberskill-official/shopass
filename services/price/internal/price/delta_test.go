package price

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func ptr64(v int64) *int64 { return &v }
func ptr32(v int32) *int32 { return &v }

var t0 = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

func TestDelta_NoChange_NotChanged(t *testing.T) {
	a := PriceSnapshot{ProductID: 1, TS: t0, Price: 100_000, FlashSale: false}
	b := PriceSnapshot{ProductID: 1, TS: t0.Add(time.Hour), Price: 100_000, FlashSale: false}
	require.False(t, changed(a, b))
}

func TestDelta_PriceChange_IsChanged(t *testing.T) {
	a := PriceSnapshot{ProductID: 1, TS: t0, Price: 100_000}
	b := PriceSnapshot{ProductID: 1, TS: t0.Add(time.Hour), Price: 89_000}
	require.True(t, changed(a, b))
}

func TestDelta_FlashFlip_IsChanged(t *testing.T) {
	a := PriceSnapshot{ProductID: 1, TS: t0, Price: 100_000, FlashSale: false}
	b := PriceSnapshot{ProductID: 1, TS: t0.Add(time.Minute), Price: 100_000, FlashSale: true}
	require.True(t, changed(a, b)) // flash flip là tín hiệu, dù giá bằng
}

func TestDelta_ListPriceChange_IsChanged(t *testing.T) {
	a := PriceSnapshot{Price: 100_000, ListPrice: ptr64(149_000)}
	b := PriceSnapshot{Price: 100_000, ListPrice: ptr64(199_000)}
	require.True(t, changed(a, b))
}

func TestDelta_ListPriceNilToSet_IsChanged(t *testing.T) {
	a := PriceSnapshot{Price: 100_000, ListPrice: nil}
	b := PriceSnapshot{Price: 100_000, ListPrice: ptr64(149_000)}
	require.True(t, changed(a, b))
}

func TestDelta_StockChange_IsChanged(t *testing.T) {
	a := PriceSnapshot{Price: 100_000, Stock: ptr32(10)}
	b := PriceSnapshot{Price: 100_000, Stock: ptr32(5)}
	require.True(t, changed(a, b))
}

func TestDelta_SoldNotTracked_NoChange(t *testing.T) {
	// sold is NOT in the delta comparison — only price, list_price, stock, flash_sale
	a := PriceSnapshot{Price: 100_000, Sold: ptr32(100)}
	b := PriceSnapshot{Price: 100_000, Sold: ptr32(200)}
	require.False(t, changed(a, b))
}

func TestDelta_AllFieldsSame_NotChanged(t *testing.T) {
	a := PriceSnapshot{
		Price: 100_000, ListPrice: ptr64(149_000),
		Stock: ptr32(10), FlashSale: false,
	}
	b := PriceSnapshot{
		Price: 100_000, ListPrice: ptr64(149_000),
		Stock: ptr32(10), FlashSale: false,
	}
	require.False(t, changed(a, b))
}

func TestEqPtr64(t *testing.T) {
	require.True(t, eqPtr64(nil, nil))
	require.False(t, eqPtr64(nil, ptr64(1)))
	require.False(t, eqPtr64(ptr64(1), nil))
	require.True(t, eqPtr64(ptr64(42), ptr64(42)))
	require.False(t, eqPtr64(ptr64(1), ptr64(2)))
}

func TestEqPtr32(t *testing.T) {
	require.True(t, eqPtr32(nil, nil))
	require.False(t, eqPtr32(nil, ptr32(1)))
	require.True(t, eqPtr32(ptr32(7), ptr32(7)))
}
