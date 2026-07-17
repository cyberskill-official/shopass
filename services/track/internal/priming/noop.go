// Package priming contains the explicit closed-beta scrape-priming boundary.
package priming

import (
	"context"
	"errors"
)

// ErrDeferred tells the caller that no synchronous queue integration has been
// configured. pricesvc has already committed the registry product by this
// point, and the scrape feeder will register it on its next pass.
var ErrDeferred = errors.New("scrape priming deferred: no queue integration configured")

// NoopQueue deliberately does not reach into scrape service tables. It makes
// the missing asynchronous integration visible to logs while retaining the
// durable user-track action.
type NoopQueue struct{}

func NewNoopQueue() NoopQueue { return NoopQueue{} }

func (NoopQueue) EnqueuePriming(_ context.Context, _ int64) error {
	return ErrDeferred
}
