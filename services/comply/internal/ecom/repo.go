package ecom

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) txCount(ctx context.Context, year int) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, "SELECT count FROM yearly_transaction_count WHERE year=$1", year).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repo) threshold(ctx context.Context, key string) (int64, error) {
	var val int64
	err := r.pool.QueryRow(ctx, "SELECT value FROM compliance_threshold WHERE key=$1 ORDER BY version DESC LIMIT 1", key).Scan(&val)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (r *Repo) Obligations(ctx context.Context) ([]EcommerceObligation, error) {
	rows, err := r.pool.Query(ctx, "SELECT obligation_key, description_vi, source_law, status FROM ecommerce_obligation")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var obs []EcommerceObligation
	for rows.Next() {
		var o EcommerceObligation
		if err := rows.Scan(&o.ObligationKey, &o.DescriptionVi, &o.SourceLaw, &o.Status); err != nil {
			return nil, err
		}
		obs = append(obs, o)
	}
	return obs, rows.Err()
}

func (r *Repo) MarkObligation(ctx context.Context, key string, status string) error {
	if status != "not_started" && status != "submitted" && status != "approved" && status != "done" && status != "n_a" {
		return fmt.Errorf("invalid status") // CHECK constraint equivalent
	}

	res, err := r.pool.Exec(ctx, "UPDATE ecommerce_obligation SET status=$1 WHERE obligation_key=$2", status, key)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("obligation not found")
	}
	return nil
}
