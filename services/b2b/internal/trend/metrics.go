package trend

import (
	"context"
	"sync"
	"time"
)

// Metrics records SHOULD OTel counters/histograms for the trend job.
// When no OTel exporter is configured the global meter is a no-op; Snapshot
// still exposes in-process totals for unit tests.
type Metrics struct {
	mu              sync.Mutex
	PublishedTotal  int64
	SuppressedTotal int64
	LastDurationMs  float64
	JobRuns         int64
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordJob(_ context.Context, published, suppressed int64, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedTotal += published
	m.SuppressedTotal += suppressed
	m.LastDurationMs = float64(d.Milliseconds())
	m.JobRuns++
}

func (m *Metrics) Snapshot() (published, suppressed, runs int64, durationMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.PublishedTotal, m.SuppressedTotal, m.JobRuns, m.LastDurationMs
}
