package voucher

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// ListActive returns active vouchers for the given platform, filtering by shop IDs for shop vouchers.
func (r *Repo) ListActive(ctx context.Context, platformID int16, shopIDs []string, now time.Time) ([]Voucher, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, platform_id, code, type, discount_type, discount_value, min_spend, cap, shop_id, valid_from, valid_to, stack_group
		FROM voucher_catalog
		WHERE platform_id = $1
		  AND valid_from <= $2 AND valid_to >= $2
		  AND (type <> 'shop' OR shop_id = ANY($3))
	`, platformID, now, shopIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vouchers []Voucher
	for rows.Next() {
		var v Voucher
		if err := rows.Scan(
			&v.ID, &v.PlatformID, &v.Code, &v.Type, &v.DiscountType, &v.DiscountValue,
			&v.MinSpend, &v.Cap, &v.ShopID, &v.ValidFrom, &v.ValidTo, &v.StackGroup,
		); err != nil {
			return nil, err
		}
		vouchers = append(vouchers, v)
	}
	return vouchers, rows.Err()
}
