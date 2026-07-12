---
id: FR-DEAL-001
title: "Phát hiện sale ảo bằng thống kê - phân loại SALE_AO / SALE_XIN / TAM_DUOC / UNKNOWN từ 90 ngày price_snapshot (median90, p10, trailing_min)"
module: DEAL
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-PRICE-002, FR-DEAL-002, FR-DEAL-003, FR-TRACK-003]
depends_on: [FR-PRICE-002]
blocks: [FR-DEAL-002]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5 (thuật toán phát hiện sale ảo, pseudo-code)"
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.6 (BeeCost cạnh tranh trên sale ảo/sale xịn)"
source_decisions:
  - "DEC-DEAL-01: ngưỡng sale ảo chốt từ §3.5 - list_price > median90*1.15 (giá gốc bị thổi), current_price >= median90*0.97 (không giảm thật), current_price <= trailing_min*1.02 (chạm đáy)"
  - "DEC-DEAL-02: dưới 14 ngày dữ liệu trả UNKNOWN (cold-start), bàn giao cho FR-DEAL-002 - KHÔNG đoán bừa khi thiếu lịch sử"
  - "DEC-DEAL-03: dùng integer-safe math trên BIGINT VND (so current*100 >= median90*97) để tránh sai số float, đồng bộ FR-PRICE-002 DEC-PRICE-05"
  - "DEC-DEAL-04: verdict là enum đóng SALE_AO / SALE_XIN / TAM_DUOC / UNKNOWN - không trả chuỗi tự do"
  - "DEC-DEAL-05: cửa sổ lịch sử cố định 90 ngày, đọc từ price_snapshot (FR-PRICE-002)"

language: "Go 1.22 (deal-svc); reads price_snapshot via price repo"
service: shopass/services/deal/
new_files:
  - services/deal/internal/fakesale/types.go
  - services/deal/internal/fakesale/stats.go
  - services/deal/internal/fakesale/detect.go
  - services/deal/internal/fakesale/detect_test.go
modified_files:
  - services/deal/internal/deal/service.go        # gọi DetectFakeSale sau khi nạp lịch sử
allowed_tools:
  - file_read: services/deal/**
  - file_write: services/deal/**
  - bash: cd services/deal && go test ./...
disallowed_tools:
  - dùng float/float64 cho phép so ngưỡng giá (vi phạm DEC-DEAL-03, sai số trên tiền tệ)
  - trả verdict tự do thay vì 4 hằng enum (vi phạm DEC-DEAL-04)
  - đoán verdict khi dưới 14 ngày dữ liệu thay vì UNKNOWN (vi phạm DEC-DEAL-02)
  - đổi 3 hằng ngưỡng 1.15 / 0.97 / 1.02 mà không cập nhật §3.5 và audit (chúng là hằng load-bearing)

effort_hours: 8
sub_tasks:
  - "0.5h: types.go - type Verdict string + 4 hằng SALE_AO/SALE_XIN/TAM_DUOC/UNKNOWN"
  - "1.5h: stats.go - percentile nearest-rank trên []int64 (sort + chỉ số) + minInt64"
  - "1.5h: detect.go - DetectFakeSale(hist, current, list) với so sánh integer-safe"
  - "0.5h: chốt cold-start len(hist) < 14 -> UNKNOWN, bàn giao FR-DEAL-002"
  - "2.0h: detect_test.go - bảng case cold-start, sale ảo, sale xịn, tạm được, biên ngưỡng"
  - "1.0h: TestPercentile_NearestRank + minInt64 edge (slice 1 phần tử, p0/p100)"
  - "1.0h: nối service.go nạp QueryRange 90d rồi gọi DetectFakeSale, OTel counter fake_sale_verdict_total{verdict}"
risk_if_skipped: "Phát hiện sale ảo là tính năng đinh của SănDeal - lý do người dùng cài thay vì tin nhãn giảm giá của sàn. Sàn TMĐT VN thường thổi giá gốc (list_price) rồi gắn nhãn -50% trong khi giá bán vẫn xấp xỉ giá thường 90 ngày. Không có bộ phát hiện này, sản phẩm chỉ là một công cụ theo dõi giá nữa, không khác BeeCost (§5.6), và mất luôn tín hiệu để FR-TRACK-003 bắn cảnh báo sale thật. Đặc tả sai 3 hằng ngưỡng (1.15 / 0.97 / 1.02) làm phân loại lệch hàng loạt, mất niềm tin người dùng. Dùng float trên VND gây sai số ở đúng các biên so sánh phần trăm, cho kết quả không ổn định giữa các lần chạy trên cùng chuỗi giá."
---

## §1 - Mô tả (BCP-14 normative)

Service DEAL **MUST** cung cấp một hàm thuần phân loại trạng thái giảm giá của một sản phẩm dựa trên 90 ngày lịch sử giá đọc từ `price_snapshot` (FR-PRICE-002). Hàm hiện thực đúng pseudo-code §3.5(1), không thêm bớt nhánh. Hợp đồng:

1. **MUST** nạp lịch sử giá `hist` của `product_id` trong cửa sổ 90 ngày từ `price_snapshot` qua price repo (`QueryRange`), lấy cột `price` thành chuỗi `[]int64` (DEC-DEAL-05).
2. **MUST** nếu `len(hist) < 14` thì trả `UNKNOWN` ngay (cold-start) và KHÔNG tính tiếp - đây là điểm bàn giao cho FR-DEAL-002 (DEC-DEAL-02).
3. **MUST** tính `median90` = percentile thứ 50 của `hist.price` theo phương pháp nearest-rank.
4. **MUST** tính `p10` = percentile thứ 10 của `hist.price` theo cùng phương pháp.
5. **MUST** tính `trailing_min` = giá trị nhỏ nhất trong `hist.price`.
6. **MUST** tính `inflated` = (`list_price > median90 * 1.15`): giá niêm yết bị thổi quá 15 phần trăm trên trung vị 90 ngày (DEC-DEAL-01).
7. **MUST** tính `not_real_discount` = (`current_price >= median90 * 0.97`): giá bán hiện tại vẫn nằm trong 3 phần trăm quanh trung vị, tức không giảm thật (DEC-DEAL-01).
8. **MUST** trả `SALE_AO` khi `inflated AND not_real_discount` (giá gốc bị thổi và giá bán không giảm thật).
9. **MUST** trả `SALE_XIN` khi `current_price <= p10 AND current_price <= trailing_min * 1.02` (chạm vùng 10 phần trăm thấp nhất và sát đáy lịch sử trong biên 2 phần trăm).
10. **MUST** trả `TAM_DUOC` cho mọi trường hợp còn lại (không phải sale ảo, cũng chưa phải sale xịn).
11. **MUST** thực hiện mọi phép so ngưỡng bằng integer-safe math trên `int64` (ví dụ `current*100 >= median90*97`, `list*100 > median90*115`, `current*100 <= trailingMin*102`), KHÔNG dùng float (DEC-DEAL-03), nhất quán với BIGINT VND của FR-PRICE-002.
12. **MUST** là hàm thuần và tất định: cùng một chuỗi `hist` cùng `current`, `list` luôn cho cùng verdict, không phụ thuộc thời gian gọi hay trạng thái ngoài.
13. **MUST** trả kiểu `Verdict` là enum đóng gồm đúng bốn giá trị `SALE_AO`, `SALE_XIN`, `TAM_DUOC`, `UNKNOWN` (DEC-DEAL-04). Verdict này là đầu vào cho luật cảnh báo sale thật của FR-TRACK-003.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đúng ba ngưỡng 1.15 / 0.97 / 1.02 (DEC-DEAL-01)?** Chiêu sale ảo kinh điển trên sàn VN là thổi `list_price` rồi gắn nhãn "-50%", trong khi giá bán thực gần như đứng yên so với 90 ngày. Ngưỡng `list_price > median90 * 1.15` bắt đúng phần "giá gốc bị thổi", còn `current_price >= median90 * 0.97` bắt phần "giảm giả". Hai điều kiện phải xảy ra cùng lúc mới kết luận sale ảo, tránh báo nhầm khi sàn chỉ niêm yết giá gốc cao theo thói quen. Ngưỡng `trailing_min * 1.02` cho sale xịn nói rằng giá hiện tại phải sát đáy lịch sử, không chỉ thấp tương đối. Ba con số này chốt từ §3.5 và là hằng load-bearing: đổi chúng là đổi định nghĩa sản phẩm, phải sửa kèm đặc tả và audit.

**Vì sao dưới 14 ngày trả UNKNOWN thay vì đoán (DEC-DEAL-02)?** Với ít hơn 14 điểm, trung vị và percentile không đại diện cho mặt bằng giá thật, dễ cho kết luận sai và làm người dùng mất tin. Trả `UNKNOWN` là một câu trả lời trung thực, đồng thời là tín hiệu bàn giao rõ ràng để FR-DEAL-002 xử lý cold-start bằng nguồn khác (giá danh mục, sản phẩm tương tự). Đoán bừa lúc thiếu dữ liệu là rủi ro lớn hơn nhiều so với việc thừa nhận chưa đủ cơ sở.

**Vì sao integer-safe math trên tiền tệ (DEC-DEAL-03)?** Giá VN luôn là số nguyên đồng và FR-PRICE-002 (DEC-PRICE-05) lưu `price` dạng BIGINT. Nếu nhân với `0.97` kiểu float, sai số dấu phẩy động xuất hiện đúng tại các biên so sánh phần trăm, làm cùng một chuỗi giá có thể cho verdict khác nhau giữa các kiến trúc hoặc bản dựng. Nhân chéo bằng số nguyên (`current*100 >= median90*97`) cho kết quả chính xác tuyệt đối và tất định, giữ đúng cam kết §1 #12.

**Vì sao dùng trung vị (median90) chứ không phải trung bình?** Trung vị ít nhạy với ngoại lai: một vài lần giá nhảy bất thường (lỗi parse, flash sale chớp nhoáng) không kéo lệch mặt bằng. Trung bình bị các giá trị cực đoan kéo đi, làm ngưỡng `median90 * 1.15` và `median90 * 0.97` mất ý nghĩa. Với chuỗi giá có nhiễu, trung vị là mốc tham chiếu ổn định hơn cho cả hai vế của phép phân loại.

---

## §3 - Hợp đồng API / DDL

### Types (Go)

```go
// services/deal/internal/fakesale/types.go
package fakesale

// Verdict là enum đóng cho kết quả phân loại sale ảo (DEC-DEAL-04).
type Verdict string

const (
    SaleAo   Verdict = "SALE_AO"   // giá gốc bị thổi, giảm giả
    SaleXin  Verdict = "SALE_XIN"  // giảm thật, sát đáy lịch sử
    TamDuoc  Verdict = "TAM_DUOC"  // không ảo, cũng chưa phải sale xịn
    Unknown  Verdict = "UNKNOWN"   // dưới 14 ngày dữ liệu (cold-start)
)

const minHistoryPoints = 14
```

### Helper thống kê (Go)

```go
// services/deal/internal/fakesale/stats.go
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
```

### Hàm phát hiện (Go)

```go
// services/deal/internal/fakesale/detect.go
package fakesale

// DetectFakeSale phân loại trạng thái giảm giá theo §3.5(1).
// Hàm thuần, tất định: cùng (hist, current, list) luôn cho cùng Verdict.
// Mọi so ngưỡng dùng integer-safe math trên int64 (DEC-DEAL-03).
func DetectFakeSale(hist []int64, current, list int64) Verdict {
    if len(hist) < minHistoryPoints {
        return Unknown // cold-start, bàn giao FR-DEAL-002
    }
    median90 := percentile(hist, 50)
    p10 := percentile(hist, 10)
    trailingMin := minInt64(hist)

    // inflated: list_price > median90 * 1.15  ->  list*100 > median90*115
    inflated := list*100 > median90*115
    // not_real_discount: current_price >= median90 * 0.97  ->  current*100 >= median90*97
    notRealDiscount := current*100 >= median90*97

    if inflated && notRealDiscount {
        return SaleAo
    }
    // SALE_XIN: current <= p10 AND current <= trailing_min * 1.02
    if current <= p10 && current*100 <= trailingMin*102 {
        return SaleXin
    }
    return TamDuoc
}
```

---

## §4 - Acceptance criteria

1. `DetectFakeSale` nạp được chuỗi `price` 90 ngày từ `QueryRange` của price repo (kiểm qua mock repo trong service test).
2. `len(hist) < 14` -> trả `UNKNOWN`, không tính median/p10/min.
3. `median90` bằng đúng percentile thứ 50 nearest-rank của chuỗi đầu vào.
4. `p10` bằng đúng percentile thứ 10 nearest-rank của chuỗi đầu vào.
5. `trailing_min` bằng đúng `min` của chuỗi đầu vào.
6. `inflated` đúng true khi và chỉ khi `list*100 > median90*115`.
7. `not_real_discount` đúng true khi và chỉ khi `current*100 >= median90*97`.
8. `inflated AND not_real_discount` -> trả `SALE_AO`.
9. `current <= p10 AND current*100 <= trailingMin*102` -> trả `SALE_XIN`.
10. Mọi trường hợp còn lại -> trả `TAM_DUOC`.
11. Không có float trong đường tính ngưỡng (kiểm qua đọc mã: chỉ phép `int64`); cùng chuỗi cho verdict ổn định qua nhiều lần chạy.
12. Hàm tất định: gọi hai lần cùng đầu vào cho cùng kết quả; không sửa slice `hist` đầu vào.
13. Kiểu trả về là `Verdict` và chỉ nhận một trong bốn hằng đã định nghĩa.

---

## §5 - Kiểm thử (verification)

```go
// services/deal/internal/fakesale/detect_test.go
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
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `types.go` (enum Verdict) -> `stats.go` (percentile + minInt64) -> `detect.go` (DetectFakeSale) -> `detect_test.go`. Sau khi gói `fakesale` xanh, nối vào `service.go`: nạp `QueryRange(productID, now-90d, now)` từ price repo (FR-PRICE-002), rút cột `price` thành `[]int64`, gọi `DetectFakeSale`, phát OTel counter `fake_sale_verdict_total{verdict}`. Gói `fakesale` không phụ thuộc DB - chỉ là hàm thuần trên chuỗi - nên test chạy nhanh, không cần TimescaleDB.

---

## §7 - Phụ thuộc

- **FR-PRICE-002** - nguồn 90 ngày lịch sử giá (`price_snapshot.price`, BIGINT VND) qua `QueryRange`. Hằng integer-safe math đồng bộ DEC-PRICE-05.
- **FR-DEAL-002 (downstream, blocks)** - mở rộng nhánh cold-start: khi `DetectFakeSale` trả `UNKNOWN`, FR-DEAL-002 dùng nguồn thay thế.
- **FR-TRACK-003 (consumer)** - luật cảnh báo `real_sale` tiêu thụ verdict (`SALE_XIN` kích cảnh báo sale thật).
- **FR-DEAL-003 (downstream)** - hiển thị nhãn verdict kèm biểu đồ giá.
- Thư viện: chỉ `sort` chuẩn của Go; không phụ thuộc ngoài.

---

## §8 - Payload ví dụ

### Ví dụ SALE_AO (giá gốc bị thổi)

Lịch sử 90 điểm với mặt bằng quanh `100000` đồng nên `median90 ~ 100000`. Sản phẩm gắn nhãn giảm sâu:

- `list_price = 149000` -> `149000 * 100 = 14900000 > 100000 * 115 = 11500000` => `inflated = true` (giá gốc thổi hơn 15 phần trăm).
- `current_price = 120000` -> `120000 * 100 = 12000000 >= 100000 * 97 = 9700000` => `not_real_discount = true` (giá bán vẫn trên mặt bằng).
- `inflated AND not_real_discount` => verdict = `SALE_AO`.

### Ví dụ SALE_XIN (giảm thật, sát đáy)

Lịch sử 90 điểm chủ yếu quanh `100000` nhưng có một đợt chạm `80000`, nên `trailing_min = 80000` và `p10 = 80000`:

- `current_price = 80000` -> `80000 <= p10 (80000)` đúng, và `80000 * 100 = 8000000 <= 80000 * 102 = 8160000` đúng (sát đáy trong biên 2 phần trăm).
- => verdict = `SALE_XIN`. FR-TRACK-003 bắn cảnh báo sale thật.

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Ngưỡng theo từng ngành hàng (thời trang biến động khác điện máy) - hoãn sang FR-DEAL-004/005 (ML), giai đoạn này dùng ba hằng cố định toàn cục.
- Trọng số theo thời gian (giá gần đây quan trọng hơn) - cân nhắc weighted percentile ở slice sau.
- Phát hiện chuỗi giá có khoảng trống dài (sản phẩm ngừng bán rồi bán lại) - xử lý ở tầng nạp dữ liệu trước khi gọi hàm.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Lịch sử có khoảng trống (gap nhiều ngày) | đếm điểm khi nạp | percentile lệch theo mật độ điểm | Chuẩn hóa/lấy mẫu đều ở tầng nạp trước khi gọi hàm |
| `len(hist) < 14` | kiểm tra đầu hàm | thiếu cơ sở thống kê | Trả UNKNOWN, bàn giao FR-DEAL-002 (theo thiết kế) |
| `list_price` null hoặc 0 | tầng service trước khi gọi | `inflated` tính sai (so với 0) | Bỏ qua nhánh inflated hoặc đặt list = current ở tầng gọi; ghi rõ ở service.go |
| Dùng float làm tròn sai ở biên | đọc mã + test biên | verdict không ổn định | Bắt buộc integer-safe math (DEC-DEAL-03), test `current_bang_97` |
| Tràn số khi nhân `*115` / `*102` | giá tối đa << int64 | tràn lý thuyết | Giá VND thực tế << 9,2e18 nên an toàn; ghi chú giới hạn |
| Ngưỡng cần tinh chỉnh theo ngành hàng | phản hồi người dùng | phân loại lệch một số ngành | Hoãn sang ML FR-DEAL-004/005, giữ ba hằng cố định ở v1 |
| Chuỗi toàn giá bằng nhau (phẳng) | test `flat` | p10 = median90 = min, dễ ra SALE_XIN giả | Logic đúng theo §3.5; tài liệu hóa hành vi chuỗi phẳng |
| Slice `hist` bị sửa ngoài ý muốn | test percentile | nhiễm bẩn caller | `percentile` copy trước khi sort (hàm thuần) |

---

## §11 - Ghi chú

- `DetectFakeSale` là tính năng đinh của SănDeal - phân biệt với một công cụ theo dõi giá thuần (so §5.6 BeeCost).
- Ba hằng `1.15 / 0.97 / 1.02` là load-bearing: đổi chúng là đổi định nghĩa sản phẩm, phải sửa kèm §3.5 và audit.
- Integer-safe math trên BIGINT VND đảm bảo verdict tất định, đồng bộ DEC-PRICE-05 của FR-PRICE-002.
- Trung vị ít nhạy với ngoại lai hơn trung bình, nên median90 là mốc tham chiếu ổn định cho chuỗi giá có nhiễu.
- UNKNOWN không phải lỗi mà là câu trả lời trung thực khi thiếu dữ liệu, đồng thời là điểm bàn giao sạch cho FR-DEAL-002.
- Gói `fakesale` cố ý không chạm DB: hàm thuần trên `[]int64`, dễ test và tái dùng cho mô phỏng/backtest sau này.

---

*Hết FR-DEAL-001. Status: ready_to_implement (mục tiêu audit 10/10).*
