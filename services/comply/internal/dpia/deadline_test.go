package dpia

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDeadline_FilingOverdueAfter60d(t *testing.T) {
	now := time.Now()
	a := ProcessingActivity{StartedAt: now.Add(-61 * 24 * time.Hour)}
	d := DPIA{FiledAt: nil}
	require.Equal(t, "overdue", Status(a, d, now))
}

func TestDeadline_DraftWithin60d(t *testing.T) {
	now := time.Now()
	a := ProcessingActivity{StartedAt: now.Add(-10 * 24 * time.Hour)}
	require.Equal(t, "draft", Status(a, DPIA{FiledAt: nil}, now))
}

func TestDeadline_ReviewOverdueAfter6m(t *testing.T) {
	now := time.Now()
	filed := now.Add(-200 * 24 * time.Hour) // > 6 thang
	a := ProcessingActivity{StartedAt: now.Add(-220 * 24 * time.Hour)}
	d := DPIA{FiledAt: &filed}
	require.Equal(t, "review_overdue", Status(a, d, now))
}

func TestDeadline_Submitted(t *testing.T) {
	now := time.Now()
	filed := now.Add(-10 * 24 * time.Hour)
	a := ProcessingActivity{StartedAt: now.Add(-20 * 24 * time.Hour)}
	d := DPIA{FiledAt: &filed}
	require.Equal(t, "submitted", Status(a, d, now))
}
