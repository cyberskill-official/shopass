package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"regexp"
	"testing"
)

var base64url = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func TestNewVerifier_ShapeAndUniqueness(t *testing.T) {
	v1, err := NewVerifier()
	if err != nil {
		t.Fatal(err)
	}
	v2, err := NewVerifier()
	if err != nil {
		t.Fatal(err)
	}
	if len(v1) < 43 || len(v1) > 128 {
		t.Fatalf("verifier length %d outside RFC 7636 range 43-128", len(v1))
	}
	if !base64url.MatchString(v1) {
		t.Fatalf("verifier not base64url: %q", v1)
	}
	if v1 == v2 {
		t.Fatal("two verifiers must differ (random)")
	}
}

func TestChallenge_IsS256OfVerifier(t *testing.T) {
	v := "test-verifier-abc"
	want := func() string {
		s := sha256.Sum256([]byte(v))
		return base64.RawURLEncoding.EncodeToString(s[:])
	}()
	if got := Challenge(v); got != want {
		t.Fatalf("Challenge = %q, want S256 %q", got, want)
	}
	// Determinism + sensitivity.
	if Challenge(v) != Challenge(v) {
		t.Fatal("Challenge must be deterministic")
	}
	if Challenge("other") == Challenge(v) {
		t.Fatal("different verifiers must produce different challenges")
	}
}
