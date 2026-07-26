package sms

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
}

func (m *mockRepo) ClaimSMSBatch(ctx context.Context, n int) ([]Job, error) {
	if n > len(m.jobs) {
		n = len(m.jobs)
	}
	out := m.jobs[:n]
	m.jobs = m.jobs[n:]
	return out, nil
}
func (m *mockRepo) MarkSent(ctx context.Context, id int64) error  { m.sent = append(m.sent, id); return nil }
func (m *mockRepo) MarkFailed(ctx context.Context, id int64) error { m.fail = append(m.fail, id); return nil }

func TestGuard_RejectsLowValue(t *testing.T) {
	require.Error(t, Guard(Message{Body: "hi"}))
	require.NoError(t, Guard(Message{HighValue: true}))
	require.NoError(t, Guard(Message{OTP: true}))
}

func TestSMSDispatch_FailsGuard(t *testing.T) {
	repo := &mockRepo{jobs: []Job{{NotifID: 1, Address: "+84", Payload: json.RawMessage(`{"body":"x"}`)}}}
	d := NewDispatcher(SpeedSMS{}, nil, "SHOPASS", repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{1}, repo.fail)
}

func TestSMSDispatch_HighValueSends(t *testing.T) {
	repo := &mockRepo{jobs: []Job{{NotifID: 2, Address: "+84", Payload: json.RawMessage(`{"body":"deal","high_value":true}`)}}}
	d := NewDispatcher(SpeedSMS{}, nil, "SHOPASS", repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{2}, repo.sent)
}
