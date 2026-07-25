package cart

import (
	"context"
	"database/sql"
	"fmt"
)

type SnapshotRepo struct {
	db *sql.DB
}

func NewSnapshotRepo(db *sql.DB) *SnapshotRepo {
	return &SnapshotRepo{db: db}
}

func (r *SnapshotRepo) InsertSnapshot(ctx context.Context, s *CartSnapshot) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Check if snapshot_ref already exists for this user (idempotent)
	var existingID int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM cart_snapshot WHERE user_id = $1 AND snapshot_ref = $2", s.UserID, s.SnapshotRef).Scan(&existingID)
	if err == nil {
		s.ID = existingID
		return nil // already exists, idempotent return
	} else if err != sql.ErrNoRows {
		return err
	}

	// 2. Insert snapshot
	err = tx.QueryRowContext(ctx,
		"INSERT INTO cart_snapshot (user_id, platform_id, snapshot_ref, captured_at) VALUES ($1, $2, $3, $4) RETURNING id",
		s.UserID, s.PlatformID, s.SnapshotRef, s.CapturedAt,
	).Scan(&s.ID)
	if err != nil {
		return err
	}

	// 3. Insert items
	if len(s.Items) > 0 {
		stmt, err := tx.PrepareContext(ctx, "INSERT INTO cart_item (cart_snapshot_id, product_id, platform_item_id, shop_id, qty, unit_price) VALUES ($1, $2, $3, $4, $5, $6)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for i, item := range s.Items {
			_, err = stmt.ExecContext(ctx, s.ID, item.ProductID, item.PlatformItemID, item.ShopID, item.Qty, item.UnitPrice)
			if err != nil {
				return err
			}
			s.Items[i].CartSnapshotID = s.ID
		}
	}

	return tx.Commit()
}

func (r *SnapshotRepo) GetSnapshot(ctx context.Context, id int64, userID int64) (*CartSnapshot, error) {
	s := &CartSnapshot{}
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, platform_id, snapshot_ref, captured_at FROM cart_snapshot WHERE id = $1 AND user_id = $2", id, userID).
		Scan(&s.ID, &s.UserID, &s.PlatformID, &s.SnapshotRef, &s.CapturedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("snapshot not found or unauthorized")
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, "SELECT id, product_id, platform_item_id, shop_id, qty, unit_price FROM cart_item WHERE cart_snapshot_id = $1", s.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.PlatformItemID, &item.ShopID, &item.Qty, &item.UnitPrice); err != nil {
			return nil, err
		}
		item.CartSnapshotID = s.ID
		s.Items = append(s.Items, item)
	}

	return s, nil
}
