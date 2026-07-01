package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shopass/services/scrape/internal/orchestrator"
)

const (
	shopeePlatformID int16 = 1 // platform.id for Shopee
)

// Farm is the Playwright fallback interface (FR-SCRAPE-003).
// Stubbed here as an interface so this adapter compiles without the farm package.
type Farm interface {
	RenderPrice(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error)
}

// HTTPClient abstracts the HTTP layer so tests can inject stubs.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ShopeeAdapter implements orchestrator.PlatformAdapter for Shopee.
type ShopeeAdapter struct {
	base string
	http HTTPClient
	farm Farm
}

// NewShopeeAdapter creates an adapter targeting the given Shopee API base URL.
func NewShopeeAdapter(base string, client HTTPClient, farm Farm) *ShopeeAdapter {
	return &ShopeeAdapter{base: base, http: client, farm: farm}
}

// PlatformID returns the Shopee platform identifier.
func (a *ShopeeAdapter) PlatformID() int16 { return shopeePlatformID }

// Fetch lấy giá qua internal JSON endpoint; rơi xuống Playwright farm nếu challenge.
func (a *ShopeeAdapter) Fetch(ctx context.Context, job orchestrator.ScrapeJob) (orchestrator.PriceSnapshot, error) {
	itemID, shopID, err := splitRef(job.PlatformItemID)
	if err != nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("shopee: bad platform_item_id %q: %w", job.PlatformItemID, err)
	}

	url := pdpURL(a.base, itemID, shopID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return orchestrator.PriceSnapshot{}, err
	}
	// §1 #3: KHÔNG đính cookie phiên người dùng — is_login:false
	req.Header.Set("User-Agent", "SanDeal-Scraper/1.0")

	resp, err := a.http.Do(req)
	if err != nil {
		// Network error → fallback farm
		return a.fallback(ctx, job, "network_error")
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !isJSON(ct) {
		// HTML challenge / WAF → fallback Playwright (DEC-SCRAPE-08)
		return a.fallback(ctx, job, "challenge_or_html")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return a.fallback(ctx, job, "read_error")
	}

	var pdp pdpResponse
	if err := json.Unmarshal(body, &pdp); err != nil {
		return a.fallback(ctx, job, "parse_error")
	}

	snap, err := pdp.toSnapshot(job.ProductID, time.Now())
	if err != nil {
		return orchestrator.PriceSnapshot{}, err // ErrItemGone — don't fallback, item is dead
	}
	return snap, nil
}

// fallback gọi Playwright farm (FR-SCRAPE-003). Nếu farm cũng fail → trả error.
func (a *ShopeeAdapter) fallback(ctx context.Context, job orchestrator.ScrapeJob, reason string) (orchestrator.PriceSnapshot, error) {
	if a.farm == nil {
		return orchestrator.PriceSnapshot{}, fmt.Errorf("shopee: %s and no farm configured", reason)
	}
	return a.farm.RenderPrice(ctx, job)
}

// splitRef parses "itemID:shopID" from tracked_product.platform_item_id.
func splitRef(ref string) (itemID, shopID int64, err error) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected itemID:shopID, got %q", ref)
	}
	itemID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	shopID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return itemID, shopID, nil
}

// isJSON checks if content-type header indicates JSON.
func isJSON(ct string) bool {
	return strings.Contains(ct, "application/json")
}
