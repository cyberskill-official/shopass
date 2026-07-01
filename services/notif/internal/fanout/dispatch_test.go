package fanout

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"shopass/services/notif/internal/notif"
)

type mockDispatcher struct {
	channel string
	class   ErrClass
	err     error
}

func (m *mockDispatcher) Dispatch(ctx context.Context, n notif.Notification) (ErrClass, error) {
	return m.class, m.err
}

func (m *mockDispatcher) Channel() string {
	return m.channel
}

func TestRouter_Route(t *testing.T) {
	pushD := &mockDispatcher{channel: "push"}
	emailD := &mockDispatcher{channel: "email"}

	r := NewRouter(pushD, emailD)

	d, ok := r.Route("push")
	require.True(t, ok)
	require.Equal(t, "push", d.Channel())

	d, ok = r.Route("email")
	require.True(t, ok)
	require.Equal(t, "email", d.Channel())

	_, ok = r.Route("sms")
	require.False(t, ok) // not registered
}
