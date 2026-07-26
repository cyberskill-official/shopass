package apns

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultTokenDead
	ResultFailed
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// TokenSource provides a short-lived APNs JWT (ES256 .p8).
type TokenSource interface {
	Bearer(ctx context.Context) (string, error)
}

type Client struct {
	host  string // api.push.apple.com or api.sandbox.push.apple.com
	topic string
	oauth TokenSource
	http  HTTPDoer
}

func NewClient(host, topic string, oauth TokenSource, httpClient HTTPDoer) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{host: host, topic: topic, oauth: oauth, http: httpClient}
}

func (c *Client) Send(ctx context.Context, deviceToken string, payload []byte) (SendResult, error) {
	tok, err := c.oauth.Bearer(ctx)
	if err != nil {
		return ResultRetry, err
	}
	url := fmt.Sprintf("https://%s/3/device/%s", c.host, deviceToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return ResultFailed, err
	}
	req.Header.Set("authorization", "bearer "+tok)
	req.Header.Set("apns-topic", c.topic)
	req.Header.Set("apns-push-type", "alert")
	req.Header.Set("apns-priority", "10")
	req.Header.Set("apns-expiration", "0")
	req.Header.Set("content-type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return ResultRetry, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case 200:
		return ResultSent, nil
	case 410:
		return ResultTokenDead, nil
	case 400:
		return ResultTokenDead, nil // BadDeviceToken
	case 429, 500, 503:
		return ResultRetry, nil
	default:
		return ResultFailed, nil
	}
}
