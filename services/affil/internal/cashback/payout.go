package cashback

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Payer sends a cashback payout (VietQR / internal balance). CI uses the noop stub.
type Payer interface {
	Pay(ctx context.Context, userID, amount int64, orderRef string) (gatewayRef string, err error)
}

// VietQRStub returns a synthetic VietQR payload (TASK-BILL-002 pattern; no live creds).
type VietQRStub struct {
	Log *slog.Logger
}

func NewVietQRStub(log *slog.Logger) *VietQRStub {
	return &VietQRStub{Log: log}
}

func (v *VietQRStub) Pay(ctx context.Context, userID, amount int64, orderRef string) (string, error) {
	ref := fmt.Sprintf("vietqr://cashback/%d?amount=%d&addInfo=%s", userID, amount, orderRef)
	if v != nil && v.Log != nil {
		v.Log.Info("cashback payout noop", "user_id", userID, "amount", amount, "ref", ref)
	}
	return ref, nil
}

// MaybeRequestPayout aggregates available entries when sum >= threshold and pays once.
func (l *Ledger) MaybeRequestPayout(ctx context.Context, userID int64) (bool, error) {
	if l == nil || l.Store == nil {
		return false, nil
	}
	sum, err := l.Store.SumAvailable(ctx, userID)
	if err != nil {
		return false, err
	}
	threshold := l.Cfg.PayoutThreshold
	if threshold <= 0 {
		threshold = DefaultConfig().PayoutThreshold
	}
	if sum < threshold {
		return false, nil
	}
	entries, err := l.Store.ListAvailable(ctx, userID)
	if err != nil {
		return false, err
	}
	if len(entries) == 0 {
		return false, nil
	}
	ids := make([]int64, 0, len(entries))
	var total int64
	for _, e := range entries {
		ids = append(ids, e.ConversionID)
		total += e.UserShare
	}
	if total < threshold {
		return false, nil
	}
	orderRef := fmt.Sprintf("cb-%d-%d", userID, time.Now().UTC().Unix())
	gatewayRef := ""
	if l.Payer != nil {
		ref, err := l.Payer.Pay(ctx, userID, total, orderRef)
		if err != nil {
			_, _ = l.Store.CreatePayoutRequest(ctx, userID, total, "failed:"+err.Error())
			return false, err
		}
		gatewayRef = ref
	} else {
		gatewayRef = "noop:" + orderRef
	}
	if _, err := l.Store.CreatePayoutRequest(ctx, userID, total, gatewayRef); err != nil {
		return false, err
	}
	paidAt := time.Now().UTC()
	if err := l.Store.MarkPaid(ctx, ids, paidAt); err != nil {
		return false, err
	}
	if l.Metrics != nil {
		l.Metrics.Paid(total)
	}
	return true, nil
}
