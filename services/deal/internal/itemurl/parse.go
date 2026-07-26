package itemurl

import (
	"net/url"
	"regexp"
	"strings"
)

// Parsed is the platform registry key extracted from a product URL.
type Parsed struct {
	Platform       string // shopee | lazada | tiktok
	PlatformItemID string
}

var (
	shopeeItemRe = regexp.MustCompile(`(^|[-/])i\.(\d+)\.(\d+)($|/)`)
	lazadaItemRe = regexp.MustCompile(`/products/.*-i(\d+)`)
	tiktokItemRe = regexp.MustCompile(`/product/(\d+)`)
)

// Parse detects the marketplace from the host and extracts platform_item_id.
func Parse(rawURL string) (Parsed, bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return Parsed{}, false
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "shopee.vn" || host == "www.shopee.vn":
		if u.Scheme != "https" || u.User != nil || u.Port() != "" {
			return Parsed{}, false
		}
		m := shopeeItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return Parsed{}, false
		}
		return Parsed{Platform: "shopee", PlatformItemID: m[3] + ":" + m[2]}, true
	case strings.Contains(host, "lazada."):
		m := lazadaItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return Parsed{}, false
		}
		return Parsed{Platform: "lazada", PlatformItemID: m[1]}, true
	case strings.Contains(host, "tiktok.com"):
		m := tiktokItemRe.FindStringSubmatch(u.Path)
		if m == nil {
			return Parsed{}, false
		}
		return Parsed{Platform: "tiktok", PlatformItemID: m[1]}, true
	default:
		return Parsed{}, false
	}
}
