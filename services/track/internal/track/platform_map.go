package track

import "strings"

// ShopeeVNPlatformID is the canonical platform.id seeded for Shopee Vietnam.
// Closed beta deliberately exposes only this marketplace until the other
// adapters and operational controls are ready.
const ShopeeVNPlatformID int16 = 1

// ShopeeVNPlatformMap is a static, fail-closed platform lookup for closed beta.
// It avoids a read dependency on the shared platform table in tracksvc; pricesvc
// remains the owner of the product registry and its FK is the final guard.
type ShopeeVNPlatformMap struct{}

func NewShopeeVNPlatformMap() ShopeeVNPlatformMap {
	return ShopeeVNPlatformMap{}
}

func (ShopeeVNPlatformMap) IDByCode(code string) (int16, bool) {
	if strings.ToLower(strings.TrimSpace(code)) != "shopee" {
		return 0, false
	}
	return ShopeeVNPlatformID, true
}
