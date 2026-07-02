package price

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// CompareRow is one platform's current price for a canonical_key (FR-PRICE-004).
// platform has only a `code` column (FR-INFRA-002 owns id/code/country/base_url) -
// there is no `name`, so the display name is derived from `code` server-side.
type CompareRow struct {
	PlatformCode string
	ProductID    int64
	Price        int64     // VND, no decimals (DEC-PRICE-05)
	TS           time.Time // freshness of this platform's price (DEC-PRICE-43)
	PlatformItem string
}

// CompareByCanonicalKey returns the current price of every platform listing that
// shares a canonical_key: the latest snapshot per product (a per-product LATERAL
// picking the newest ts, DEC-PRICE-41), joined to tracked_product filtered by
// canonical_key and to platform for its code, ordered cheapest first. It reads
// price_snapshot directly (not price_daily) so the real second/minute ts is kept.
func (r *Repo) CompareByCanonicalKey(ctx context.Context, key string) ([]CompareRow, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pf.code AS platform_code,
		       tp.id   AS product_id,
		       ls.price,
		       ls.ts,
		       tp.platform_item_id
		FROM tracked_product tp
		JOIN platform pf ON pf.id = tp.platform_id
		JOIN LATERAL (
			SELECT ps.price, ps.ts
			FROM price_snapshot ps
			WHERE ps.product_id = tp.id
			ORDER BY ps.ts DESC
			LIMIT 1
		) ls ON true
		WHERE tp.canonical_key = $1
		ORDER BY ls.price ASC`, key)
	if err != nil {
		return nil, err
	}
	return scanCompareRows(rows)
}

func scanCompareRows(rows pgx.Rows) ([]CompareRow, error) {
	defer rows.Close()
	var out []CompareRow
	for rows.Next() {
		var c CompareRow
		if err := rows.Scan(&c.PlatformCode, &c.ProductID, &c.Price, &c.TS, &c.PlatformItem); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DisplayName derives the human-facing platform name from platform.code, since
// the schema (FR-INFRA-002) has no name column. Server-side so web and app show
// the same label (DEC-PRICE-42 spirit: one source of truth).
func DisplayName(code string) string {
	switch code {
	case "shopee":
		return "Shopee"
	case "tiktok":
		return "TikTok Shop"
	case "lazada":
		return "Lazada"
	default:
		return code
	}
}
