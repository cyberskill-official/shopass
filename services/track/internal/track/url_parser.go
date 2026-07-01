package track

import (
	"net/url"
	"regexp"
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
	if err != nil || u.Host == "" {
		return ParsedItem{}, false
	}
	switch platform {
	case "shopee":
		m := shopeeItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return ParsedItem{}, false
		}
		return ParsedItem{ShopID: m[1], PlatformItemID: m[2]}, true
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
