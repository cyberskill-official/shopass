package fanout

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"shopass/services/notif/internal/notif"
)

// --- mock repo ---

type mockNotifRepo struct {
	mu           sync.Mutex
	notifications map[int64]*notif.Notification
	dlqs         []dlqEntry
}

type dlqEntry struct {
	id     int64
	reason string
}

func (m *mockNotifRepo) ClaimPending(ctx context.Context, id int64, lease time.Duration) (notif.Notification, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	n, ok := m.notifications[id]
	if !ok {
		return notif.Notification{}, false, nil
	}

	if n.Status == "pending" || (n.Status == "queued" && n.LeaseUntil != nil && n.LeaseUntil.Before(time.Now())) {
		n.Status = "queued"
		n.Attempts++
		t := time.Now().Add(lease)
		n.LeaseUntil = &t
		return *n, true, nil
	}
	return notif.Notification{}, false, nil
}

func (m *mockNotifRepo) MarkSent(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.notifications[id]
	if n.Status == "queued" {
		n.Status = "sent"
		t := time.Now()
		n.SentAt = &t
	}
	return nil
}

func (m *mockNotifRepo) Requeue(ctx context.Context, id int64, backoff time.Duration, lastErr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := m.notifications[id]
	if n.Status == "queued" {
		t := time.Now().Add(backoff)
		n.LeaseUntil = &t
		n.LastError = lastErr
	}
	return nil
}

func (m *mockNotifRepo) PublishDLQ(ctx context.Context, n notif.Notification, reason, msg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry := m.notifications[n.ID]
	if entry.Status == "queued" {
		entry.Status = "failed"
		entry.LastError = msg
		m.dlqs = append(m.dlqs, dlqEntry{id: n.ID, reason: reason})
	}
	return nil
}

func newTestWorker(t *testing.T) (*Worker, *mockNotifRepo) {
	t.Helper()
	repo := &mockNotifRepo{
		notifications: make(map[int64]*notif.Notification),
	}
	w := NewWorker(repo, NewRouter(), time.Second, time.Second, 5*time.Minute, 3)
	return w, repo
}

func seedPending(t *testing.T, repo *mockNotifRepo, channel string) int64 {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	id := int64(len(repo.notifications) + 1)
	repo.notifications[id] = &notif.Notification{
		ID:      id,
		Channel: channel,
		Status:  "pending",
	}
	return id
}

func statusOf(t *testing.T, repo *mockNotifRepo, id int64) string {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	return repo.notifications[id].Status
}

func attemptsOf(t *testing.T, repo *mockNotifRepo, id int64) int {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	return repo.notifications[id].Attempts
}

func dlqCount(t *testing.T, repo *mockNotifRepo, id int64) int {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	count := 0
	for _, entry := range repo.dlqs {
		if entry.id == id {
			count++
		}
	}
	return count
}

func dlqReason(t *testing.T, repo *mockNotifRepo, id int64) string {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	for _, entry := range repo.dlqs {
		if entry.id == id {
			return entry.reason
		}
	}
	return ""
}

func expireLease(t *testing.T, repo *mockNotifRepo, id int64) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	past := time.Now().Add(-time.Hour)
	repo.notifications[id].LeaseUntil = &past
}

type funcDispatcher struct {
	channel string
	f       func() (ErrClass, error)
}

func (f *funcDispatcher) Dispatch(ctx context.Context, n notif.Notification) (ErrClass, error) {
	return f.f()
}

func (f *funcDispatcher) Channel() string { return f.channel }

func spyDispatcher(channel string, f func() (ErrClass, error)) *funcDispatcher {
	return &funcDispatcher{channel: channel, f: f}
}

func constDispatcher(channel string, class ErrClass, err error) *funcDispatcher {
	return &funcDispatcher{channel: channel, f: func() (ErrClass, error) { return class, err }}
}

func routeAll(d ChannelDispatcher) *Router {
	return NewRouter(d)
}

// --- tests ---

func TestFanout_AtLeastOnce_NoDoubleSend(t *testing.T) {
	w, repo := newTestWorker(t)
	id := seedPending(t, repo, "push")
	var sent int32
	w.router = routeAll(spyDispatcher("push", func() (ErrClass, error) {
		atomic.AddInt32(&sent, 1)
		return ClassOK, nil
	}))

	ctx := context.Background()
	require.NoError(t, w.Handle(ctx, id)) // lần giao 1
	require.NoError(t, w.Handle(ctx, id)) // lần giao 2 (giao lại) -> thua CAS, bỏ qua

	require.Equal(t, int32(1), atomic.LoadInt32(&sent)) // gửi đúng 1 lần
	require.Equal(t, "sent", statusOf(t, repo, id))
}

func TestFanout_ConcurrentClaim_SingleSend(t *testing.T) {
	w, repo := newTestWorker(t)
	id := seedPending(t, repo, "email")
	var sent int32
	w.router = routeAll(spyDispatcher("email", func() (ErrClass, error) {
		atomic.AddInt32(&sent, 1)
		return ClassOK, nil
	}))
	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _ = w.Handle(ctx, id) }()
	}
	wg.Wait()
	require.Equal(t, int32(1), atomic.LoadInt32(&sent)) // chỉ một worker thắng CAS
}

func TestFanout_TransientRetries(t *testing.T) {
	w, repo := newTestWorker(t)
	id := seedPending(t, repo, "push")
	w.router = routeAll(constDispatcher("push", ClassTransient, errors.New("429")))
	
	require.NoError(t, w.Handle(context.Background(), id))
	require.Equal(t, "queued", statusOf(t, repo, id))
	require.GreaterOrEqual(t, attemptsOf(t, repo, id), 1)
	require.Equal(t, 0, dlqCount(t, repo, id)) // chưa cạn retry
}

func TestFanout_MaxAttempts_ToDLQ(t *testing.T) {
	w, repo := newTestWorker(t)
	w.maxAttempts = 3
	id := seedPending(t, repo, "sms")
	w.router = routeAll(constDispatcher("sms", ClassTransient, errors.New("timeout")))
	
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		expireLease(t, repo, id) // giả lập tới hạn re-claim giữa các lần
		_ = w.Handle(ctx, id)
	}
	require.Equal(t, "failed", statusOf(t, repo, id))
	require.Equal(t, "max_attempts", dlqReason(t, repo, id))
}

func TestFanout_PermanentStraightToDLQ(t *testing.T) {
	w, repo := newTestWorker(t)
	id := seedPending(t, repo, "push")
	w.router = routeAll(constDispatcher("push", ClassPermanent, errors.New("token gỡ app")))
	
	require.NoError(t, w.Handle(context.Background(), id))
	require.Equal(t, "failed", statusOf(t, repo, id))
	require.Equal(t, "permanent", dlqReason(t, repo, id))
}
