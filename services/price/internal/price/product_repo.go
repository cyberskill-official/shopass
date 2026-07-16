package price

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

// UnexportedPoolForTest exposes the pgxpool only for tests
func (r *Repo) UnexportedPoolForTest() *pgxpool.Pool {
	return r.pool
}

// Upsert chèn SKU mới hoặc cập nhật metadata khi (platform_id, platform_item_id) đã có.
// canonical_key KHÔNG được ghi ở đây - để NULL, TASK-PRICE-005 điền sau.
// first_seen KHÔNG bị nhánh DO UPDATE ghi đè (giữ mốc cold-start).
func (r *Repo) Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error) {
	var out TrackedProduct
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tracked_product
            (platform_id, platform_item_id, shop_id, title, category_id)
         VALUES ($1, $2, $3, $4, $5)
         ON CONFLICT (platform_id, platform_item_id) DO UPDATE
            SET shop_id     = EXCLUDED.shop_id,
                title       = EXCLUDED.title,
                category_id = EXCLUDED.category_id
         RETURNING id, platform_id, platform_item_id, shop_id, title,
                   category_id, canonical_key, first_seen`,
		p.PlatformID, p.PlatformItemID, p.ShopID, p.Title, p.CategoryID).
		Scan(&out.ID, &out.PlatformID, &out.PlatformItemID, &out.ShopID,
			&out.Title, &out.CategoryID, &out.CanonicalKey, &out.FirstSeen)
	if err != nil {
		return TrackedProduct{}, err
	}
	return out, nil
}

func (r *Repo) GetByID(ctx context.Context, id int64) (TrackedProduct, error) {
	var out TrackedProduct
	err := r.pool.QueryRow(ctx,
		`SELECT id, platform_id, platform_item_id, shop_id, title,
                category_id, canonical_key, first_seen
         FROM tracked_product WHERE id = $1`, id).
		Scan(&out.ID, &out.PlatformID, &out.PlatformItemID, &out.ShopID,
			&out.Title, &out.CategoryID, &out.CanonicalKey, &out.FirstSeen)
	return out, err
}

func (r *Repo) FindByPlatformItem(ctx context.Context, platformID int16, platformItemID string) (TrackedProduct, error) {
	var out TrackedProduct
	err := r.pool.QueryRow(ctx,
		`SELECT id, platform_id, platform_item_id, shop_id, title,
                category_id, canonical_key, first_seen
         FROM tracked_product WHERE platform_id = $1 AND platform_item_id = $2`, platformID, platformItemID).
		Scan(&out.ID, &out.PlatformID, &out.PlatformItemID, &out.ShopID,
			&out.Title, &out.CategoryID, &out.CanonicalKey, &out.FirstSeen)
	return out, err
}

func scanProducts(rows pgx.Rows) ([]TrackedProduct, error) {
	var out []TrackedProduct
	for rows.Next() {
		var p TrackedProduct
		if err := rows.Scan(&p.ID, &p.PlatformID, &p.PlatformItemID, &p.ShopID,
			&p.Title, &p.CategoryID, &p.CanonicalKey, &p.FirstSeen); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) GetByCanonicalKey(ctx context.Context, key string) ([]TrackedProduct, error) {
	if key == "" {
		return nil, fmt.Errorf("canonical_key rỗng không hợp lệ")
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, platform_id, platform_item_id, shop_id, title,
                category_id, canonical_key, first_seen
         FROM tracked_product WHERE canonical_key = $1`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

// SetCanonicalKey ghi canonical_key ngược, idempotent. (TASK-PRICE-005)
func (r *Repo) SetCanonicalKey(ctx context.Context, productID int64, key string) error {
	if key == "" {
		return fmt.Errorf("canonical_key cannot be empty")
	}
	// Đảm bảo idempotent: nếu key không đổi thì không cập nhật gì
	tag, err := r.pool.Exec(ctx,
		`UPDATE tracked_product SET canonical_key = $1 WHERE id = $2 AND (canonical_key IS NULL OR canonical_key != $1)`,
		key, productID)
	if err != nil {
		return err
	}
	_ = tag.RowsAffected() // could check this, but idempotent means it might be 0
	return nil
}
