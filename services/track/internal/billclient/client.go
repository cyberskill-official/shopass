// Package billclient is the private HTTP boundary between tracksvc and billsvc
// for subscription feature gating.
package billclient

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

const checkPath = "/internal/v1/gating/check"

type Client struct {
	baseURL      string
	serviceToken string
	http         *http.Client
}

func New(baseURL, serviceToken string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.TrimSpace(serviceToken) == "" {
		return nil, fmt.Errorf("BILL_INTERNAL_SERVICE_TOKEN is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid BILL_INTERNAL_URL")
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

type CheckResult struct {
	Allowed      bool `json:"allowed"`
	LimitReached bool `json:"limit_reached"`
}

func (c *Client) Check(ctx context.Context, userID int64, featureKey string, usage *int64) (CheckResult, error) {
	body, err := json.Marshal(map[string]any{
		"user_id":     userID,
		"feature_key": featureKey,
		"usage":       usage,
	})
	if err != nil {
		return CheckResult{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+checkPath, bytes.NewReader(body))
	if err != nil {
		return CheckResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.serviceToken)
	resp, err := c.http.Do(req)
	if err != nil {
		return CheckResult{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return CheckResult{}, fmt.Errorf("billsvc gating status %d", resp.StatusCode)
	}
	var out CheckResult
	if err := json.Unmarshal(raw, &out); err != nil {
		return CheckResult{}, err
	}
	return out, nil
}
