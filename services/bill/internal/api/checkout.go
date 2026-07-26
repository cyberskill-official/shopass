package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"shopass/services/bill/internal/pay"
)

type Plan struct {
	Tier  string
	Price int64
}

type PlanCatalog interface {
	ByTier(ctx context.Context, tier string) (Plan, bool)
}

type PaymentRepo interface {
	ByOrderRef(ctx context.Context, orderRef string) (pay.PaymentResult, bool)
	InsertPending(ctx context.Context, orderRef string, userID int64, amount int64, gateway string)
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value("user_id")
	id, ok := v.(int64)
	return id, ok && id > 0
}

func userIDFromRequest(req *http.Request) (int64, bool) {
	if id, ok := userIDFromContext(req.Context()); ok {
		return id, true
	}
	raw := req.Header.Get("X-User-Id")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(`{"error":"` + msg + `"}`))
}

type metricsRecorder interface {
	GatewayError(gw string)
	IntentCreated(gw string)
}

type dummyMetrics struct{}

func (m dummyMetrics) GatewayError(gw string)  {}
func (m dummyMetrics) IntentCreated(gw string) {}

type Handler struct {
	plans    PlanCatalog
	gateways *pay.Registry
	payments PaymentRepo
	metrics  metricsRecorder
}

func NewHandler(plans PlanCatalog, gateways *pay.Registry, payments PaymentRepo) *Handler {
	return &Handler{
		plans:    plans,
		gateways: gateways,
		payments: payments,
		metrics:  dummyMetrics{},
	}
}

func (h *Handler) HandleCheckout(w http.ResponseWriter, req *http.Request) {
	userID, ok := userIDFromRequest(req)
	if !ok {
		writeErr(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		PlanTier string `json:"plan_tier"`
		Gateway  string `json:"gateway"`
	}
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid body")
		return
	}
	plan, ok := h.plans.ByTier(req.Context(), body.PlanTier) // TASK-BILL-001
	if !ok {
		writeErr(w, 400, "unknown plan")
		return
	}
	gw, ok := h.gateways.Get(body.Gateway)
	if !ok {
		writeErr(w, 400, "unsupported gateway")
		return
	}

	orderRef := pay.NewOrderRef(userID, plan.Tier) // duy nhất + idempotent
	if existing, found := h.payments.ByOrderRef(req.Context(), orderRef); found {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(existing)
		return // không tạo lần hai
	}

	res, err := gw.CreatePayment(req.Context(), pay.PaymentRequest{
		OrderRef: orderRef, Amount: plan.Price, UserID: userID, PlanTier: plan.Tier, // BIGINT VND
	})
	if err != nil {
		h.metrics.GatewayError(body.Gateway)
		writeErr(w, 502, "payment gateway error")
		return // KHÔNG coi là thành công
	}
	h.payments.InsertPending(req.Context(), orderRef, userID, plan.Price, body.Gateway) // ghi vào DB
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(res)
	h.metrics.IntentCreated(body.Gateway)
}
