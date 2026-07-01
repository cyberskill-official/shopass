package track

import (
	"context"
	"database/sql"
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
