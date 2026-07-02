package optimizer

// StackRules trừu tượng luật xếp chồng per-country; FR-CART-004 hiện thực cụ thể.
type StackRules interface {
	// ValidStack trả true nếu tổ hợp (platform voucher, freeship, các shop voucher) được phép
	// theo luật của nước; FR-CART-003 KHÔNG biết luật cụ thể (DEC-CART-14).
	ValidStack(pv *Voucher, fs *Voucher, shopVouchers []Voucher) bool
}
