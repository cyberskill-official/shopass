package report

import "sync"

type Metrics struct {
	mu             sync.Mutex
	ServedByTier   map[string]int64
	DeniedByReason map[string]int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		ServedByTier:   make(map[string]int64),
		DeniedByReason: make(map[string]int64),
	}
}

func (m *Metrics) Served(tier string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.ServedByTier[tier]++
	m.mu.Unlock()
}

func (m *Metrics) Denied(reason string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	m.DeniedByReason[reason]++
	m.mu.Unlock()
}
