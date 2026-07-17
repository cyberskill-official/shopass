package track

import "testing"

func TestShopeeVNPlatformMapFailsClosedForOtherPlatforms(t *testing.T) {
	platforms := NewShopeeVNPlatformMap()
	id, ok := platforms.IDByCode(" Shopee ")
	if !ok || id != ShopeeVNPlatformID {
		t.Fatalf("Shopee map = (%d, %t), want (%d, true)", id, ok, ShopeeVNPlatformID)
	}
	for _, code := range []string{"tiktok", "lazada", "amazon", ""} {
		if _, ok := platforms.IDByCode(code); ok {
			t.Fatalf("platform %q must not be enabled in closed beta", code)
		}
	}
}
