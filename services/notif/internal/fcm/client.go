package fcm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SendResult classifies FCM response for dispatcher action (§1 #6).
type SendResult int

const (
	ResultSent      SendResult = iota // 200 OK, has name
	ResultRetry                       // 429/500/503/timeout → backoff
	ResultTokenDead                   // UNREGISTERED / INVALID_ARGUMENT token
	ResultFailed                      // other permanent error
)

// SendOutcome carries the result classification and optional Retry-After.
type SendOutcome struct {
	Result     SendResult
	RetryAfter time.Duration // > 0 when FCM returns Retry-After header (§1 #5)
}

// OAuthTokenSource provides OAuth2 access tokens for FCM HTTP v1.
type OAuthTokenSource interface {
	Token(ctx context.Context) (string, error)
}

// HTTPDoer abstracts the HTTP client for testing.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is the FCM HTTP v1 API client (DEC-NOTIF-10).
type Client struct {
	projectID string
	baseURL   string // default: "https://fcm.googleapis.com"
	oauth     OAuthTokenSource
	http      HTTPDoer
}

// NewClient creates a new FCM client.
func NewClient(projectID string, oauth OAuthTokenSource, httpClient HTTPDoer) *Client {
	return &Client{
		projectID: projectID,
		baseURL:   "https://fcm.googleapis.com",
		oauth:     oauth,
		http:      httpClient,
	}
}

// Send sends a message via FCM HTTP v1 (DEC-NOTIF-10).
func (c *Client) Send(ctx context.Context, msg Message) (SendOutcome, error) {
	body, err := json.Marshal(map[string]any{"message": msg})
	if err != nil {
		return SendOutcome{Result: ResultFailed}, err
	}

	tok, err := c.oauth.Token(ctx)
	if err != nil {
		return SendOutcome{Result: ResultRetry}, fmt.Errorf("fcm: oauth token: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/messages:send", c.baseURL, c.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return SendOutcome{Result: ResultFailed}, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return SendOutcome{Result: ResultRetry}, err // timeout/network → retry
	}
	defer resp.Body.Close()
	// Drain body to allow connection reuse
	io.Copy(io.Discard, resp.Body)

	return classify(resp), nil
}

// classify maps HTTP status + FCM error to SendOutcome (§1 #6).
func classify(resp *http.Response) SendOutcome {
	switch resp.StatusCode {
	case 200:
		return SendOutcome{Result: ResultSent}
	case 429:
		return SendOutcome{Result: ResultRetry, RetryAfter: parseRetryAfter(resp.Header)}
	case 500, 503:
		return SendOutcome{Result: ResultRetry, RetryAfter: parseRetryAfter(resp.Header)}
	case 404:
		return SendOutcome{Result: ResultTokenDead} // UNREGISTERED
	case 400:
		return SendOutcome{Result: ResultTokenDead} // INVALID_ARGUMENT (bad token format)
	default:
		return SendOutcome{Result: ResultFailed}
	}
}

// parseRetryAfter reads the Retry-After header (seconds or date).
func parseRetryAfter(h http.Header) time.Duration {
	v := h.Get("Retry-After")
	if v == "" {
		return 0
	}
	v = strings.TrimSpace(v)
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(v); err == nil {
		d := time.Until(t)
		if d > 0 {
			return d
		}
	}
	return 0
}
