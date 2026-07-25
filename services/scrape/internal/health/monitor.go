package health

import (
	"sync"
)

type adapterKey struct {
	platformID int16
	version    string
}

type Monitor struct {
	mu        sync.RWMutex
	windows   map[adapterKey]*Window
	states    map[adapterKey]Health
	baselines map[adapterKey]float64
	alerter   *AlertDedup
}

func NewMonitor(alerter *AlertDedup) *Monitor {
	return &Monitor{
		windows:   make(map[adapterKey]*Window),
		states:    make(map[adapterKey]Health),
		baselines: make(map[adapterKey]float64),
		alerter:   alerter,
	}
}

func (m *Monitor) getOrCreateWindow(key adapterKey) *Window {
	m.mu.Lock()
	defer m.mu.Unlock()
	if w, ok := m.windows[key]; ok {
		return w
	}
	w := NewWindow(1000)
	m.windows[key] = w
	m.states[key] = Healthy
	m.baselines[key] = 0.05 // default baseline
	return w
}

func (m *Monitor) Report(platformID int16, version string, outcome Outcome) {
	key := adapterKey{platformID, version}
	w := m.getOrCreateWindow(key)
	w.Record(outcome)

	rate, n := w.ParseFailRate()

	m.mu.Lock()
	curState := m.states[key]
	baseline := m.baselines[key] // Giả sử tính động ở đây hoặc update định kỳ

	newState := Next(curState, rate, baseline, n)
	m.states[key] = newState
	m.mu.Unlock()

	if newState != Healthy && newState != curState {
		m.alerter.Alert(platformID, version, newState, rate, n, baseline)
	}
}

func (m *Monitor) stateOf(platformID int16, version string) Health {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.states[adapterKey{platformID, version}]
}

// ShouldThrottle: khi adapter broken, orchestrator giảm tần suất quét target.
func (m *Monitor) ShouldThrottle(platformID int16, version string) (bool, float64) {
	switch m.stateOf(platformID, version) {
	case Broken:
		return true, 0.1 // còn ~10% tần suất để dò sàn revert
	case Degraded:
		return true, 0.5
	default:
		return false, 1.0
	}
}
