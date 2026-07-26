package pay

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// sandboxAware wraps a stub gateway and, when partner credentials are present,
// returns a signed sandbox redirect URL instead of an unsigned placeholder.
// Full partner HTTP create-order APIs can replace the signed-URL shape later
// without changing the PaymentGateway contract.
type sandboxAware struct {
	code    string
	secrets SecretReader
	baseURL string
	stub    PaymentGateway
}

func (s *sandboxAware) Code() string { return s.code }

func (s *sandboxAware) CreatePayment(ctx context.Context, r PaymentRequest) (PaymentResult, error) {
	if s.secrets == nil || !s.credsPresent() {
		return s.stub.CreatePayment(ctx, r)
	}
	raw := fmt.Sprintf("amount=%d&orderId=%s&requestId=%d", r.Amount, r.OrderRef, time.Now().UnixNano())
	var sig string
	var err error
	switch s.code {
	case "momo":
		sig, err = SignMoMo(ctx, s.secrets, raw)
	case "zalopay":
		sig, err = SignZaloPay(ctx, s.secrets, raw)
	case "vnpay":
		sig, err = SignVNPay(ctx, s.secrets, raw)
	default:
		return s.stub.CreatePayment(ctx, r)
	}
	if err != nil {
		return PaymentResult{}, err
	}
	u, err := url.Parse(s.baseURL)
	if err != nil {
		return PaymentResult{}, err
	}
	q := u.Query()
	q.Set("amount", strconv.FormatInt(r.Amount, 10))
	if s.code == "vnpay" {
		q.Set("vnp_Amount", strconv.FormatInt(r.Amount*100, 10))
		q.Set("vnp_TxnRef", r.OrderRef)
	} else {
		q.Set("orderId", r.OrderRef)
	}
	q.Set("signature", sig)
	u.RawQuery = q.Encode()
	return PaymentResult{
		OrderRef: r.OrderRef,
		Gateway:  s.code,
		Amount:   r.Amount,
		PayURL:   u.String(),
	}, nil
}

func (s *sandboxAware) credsPresent() bool {
	switch s.code {
	case "momo":
		return os.Getenv("MOMO_SECRET_KEY") != "" && os.Getenv("MOMO_PARTNER_CODE") != ""
	case "zalopay":
		return os.Getenv("ZALOPAY_MAC_KEY") != "" && os.Getenv("ZALOPAY_APP_ID") != ""
	case "vnpay":
		return os.Getenv("VNPAY_HASH_SECRET") != "" && os.Getenv("VNPAY_TMN_CODE") != ""
	}
	return false
}

// NewMoMoSandbox returns MoMo gateway: stub URL without creds, signed sandbox URL with creds.
func NewMoMoSandbox(secrets SecretReader) PaymentGateway {
	base := envOr("MOMO_PAY_URL", "https://test-payment.momo.vn/v2/gateway/pay")
	return &sandboxAware{code: "momo", secrets: secrets, baseURL: base, stub: NewMoMo()}
}

// NewZaloPaySandbox returns ZaloPay gateway with the same env-gated behavior.
func NewZaloPaySandbox(secrets SecretReader) PaymentGateway {
	base := envOr("ZALOPAY_PAY_URL", "https://sandbox.zalopay.vn/v001/tpe/redirectpay")
	return &sandboxAware{code: "zalopay", secrets: secrets, baseURL: base, stub: NewZaloPay()}
}

// NewVNPaySandbox returns VNPay gateway with the same env-gated behavior.
func NewVNPaySandbox(secrets SecretReader) PaymentGateway {
	base := envOr("VNPAY_PAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html")
	return &sandboxAware{code: "vnpay", secrets: secrets, baseURL: base, stub: NewVNPay()}
}

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

// SignRaw is a test helper for HMAC-SHA256 hex digests.
func SignRaw(key, raw string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}
