package batch

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// RunNightlyScore chấm điểm mọi SKU MATURE, bắn cảnh báo đáy khi p_bottom_14d > 0.7.
// Model-agnostic: chỉ đọc price_forecast.p_bottom_14d (Prophet hoặc LightGBM).
func (b *Batch) RunNightlyScore(ctx context.Context, today time.Time) error {
	startTime := time.Now()
	defer func() {
		batchDuration.Record(ctx, float64(time.Since(startTime).Milliseconds()))
	}()

	// §1 #2 #3 #4 #5: chỉ SKU MATURE, dự báo còn tươi, p_bottom_14d > 0.7 (strict).
	rows, err := b.pool.Query(ctx, `
        SELECT f.product_id, f.p_bottom_14d
        FROM price_forecast f
        JOIN tracked_product tp ON tp.id = f.product_id
        WHERE f.p_bottom_14d > 0.7                              -- DEC-DEAL-50, strict
          AND f.scored_at  >= now() - INTERVAL '36 hours'       -- §1 #4, dự báo tươi
          AND now() - tp.first_seen >= INTERVAL '90 days'`)     // §1 #2, maturity gate
	if err != nil {
		return err
	}
	defer rows.Close()

	var scored, fired, skipped int
	for rows.Next() {
		var productID int64
		var pBottom float64
		if err := rows.Scan(&productID, &pBottom); err != nil {
			return err
		}
		scored++
		
		// §1 #6: gom user từ alert_rule type 'bottom_predicted' đang bật.
		users, err := b.matchBottomRules(ctx, productID)
		if err != nil {
			b.log.Warn("match rules failed", "product", productID, "err", err)
			continue // §1 #11: một SKU lỗi không làm hỏng batch
		}
		for _, userID := range users {
			if b.shouldSkip(ctx, userID, productID, today) { // §1 #7 #8 dedupe + cooldown
				skipped++
				continue
			}
			if err := b.enqueueAndLog(ctx, userID, productID, pBottom, today); err != nil {
				b.log.Warn("enqueue failed", "user", userID, "product", productID, "err", err)
				continue // §1 #11: không ghi log -> đêm sau thử lại
			}
			fired++
		}
	}
	
	scoredTotal.Add(ctx, int64(scored))
	firedTotal.Add(ctx, int64(fired))
	skippedTotal.Add(ctx, int64(skipped), metric.WithAttributes())
	return rows.Err()
}
