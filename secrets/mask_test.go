package secrets

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSecret_NeverLeaksRaw(t *testing.T) {
	s := NewSecret("super-secret-token-123", "v3")
	if contains(fmt.Sprintf("%v", s), "super-secret-token-123") {
		t.Error("String() leaked raw value")
	}
	if contains(fmt.Sprintf("%+v", s), "super-secret-token-123") {
		t.Error("detailed format leaked raw value")
	}
	b, _ := json.Marshal(s)
	if contains(string(b), "super-secret-token-123") {
		t.Error("MarshalJSON() leaked raw value")
	}
	if s.Reveal() != "super-secret-token-123" {
		t.Error("Reveal() should return raw value")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
