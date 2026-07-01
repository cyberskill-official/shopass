package notif

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouting_PicksCheapestAvailable(t *testing.T) {
	c, ok := ResolveChannel([]string{"push", "email"}, UserChannels{Push: true, Email: true}, false)
	require.True(t, ok)
	require.Equal(t, "push", c)
}

func TestRouting_DowngradesWhenPushUnavailable(t *testing.T) {
	c, ok := ResolveChannel([]string{"push", "email"}, UserChannels{Push: false, Email: true}, false)
	require.True(t, ok)
	require.Equal(t, "email", c)
}

func TestRouting_SMSNotChosenWhenPushAvailable(t *testing.T) {
	c, ok := ResolveChannel([]string{"push", "sms"}, UserChannels{Push: true, SMS: true}, false)
	require.True(t, ok)
	require.Equal(t, "push", c)
}

func TestRouting_SMSAllowedForHighValue(t *testing.T) {
	c, ok := ResolveChannel([]string{"sms"}, UserChannels{SMS: true}, true)
	require.True(t, ok)
	require.Equal(t, "sms", c)
}

func TestRouting_NoChannelAvailable(t *testing.T) {
	_, ok := ResolveChannel([]string{"push"}, UserChannels{Push: false}, false)
	require.False(t, ok)
}
