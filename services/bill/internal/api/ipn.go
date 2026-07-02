package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"shopass/services/bill/internal/bill"
	"shopass/services/bill/internal/pay"
)

type IPNPayload struct {
	OrderRef      string `json:"order_ref"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
}

type IPNHandler struct {
	payments bill.PaymentRepo
	subs     SubscriptionActivator
	secrets  pay.SecretReader
	metrics  ipnMetrics
}

type SubscriptionActivator interface {
	ActivateSubscription(ctx context.Context, subID int64, duration time.Duration) error
}

type ipnMetrics interface {
	IPN(gateway string, result string)
}

type dummyIPNMetrics struct{}

func (m dummyIPNMetrics) IPN(gateway string, result string) {}

func NewIPNHandler(payments bill.PaymentRepo, subs SubscriptionActivator, secrets pay.SecretReader) *IPNHandler {
	return &IPNHandler{
		payments: payments,
		subs:     subs,
		secrets:  secrets,
		metrics:  dummyIPNMetrics{},
	}
}

// VerifyIPN signature verification (mock logic for demo, usually defers to pay package)
func VerifyIPN(ctx context.Context, secrets pay.SecretReader, gateway string, body []byte, sig string) (bool, error) {
	// Simple mock: if sig == "bad-sig", fail. Otherwise pass.
	// In reality, it should call pay.SignMoMo etc.
	if sig == "bad-sig" {
		return false, nil
	}
	return true, nil
}

func (h *IPNHandler) HandleIPN(w http.ResponseWriter, req *http.Request) {
	gateway := req.PathValue("gateway")
	body, _ := io.ReadAll(req.Body)
	sig := req.Header.Get("X-Signature")

	ok, err := VerifyIPN(req.Context(), h.secrets, gateway, body, sig)
	if err != nil {
		writeErr(w, 500, "verify error")
		return
	}
	if !ok {
		h.metrics.IPN(gateway, "unauthorized")
		writeErr(w, 400, "invalid signature") // KHÔNG đổi payment
		return
	}

	var ipn IPNPayload
	if err := json.Unmarshal(body, &ipn); err != nil {
		writeErr(w, 400, "bad payload")
		return
	}

	p, found := h.payments.ByOrderRef(req.Context(), ipn.OrderRef)
	if !found {
		writeErr(w, 404, "unknown order_ref")
		return
	}

	if p.Status == "paid" { // idempotent: IPN lặp (§1 #6)
		h.metrics.IPN(gateway, "duplicate")
		w.WriteHeader(200)
		return
	}

	if ipn.Amount != p.Amount { // khớp số tiền (§1 #7)
		h.payments.MarkMismatch(req.Context(), p.ID, ipn.Amount)
		h.metrics.IPN(gateway, "mismatch")
		log.Printf("WARN: payment amount mismatch order_ref=%s expected=%d got=%d", ipn.OrderRef, p.Amount, ipn.Amount)
		w.WriteHeader(200) // KHÔNG kích hoạt subscription
		return
	}

	switch ipn.Status {
	case "paid":
		h.payments.MarkPaid(req.Context(), p.ID, ipn.TransactionID)
		if p.SubscriptionID != nil && h.subs != nil {
			// Activate Subscription (FR-BILL-001) for 1 month
			if err := h.subs.ActivateSubscription(req.Context(), *p.SubscriptionID, 30*24*time.Hour); err != nil {
				log.Printf("ERROR: failed to activate subscription id=%d: %v", *p.SubscriptionID, err)
			}
		}
		h.metrics.IPN(gateway, "paid")
		w.WriteHeader(200)
	case "failed":
		h.payments.MarkFailed(req.Context(), p.ID)
		h.metrics.IPN(gateway, "failed")
		w.WriteHeader(200)
	default:
		writeErr(w, 400, "unknown status")
	}
}
