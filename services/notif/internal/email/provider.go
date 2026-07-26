package email

import "context"

type Message struct {
	To      string
	Subject string
	Body    string
}

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultPermanent
)

type Provider interface {
	Send(ctx context.Context, msg Message) (SendResult, error)
}

// SESProvider is a thin HTTP-shaped stub that can be swapped for AWS SES SDK.
type SESProvider struct {
	SendFn func(ctx context.Context, msg Message) (SendResult, error)
}

func (s SESProvider) Send(ctx context.Context, msg Message) (SendResult, error) {
	if s.SendFn != nil {
		return s.SendFn(ctx, msg)
	}
	return ResultSent, nil
}
