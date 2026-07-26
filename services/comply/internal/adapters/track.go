package adapters

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"shopass/services/comply/internal/dsar"
)

type Track struct {
	pool *pgxpool.Pool
}

func NewTrack(pool *pgxpool.Pool) *Track {
	return &Track{pool: pool}
}

func (t *Track) ByUser(ctx context.Context, userID int64) ([]dsar.ProductView, error) {
	rows, err := t.pool.Query(ctx, `
		SELECT tp.id, p.code, COALESCE(tp.title, '')
		FROM user_tracked_product utp
		JOIN tracked_product tp ON tp.id = utp.product_id
		JOIN platform p ON p.id = tp.platform_id
		WHERE utp.user_id = $1
		ORDER BY utp.tracked_at DESC, tp.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []dsar.ProductView
	for rows.Next() {
		var product dsar.ProductView
		if err := rows.Scan(&product.ID, &product.Platform, &product.Name); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (t *Track) HardDeleteByUser(ctx context.Context, userID int64) (int, error) {
	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	total := int64(0)
	statements := []string{
		`DELETE FROM alert_fired_state afs
		 USING alert_rule ar
		 WHERE afs.alert_rule_id = ar.id AND ar.user_id = $1`,
		`DELETE FROM alert a
		 USING alert_rule ar
		 WHERE a.alert_rule_id = ar.id AND ar.user_id = $1`,
		`DELETE FROM alert_rule WHERE user_id = $1`,
		`DELETE FROM wishlist_item wi
		 USING wishlist w
		 WHERE wi.wishlist_id = w.id AND w.user_id = $1`,
		`DELETE FROM wishlist WHERE user_id = $1`,
		`DELETE FROM user_tracked_product WHERE user_id = $1`,
	}
	for _, stmt := range statements {
		tag, err := tx.Exec(ctx, stmt, userID)
		if err != nil {
			return 0, err
		}
		total += tag.RowsAffected()
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return int(total), nil
}
