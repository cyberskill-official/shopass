package pay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type SecretReader interface {
	Get(ctx context.Context, path string) (string, error)
}

func SignMoMo(ctx context.Context, secrets SecretReader, raw string) (string, error) {
	key, err := secrets.Get(ctx, "bill/momo/secret_key")
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func SignZaloPay(ctx context.Context, secrets SecretReader, raw string) (string, error) {
	key, err := secrets.Get(ctx, "bill/zalopay/mac_key")
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func SignVNPay(ctx context.Context, secrets SecretReader, raw string) (string, error) {
	key, err := secrets.Get(ctx, "bill/vnpay/hash_secret")
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
