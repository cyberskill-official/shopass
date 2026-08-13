package email

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	outcome SendOutcome
	err     error
	sent    []EmailMessage
}

func (f *fakeProvider) Send(ctx context.Context, msg EmailMessage) (SendOutcome, error) {
	f.sent = append(f.sent, msg)
	return f.outcome, f.err
}

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
	provider := &fakeProvider{outcome: SendOutcome{Result: ResultSent}}
	d := NewDispatcher(provider, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{1}, repo.sent)
	require.Len(t, provider.sent, 1)
	require.Equal(t, "a@b.c", provider.sent[0].To)
	require.Equal(t, "t", provider.sent[0].Subject)
	require.Equal(t, "b", provider.sent[0].TextBody)
}

func TestEmailDispatch_RetryLeavesQueued(t *testing.T) {
	repo := &mockRepo{jobs: []Job{{NotifID: 1, UserID: 2, Address: "a@b.c", Payload: json.RawMessage(`{"title":"t","body":"b"}`)}}}
	provider := &fakeProvider{outcome: SendOutcome{Result: ResultRetry}}
	d := NewDispatcher(provider, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Empty(t, repo.sent)
	require.Empty(t, repo.fail)
	require.Empty(t, repo.inv)
}

func TestEmailDispatch_DeadAddressInvalidatesAndFails(t *testing.T) {
	repo := &mockRepo{jobs: []Job{{NotifID: 1, UserID: 2, Address: "a@b.c", Payload: json.RawMessage(`{"title":"t","body":"b"}`)}}}
	provider := &fakeProvider{outcome: SendOutcome{Result: ResultPermanent}}
	d := NewDispatcher(provider, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{2}, repo.inv)
	require.Equal(t, []int64{1}, repo.fail)
}

func TestLogProvider_DefaultNoopFailsClosed(t *testing.T) {
	out, err := NewLogProvider(nil, "").Send(context.Background(), EmailMessage{
		To:       "user@example.com",
		Subject:  "subject",
		TextBody: "body",
	})
	require.NoError(t, err)
	require.Equal(t, ResultFailed, out.Result)
	require.Equal(t, "noop", out.ProviderMessageID)
}

func TestHandleBounce_HardOnly(t *testing.T) {
	repo := &mockRepo{}
	require.NoError(t, HandleBounce(context.Background(), repo, 9, false))
	require.Empty(t, repo.inv)
	require.NoError(t, HandleBounce(context.Background(), repo, 9, true))
	require.Equal(t, []int64{9}, repo.inv)
}
