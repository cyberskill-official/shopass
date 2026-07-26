package sms

import (
	"context"
	"encoding/json"
)

type Job struct {
	NotifID int64
	UserID  int64
	Address string
	Payload json.RawMessage
}

type NotifRepo interface {
	ClaimSMSBatch(ctx context.Context, n int) ([]Job, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type Dispatcher struct {
	primary   Provider
	fallback  Provider
	brand     string
	repo      NotifRepo
	batchSize int
}

func NewDispatcher(primary, fallback Provider, brand string, repo NotifRepo, batchSize int) *Dispatcher {
	if batchSize <= 0 {
		batchSize = 20
	}
	return &Dispatcher{primary: primary, fallback: fallback, brand: brand, repo: repo, batchSize: batchSize}
}

func (d *Dispatcher) RunOnce(ctx context.Context) error {
	jobs, err := d.repo.ClaimSMSBatch(ctx, d.batchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		var p struct {
			Body      string `json:"body"`
			HighValue bool   `json:"high_value"`
			OTP       bool   `json:"otp"`
		}
		_ = json.Unmarshal(job.Payload, &p)
		msg := Message{To: job.Address, Body: p.Body, Brand: d.brand, HighValue: p.HighValue, OTP: p.OTP}
		if err := Guard(msg); err != nil {
			_ = d.repo.MarkFailed(ctx, job.NotifID)
			continue
		}
		res, err := d.primary.Send(ctx, msg)
		if (err != nil || res == ResultRetry || res == ResultPermanent) && d.fallback != nil && (msg.HighValue || msg.OTP) {
			res, err = d.fallback.Send(ctx, msg)
		}
		if err != nil || res == ResultRetry {
			continue
		}
		if res == ResultPermanent || res == ResultRejected {
			_ = d.repo.MarkFailed(ctx, job.NotifID)
			continue
		}
		_ = d.repo.MarkSent(ctx, job.NotifID)
	}
	return nil
}
