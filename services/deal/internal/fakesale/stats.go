package fakesale

import "sort"

// percentile trả phân vị thứ p (0..100) theo nearest-rank trên chuỗi giá int64.
// Sao chép trước khi sort để giữ hàm thuần, không sửa slice đầu vào.
func percentile(hist []int64, p int) int64 {
	n := len(hist)
	s := make([]int64, n)
	copy(s, hist)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	// rank nearest-rank: ceil(p/100 * n), kẹp trong [1, n].
	rank := (p*n + 99) / 100
	if rank < 1 {
		rank = 1
	}
	if rank > n {
		rank = n
	}
	return s[rank-1]
}

func minInt64(hist []int64) int64 {
	m := hist[0]
	for _, v := range hist[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
