package price

import "time"

// PriceSnapshot represents a single price observation for a tracked product.
// price and list_price are BIGINT (VND, không thập phân) per DEC-PRICE-05.
type PriceSnapshot struct {
	ProductID int64     `db:"product_id"`
	TS        time.Time `db:"ts"`
	Price     int64     `db:"price"`      // VND, > 0
	ListPrice *int64    `db:"list_price"` // nil hoặc >= price
	Stock     *int32    `db:"stock"`
	Sold      *int32    `db:"sold"`
	FlashSale bool      `db:"flash_sale"`
}

// DailyBucket is a row from the price_daily continuous aggregate.
type DailyBucket struct {
	ProductID int64     `db:"product_id"`
	Day       time.Time `db:"day"`
	MinP      int64     `db:"min_p"`
	MaxP      int64     `db:"max_p"`
	CloseP    int64     `db:"close_p"`
}
