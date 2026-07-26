package pay

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvSecrets reads gateway HMAC secrets from process environment.
// Paths match SignMoMo / SignZaloPay / SignVNPay:
//
//	bill/momo/secret_key      → MOMO_SECRET_KEY
//	bill/zalopay/mac_key      → ZALOPAY_MAC_KEY
//	bill/vnpay/hash_secret    → VNPAY_HASH_SECRET
type EnvSecrets struct{}

func NewEnvSecrets() *EnvSecrets { return &EnvSecrets{} }

func (e *EnvSecrets) Get(_ context.Context, path string) (string, error) {
	envKey, ok := map[string]string{
		"bill/momo/secret_key":   "MOMO_SECRET_KEY",
		"bill/zalopay/mac_key":   "ZALOPAY_MAC_KEY",
		"bill/vnpay/hash_secret": "VNPAY_HASH_SECRET",
	}[path]
	if !ok {
		return "", fmt.Errorf("unknown secret path %q", path)
	}
	v := strings.TrimSpace(os.Getenv(envKey))
	if v == "" {
		return "", fmt.Errorf("%s not set", envKey)
	}
	return v, nil
}
