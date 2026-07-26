package email

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	jobs []Job
	sent []int64
	fail []int64
	inv  []int64
}

func (m *mockRepo) ClaimEmailBatch(ctx context.Context, n int) ([]Job, error) {
	if n > len(m.jobs) {
		n = len(m.jobs)
	}
	out := m.jobs[:n]
	m.jobs = m.jobs[n:]
	return out, nil
}
func (m *mockRepo) MarkSent(ctx context.Context, id int64) error {
	m.sent = append(m.sent, id)
	return nil
}
func (m *mockRepo) MarkFailed(ctx context.Context, id int64) error {
	m.fail = append(m.fail, id)
	return nil
}
func (m *mockRepo) InvalidateEmail(ctx context.Context, userID int64) error {
	m.inv = append(m.inv, userID)
	return nil
}

func TestEmailDispatch_Success(t *testing.T) {
	repo := &mockRepo{jobs: []Job{{NotifID: 1, UserID: 2, Address: "a@b.c", Payload: json.RawMessage(`{"title":"t","body":"b"}`)}}}
	d := NewDispatcher(SESProvider{}, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{1}, repo.sent)
}

func TestHandleBounce_HardOnly(t *testing.T) {
	repo := &mockRepo{}
	require.NoError(t, HandleBounce(context.Background(), repo, 9, false))
	require.Empty(t, repo.inv)
	require.NoError(t, HandleBounce(context.Background(), repo, 9, true))
	require.Equal(t, []int64{9}, repo.inv)
}
