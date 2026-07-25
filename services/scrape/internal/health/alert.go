package health

import (
	"sync"
	"time"
)

type Alerter interface {
	SendAlert(platformID int16, version string, state Health, failRate float64, n int, baseline float64)
}

type AlertDedup struct {
	mu       sync.Mutex
	lastSent map[string]time.Time
	cooldown time.Duration
	sink     Alerter
}

func NewAlertDedup(sink Alerter, cooldown time.Duration) *AlertDedup {
	return &AlertDedup{
		lastSent: make(map[string]time.Time),
		cooldown: cooldown,
		sink:     sink,
	}
}

func (a *AlertDedup) Alert(platformID int16, version string, state Health, failRate float64, n int, baseline float64) {
	if state == Healthy {
		return // No alert for healthy, though could alert for recovery
	}

	key := string(rune(platformID)) + ":" + version
	a.mu.Lock()
	defer a.mu.Unlock()

	last, ok := a.lastSent[key]
	if !ok || time.Since(last) >= a.cooldown {
		a.sink.SendAlert(platformID, version, state, failRate, n, baseline)
		a.lastSent[key] = time.Now()
	}
}
