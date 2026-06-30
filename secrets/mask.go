package secrets

import (
	"encoding/json"
	"fmt"
)

func mask(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return "****" + val[len(val)-4:]
}

// String masks the secret.
func (s Secret) String() string {
	return mask(s.value)
}

// MarshalJSON masks the secret.
func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(mask(s.value))
}

// Format ensures it's masked with formatting verbs like %+v or %v.
func (s Secret) Format(f fmt.State, c rune) {
	fmt.Fprint(f, mask(s.value))
}
