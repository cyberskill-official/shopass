package affil

import (
	"crypto/rand"
	"encoding/hex"
)

// NewSubID generates a random sub_id with the prefix "sd_".
// It uses 12 random bytes resulting in a 24-character hex string.
func NewSubID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "sd_" + hex.EncodeToString(b)
}
