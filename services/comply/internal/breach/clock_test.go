package breach

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClock_OverdueAfter72h(t *testing.T) {
	now := time.Now()
	b := BreachIncident{AcknowledgedAt: now.Add(-73 * time.Hour), NotifiedAuthorityAt: nil}
	require.Equal(t, "breach_overdue", DeadlineFlag(b, now))
}

func TestClock_WithinWindow(t *testing.T) {
	now := time.Now()
	b := BreachIncident{AcknowledgedAt: now.Add(-10 * time.Hour), NotifiedAuthorityAt: nil}
	require.Equal(t, "within_window", DeadlineFlag(b, now))
}

func TestClock_Notified(t *testing.T) {
	now := time.Now()
	no := now.Add(-10 * time.Hour)
	b := BreachIncident{AcknowledgedAt: now.Add(-20 * time.Hour), NotifiedAuthorityAt: &no}
	require.Equal(t, "notified", DeadlineFlag(b, now))
}
