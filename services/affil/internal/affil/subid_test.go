package affil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubID_UniqueAndNoPII(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 1000; i++ {
		s := NewSubID()
		require.False(t, seen[s]) // duy nhat
		seen[s] = true
		require.NotContains(t, s, "@") // khong giong email
		// s format is "sd_" + 24 chars hex = 27 chars
		require.True(t, strings.HasPrefix(s, "sd_"))
		require.GreaterOrEqual(t, len(s), 27) // du entropy
	}
}
