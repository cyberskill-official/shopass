package health

import (
	"sync"
	"testing"
)

func TestWindow_RateAndCount(t *testing.T) {
	w := NewWindow(100)
	for i := 0; i < 30; i++ {
		w.Record(OutcomeSuccess)
	}
	for i := 0; i < 10; i++ {
		w.Record(OutcomeParseFail)
	}
	rate, n := w.ParseFailRate()
	if n != 40 {
		t.Errorf("Expected 40 samples, got %d", n)
	}
	if rate < 0.24 || rate > 0.26 {
		t.Errorf("Expected rate ~0.25, got %v", rate)
	}
}

func TestWindow_RaceSafe(t *testing.T) {
	w := NewWindow(1000)
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w.Record(OutcomeSuccess)
			}
		}()
	}
	wg.Wait()
	_, n := w.ParseFailRate()
	if n != 1000 { // Max capacity is 1000, we wrote 1600 items. Cap is 1000. Wait, actually `n` represents count. Since it overflows it should cap at 1000.
		t.Errorf("Expected 1000 samples, got %d", n)
	}
}
