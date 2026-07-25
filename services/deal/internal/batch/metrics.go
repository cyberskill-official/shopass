package batch

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter         = otel.Meter("deal-batch")
	scoredTotal   metric.Int64Counter
	firedTotal    metric.Int64Counter
	skippedTotal  metric.Int64Counter
	batchDuration metric.Float64Histogram
)

func init() {
	var err error
	scoredTotal, err = meter.Int64Counter("deal_nightly_scored_total")
	if err != nil {
		panic(err)
	}
	firedTotal, err = meter.Int64Counter("deal_bottom_alert_fired_total")
	if err != nil {
		panic(err)
	}
	skippedTotal, err = meter.Int64Counter("deal_alert_dedupe_skipped_total")
	if err != nil {
		panic(err)
	}
	batchDuration, err = meter.Float64Histogram("deal_nightly_batch_duration_ms")
	if err != nil {
		panic(err)
	}
}
