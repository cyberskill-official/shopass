package referral

import (
	"crypto/rand"
)

const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func NewCode(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // in practice, crypto/rand should not fail
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return "SD" + string(b)
}
