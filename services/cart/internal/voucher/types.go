package voucher

import "time"

type VoucherType string

const (
	TypeShop     VoucherType = "shop"
	TypePlatform VoucherType = "platform"
	TypeFreeship VoucherType = "freeship"
)

type DiscountType string

const (
	DiscountAmount  DiscountType = "amount"
	DiscountPercent DiscountType = "percent"
)

type Voucher struct {
	ID            int64        `db:"id"`
	PlatformID    int16        `db:"platform_id"`
	Code          string       `db:"code"`
	Type          VoucherType  `db:"type"`
	DiscountType  DiscountType `db:"discount_type"`
	DiscountValue int64        `db:"discount_value"` // VND (amount) hoac % nguyen (percent)
	MinSpend      *int64       `db:"min_spend"`      // VND
	Cap           *int64       `db:"cap"`            // VND
	ShopID        *string      `db:"shop_id"`
	ValidFrom     time.Time    `db:"valid_from"`
	ValidTo       time.Time    `db:"valid_to"`
	StackGroup    *string      `db:"stack_group"`
}
