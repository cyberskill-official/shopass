package affil

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUnknownSubID     = errors.New("unknown sub_id")
	ErrConversionExists = errors.New("conversion already exists for this click")
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// RecordClick inserts an affiliate click that is user-initiated.
func (r *Repo) RecordClick(ctx context.Context, c AffiliateClick) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO affiliate_click (user_id, platform_id, product_id, sub_id, network)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		c.UserID, c.PlatformID, c.ProductID, c.SubID, c.Network).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// RecordConversion finds click by sub_id (last-click), inserts pending conversion.
func (r *Repo) RecordConversion(ctx context.Context, subID string, orderValue, commission int64, network string) (int64, error) {
	var clickID int64
	err := r.pool.QueryRow(ctx,
		`SELECT id FROM affiliate_click WHERE sub_id = $1`, subID).Scan(&clickID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUnknownSubID
	} else if err != nil {
		return 0, err
	}

	var id int64
	err = r.pool.QueryRow(ctx,
		`INSERT INTO affiliate_conversion (click_id, order_value, commission)
		 VALUES ($1,$2,$3)
		 ON CONFLICT (click_id) DO NOTHING
		 RETURNING id`,
		clickID, orderValue, commission).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrConversionExists
	}
	return id, err
}

func (r *Repo) ConfirmConversion(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE affiliate_conversion
		 SET status='confirmed', confirmed_at=now()
		 WHERE id=$1 AND status='pending'`, id)
	return err
}

func (r *Repo) RejectConversion(ctx context.Context, id int64, reason string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE affiliate_conversion
		 SET status='rejected'
		 WHERE id=$1 AND status='pending'`, id) // status can also just go to rejected
	return err
}
