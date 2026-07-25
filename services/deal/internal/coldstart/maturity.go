package coldstart

import "shopass/services/deal/internal/fakesale"

type State int

const (
	StateNew     State = iota // < 14 ngày  -> UNKNOWN
	StateWarming              // 14-90 ngày -> low-confidence + "đang tích lũy"
	StateMature               // >= 90 ngày -> full confidence
)

const (
	warmingDays = 14
	matureDays  = 90
	minSamples  = 30 // sàn sample_count cho prior (§1 #12)
)

// Product interface để test dễ dàng
type Product interface {
	DaysOfHistory() int
}

// Maturity ánh xạ số ngày lịch sử sang trạng thái trưởng thành (DEC-DEAL-10).
func Maturity(daysOfHistory int) State {
	switch {
	case daysOfHistory < warmingDays:
		return StateNew
	case daysOfHistory < matureDays:
		return StateWarming
	default:
		return StateMature
	}
}

// IsFeatureReady: cổng baseline 90 ngày cho sale ảo/biểu đồ công khai (DEC-DEAL-11).
func IsFeatureReady(p Product) bool {
	return Maturity(p.DaysOfHistory()) == StateMature
}

// EvaluateVerdict áp dụng logic maturity vào kết quả raw của DetectFakeSale.
func EvaluateVerdict(p Product, rawVerdict fakesale.Verdict) (fakesale.Verdict, bool) {
	state := Maturity(p.DaysOfHistory())
	if state == StateNew {
		return fakesale.Unknown, false
	}
	ready := IsFeatureReady(p)
	if !ready {
		// Trả về low_confidence hoặc Unknown nếu không muốn surface
		// Theo §1 #4: WARMING -> low_confidence + "đang tích lũy", không phát verdict đầy đủ.
		// Để dễ implement, có thể trả verdict với cờ ready=false.
		return rawVerdict, false
	}
	return rawVerdict, true
}
