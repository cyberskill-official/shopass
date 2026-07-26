package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Twilio is the OTP/high-value fallback provider.
type Twilio struct {
	sid    string
	token  string
	from   string
	http   *http.Client
	SendFn func(ctx context.Context, msg Message) (SendResult, error)
}

func NewTwilioFromEnv() *Twilio {
	sid := os.Getenv("TWILIO_ACCOUNT_SID")
	token := os.Getenv("TWILIO_AUTH_TOKEN")
	from := os.Getenv("TWILIO_FROM")
	if sid == "" || token == "" || from == "" {
		return nil
	}
	return &Twilio{
		sid:   sid,
		token: token,
		from:  from,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (t Twilio) Name() string { return "twilio" }

func (t Twilio) Send(ctx context.Context, msg Message) (SendResult, error) {
	if t.SendFn != nil {
		return t.SendFn(ctx, msg)
	}
	if t.sid == "" || t.token == "" {
		return ResultPermanent, fmt.Errorf("twilio: credentials unset")
	}
	form := url.Values{}
	form.Set("To", msg.To)
	form.Set("From", t.from)
	form.Set("Body", msg.Body)
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.sid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return ResultPermanent, err
	}
	req.SetBasicAuth(t.sid, t.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := t.http
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return ResultRetry, err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return ResultSent, nil
	case resp.StatusCode == 429 || resp.StatusCode >= 500:
		return ResultRetry, nil
	default:
		return ResultPermanent, nil
	}
}
