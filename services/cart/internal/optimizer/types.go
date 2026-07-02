package optimizer

import (
	"shopass/services/cart/internal/cart"
	"shopass/services/cart/internal/voucher"
)

type CartItem = cart.CartItem
type Voucher = voucher.Voucher
type VoucherType = voucher.VoucherType
type DiscountType = voucher.DiscountType

const (
	TypeShop     = voucher.TypeShop
	TypePlatform = voucher.TypePlatform
	TypeFreeship = voucher.TypeFreeship

	DiscountAmount  = voucher.DiscountAmount
	DiscountPercent = voucher.DiscountPercent
)

type Vouchers struct {
	Shop     []Voucher
	Platform []Voucher
	Freeship []Voucher
}

type Breakdown struct {
	Shop     int64 `json:"shop"`
	Platform int64 `json:"platform"`
	Freeship int64 `json:"freeship"`
}

type OptimizeResult struct {
	Discount  int64     `json:"discount"`
	Combo     []Voucher `json:"combo"`
	Breakdown Breakdown `json:"breakdown"`
}
