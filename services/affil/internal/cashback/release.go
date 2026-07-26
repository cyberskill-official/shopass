package cashback

import (
	"context"
	"time"
)

// Releaser moves due pending entries to available when TRUST-005 is clear.
type Releaser struct {
	Ledger *Ledger
}

func NewReleaser(l *Ledger) *Releaser {
	return &Releaser{Ledger: l}
}

// ReleaseDue promotes pending entries with available_at <= now that are not trust-blocked.
func (r *Releaser) ReleaseDue(ctx context.Context, now time.Time) (int, error) {
	if r == nil || r.Ledger == nil || r.Ledger.Store == nil {
		return 0, nil
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	due, err := r.Ledger.Store.ListDuePending(ctx, now)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, e := range due {
		if r.Ledger.Hold != nil {
			blocked, err := r.Ledger.Hold.Blocked(ctx, e.ConversionID)
			if err != nil {
				return n, err
			}
			if blocked {
				continue // keep pending for investigation
			}
		}
		if err := r.Ledger.Store.MarkAvailable(ctx, e.ConversionID); err != nil {
			return n, err
		}
		if r.Ledger.Metrics != nil {
			r.Ledger.Metrics.Released(e.UserShare)
		}
		n++
	}
	return n, nil
}
