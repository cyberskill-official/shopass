package health

type Health int

const (
	Healthy Health = iota
	Degraded
	Broken
)

const minSamples = 30 // số mẫu tối thiểu trước khi đổi state

// Next suy ra state mới từ tỷ lệ fail, baseline, và state hiện tại (có hysteresis).
func Next(cur Health, failRate, baseline float64, n int) Health {
	if n < minSamples {
		return cur // chưa đủ mẫu, giữ nguyên
	}
	up := baseline + 0.25   // ngưỡng lên
	upHard := 0.70          // phần lớn fail -> broken
	down := baseline + 0.10 // ngưỡng xuống (hysteresis)
	switch {
	case failRate >= upHard:
		return Broken
	case failRate >= up:
		if cur == Broken {
			// still > down but not >= upHard, wait. Actually wait:
			// If cur == Broken and failRate is >= up, shouldn't it transition to Degraded?
			// Let's think:
			// if it was Broken, it goes to Degraded if failRate < upHard? Yes, if it is >= up.
			// Let's just return Degraded here, so it steps down.
			return Degraded
		}
		return Degraded
	case failRate <= down:
		return Healthy
	default:
		return cur // vùng hysteresis -> giữ nguyên
	}
}
