package fakesale

import "testing"

// flat tạo chuỗi n điểm cùng giá v (mặt bằng giá ổn định).
func flat(v int64, n int) []int64 {
	h := make([]int64, n)
	for i := range h {
		h[i] = v
	}
	return h
}

// mixed tạo chuỗi 81 điểm ở giá hi + 9 điểm ở giá lo: median90 = hi
// nhưng p10 = trailing_min = lo (đáy lịch sử nằm thấp hơn mặt bằng).
func mixed(hi, lo int64) []int64 {
	return append(flat(hi, 81), flat(lo, 9)...)
}

func TestDetect_ColdStart_Unknown(t *testing.T) {
	if got := DetectFakeSale(flat(100_000, 13), 50_000, 200_000); got != Unknown {
		t.Fatalf("13 điểm phải UNKNOWN, được %s", got)
	}
}

func TestDetect_InflatedListPrice_SaleAo(t *testing.T) {
	hist := flat(100_000, 90) // median90 = 100_000
	// list 149_000 > 115_000 (thổi) ; current 120_000 >= 97_000 (giảm giả)
	if got := DetectFakeSale(hist, 120_000, 149_000); got != SaleAo {
		t.Fatalf("phải SALE_AO, được %s", got)
	}
}

func TestDetect_GenuineLow_SaleXin(t *testing.T) {
	hist := append(flat(100_000, 89), 80_000) // trailing_min = 80_000, p10 thấp
	// current 80_000 <= p10 và 80_000*100 <= 80_000*102 -> sát đáy
	if got := DetectFakeSale(hist, 80_000, 100_000); got != SaleXin {
		t.Fatalf("phải SALE_XIN, được %s", got)
	}
}

func TestDetect_Middle_TamDuoc(t *testing.T) {
	// median90 = 100_000, đáy lịch sử = 60_000. current 90_000 dưới biên 0.97
	// (không ảo) nhưng vẫn cao hơn p10/đáy -> chưa phải sale xịn -> TAM_DUOC.
	hist := mixed(100_000, 60_000)
	if got := DetectFakeSale(hist, 90_000, 100_000); got != TamDuoc {
		t.Fatalf("phải TAM_DUOC, được %s", got)
	}
}

func TestDetect_BoundaryThresholds(t *testing.T) {
	// median90 = 100_000, p10 = trailing_min = 60_000 (đáy thấp hơn mặt bằng).
	hist := mixed(100_000, 60_000)
	cases := []struct {
		name          string
		current, list int64
		want          Verdict
	}{
		// list*100 == median90*115 (bằng, không lớn hơn) -> không inflated -> không SALE_AO;
		// current cao hơn p10 -> không sale xịn -> TAM_DUOC.
		{"list_bang_115", 100_000, 115_000, TamDuoc},
		// list*100 > median90*115 và current == median90*0.97 đúng biên -> SALE_AO.
		{"current_bang_97", 97_000, 116_000, SaleAo},
		// current ngay dưới biên 0.97 -> không not_real_discount, lại cao hơn p10 -> TAM_DUOC.
		{"current_duoi_97", 96_999, 116_000, TamDuoc},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DetectFakeSale(hist, c.current, c.list); got != c.want {
				t.Fatalf("%s: muốn %s, được %s", c.name, c.want, got)
			}
		})
	}
}

func TestDetect_SaleXin_BoundaryAtFloor(t *testing.T) {
	// Chuỗi 90 điểm: một đáy duy nhất 50_000, mười hai điểm 78_000, còn lại 100_000.
	// -> trailing_min = 50_000, p10 = 78_000 (đáy thấp hơn p10 để biên 1.02 thành ràng buộc).
	hist := append(append([]int64{50_000}, flat(78_000, 12)...), flat(100_000, 77)...)
	// current == trailing_min*1.02 = 51_000 đúng biên, và 51_000 <= p10 (78_000) -> SALE_XIN.
	if got := DetectFakeSale(hist, 51_000, 100_000); got != SaleXin {
		t.Fatalf("biên 1.02 phải SALE_XIN, được %s", got)
	}
	// Ngay trên biên (51_001) -> vượt trailing_min*1.02 -> không còn SALE_XIN.
	if got := DetectFakeSale(hist, 51_001, 100_000); got == SaleXin {
		t.Fatal("vượt biên 1.02 không được là SALE_XIN")
	}
}

func TestPercentile_NearestRank(t *testing.T) {
	h := []int64{10, 20, 30, 40, 50}
	cases := []struct {
		p    int
		want int64
	}{
		{0, 10}, {10, 10}, {50, 30}, {100, 50},
	}
	for _, c := range cases {
		if got := percentile(h, c.p); got != c.want {
			t.Fatalf("p%d: muốn %d, được %d", c.p, c.want, got)
		}
	}
	// percentile không được sửa slice đầu vào
	if h[0] != 10 || h[4] != 50 {
		t.Fatal("percentile đã sửa slice đầu vào")
	}
}
