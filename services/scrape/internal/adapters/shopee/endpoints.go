package shopee

import (
	"fmt"
)

const (
	pdpPath       = "/api/v4/pdp/get_pc"
	recommendPath = "/api/v4/recommend/recommend"
)

// pdpURL dựng URL lấy giá chính cho một item Shopee.
func pdpURL(base string, itemID, shopID int64) string {
	return fmt.Sprintf("%s%s?item_id=%d&shop_id=%d&detail_level=0",
		base, pdpPath, itemID, shopID)
}
