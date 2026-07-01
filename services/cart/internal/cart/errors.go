package cart

import "errors"

var (
	ErrShopVoucherNeedsShopID   = errors.New("shop voucher must have shop_id")
	ErrPlatformVoucherHasShopID = errors.New("platform/freeship voucher must not have shop_id")
	ErrUnknownVoucherType       = errors.New("unknown voucher type")
	ErrPercentOutOfRange        = errors.New("discount percent must be between 1 and 100")
	ErrNonPositiveDiscount      = errors.New("discount value must be positive")
	ErrInvalidWindow            = errors.New("valid_to must be after valid_from")
)
