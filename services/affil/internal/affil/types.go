package affil

import "time"

type AffiliateClick struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	PlatformID int16     `db:"platform_id"`
	ProductID  *int64    `db:"product_id"`
	SubID      string    `db:"sub_id"`
	Network    string    `db:"network"`
	ClickedAt  time.Time `db:"clicked_at"`
}

type AffiliateConversion struct {
	ID          int64      `db:"id"`
	ClickID     int64      `db:"click_id"`
	OrderValue  int64      `db:"order_value"` // VND
	Commission  int64      `db:"commission"`  // VND
	Status      string     `db:"status"`      // pending|confirmed|rejected
	ConfirmedAt *time.Time `db:"confirmed_at"`
}
