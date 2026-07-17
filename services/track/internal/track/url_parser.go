package track

import (
	"net/url"
	"regexp"
	"strings"
)

// ParsedItem là kết quả bóc từ item_url.
type ParsedItem struct {
	PlatformItemID string
	ShopID         string // có thể rỗng nếu sàn không nhúng shop trong url
}

// shopeeItemRe khớp .../<shop_id>.<item_id> hoặc ...-i.<shop_id>.<item_id>
var shopeeItemRe = regexp.MustCompile(`i\.(\d+)\.(\d+)`)
var lazadaItemRe = regexp.MustCompile(`/products/.*-i(\d+)`)
var tiktokItemRe = regexp.MustCompile(`/product/(\d+)`)

// ParseItemURL bóc (platform_item_id, shop_id) theo từng sàn.
func ParseItemURL(platform, rawURL string) (ParsedItem, bool) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return ParsedItem{}, false
	}
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "shopee":
		// Closed beta only accepts a Shopee Vietnam product link. Matching an
		// ID-looking path on an unrelated host would otherwise create a bogus
		// product that the scraper can never retrieve.
		host := strings.ToLower(u.Hostname())
		if host != "shopee.vn" && host != "www.shopee.vn" {
			return ParsedItem{}, false
		}
		m := shopeeItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return ParsedItem{}, false
		}
		// Shopee item_id is only unique within a shop. The scraper expects the
		// registry key in itemID:shopID form (not the old item-only value).
		return ParsedItem{ShopID: m[1], PlatformItemID: m[2] + ":" + m[1]}, true
	case "lazada":
		m := lazadaItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return ParsedItem{}, false
		}
		return ParsedItem{PlatformItemID: m[1]}, true
	case "tiktok":
		m := tiktokItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return ParsedItem{}, false
		}
		return ParsedItem{PlatformItemID: m[1]}, true
	default:
		return ParsedItem{}, false
	}
}
