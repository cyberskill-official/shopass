package affil

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeVault map[string]string

func (v fakeVault) Get(ctx context.Context, keyPath string) (string, error) {
	if val, ok := v[keyPath]; ok {
		return val, nil
	}
	return "", nil
}

func hmacHex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerify_GoodSignature(t *testing.T) {
	secrets := fakeVault{"affil/involve_asia/postback_secret": "shh-secret"}
	body := []byte(`{"sub_id":"sd_x"}`)
	sig := hmacHex("shh-secret", body)
	ok, err := VerifyPostback(context.Background(), secrets, "involve_asia", body, sig)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestVerify_TamperedBody(t *testing.T) {
	secrets := fakeVault{"affil/involve_asia/postback_secret": "shh-secret"}
	sig := hmacHex("shh-secret", []byte(`{"commission":1000}`))
	ok, _ := VerifyPostback(context.Background(), secrets, "involve_asia", []byte(`{"commission":9999999}`), sig)
	require.False(t, ok) // body đổi -> chữ ký không khớp
}
