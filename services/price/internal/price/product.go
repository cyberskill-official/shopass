package price

import (
	"time"
)

type TrackedProduct struct {
	ID             int64     `db:"id"`
	PlatformID     int16     `db:"platform_id"`
	PlatformItemID string    `db:"platform_item_id"`
	ShopID         *string   `db:"shop_id"`
	Title          *string   `db:"title"`
	CategoryID     *int64    `db:"category_id"`
	CanonicalKey   *string   `db:"canonical_key"` // nil khi chưa so khớp (FR-PRICE-005 điền)
	FirstSeen      time.Time `db:"first_seen"`
}
