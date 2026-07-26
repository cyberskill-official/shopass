package email

import (
	"context"
	"encoding/json"

	"shopass/services/notif/internal/notif"
)

// RepoAdapter adapts notif.Repo's shared queued-job shape to email.Job.
type RepoAdapter struct {
	Repo *notif.Repo
}

func (a RepoAdapter) ClaimEmailBatch(ctx context.Context, n int) ([]Job, error) {
	jobs, err := a.Repo.ClaimEmailBatch(ctx, n)
	if err != nil {
		return nil, err
	}
	out := make([]Job, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, Job{
			NotifID: j.NotifID,
			UserID:  j.UserID,
			Address: j.Token,
			Payload: json.RawMessage(j.Payload),
		})
	}
	return out, nil
}

func (a RepoAdapter) MarkSent(ctx context.Context, id int64) error {
	return a.Repo.MarkSent(ctx, id)
}

func (a RepoAdapter) MarkFailed(ctx context.Context, id int64) error {
	return a.Repo.MarkFailed(ctx, id)
}

func (a RepoAdapter) InvalidateEmail(ctx context.Context, userID int64) error {
	return a.Repo.InvalidateEmail(ctx, userID)
}
