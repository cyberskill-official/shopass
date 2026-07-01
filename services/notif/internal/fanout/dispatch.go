package fanout

import (
	"context"

	"shopass/services/notif/internal/notif"
)

// ErrClass categorizes dispatch errors for the pipeline's DLQ/retry logic (§3).
type ErrClass int

const (
	ClassOK        ErrClass = iota // Success
	ClassTransient                 // Retryable: 429, 5xx, network timeout
	ClassPermanent                 // DLQ immediately: bad token, invalid payload
)

// ChannelDispatcher is the interface for per-channel senders (FR-NOTIF-002, 005, 006, 007).
type ChannelDispatcher interface {
	Dispatch(ctx context.Context, n notif.Notification) (ErrClass, error)
	Channel() string // e.g. "push", "email", "sms"
}

// Router routes notifications to the appropriate dispatcher based on their channel.
type Router struct {
	byChannel map[string]ChannelDispatcher
}

// NewRouter creates a new Router with the provided dispatchers.
func NewRouter(dispatchers ...ChannelDispatcher) *Router {
	m := make(map[string]ChannelDispatcher)
	for _, d := range dispatchers {
		m[d.Channel()] = d
	}
	return &Router{byChannel: m}
}

// Route returns the dispatcher for a channel, if registered.
func (r *Router) Route(channel string) (ChannelDispatcher, bool) {
	d, ok := r.byChannel[channel]
	return d, ok
}
