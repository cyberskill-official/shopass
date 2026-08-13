package apns

import (
	"context"
	"log/slog"
)

// NoopClient records intent and reports ResultFailed without contacting Apple.
type NoopClient struct {
	log *slog.Logger
}

func NewNoopClient(log *slog.Logger) *NoopClient {
	if log == nil {
		log = slog.Default()
	}
	return &NoopClient{log: log}
}

func (c *NoopClient) Send(_ context.Context, deviceToken string, payload []byte) (SendResult, error) {
	c.log.Info("apns noop refused push",
		"token_prefix", truncate(deviceToken, 8),
		"payload_bytes", len(payload),
	)
	return ResultFailed, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// Sender is implemented by *Client and *NoopClient.
type Sender interface {
	Send(ctx context.Context, deviceToken string, payload []byte) (SendResult, error)
}
