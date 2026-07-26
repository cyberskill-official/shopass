package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"shopass/services/affil/internal/affil"
	"shopass/services/affil/internal/cashback"
)

type PostbackPayload struct {
	SubID      string `json:"sub_id"`
	OrderValue int64  `json:"order_value"`
	Commission int64  `json:"commission"`
	Status     string `json:"status"` // e.g. "approved", "rejected", "pending"
	UserTier   string `json:"user_tier,omitempty"`
}

// PayoutHoldCreator is TRUST-005: create payout_hold on confirm (never pay immediately).
type PayoutHoldCreator interface {
	OnConversionConfirmed(ctx context.Context, conversionID, beneficiaryID, amount int64) error
}

// CashbackLedger is TASK-AFFIL-005: pending entry on confirm, clawback on reject.
type CashbackLedger interface {
	OnConfirmed(ctx context.Context, c cashback.Conversion) (cashback.Entry, error)
	Clawback(ctx context.Context, conversionID int64) error
}

// Handler contains dependencies for the API
type PostbackHandler struct {
	repo     *affil.Repo
	secrets  affil.SecretReader
	holds    PayoutHoldCreator
	cashback CashbackLedger
}

func NewPostbackHandler(repo *affil.Repo, secrets affil.SecretReader) *PostbackHandler {
	return &PostbackHandler{
		repo:    repo,
		secrets: secrets,
	}
}

// WithPayoutHolds wires TRUST-005 hold creation after network confirm.
func (h *PostbackHandler) WithPayoutHolds(holds PayoutHoldCreator) *PostbackHandler {
	h.holds = holds
	return h
}

// WithCashback wires TASK-AFFIL-005 ledger hooks.
func (h *PostbackHandler) WithCashback(ledger CashbackLedger) *PostbackHandler {
	h.cashback = ledger
	return h
}

func (h *PostbackHandler) HandlePostback(w http.ResponseWriter, req *http.Request) {
	network := req.PathValue("network")
	body, _ := io.ReadAll(req.Body)
	sig := req.Header.Get("X-Signature")

	ok, err := affil.VerifyPostback(req.Context(), h.secrets, network, body, sig)
	_ = h.repo.LogPostback(req.Context(), network, body, sig, ok) // luôn ghi raw (§1 #6)

	if err != nil {
		writeErr(w, 500, "verify error")
		return
	}
	if !ok {
		writeErr(w, 401, "invalid signature") // KHÔNG ghi conversion (§1 #5)
		return
	}

	var p PostbackPayload
	if err := json.Unmarshal(body, &p); err != nil {
		writeErr(w, 400, "bad payload")
		return
	}

	cid, err := h.repo.RecordConversion(req.Context(), p.SubID, p.OrderValue, p.Commission, network)
	switch {
	case errors.Is(err, affil.ErrUnknownSubID):
		writeErr(w, 404, "unknown sub_id") // không conversion mồ côi (§1 #7)
		return
	case errors.Is(err, affil.ErrConversionExists):
		cid = h.repo.ConversionIDBySubID(req.Context(), p.SubID) // idempotent (§1 #9)
	case err != nil:
		writeErr(w, 500, "internal error")
		return
	}

	switch p.Status { // map trạng thái network -> vòng đời (§1 #8)
	case "approved":
		_ = h.repo.ConfirmConversion(req.Context(), cid)
		if h.holds != nil {
			if seed, err := h.repo.HoldSeedByID(req.Context(), cid); err == nil {
				_ = h.holds.OnConversionConfirmed(req.Context(), seed.ConversionID, seed.BeneficiaryID, seed.Commission)
			}
		}
		if h.cashback != nil {
			if seed, err := h.repo.HoldSeedByID(req.Context(), cid); err == nil {
				tier := p.UserTier
				if tier == "" {
					tier = cashback.TierFree
				}
				_, _ = h.cashback.OnConfirmed(req.Context(), cashback.Conversion{
					ID:          seed.ConversionID,
					UserID:      seed.BeneficiaryID,
					Commission:  seed.Commission,
					UserTier:    tier,
					ConfirmedAt: time.Now().UTC(),
				})
			}
		}
	case "rejected":
		_ = h.repo.RejectConversion(req.Context(), cid, "network rejected")
		if h.cashback != nil {
			_ = h.cashback.Clawback(req.Context(), cid)
		}
	default:
		// giữ pending
	}
	w.WriteHeader(200)
}
