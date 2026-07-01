package chart

import (
	"sort"
	"time"
)

type DailyPoint struct {
    Day    time.Time `json:"day"`
    MinP   int64     `json:"min_p"`   // VND
    MaxP   int64     `json:"max_p"`   // VND
    CloseP int64     `json:"close_p"` // VND, giá hiển thị của điểm ngày
}

type Annotations struct {
    Median90     int64    `json:"median90"`     // VND, trung vị close_p 90 ngày
    TrailingMin  int64    `json:"trailing_min"` // VND, đáy min_p trong khoảng
    Verdict      string   `json:"verdict"`      // từ FR-DEAL-001
    Accumulating bool     `json:"accumulating"` // true khi maturity=WARMING
    DoubleDates  []string `json:"double_dates"` // YYYY-MM-DD trong khoảng
}

// Build tính median90 + trailing_min từ chuỗi daily và sinh mốc ngày đôi (DEC-DEAL-21).
// verdict + accumulating được nhồi bởi caller (chart.go) từ FR-DEAL-001/002.
func Build(daily []DailyPoint, from, to time.Time) Annotations {
    return Annotations{
        Median90:    median90(daily),
        TrailingMin: trailingMin(daily),
        DoubleDates: doubleDates(from, to),
    }
}

// median90 lấy trung vị close_p của các điểm trong 90 ngày gần nhất.
func median90(d []DailyPoint) int64 {
    cut := time.Now().AddDate(0, 0, -90)
    var v []int64
    for _, p := range d {
        if !p.Day.Before(cut) {
            v = append(v, p.CloseP)
        }
    }
    if len(v) == 0 {
        return 0
    }
    sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
    return v[len(v)/2]
}

// trailingMin lấy đáy min_p trong toàn chuỗi được trả.
func trailingMin(d []DailyPoint) int64 {
    if len(d) == 0 {
        return 0
    }
    m := d[0].MinP
    for _, p := range d[1:] {
        if p.MinP < m {
            m = p.MinP
        }
    }
    return m
}

// doubleDates liệt kê các ngày dd==mm (1.1...12.12) rơi trong [from, to] (DEC-DEAL-21).
func doubleDates(from, to time.Time) []string {
    var out []string
    for y := from.Year(); y <= to.Year(); y++ {
        for m := 1; m <= 12; m++ {
            d := time.Date(y, time.Month(m), m, 0, 0, 0, 0, time.UTC)
            if !d.Before(from) && !d.After(to) {
                out = append(out, d.Format("2006-01-02"))
            }
        }
    }
	if out == nil {
		out = []string{}
	}
    return out
}
