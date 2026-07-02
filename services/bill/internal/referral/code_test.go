package referral

import (
	"strings"
	"testing"
)

func TestNewCode_NoConfusingChars(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		c := NewCode(6)
		if seen[c] {
			t.Fatalf("expected unique code, got duplicate: %s", c)
		}
		seen[c] = true

		if strings.ContainsAny(c[2:], "O0I1") {
			t.Fatalf("code contains confusing char: %s", c)
		}
		if !strings.HasPrefix(c, "SD") {
			t.Fatalf("code should start with SD: %s", c)
		}
	}
}
