package priceclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"shopass/services/scrape/internal/orchestrator"
)

// Client posts scraped snapshots to the price service ingest endpoint. It
// implements orchestrator.PriceRepo: price owns price_snapshot, so the scraper
// writes through price rather than touching the table directly.
type Client struct {
	base         string
	serviceToken string
	http         *http.Client
}

func New(base, serviceToken string, timeout time.Duration) *Client {
	return &Client{base: base, serviceToken: serviceToken, http: &http.Client{Timeout: timeout}}
}

type ingestReq struct {
	ProductID int64     `json:"product_id"`
	TS        time.Time `json:"ts"`
	Price     int64     `json:"price"`
	ListPrice *int64    `json:"list_price,omitempty"`
	Stock     *int32    `json:"stock,omitempty"`
	Sold      *int32    `json:"sold,omitempty"`
	FlashSale bool      `json:"flash_sale"`
}

type ingestResp struct {
	Written bool `json:"written"`
}

// InsertSnapshot POSTs the snapshot; written=true means price stored a new row
// (delta-only applies on the price side).
func (c *Client) InsertSnapshot(ctx context.Context, s orchestrator.PriceSnapshot) (bool, error) {
	body, err := json.Marshal(ingestReq{
		ProductID: s.ProductID, TS: s.TS, Price: s.Price,
		ListPrice: s.ListPrice, Stock: s.Stock, Sold: s.Sold, FlashSale: s.FlashSale,
	})
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/v1/price/snapshots", bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.serviceToken)
	resp, err := c.http.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("price ingest: unexpected status %d", resp.StatusCode)
	}
	var out ingestResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Written, nil
}
