package apikey

import (
	"context"
	"sync"
	"time"
)

type UsageEvent struct {
	APIKeyID   int64
	Endpoint   string
	TS         time.Time
	StatusCode int
	// Intentionally no response body / payload fields (DEC-B2B-35).
}

type UsageStore interface {
	Record(ctx context.Context, e UsageEvent) error
	CountMonth(ctx context.Context, apiKeyID int64, monthStart time.Time) (int, error)
}

type MemoryUsage struct {
	mu     sync.Mutex
	events []UsageEvent
}

func NewMemoryUsage() *MemoryUsage {
	return &MemoryUsage{}
}

func (m *MemoryUsage) Record(_ context.Context, e UsageEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
	return nil
}

func (m *MemoryUsage) CountMonth(_ context.Context, apiKeyID int64, monthStart time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	end := monthStart.AddDate(0, 1, 0)
	for _, e := range m.events {
		if e.APIKeyID == apiKeyID && !e.TS.Before(monthStart) && e.TS.Before(end) {
			n++
		}
	}
	return n, nil
}

func (m *MemoryUsage) Events() []UsageEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]UsageEvent, len(m.events))
	copy(out, m.events)
	return out
}
