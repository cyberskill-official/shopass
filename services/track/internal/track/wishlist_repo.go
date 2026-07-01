package track

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type Wishlist struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type WishlistItem struct {
	ID          int64     `json:"id"`
	WishlistID  int64     `json:"wishlist_id"`
	ProductID   int64     `json:"product_id"`
	TargetPrice *int64    `json:"target_price"`
	AddedAt     time.Time `json:"added_at"`
}

// WishlistRepo interface for dependency injection
type WishlistRepo interface {
	CreateWishlist(ctx context.Context, userID int64, name string) (Wishlist, error)
	ListWishlists(ctx context.Context, userID int64) ([]Wishlist, error)
	OwnsWishlist(ctx context.Context, userID, wishlistID int64) (bool, error)
	AddItem(ctx context.Context, wishlistID, productID int64, target *int64) error
	RemoveItem(ctx context.Context, wishlistID, productID int64) error
	DeleteWishlist(ctx context.Context, wishlistID int64) error
	ListItems(ctx context.Context, wishlistID int64) ([]WishlistItem, error)
}

type wishlistRepoImpl struct {
	db *sql.DB
}

func NewWishlistRepo(db *sql.DB) WishlistRepo {
	return &wishlistRepoImpl{db: db}
}

func (r *wishlistRepoImpl) CreateWishlist(ctx context.Context, userID int64, name string) (Wishlist, error) {
	var w Wishlist
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO wishlist (user_id, name)
		VALUES ($1, $2)
		RETURNING id, user_id, name, created_at
	`, userID, name).Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt)
	return w, err
}

func (r *wishlistRepoImpl) ListWishlists(ctx context.Context, userID int64) ([]Wishlist, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, name, created_at
		FROM wishlist
		WHERE user_id = $1
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []Wishlist
	for rows.Next() {
		var w Wishlist
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, w)
	}
	return lists, rows.Err()
}

func (r *wishlistRepoImpl) OwnsWishlist(ctx context.Context, userID, wishlistID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM wishlist WHERE id = $1 AND user_id = $2)
	`, wishlistID, userID).Scan(&ok)
	return ok, err
}

func (r *wishlistRepoImpl) AddItem(ctx context.Context, wishlistID, productID int64, target *int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wishlist_item (wishlist_id, product_id, target_price)
		VALUES ($1, $2, $3)
		ON CONFLICT (wishlist_id, product_id)
		DO UPDATE SET target_price = EXCLUDED.target_price
	`, wishlistID, productID, target)
	return err
}

func (r *wishlistRepoImpl) RemoveItem(ctx context.Context, wishlistID, productID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM wishlist_item WHERE wishlist_id = $1 AND product_id = $2
	`, wishlistID, productID)
	return err
}

func (r *wishlistRepoImpl) DeleteWishlist(ctx context.Context, wishlistID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM wishlist WHERE id = $1
	`, wishlistID)
	return err
}

func (r *wishlistRepoImpl) ListItems(ctx context.Context, wishlistID int64) ([]WishlistItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, wishlist_id, product_id, target_price, added_at
		FROM wishlist_item
		WHERE wishlist_id = $1
		ORDER BY id DESC
	`, wishlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WishlistItem
	for rows.Next() {
		var it WishlistItem
		if err := rows.Scan(&it.ID, &it.WishlistID, &it.ProductID, &it.TargetPrice, &it.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// IsFKViolation checks if err is a postgres foreign key violation (23503)
func IsFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}
