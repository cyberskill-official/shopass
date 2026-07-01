package health

import (
	"sync"
)

type Outcome int

const (
	OutcomeSuccess Outcome = iota
	OutcomeParseFail
	OutcomeChallenge
	OutcomeNetworkErr
)

// Window là rolling window lưu các kết quả (outcomes) gần đây
type Window struct {
	mu      sync.Mutex
	samples []Outcome
	cap     int
	idx     int
	count   int
	fails   int
}

func NewWindow(cap int) *Window {
	return &Window{
		samples: make([]Outcome, cap),
		cap:     cap,
	}
}

// Record thêm outcome; thread-safe (nhiều worker song song).
func (w *Window) Record(o Outcome) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.count == w.cap {
		// Xóa phần tử cũ
		old := w.samples[w.idx]
		if old == OutcomeParseFail {
			w.fails--
		}
	} else {
		w.count++
	}

	w.samples[w.idx] = o
	if o == OutcomeParseFail {
		w.fails++
	}

	w.idx = (w.idx + 1) % w.cap
}

// ParseFailRate trả tỷ lệ parse_fail hiện tại và số mẫu.
func (w *Window) ParseFailRate() (rate float64, n int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.count == 0 {
		return 0, 0
	}
	return float64(w.fails) / float64(w.count), w.count
}
