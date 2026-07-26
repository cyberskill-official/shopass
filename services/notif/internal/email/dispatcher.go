package email

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
	ClaimEmailBatch(ctx context.Context, n int) ([]Job, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	InvalidateEmail(ctx context.Context, userID int64) error
}

type Dispatcher struct {
	provider  Provider
	repo      NotifRepo
	batchSize int
}

func NewDispatcher(p Provider, repo NotifRepo, batchSize int) *Dispatcher {
	if batchSize <= 0 {
		batchSize = 50
	}
	return &Dispatcher{provider: p, repo: repo, batchSize: batchSize}
}

func (d *Dispatcher) RunOnce(ctx context.Context) error {
	jobs, err := d.repo.ClaimEmailBatch(ctx, d.batchSize)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		var p struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		_ = json.Unmarshal(job.Payload, &p)
		out, err := d.provider.Send(ctx, EmailMessage{
			To:       job.Address,
			Subject:  p.Title,
			HTMLBody: p.Body,
			TextBody: p.Body,
		})
		if err != nil || out.Result == ResultRetry {
			_ = BackoffDelay(1, out.RetryAfter)
			continue
		}
		if out.Result == ResultPermanent {
			_ = d.repo.InvalidateEmail(ctx, job.UserID)
			_ = d.repo.MarkFailed(ctx, job.NotifID)
			continue
		}
		if out.Result == ResultFailed {
			_ = d.repo.MarkFailed(ctx, job.NotifID)
			continue
		}
		if out.Result == ResultSent {
			_ = d.repo.MarkSent(ctx, job.NotifID)
		}
	}
	return nil
}

// HandleBounce marks hard bounces / complaints as unverified (not soft bounces).
func HandleBounce(ctx context.Context, repo NotifRepo, userID int64, hard bool) error {
	if !hard {
		return nil
	}
	return repo.InvalidateEmail(ctx, userID)
}
