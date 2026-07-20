package track

import (
	"context"
	"database/sql"
	"time"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

// LinkUserProduct links a user to a tracked product idempotently.
// Returns true if a new link was created, false if it already existed.
func (r *Repo) LinkUserProduct(ctx context.Context, userID, productID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO user_tracked_product (user_id, product_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, product_id) DO NOTHING
	`, userID, productID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// UserCanViewProduct proves ownership without exposing a product registry row
// to another account. It is used before a browser-assisted price is accepted.
func (r *Repo) UserCanViewProduct(ctx context.Context, userID, productID int64) (bool, error) {
	var allowed bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM user_tracked_product
			WHERE user_id = $1 AND product_id = $2
		)
	`, userID, productID).Scan(&allowed)
	return allowed, err
}

// UserTrackedProduct is the owner-scoped product summary used by the closed
// beta dashboard. It intentionally contains registry metadata only: price and
// alert state have their own service-owned read paths.
type UserTrackedProduct struct {
	ProductID      int64     `json:"product_id"`
	Platform       string    `json:"platform"`
	PlatformItemID string    `json:"platform_item_id"`
	FirstSeen      time.Time `json:"first_seen"`
	TrackedAt      time.Time `json:"tracked_at"`
}

// ListUserTrackedProducts returns only products linked to the supplied user.
// The user_id predicate is deliberately in the query rather than accepted as
// a URL parameter, so callers cannot enumerate another user's dashboard.
func (r *Repo) ListUserTrackedProducts(ctx context.Context, userID int64) ([]UserTrackedProduct, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tp.id, p.code, tp.platform_item_id, tp.first_seen, utp.tracked_at
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

	products := make([]UserTrackedProduct, 0)
	for rows.Next() {
		var product UserTrackedProduct
		if err := rows.Scan(
			&product.ProductID,
			&product.Platform,
			&product.PlatformItemID,
			&product.FirstSeen,
			&product.TrackedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}
