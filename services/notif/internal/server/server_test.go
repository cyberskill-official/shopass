package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"shopass/services/notif/internal/notif"
)

type mockStore struct {
	caps      notif.UserChannels
	capsErr   error
	inserted  []notif.Notification
	queuedIDs []int64
	tokens    []struct {
		userID                     int64
		channel, platform, address string
	}
	nextID int64
}

func (m *mockStore) GetUserChannels(ctx context.Context, userID int64) (notif.UserChannels, error) {
	return m.caps, m.capsErr
}

func (m *mockStore) InsertNotification(ctx context.Context, n notif.Notification) (int64, error) {
	m.nextID++
	m.inserted = append(m.inserted, n)
	return m.nextID, nil
}

func (m *mockStore) MarkQueued(ctx context.Context, id int64) error {
	m.queuedIDs = append(m.queuedIDs, id)
	return nil
}

func (m *mockStore) UpsertToken(ctx context.Context, userID int64, channel, platform, address string) error {
	m.tokens = append(m.tokens, struct {
		userID                     int64
		channel, platform, address string
	}{userID, channel, platform, address})
	return nil
}

func (m *mockStore) DeleteToken(ctx context.Context, userID int64, address string) error {
	return nil
}

func TestNotify_EnqueuesWhenPushTokenPresent(t *testing.T) {
	store := &mockStore{caps: notif.UserChannels{Push: true}}
	h := New(store, nil).Handler()

	body := `{"user_id":42,"product_id":7,"reason":"bottom_predicted","payload":{"p_bottom_14d":0.82}}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusAccepted, rr.Code)
	require.Len(t, store.inserted, 1)
	require.Equal(t, int64(42), store.inserted[0].UserID)
	require.Equal(t, "push", store.inserted[0].Channel)
	require.Equal(t, "bottom_predicted", store.inserted[0].Template)
	require.Equal(t, []int64{1}, store.queuedIDs)
	require.Contains(t, store.inserted[0].Payload["title"], "đáy")
	require.NotEmpty(t, store.inserted[0].Payload["body"])
}

func TestNotify_RejectsWithoutPushChannel(t *testing.T) {
	store := &mockStore{caps: notif.UserChannels{}}
	h := New(store, nil).Handler()

	body := `{"user_id":1,"reason":"bottom_predicted","payload":{"p_bottom_14d":0.9}}`
	req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	require.Empty(t, store.inserted)
}

func TestRegisterDevice_RequiresUserHeader(t *testing.T) {
	h := New(&mockStore{}, nil).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v1/devices", bytes.NewBufferString(`{"fcm_token":"t"}`))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestRegisterDevice_UpsertsWebToken(t *testing.T) {
	store := &mockStore{}
	h := New(store, nil).Handler()
	req := httptest.NewRequest(http.MethodPost, "/v1/devices", bytes.NewBufferString(`{"fcm_token":"abc","platform":"web"}`))
	req.Header.Set("X-User-Id", "9")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Len(t, store.tokens, 1)
	require.Equal(t, int64(9), store.tokens[0].userID)
	require.Equal(t, "push", store.tokens[0].channel)
	require.Equal(t, "web", store.tokens[0].platform)
	require.Equal(t, "abc", store.tokens[0].address)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, true, resp["ok"])
}
