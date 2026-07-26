package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/deal/internal/api"
	"shopass/services/deal/internal/chart"
	"shopass/services/deal/internal/coldstart"
	"shopass/services/deal/internal/fakesale"
)

type chartQueryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// chartRepo implements api.Repo for the chart endpoint. Chart access is scoped
// through user_tracked_product before price_daily is read.
type chartRepo struct{ pool chartQueryer }

func (r *chartRepo) UserCanViewProduct(ctx context.Context, userID, productID int64) (bool, error) {
	var ok bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1
			FROM user_tracked_product
			WHERE user_id = $1 AND product_id = $2
		)`, userID, productID).Scan(&ok)
	return ok, err
}

func (r *chartRepo) QueryDaily(ctx context.Context, productID int64, from time.Time) ([]chart.DailyPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT day, min_p, max_p, close_p
		 FROM price_daily
		 WHERE product_id = $1 AND day >= $2
		 ORDER BY day`, productID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []chart.DailyPoint
	for rows.Next() {
		var p chart.DailyPoint
		if err := rows.Scan(&p.Day, &p.MinP, &p.MaxP, &p.CloseP); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *chartRepo) FindProductID(ctx context.Context, platformCode, platformItemID string) (int64, bool, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `
		SELECT tp.id
		FROM tracked_product tp
		JOIN platform p ON p.id = tp.platform_id
		WHERE lower(p.code) = lower($1) AND tp.platform_item_id = $2
		LIMIT 1
	`, platformCode, platformItemID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return id, true, nil
}

func (r *chartRepo) QueryRawTail(ctx context.Context, productID int64) ([]api.SnapshotPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ts, price
		 FROM price_snapshot
		 WHERE product_id = $1 AND ts >= now() - INTERVAL '2 hours'
		 ORDER BY ts`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []api.SnapshotPoint
	for rows.Next() {
		var point api.SnapshotPoint
		if err := rows.Scan(&point.TS, &point.Price); err != nil {
			return nil, err
		}
		out = append(out, point)
	}
	return out, rows.Err()
}

// dealService implements api.DealService: maturity from days of history
// (FR-DEAL-002) and the fake-sale verdict from the daily series (FR-DEAL-001).
type dealService struct{ pool *pgxpool.Pool }

func (s *dealService) daysOfHistory(ctx context.Context, productID int64) int {
	var days int
	if err := s.pool.QueryRow(ctx,
		`SELECT GREATEST(0, (CURRENT_DATE - first_seen::date))::int
		 FROM tracked_product WHERE id = $1`, productID).Scan(&days); err != nil {
		return 0
	}
	return days
}

func (s *dealService) Maturity(ctx context.Context, productID int64) string {
	switch coldstart.Maturity(s.daysOfHistory(ctx, productID)) {
	case coldstart.StateMature:
		return "MATURE"
	case coldstart.StateWarming:
		return "WARMING"
	default:
		return "NEW"
	}
}

func (s *dealService) Verdict(ctx context.Context, productID int64) string {
	// hist = daily closes over ~90 days; current = latest close; list = window max
	// (a server-side proxy for list price). HandleChart overrides this to UNKNOWN
	// when maturity is NEW, so a thin cold-start series stays honest.
	rows, err := s.pool.Query(ctx,
		`SELECT close_p, max_p
		 FROM price_daily
		 WHERE product_id = $1 AND day >= CURRENT_DATE - 90
		 ORDER BY day`, productID)
	if err != nil {
		return string(fakesale.Unknown)
	}
	defer rows.Close()
	var hist []int64
	var current, list int64
	for rows.Next() {
		var closeP, maxP int64
		if err := rows.Scan(&closeP, &maxP); err != nil {
			return string(fakesale.Unknown)
		}
		hist = append(hist, closeP)
		current = closeP
		if maxP > list {
			list = maxP
		}
	}
	if len(hist) == 0 {
		return string(fakesale.Unknown)
	}
	return string(fakesale.DetectFakeSale(hist, current, list))
}
