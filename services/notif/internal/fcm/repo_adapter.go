package fcm

import (
	"context"
	"encoding/json"

	"shopass/services/notif/internal/notif"
)

// RepoAdapter adapts notif.Repo to the fcm.NotifRepo interface.
type RepoAdapter struct {
	Repo *notif.Repo
}

func (a RepoAdapter) ClaimPushBatch(ctx context.Context, n int) ([]PushJob, error) {
	jobs, err := a.Repo.ClaimPushBatch(ctx, n)
	if err != nil {
		return nil, err
	}
	out := make([]PushJob, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, PushJob{
			NotifID: j.NotifID,
			UserID:  j.UserID,
			Token:   j.Token,
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

func (a RepoAdapter) InvalidateToken(ctx context.Context, userID int64) error {
	return a.Repo.InvalidateToken(ctx, userID)
}
