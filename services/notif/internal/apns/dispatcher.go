package apns

import (
	"context"
	"encoding/json"
)

type PushJob struct {
	NotifID int64
	UserID  int64
	Token   string
	Payload json.RawMessage
}

type NotifRepo interface {
	ClaimIOSPushBatch(ctx context.Context, n int) ([]PushJob, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	InvalidateToken(ctx context.Context, userID int64) error
}

type Dispatcher struct {
	client    Sender
	repo      NotifRepo
	batchSize int
}

func NewDispatcher(client Sender, repo NotifRepo, batchSize int) *Dispatcher {
	if batchSize <= 0 {
		batchSize = 50
	}
	return &Dispatcher{client: client, repo: repo, batchSize: batchSize}
}

func (d *Dispatcher) RunOnce(ctx context.Context) error {
	jobs, err := d.repo.ClaimIOSPushBatch(ctx, d.batchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		payload, err := EnsurePayload(job.Payload)
		if err != nil {
			_ = d.repo.MarkFailed(ctx, job.NotifID)
			continue
		}
		res, err := d.client.Send(ctx, job.Token, payload)
		if err != nil {
			continue
		}
		switch res {
		case ResultSent:
			_ = d.repo.MarkSent(ctx, job.NotifID)
		case ResultTokenDead:
			_ = d.repo.InvalidateToken(ctx, job.UserID)
			_ = d.repo.MarkFailed(ctx, job.NotifID)
		case ResultFailed:
			_ = d.repo.MarkFailed(ctx, job.NotifID)
		case ResultRetry:
			// leave queued for retry
		}
	}
	return nil
}
