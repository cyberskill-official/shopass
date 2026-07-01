package price

// changed returns true if any price-relevant field differs between two snapshots.
// Delta-only (DEC-PRICE-04): only INSERT when at least one of
// (price, list_price, stock, flash_sale) changes.
func changed(a, b PriceSnapshot) bool {
	if a.Price != b.Price {
		return true
	}
	if !eqPtr64(a.ListPrice, b.ListPrice) {
		return true
	}
	if a.FlashSale != b.FlashSale {
		return true
	}
	if !eqPtr32(a.Stock, b.Stock) {
		return true
	}
	return false
}

func eqPtr64(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func eqPtr32(a, b *int32) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
