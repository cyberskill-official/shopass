package optimizer

// discountAmount tính phần giảm của một voucher trên một giá trị đơn (int64 VND).
func discountAmount(v Voucher, base int64) int64 {
	var d int64
	switch v.DiscountType {
	case DiscountAmount:
		d = v.DiscountValue
	case DiscountPercent:
		d = base * v.DiscountValue / 100 // chia nguyên, KHÔNG float (DEC-CART-15)
	}
	return applyCap(d, v.Cap)
}

// applyCap kẹp phần giảm theo trần tuyệt đối của voucher (DEC-CART-15, §1 #7).
func applyCap(d int64, cap *int64) int64 {
	if cap != nil && d > *cap {
		return *cap
	}
	return d
}

// freeshipValue tính phần giảm của voucher freeship trên một giá trị đơn (int64 VND).
func freeshipValue(fs *Voucher, base int64) int64 {
	if fs == nil {
		return 0
	}
	return discountAmount(*fs, base)
}
