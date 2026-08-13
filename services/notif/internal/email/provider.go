package email

import (
	"context"
	"log/slog"
	"time"
)

type EmailMessage struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultPermanent
	ResultFailed
)

type Provider interface {
	Send(ctx context.Context, msg EmailMessage) (SendOutcome, error)
}

type SendOutcome struct {
	Result            SendResult
	RetryAfter        time.Duration
	ProviderMessageID string
}

// LogProvider is the default CI/dev-safe provider: it records intent and
// reports failure so a missing SMTP/SES credential can never look delivered.
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

func (p LogProvider) Send(_ context.Context, msg EmailMessage) (SendOutcome, error) {
	p.log.Info("email noop provider refused message",
		"provider", p.name,
		"to", msg.To,
		"subject", msg.Subject,
	)
	return SendOutcome{Result: ResultFailed, ProviderMessageID: "noop"}, nil
}

// SESProvider is intentionally only an interface-shaped stub here; wiring the
// AWS SES SDK belongs behind Provider and must not be required by CI.
type SESProvider struct {
	SendFn func(ctx context.Context, msg EmailMessage) (SendOutcome, error)
}

func (s SESProvider) Send(ctx context.Context, msg EmailMessage) (SendOutcome, error) {
	if s.SendFn != nil {
		return s.SendFn(ctx, msg)
	}
	return SendOutcome{Result: ResultFailed}, nil
}
