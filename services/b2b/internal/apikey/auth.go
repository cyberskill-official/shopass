package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidKey = errors.New("apikey: invalid")

// NewKey returns prefix.secret and the hash to store (never store cleartext secret).
func NewKey() (prefix, secret, hash string, err error) {
	var b [24]byte
	if _, err = rand.Read(b[:]); err != nil {
		return "", "", "", err
	}
	prefix = hex.EncodeToString(b[:8])
	secret = hex.EncodeToString(b[8:])
	sum := sha256.Sum256([]byte(secret))
	hash = hex.EncodeToString(sum[:])
	return prefix, secret, hash, nil
}

func ParsePresented(raw string) (prefix, secret string, err error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidKey
	}
	return parts[0], parts[1], nil
}

func Verify(secret, storedHash string) bool {
	sum := sha256.Sum256([]byte(secret))
	got := hex.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHash)) == 1
}

func Format(prefix, secret string) string {
	return fmt.Sprintf("%s.%s", prefix, secret)
}
