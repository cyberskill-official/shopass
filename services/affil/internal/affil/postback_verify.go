package affil

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SecretReader is an interface for reading secrets from Vault (FR-INFRA-003).
type SecretReader interface {
	Get(ctx context.Context, keyPath string) (string, error)
}

func secretRefFor(network string) string {
	return fmt.Sprintf("affil/%s/postback_secret", network)
}

// VerifyPostback kiểm HMAC-SHA256 của body bằng secret đọc từ Vault cho network.
// Trả false nếu thiếu/sai chữ ký -> handler trả 401, KHÔNG ghi conversion (§1 #5).
func VerifyPostback(ctx context.Context, secrets SecretReader, network string, body []byte, gotSig string) (bool, error) {
	secret, err := secrets.Get(ctx, secretRefFor(network))
	if err != nil {
		return false, err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(want), []byte(gotSig)), nil // so sánh hằng thời gian
}
