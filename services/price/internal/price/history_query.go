package price

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// rangeWindows là allowlist range -> khoảng thời gian (DEC-PRICE-31).
var rangeWindows = map[string]time.Duration{
	"7d":   7 * 24 * time.Hour,
	"30d":  30 * 24 * time.Hour,
	"90d":  90 * 24 * time.Hour,
	"180d": 180 * 24 * time.Hour,
	"1y":   365 * 24 * time.Hour,
}

// ParseRange validate range, trả khoảng + ok=false nếu ngoài allowlist.
func ParseRange(raw string) (time.Duration, bool) {
	if raw == "" {
		return rangeWindows["90d"], true // default 90d
	}
	d, ok := rangeWindows[raw]
	return d, ok
}

type DailyPoint struct {
	Day    time.Time `json:"day"`
	MinP   int64     `json:"min_p"`   // VND
	MaxP   int64     `json:"max_p"`   // VND
	CloseP int64     `json:"close_p"` // VND, giá hiển thị (DEC-PRICE-33)
}

func scanDaily(rows pgx.Rows) ([]DailyPoint, error) {
	defer rows.Close()
	var out []DailyPoint
	for rows.Next() {
		var p DailyPoint
		if err := rows.Scan(&p.Day, &p.MinP, &p.MaxP, &p.CloseP); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// QueryDailyBody đọc phần thân từ price_daily (DEC-PRICE-30).
func (r *Repo) QueryDailyBody(ctx context.Context, productID int64, from time.Time) ([]DailyPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT day, min_p, max_p, close_p
         FROM price_daily
         WHERE product_id = $1 AND day >= $2
         ORDER BY day`, productID, from)
	if err != nil {
		return nil, err
	}
	return scanDaily(rows)
}

type TailPoint struct {
	TS        time.Time `json:"ts"`
	Price     int64     `json:"price"` // VND
	FlashSale bool      `json:"flash_sale"`
}

func scanTail(rows pgx.Rows) ([]TailPoint, error) {
	defer rows.Close()
	var out []TailPoint
	for rows.Next() {
		var p TailPoint
		if err := rows.Scan(&p.TS, &p.Price, &p.FlashSale); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// QueryRawTail đọc raw kể từ đầu ngày hôm nay - tối đa một chunk nóng (DEC-PRICE-32).
func (r *Repo) QueryRawTail(ctx context.Context, productID int64) ([]TailPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ts, price, flash_sale
         FROM price_snapshot
         WHERE product_id = $1 AND ts >= date_trunc('day', now())
         ORDER BY ts`, productID)
	if err != nil {
		return nil, err
	}
	return scanTail(rows)
}

// ProductExists phân biệt 404 với 200-rỗng (DEC-PRICE-34).
func (r *Repo) ProductExists(ctx context.Context, productID int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tracked_product WHERE id = $1)`, productID).Scan(&ok)
	return ok, err
}
