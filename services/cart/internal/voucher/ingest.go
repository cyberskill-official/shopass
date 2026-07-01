package voucher

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/cart/internal/cart"
)

type Ingestor struct {
	pool *pgxpool.Pool
}

func NewIngestor(pool *pgxpool.Pool) *Ingestor {
	return &Ingestor{pool: pool}
}

func (i *Ingestor) Upsert(ctx context.Context, v Voucher) error {
	if err := validate(v); err != nil {
		return err // (DEC-CART-35)
	}
	_, err := i.pool.Exec(ctx,
		`INSERT INTO voucher_catalog
           (platform_id, code, type, discount_type, discount_value, min_spend, cap, shop_id, valid_from, valid_to, stack_group)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
         ON CONFLICT (platform_id, code) DO UPDATE SET
           type=EXCLUDED.type, discount_type=EXCLUDED.discount_type, discount_value=EXCLUDED.discount_value,
           min_spend=EXCLUDED.min_spend, cap=EXCLUDED.cap, shop_id=EXCLUDED.shop_id, valid_from=EXCLUDED.valid_from,
           valid_to=EXCLUDED.valid_to, stack_group=EXCLUDED.stack_group`,
		v.PlatformID, v.Code, v.Type, v.DiscountType, v.DiscountValue,
		v.MinSpend, v.Cap, v.ShopID, v.ValidFrom, v.ValidTo, v.StackGroup)
	return err
}

func validate(v Voucher) error {
	switch v.Type {
	case TypeShop:
		if v.ShopID == nil {
			return cart.ErrShopVoucherNeedsShopID
		}
	case TypePlatform, TypeFreeship:
		if v.ShopID != nil {
			return cart.ErrPlatformVoucherHasShopID
		}
	default:
		return cart.ErrUnknownVoucherType
	}

	if v.DiscountType == DiscountPercent && (v.DiscountValue < 1 || v.DiscountValue > 100) {
		return cart.ErrPercentOutOfRange
	}
	if v.DiscountType != DiscountAmount && v.DiscountType != DiscountPercent {
		return cart.ErrPercentOutOfRange // Using this as fallback for now
	}

	if v.DiscountValue <= 0 {
		return cart.ErrNonPositiveDiscount
	}
	if v.ValidTo.Before(v.ValidFrom) {
		return cart.ErrInvalidWindow
	}
	return nil
}
