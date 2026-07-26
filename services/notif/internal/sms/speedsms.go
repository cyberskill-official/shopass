package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// SpeedSMS is the primary VN SMS provider (brandname).
type SpeedSMS struct {
	token  string
	base   string
	http   *http.Client
	SendFn func(ctx context.Context, msg Message) (SendResult, error) // tests
}

func NewSpeedSMSFromEnv() *SpeedSMS {
	token := os.Getenv("SPEEDSMS_TOKEN")
	if token == "" {
		return nil
	}
	base := os.Getenv("SPEEDSMS_BASE_URL")
	if base == "" {
		base = "https://api.speedsms.vn/index.php"
	}
	return &SpeedSMS{
		token: token,
		base:  base,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (s SpeedSMS) Name() string { return "speedsms" }

func (s SpeedSMS) Send(ctx context.Context, msg Message) (SendResult, error) {
	if s.SendFn != nil {
		return s.SendFn(ctx, msg)
	}
	if s.token == "" {
		return ResultPermanent, fmt.Errorf("speedsms: token unset")
	}
	body, _ := json.Marshal(map[string]any{
		"to":       []string{msg.To},
		"content":  msg.Body,
		"sms_type": 2, // brandname
		"sender":   msg.Brand,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.base+"/sms/send", bytes.NewReader(body))
	if err != nil {
		return ResultPermanent, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.token, "")
	client := s.http
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
	case resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403:
		return ResultPermanent, nil
	default:
		return ResultPermanent, nil
	}
}
