package sms

import (
	"context"
	"log/slog"
)

type Message struct {
	To        string
	Body      string
	Brand     string
	HighValue bool
	OTP       bool
}

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultPermanent
	ResultRejected // guard reject
)

type Provider interface {
	Name() string
	Send(ctx context.Context, msg Message) (SendResult, error)
}

// LogProvider is the CI/dev-safe default: log intent and refuse delivery.
type LogProvider struct {
	log  *slog.Logger
	name string
}

func NewLogProvider(log *slog.Logger, name string) LogProvider {
	if log == nil {
		log = slog.Default()
	}
	if name == "" {
		name = "noop"
	}
	return LogProvider{log: log, name: name}
}

func (p LogProvider) Name() string { return p.name }

func (p LogProvider) Send(_ context.Context, msg Message) (SendResult, error) {
	p.log.Info("sms noop provider refused message",
		"provider", p.name,
		"to", msg.To,
		"brand", msg.Brand,
		"high_value", msg.HighValue,
		"otp", msg.OTP,
	)
	return ResultRejected, nil
}
