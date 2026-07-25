package fcm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// --- mock repo ---

type mockNotifRepo struct {
	jobs             []PushJob
	sentIDs          []int64
	failedIDs        []int64
	invalidatedUsers []int64
}

func (m *mockNotifRepo) ClaimPushBatch(ctx context.Context, n int) ([]PushJob, error) {
	if n > len(m.jobs) {
		n = len(m.jobs)
	}
	claimed := m.jobs[:n]
	m.jobs = m.jobs[n:]
	return claimed, nil
}

func (m *mockNotifRepo) MarkSent(ctx context.Context, id int64) error {
	m.sentIDs = append(m.sentIDs, id)
	return nil
}

func (m *mockNotifRepo) MarkFailed(ctx context.Context, id int64) error {
	m.failedIDs = append(m.failedIDs, id)
	return nil
}

func (m *mockNotifRepo) InvalidateToken(ctx context.Context, userID int64) error {
	m.invalidatedUsers = append(m.invalidatedUsers, userID)
	return nil
}

func setupDispatcher(t *testing.T, fcmCode int, fcmBody string, jobs []PushJob) (*Dispatcher, *mockNotifRepo) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(fcmCode)
		io.WriteString(w, fcmBody)
	}))
	t.Cleanup(srv.Close)

	c := newTestClient(t, srv.URL)
	repo := &mockNotifRepo{jobs: jobs}
	q := NewBucketWithCap(1000)
	d := NewDispatcher(c, repo, q, 10, 3)
	return d, repo
}

func TestDispatch_Success_SetsSent(t *testing.T) {
	jobs := []PushJob{
		{NotifID: 100, UserID: 1, Token: "tok1", Payload: json.RawMessage(`{"title":"Hi","body":"Test"}`)},
	}
	d, repo := setupDispatcher(t, 200, `{"name":"n"}`, jobs)

	err := d.RunOnce(context.Background())
	require.NoError(t, err)
	require.Equal(t, []int64{100}, repo.sentIDs)
	require.Empty(t, repo.failedIDs)
	require.Empty(t, repo.invalidatedUsers)
}

func TestDispatch_Unregistered_InvalidatesTokenAndFails(t *testing.T) {
	jobs := []PushJob{
		{NotifID: 200, UserID: 42, Token: "dead-tok"},
	}
	d, repo := setupDispatcher(t, 404, `{"error":{"status":"UNREGISTERED"}}`, jobs)

	d.RunOnce(context.Background())
	require.Equal(t, []int64{42}, repo.invalidatedUsers) // verified=false
	require.Equal(t, []int64{200}, repo.failedIDs)       // status=failed
	require.Empty(t, repo.sentIDs)
}

func TestDispatch_PermanentFailure_MarksFailed(t *testing.T) {
	jobs := []PushJob{
		{NotifID: 300, UserID: 5, Token: "tok"},
	}
	d, repo := setupDispatcher(t, 403, `{"error":{"status":"PERMISSION_DENIED"}}`, jobs)

	d.RunOnce(context.Background())
	require.Equal(t, []int64{300}, repo.failedIDs)
	require.Empty(t, repo.sentIDs)
}

func TestDispatch_MultipleBatch(t *testing.T) {
	jobs := []PushJob{
		{NotifID: 1, UserID: 1, Token: "t1"},
		{NotifID: 2, UserID: 2, Token: "t2"},
		{NotifID: 3, UserID: 3, Token: "t3"},
	}
	d, repo := setupDispatcher(t, 200, `{"name":"n"}`, jobs)

	d.RunOnce(context.Background())
	require.Equal(t, []int64{1, 2, 3}, repo.sentIDs)
}
