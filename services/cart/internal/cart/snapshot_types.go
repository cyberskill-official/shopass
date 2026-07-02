package cart

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID             int64   `db:"id"`
	CartSnapshotID int64   `db:"cart_snapshot_id"`
	ProductID      *int64  `db:"product_id"` // NULL nếu chưa track
	PlatformItemID *string `db:"platform_item_id"`
	ShopID         *string `db:"shop_id"`
	Qty            int32   `db:"qty"`
	UnitPrice      int64   `db:"unit_price"` // VND
}

type CartSnapshot struct {
	ID          int64       `db:"id"`
	UserID      int64       `db:"user_id"` // gắn từ JWT, KHÔNG từ payload
	PlatformID  int16       `db:"platform_id"`
	SnapshotRef uuid.UUID   `db:"snapshot_ref"`
	CapturedAt  time.Time   `db:"captured_at"`
	Items       []CartItem
}

// Request DTO from extension
type SnapshotPayloadItem struct {
	ProductID      *int64  `json:"product_id"`
	PlatformItemID *string `json:"platform_item_id"`
	ShopID         *string `json:"shop_id"`
	Qty            int32   `json:"qty"`
	UnitPrice      int64   `json:"unit_price"`
}

type SnapshotPayload struct {
	PlatformID  int16                 `json:"platform_id"`
	SnapshotRef uuid.UUID             `json:"snapshot_ref"`
	Items       []SnapshotPayloadItem `json:"items"`
}
