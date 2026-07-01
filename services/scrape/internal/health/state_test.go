package health

import (
	"testing"
)

func TestNext_SpikeToBroken(t *testing.T) {
	if Next(Healthy, 0.85, 0.05, 100) != Broken {
		t.Errorf("Expected Broken")
	}
}

func TestNext_MinSamplesGuard(t *testing.T) {
	if Next(Healthy, 0.90, 0.05, 5) != Healthy {
		t.Errorf("Expected Healthy due to min samples")
	}
}

func TestNext_Hysteresis_NoFlicker(t *testing.T) {
	if Next(Degraded, 0.20, 0.05, 100) != Degraded {
		t.Errorf("Expected Degraded due to hysteresis")
	}
}

func TestNext_Recovers(t *testing.T) {
	if Next(Broken, 0.08, 0.05, 100) != Healthy {
		t.Errorf("Expected Healthy after recovery")
	}
}
