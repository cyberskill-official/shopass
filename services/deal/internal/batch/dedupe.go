package batch

import (
	"context"
	"time"
)

const cooldownDays = 7 // cooldown rising-edge liên ngày (DEC-DEAL-53)

// shouldSkip: true nếu (user, product) đã có alert hôm nay (dedupe ngày)
// hoặc còn trong cooldown rising-edge (đã bắn trong cooldownDays ngày gần nhất).
func (b *Batch) shouldSkip(ctx context.Context, userID, productID int64, today time.Time) bool {
	var lastFired *time.Time
	err := b.pool.QueryRow(ctx, `
        SELECT max(fired_on) FROM bottom_alert_log
        WHERE user_id = $1 AND product_id = $2`, userID, productID).Scan(&lastFired)
	if err != nil || lastFired == nil {
		return false // chưa từng bắn -> cạnh lên, cho phép
	}
	days := int(today.Sub(*lastFired).Hours() / 24)
	return days < cooldownDays // trong cooldown (gồm cả cùng ngày = 0) -> bỏ qua
}

// enqueueAndLog enqueue vào notification fan-out rồi ghi bottom_alert_log.
// Ghi log CHỈ khi enqueue thành công (§1 #10 #11).
func (b *Batch) enqueueAndLog(ctx context.Context, userID, productID int64, p float64, today time.Time) error {
	if err := b.notif.Enqueue(ctx, NotifItem{
		UserID:    userID,
		ProductID: productID,
		Reason:    "bottom_predicted",
		Payload:   map[string]any{"p_bottom_14d": p},
	}); err != nil {
		return err
	}
	_, err := b.pool.Exec(ctx, `
        INSERT INTO bottom_alert_log (user_id, product_id, fired_on, p_bottom)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, product_id, fired_on) DO NOTHING`, // idempotent retry
		userID, productID, today, p)
	return err
}
