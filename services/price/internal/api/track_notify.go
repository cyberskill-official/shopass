package api

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

// HTTPPriceChangeNotifier POSTs a written snapshot to tracksvc.
type HTTPPriceChangeNotifier struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

func NewHTTPPriceChangeNotifier(baseURL, serviceToken string) (*HTTPPriceChangeNotifier, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid TRACK_INTERNAL_URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("invalid TRACK_INTERNAL_URL scheme")
	}
	return &HTTPPriceChangeNotifier{
		baseURL:      baseURL,
		serviceToken: serviceToken,
		http:         &http.Client{Timeout: 2 * time.Second},
	}, nil
}

type priceChangedBody struct {
	ProductID int64  `json:"product_id"`
	Price     int64  `json:"price"`
	ListPrice *int64 `json:"list_price,omitempty"`
}

func (n *HTTPPriceChangeNotifier) NotifyWritten(ctx context.Context, productID, price int64, listPrice *int64) error {
	body, err := json.Marshal(priceChangedBody{
		ProductID: productID,
		Price:     price,
		ListPrice: listPrice,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+"/internal/v1/price-changed", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", n.serviceToken)
	resp, err := n.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4<<10))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("track price-changed: unexpected status %d", resp.StatusCode)
	}
	return nil
}
