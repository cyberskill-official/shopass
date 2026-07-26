package sms

import "context"

type Message struct {
	To      string
	Body    string
	Brand   string
	HighValue bool
	OTP     bool
}

type SendResult int

const (
	ResultSent SendResult = iota
	ResultRetry
	ResultPermanent
	ResultRejected // guard reject
)

type Provider interface {
	Send(ctx context.Context, msg Message) (SendResult, error)
}

type SpeedSMS struct {
	SendFn func(ctx context.Context, msg Message) (SendResult, error)
}

func (s SpeedSMS) Send(ctx context.Context, msg Message) (SendResult, error) {
	if s.SendFn != nil {
		return s.SendFn(ctx, msg)
	}
	return ResultSent, nil
}

type Twilio struct {
	SendFn func(ctx context.Context, msg Message) (SendResult, error)
}

func (t Twilio) Send(ctx context.Context, msg Message) (SendResult, error) {
	if t.SendFn != nil {
		return t.SendFn(ctx, msg)
	}
	return ResultSent, nil
}

// Guard rejects non–high-value / non-OTP SMS (cost model).
func Guard(msg Message) error {
	if msg.HighValue || msg.OTP {
		return nil
	}
	return errNotHighValue
}

type guardError string

func (e guardError) Error() string { return string(e) }

const errNotHighValue = guardError("sms: only high_value or otp allowed")
