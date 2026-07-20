// Package priceclient is the private HTTP boundary between tracksvc and
// pricesvc. tracksvc never writes tracked_product directly.
package priceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	productUpsertPath  = "/internal/v1/products/upsert"
	snapshotIngestPath = "/v1/price/snapshots"
)

// TrackedProduct is the minimal registry representation tracksvc needs from
// pricesvc. It deliberately mirrors the private HTTP contract rather than the
// price service's internal Go type.
type TrackedProduct struct {
	ID             int64
	PlatformID     int16
	PlatformItemID string
	ShopID         *string
}

// Client performs authenticated private calls to pricesvc.
type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

// New validates the internal endpoint configuration up front, so a beta
// deployment cannot accidentally start with a blank service credential.
func New(baseURL, serviceToken string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.TrimSpace(serviceToken) == "" {
		return nil, fmt.Errorf("PRICE_INTERNAL_SERVICE_TOKEN is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid PRICE_INTERNAL_URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid PRICE_INTERNAL_URL scheme")
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		http:         &http.Client{Timeout: timeout},
	}, nil
}

type productUpsertRequest struct {
	PlatformID     int16   `json:"platform_id"`
	PlatformItemID string  `json:"platform_item_id"`
	ShopID         *string `json:"shop_id,omitempty"`
}

type productUpsertResponse struct {
	ID int64 `json:"id"`
}

// PriceSnapshot is deliberately small: a browser-assisted beta user can only
// confirm the price they can currently see. Scraper-only metadata such as
// stock, sold count and flash-sale status is never accepted on this path.
type PriceSnapshot struct {
	ProductID int64 `json:"product_id"`
	Price     int64 `json:"price"`
}

type snapshotIngestResponse struct {
	Written bool `json:"written"`
}

// Upsert creates or retrieves the product registry entry. The price service
// owns that table and enforces idempotency on its natural key.
func (c *Client) Upsert(ctx context.Context, p TrackedProduct) (TrackedProduct, error) {
	body, err := json.Marshal(productUpsertRequest{
		PlatformID:     p.PlatformID,
		PlatformItemID: p.PlatformItemID,
		ShopID:         p.ShopID,
	})
	if err != nil {
		return TrackedProduct{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+productUpsertPath, bytes.NewReader(body))
	if err != nil {
		return TrackedProduct{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.serviceToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return TrackedProduct{}, fmt.Errorf("price product upsert: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return TrackedProduct{}, fmt.Errorf("price product upsert: unexpected status %d", resp.StatusCode)
	}

	var out productUpsertResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&out); err != nil {
		return TrackedProduct{}, fmt.Errorf("price product upsert: decode response: %w", err)
	}
	if out.ID <= 0 {
		return TrackedProduct{}, fmt.Errorf("price product upsert: invalid product id")
	}
	p.ID = out.ID
	return p, nil
}

// RecordBrowserPrice sends an owner-authorized, user-initiated price
// confirmation through pricesvc. The public caller never reaches pricesvc
// directly: this client runs inside tracksvc on the private network.
func (c *Client) RecordBrowserPrice(ctx context.Context, s PriceSnapshot) (bool, error) {
	body, err := json.Marshal(s)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+snapshotIngestPath, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	// This header is ignored by the legacy ingest endpoint today, but retaining
	// it makes the caller an authenticated internal service when ingest is
	// tightened in a later release.
	req.Header.Set("X-Service-Token", c.serviceToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("price snapshot ingest: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
		return false, fmt.Errorf("price snapshot ingest: unexpected status %d", resp.StatusCode)
	}
	var out snapshotIngestResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&out); err != nil {
		return false, fmt.Errorf("price snapshot ingest: decode response: %w", err)
	}
	return out.Written, nil
}
