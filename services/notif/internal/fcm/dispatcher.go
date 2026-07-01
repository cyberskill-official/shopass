package fcm

import (
	"context"
	"encoding/json"
	"time"
)

// PushJob represents a claimed notification row ready for FCM dispatch.
type PushJob struct {
	NotifID int64
	UserID  int64
	Token   string
	Payload json.RawMessage
}

// NotifRepo is the interface dispatcher needs from the notif repo (§3).
type NotifRepo interface {
	ClaimPushBatch(ctx context.Context, n int) ([]PushJob, error)
	MarkSent(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	InvalidateToken(ctx context.Context, userID int64) error
}

// Dispatcher consumes queued push notifications and sends them via FCM.
// It is a downstream consumer of the fan-out pipeline (FR-NOTIF-003).
type Dispatcher struct {
	client     *Client
	repo       NotifRepo
	quota      *Bucket
	batchSize  int
	maxRetries int
}

// NewDispatcher creates a push dispatcher.
func NewDispatcher(client *Client, repo NotifRepo, quota *Bucket, batchSize, maxRetries int) *Dispatcher {
	return &Dispatcher{
		client:     client,
		repo:       repo,
		quota:      quota,
		batchSize:  batchSize,
		maxRetries: maxRetries,
	}
}

// RunOnce claims a batch and processes each job.
// In production, this would be called in a loop by each worker.
func (d *Dispatcher) RunOnce(ctx context.Context) error {
	jobs, err := d.repo.ClaimPushBatch(ctx, d.batchSize)
	if err != nil {
		return err
	}

	for _, job := range jobs {
		d.processJob(ctx, job)
	}
	return nil
}

func (d *Dispatcher) processJob(ctx context.Context, job PushJob) {
	// Check quota before sending (DEC-NOTIF-11)
	if !d.quota.Allow() {
		// Throttled — leave queued for next cycle
		return
	}

	msg := Message{
		Token: job.Token,
	}
	// Parse notification from payload if present
	if job.Payload != nil {
		var p struct {
			Title    string            `json:"title"`
			Body     string            `json:"body"`
			Data     map[string]string `json:"data"`
			Deeplink string            `json:"deeplink"`
		}
		if json.Unmarshal(job.Payload, &p) == nil {
			msg.Notification = &MsgNotification{Title: p.Title, Body: p.Body}
			msg.Data = p.Data
			if p.Deeplink != "" {
				msg.Webpush = &WebpushConfig{FCMOptions: &FCMOptions{Link: p.Deeplink}}
			}
		}
	}

	out, err := d.client.Send(ctx, msg)
	if err != nil {
		// Network/OAuth error — leave queued for retry
		return
	}

	switch out.Result {
	case ResultSent:
		d.repo.MarkSent(ctx, job.NotifID)
	case ResultTokenDead:
		d.repo.InvalidateToken(ctx, job.UserID)
		d.repo.MarkFailed(ctx, job.NotifID)
	case ResultRetry:
		// Leave status='queued' — will be retried next cycle
		// In a real system, we'd track attempt count and delay
		if out.RetryAfter > 0 {
			time.Sleep(out.RetryAfter) // simplified; real impl would use scheduler
		}
	case ResultFailed:
		d.repo.MarkFailed(ctx, job.NotifID)
	}
}
