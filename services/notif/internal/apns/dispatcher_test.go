package apns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type memRepo struct {
	jobs   []PushJob
	sent   []int64
	failed []int64
	dead   []int64
}

func (m *memRepo) ClaimIOSPushBatch(ctx context.Context, n int) ([]PushJob, error) {
	if len(m.jobs) == 0 {
		return nil, nil
	}
	if n > len(m.jobs) {
		n = len(m.jobs)
	}
	out := m.jobs[:n]
	m.jobs = m.jobs[n:]
	return out, nil
}
func (m *memRepo) MarkSent(ctx context.Context, id int64) error {
	m.sent = append(m.sent, id)
	return nil
}
func (m *memRepo) MarkFailed(ctx context.Context, id int64) error {
	m.failed = append(m.failed, id)
	return nil
}
func (m *memRepo) InvalidateToken(ctx context.Context, userID int64) error {
	m.dead = append(m.dead, userID)
	return nil
}

func TestDispatcher_200_MarksSent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		io.WriteString(w, "{}")
	}))
	defer srv.Close()
	client := NewClient("host", "topic", staticTok{}, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = srv.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	}))
	repo := &memRepo{jobs: []PushJob{{NotifID: 1, UserID: 9, Token: "tok", Payload: json.RawMessage(`{"aps":{"alert":"hi"}}`)}}}
	d := NewDispatcher(client, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{1}, repo.sent)
}

func TestDispatcher_410_InvalidatesToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(410)
		io.WriteString(w, `{"reason":"Unregistered"}`)
	}))
	defer srv.Close()
	client := NewClient("host", "topic", staticTok{}, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = srv.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	}))
	repo := &memRepo{jobs: []PushJob{{NotifID: 2, UserID: 44, Token: "dead", Payload: json.RawMessage(`{"aps":{"alert":"x"}}`)}}}
	d := NewDispatcher(client, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{44}, repo.dead)
	require.Equal(t, []int64{2}, repo.failed)
}

func TestDispatcher_500_LeavesQueued(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()
	client := NewClient("host", "topic", staticTok{}, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = srv.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	}))
	repo := &memRepo{jobs: []PushJob{{NotifID: 3, UserID: 1, Token: "t", Payload: json.RawMessage(`{"aps":{"alert":"x"}}`)}}}
	d := NewDispatcher(client, repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Empty(t, repo.sent)
	require.Empty(t, repo.failed)
}

func TestNoopClient_MarksSent(t *testing.T) {
	repo := &memRepo{jobs: []PushJob{{NotifID: 7, UserID: 1, Token: "t"}}}
	d := NewDispatcher(NewNoopClient(nil), repo, 10)
	require.NoError(t, d.RunOnce(context.Background()))
	require.Equal(t, []int64{7}, repo.sent)
}

func TestBuildAlertPayload_EnforcesSize(t *testing.T) {
	_, err := BuildAlertPayload("t", "b")
	require.NoError(t, err)
	huge := make([]byte, 5000)
	for i := range huge {
		huge[i] = 'a'
	}
	_, err = BuildAlertPayload("t", string(huge))
	require.Error(t, err)
}
