package pay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

type fakeVault map[string]string

func (f fakeVault) Get(ctx context.Context, path string) (string, error) {
	return f[path], nil
}

func hmacHex(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestSignMoMo_FixedVector(t *testing.T) {
	secrets := fakeVault{"bill/momo/secret_key": "test-key"}
	sig, err := SignMoMo(context.Background(), secrets, "amount=29000&orderId=o1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := hmacHex("test-key", "amount=29000&orderId=o1")
	if sig != expected {
		t.Fatalf("expected %s, got %s", expected, sig)
	}
}
