---
id: TASK-B2B-002
title: "Báo cáo B2B insights + subscription bán cho brand/seller - sinh report xu hướng giá category/thị trường từ market_trend_daily, gating theo gói b2b_subscription, render JSON + export, đọc DUY NHẤT dữ liệu đã ẩn danh"
module: B2B
priority: SHOULD
status: done
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-B2B-001, TASK-BILL-001, TASK-B2B-003, TASK-B2B-004]
depends_on: [TASK-B2B-001, TASK-BILL-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 mục 7 (B2B data/insights - báo cáo xu hướng giá/thị trường ẩn danh bán cho brand/seller, B2B subscription margin cao)"
  - "docs/... §1 (dòng doanh thu B2B margin cao), §4.1 (unit economics), §5.5 (PDPL)"
source_decisions:
  - "DEC-B2B-10: báo cáo B2B đọc DUY NHẤT market_trend_daily qua TASK-B2B-001 (chỉ ô suppressed=false) - KHÔNG bao giờ chạm price_snapshot, price_daily raw, hay bảng user-level"
  - "DEC-B2B-11: truy cập báo cáo gating qua bảng b2b_subscription (tier basic/pro/enterprise) - tier quyết định số category, độ sâu lịch sử, và quyền export"
  - "DEC-B2B-12: một báo cáo là tổ hợp (category_id[], platform_id[], khoảng ngày) -> chuỗi ô market_trend_daily; nếu mọi ô trong phạm vi bị suppress thì trả báo cáo rỗng-có-lý-do, KHÔNG nội suy"
  - "DEC-B2B-13: report scope bị kẹp theo entitlement của tier (max_categories, history_days); request vượt entitlement trả 403 với thông điệp nâng cấp, KHÔNG cắt im lặng"
  - "DEC-B2B-14: export (CSV/JSON) chỉ cho tier có quyền export; mọi dòng export là chỉ số tổng hợp đã qua k-anonymity, không khóa định danh"
  - "DEC-B2B-15: báo cáo có cached_at + tham chiếu computed_at của ô nguồn để khách hàng B2B biết độ tươi dữ liệu"

language: "PostgreSQL 16 + Go 1.22 (b2b-svc)"
service: shopass/services/b2b/
new_files:
  - services/b2b/migrations/0002_b2b_subscription.sql
  - services/b2b/internal/report/builder.go
  - services/b2b/internal/report/entitlement.go
  - services/b2b/internal/report/export.go
  - services/b2b/internal/api/report_handler.go
  - services/b2b/internal/report/builder_test.go
  - services/b2b/internal/report/entitlement_test.go
modified_files:
  - services/b2b/internal/api/router.go            # đăng ký route GET /v1/b2b/reports
allowed_tools:
  - file_read: services/b2b/**
  - file_write: services/b2b/**
  - bash: cd services/b2b && go test ./...
disallowed_tools:
  - đọc nguồn ngoài market_trend_daily (price_snapshot/price_daily raw/bảng user-level) cho báo cáo B2B (vi phạm DEC-B2B-10)
  - phục vụ báo cáo cho subscription không active hoặc vượt entitlement mà không trả 403 (vi phạm DEC-B2B-11/13)
  - nội suy hay bịa số cho ô bị suppress (vi phạm DEC-B2B-12)

effort_hours: 8
sub_tasks:
  - "1.0h: 0002_b2b_subscription.sql - bảng b2b_subscription + tier + entitlement (max_categories, history_days, can_export) + status"
  - "1.5h: entitlement.go - nạp entitlement theo tier, kẹp scope, trả 403 khi vượt"
  - "1.5h: builder.go - dựng báo cáo từ chuỗi ô market_trend_daily (chỉ suppressed=false), gắn cached_at + computed_at"
  - "1.0h: export.go - CSV/JSON cho tier can_export; mọi dòng là chỉ số tổng hợp"
  - "1.0h: report_handler.go - HTTP: xác thực subscription active, gọi entitlement + builder, 200/402/403"
  - "1.5h: entitlement_test.go - active/inactive, vượt max_categories -> 403, history vượt -> 403"
  - "1.0h: builder_test.go - mọi ô suppress -> báo cáo rỗng-có-lý-do; ô phát hành -> chuỗi đúng"
  - "0.5h: OTel metric b2b_report_served_total{tier} + b2b_report_denied_total{reason}"

risk_if_skipped: "TASK-B2B-002 là nơi dòng doanh thu B2B thực sự thu được tiền - báo cáo insights là sản phẩm khách hàng brand/seller trả tiền (margin cao, §1). Không có nó thì pipeline ẩn danh TASK-B2B-001 chỉ là dữ liệu nằm im, không thành doanh thu. Hai rủi ro chí mạng: (1) nếu báo cáo đọc nguồn ngoài market_trend_daily thì có thể lọt giá SKU/shop đơn lẻ ra ngoài, phá k-anonymity đã dựng ở TASK-B2B-001 và vi phạm PDPL (§5.5); (2) nếu gating sai - phục vụ báo cáo cho subscription hết hạn hoặc vượt entitlement mà cắt im lặng - thì vừa thất thu vừa làm khách hàng tier thấp lấy được giá trị của tier cao, hỏng mô hình định giá. Việc trả 403 rõ ràng thay vì cắt im lặng cũng là tín hiệu nâng cấp đúng lúc."
---

## §1 - Mô tả (BCP-14 normative)

Service B2B **MUST** sinh báo cáo insights xu hướng giá/thị trường từ `market_trend_daily`, gating truy cập theo bảng `b2b_subscription`, và đọc DUY NHẤT dữ liệu đã ẩn danh (ô `suppressed=false`). Hợp đồng:

1. **MUST** đọc dữ liệu báo cáo DUY NHẤT từ `market_trend_daily` qua repo của TASK-B2B-001, chỉ ô `suppressed=false` (DEC-B2B-10). **MUST NOT** chạm `price_snapshot`, `price_daily` raw, hay bất kỳ bảng user-level nào.
2. **MUST** định nghĩa bảng `b2b_subscription (id, org_name, tier, max_categories, history_days, can_export, status, started_at, expires_at)`; `tier` thuộc `{basic, pro, enterprise}` (DEC-B2B-11).
3. **MUST** chỉ phục vụ báo cáo khi subscription `status = 'active'` và `expires_at > now()`. Subscription không active hoặc hết hạn -> trả `402` (cần thanh toán/gia hạn).
4. **MUST** mô tả một báo cáo là tổ hợp `(category_id[], platform_id[], from, to)` ánh xạ sang chuỗi ô `market_trend_daily` (DEC-B2B-12). Nếu MỌI ô trong phạm vi bị suppress -> trả báo cáo rỗng kèm lý do (`reason = "insufficient_data"`); **MUST NOT** nội suy hay bịa số.
5. **MUST** kẹp scope theo entitlement của tier (DEC-B2B-13): số `category_id` không vượt `max_categories`; khoảng `(to - from)` không vượt `history_days`. Request vượt entitlement -> trả `403` với thông điệp nâng cấp; **MUST NOT** cắt im lặng rồi trả phần dữ liệu.
6. **MUST** chỉ cho export (CSV/JSON) khi `can_export = true` (DEC-B2B-14). Mọi dòng export **MUST** là chỉ số tổng hợp đã qua k-anonymity, không chứa khóa định danh.
7. **MUST** gắn vào báo cáo `cached_at` (lúc dựng) và `source_computed_at` (max `computed_at` của ô nguồn) để khách hàng B2B biết độ tươi (DEC-B2B-15).
8. **MUST** expose endpoint `GET /v1/b2b/reports?categories=...&platforms=...&from=...&to=...` (đọc) và `GET /v1/b2b/reports/export?...` (export, gating `can_export`).
9. **MUST** phân biệt mã trạng thái: `200` (có dữ liệu hoặc rỗng-có-lý-do), `402` (subscription không active), `403` (vượt entitlement), `400` (tham số sai).
10. **SHOULD** phát OTel metric: `b2b_report_served_total{tier}` (counter), `b2b_report_denied_total{reason}` (counter), `b2b_report_build_ms` (histogram).
11. **MUST** resolve entitlement từ chính `b2b_subscription` của caller; **MUST NOT** tin tham số tier/quyền do client gửi.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao chỉ đọc market_trend_daily (DEC-B2B-10)? TASK-B2B-001 đã dựng k-anonymity ở tầng pipeline. Nếu báo cáo lại đi đường vòng đọc raw để "làm giàu" thì phá tan công sức đó và mở lại cửa rò rỉ. Quy tắc đơn giản: tầng báo cáo chỉ tiêu thụ ô đã phát hành, không bao giờ chạm nguồn chi tiết. Mọi thứ khách hàng B2B thấy đều đã qua cổng ẩn danh.

Vì sao gating theo b2b_subscription server-side (DEC-B2B-11, §1 #11)? Quyền truy cập là tài sản doanh thu. Nếu tin tham số tier do client gửi thì ai cũng tự xưng enterprise. Entitlement phải nạp từ bản ghi subscription của chính caller trong DB. Tier quyết định ba trục giá trị: bao nhiêu category, lịch sử sâu bao nhiêu, có export không.

Vì sao trả 403 thay vì cắt im lặng (DEC-B2B-13, §1 #5)? Nếu khách tier basic xin 100 category mà ta lặng lẽ trả 5 category đầu thì họ tưởng dữ liệu thiếu, còn ta thì vừa hỏng trải nghiệm vừa mất cơ hội bán nâng cấp. Trả 403 rõ ràng "vượt gói, nâng cấp để xem thêm" biến giới hạn thành điểm chạm bán hàng.

Vì sao rỗng-có-lý-do thay vì nội suy (DEC-B2B-12, §1 #4)? Khi mọi ô trong phạm vi bị suppress (category hẹp, ngày thưa), cám dỗ là nội suy cho báo cáo "đẹp". Nhưng nội suy số ẩn danh là bịa - vừa sai vừa có thể vô tình tái dựng tín hiệu của một cá thể. Trung thực trả rỗng kèm lý do; khách hàng B2B cần số đúng, không cần số đầy.

Vì sao gắn độ tươi (DEC-B2B-15, §1 #7)? Báo cáo dựa trên job đêm (TASK-B2B-001) nên dữ liệu trễ tới một ngày. Khách hàng B2B ra quyết định giá phải biết số này tính tới khi nào. `source_computed_at` nói rõ thay vì để họ đoán.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/b2b/migrations/0002_b2b_subscription.sql
CREATE TABLE b2b_subscription (
  id             BIGSERIAL   PRIMARY KEY,
  org_name       TEXT        NOT NULL,
  tier           TEXT        NOT NULL CHECK (tier IN ('basic','pro','enterprise')),
  max_categories INTEGER     NOT NULL CHECK (max_categories > 0),
  history_days   INTEGER     NOT NULL CHECK (history_days > 0),
  can_export     BOOLEAN     NOT NULL DEFAULT false,
  status         TEXT        NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active','past_due','canceled')),
  started_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at     TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_b2bsub_active ON b2b_subscription (id)
  WHERE status = 'active';
```

### Types + entitlement (Go)

```go
// services/b2b/internal/report/entitlement.go
type Entitlement struct {
    Tier          string
    MaxCategories int
    HistoryDays   int
    CanExport     bool
}

type ReportScope struct {
    CategoryIDs []int64
    PlatformIDs []int16
    From, To    time.Time
}

// checkScope kẹp scope theo entitlement; trả lỗi 403 khi vượt (KHÔNG cắt im lặng).
func checkScope(e Entitlement, s ReportScope) error {
    if len(s.CategoryIDs) > e.MaxCategories {
        return ErrScopeExceeded{Field: "categories", Limit: e.MaxCategories}
    }
    if s.To.Sub(s.From) > time.Duration(e.HistoryDays)*24*time.Hour {
        return ErrScopeExceeded{Field: "history_days", Limit: e.HistoryDays}
    }
    return nil
}
```

### Builder (§1 #4, #7)

```go
// services/b2b/internal/report/builder.go
type Report struct {
    Scope           ReportScope     `json:"scope"`
    Cells           []TrendPoint    `json:"cells"`
    Reason          string          `json:"reason,omitempty"`     // "insufficient_data" khi rỗng
    CachedAt        time.Time       `json:"cached_at"`
    SourceComputedAt time.Time      `json:"source_computed_at"`
}

func (b *Builder) Build(ctx context.Context, s ReportScope) (Report, error) {
    var pts []TrendPoint
    var maxComputed time.Time
    for _, cat := range s.CategoryIDs {
        for _, pf := range s.PlatformIDs {
            // chỉ ô suppressed=false (QueryCells lọc mặc định)
            cells, err := b.trend.QueryCells(ctx, cat, pf, s.From, s.To)
            if err != nil {
                return Report{}, err
            }
            for _, c := range cells {
                pts = append(pts, toPoint(c))
                if c.ComputedAt.After(maxComputed) {
                    maxComputed = c.ComputedAt
                }
            }
        }
    }
    r := Report{Scope: s, Cells: pts, CachedAt: time.Now(), SourceComputedAt: maxComputed}
    if len(pts) == 0 {
        r.Reason = "insufficient_data" // mọi ô bị suppress - KHÔNG nội suy
    }
    return r, nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `b2b_subscription` tồn tại với CHECK `tier`/`status`.
2. Subscription `active` + chưa hết hạn, scope trong gói -> `200` + chuỗi ô đúng.
3. Subscription `status='past_due'` hoặc `expires_at < now()` -> `402`, KHÔNG trả dữ liệu.
4. Request có số category vượt `max_categories` -> `403` với thông điệp nâng cấp.
5. Request có khoảng ngày vượt `history_days` -> `403`.
6. Mọi ô trong phạm vi bị suppress (không có ô phát hành nào) -> `200` + `cells=[]` + `reason="insufficient_data"`; KHÔNG có số nội suy.
7. Export khi `can_export=false` -> `403`; khi `can_export=true` -> file CSV/JSON các dòng chỉ số tổng hợp.
8. Mọi dòng export KHÔNG chứa `product_id/shop_id/user_id`.
9. Báo cáo có `cached_at` và `source_computed_at` đúng (= max `computed_at` của ô nguồn).
10. Tham số tier/quyền gửi từ client bị bỏ qua; entitlement nạp từ DB của caller.
11. Builder gọi `QueryCells` (vốn lọc suppressed) - không có truy vấn nào chạm nguồn raw (review/test grep).

---

## §5 - Kiểm thử (verification)

```go
// services/b2b/internal/report/entitlement_test.go
func TestScope_WithinTier_OK(t *testing.T) {
    e := Entitlement{Tier: "pro", MaxCategories: 10, HistoryDays: 180, CanExport: true}
    s := ReportScope{CategoryIDs: ids(5), From: dNow.AddDate(0, 0, -90), To: dNow}
    require.NoError(t, checkScope(e, s))
}

func TestScope_TooManyCategories_403(t *testing.T) {
    e := Entitlement{Tier: "basic", MaxCategories: 3, HistoryDays: 30}
    s := ReportScope{CategoryIDs: ids(4), From: dNow.AddDate(0, 0, -7), To: dNow}
    err := checkScope(e, s)
    require.ErrorAs(t, err, &ErrScopeExceeded{})
}

func TestScope_HistoryTooDeep_403(t *testing.T) {
    e := Entitlement{Tier: "basic", MaxCategories: 3, HistoryDays: 30}
    s := ReportScope{CategoryIDs: ids(1), From: dNow.AddDate(0, 0, -90), To: dNow}
    err := checkScope(e, s)
    require.ErrorAs(t, err, &ErrScopeExceeded{})
}

func TestReport_InactiveSubscription_402(t *testing.T) {
    h := setupHandler(t, withSub("past_due"))
    res := h.get("/v1/b2b/reports?categories=7&platforms=1&from=...&to=...")
    require.Equal(t, 402, res.Code)
}

// services/b2b/internal/report/builder_test.go
func TestBuild_AllSuppressed_EmptyWithReason(t *testing.T) {
    b := setupBuilder(t, withTrendCells(/* mọi ô suppressed=true */))
    r, _ := b.Build(ctx, ReportScope{CategoryIDs: []int64{991}, PlatformIDs: []int16{1},
        From: d0, To: d0.AddDate(0, 0, 7)})
    require.Empty(t, r.Cells)
    require.Equal(t, "insufficient_data", r.Reason)
}

func TestBuild_PublishedCells_Series(t *testing.T) {
    b := setupBuilder(t, withTrendCells(/* category 7, platform 1, 7 ô phát hành */))
    r, _ := b.Build(ctx, ReportScope{CategoryIDs: []int64{7}, PlatformIDs: []int16{1},
        From: d0, To: d0.AddDate(0, 0, 7)})
    require.Len(t, r.Cells, 7)
    require.False(t, r.SourceComputedAt.IsZero())
}

func TestExport_NoIdentifierColumns(t *testing.T) {
    csv := exportCSV(t, /* báo cáo có dữ liệu */)
    require.NotContains(t, csv, "product_id")
    require.NotContains(t, csv, "shop_id")
    require.NotContains(t, csv, "user_id")
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0002 (b2b_subscription) -> entitlement.go (nạp + kẹp scope) -> builder.go (dựng từ ô đã phát hành) -> export.go -> report_handler.go -> tests. Handler chạy sau JWT/gateway (TASK-INFRA-001) cho xác thực; resolve subscription của org từ claim. Báo cáo có thể cache theo `(scope hash, ngày)` vì market_trend_daily chỉ đổi sau job đêm.

---

## §7 - Phụ thuộc

- TASK-B2B-001 - `market_trend_daily` + `QueryCells` (lọc suppressed) là nguồn dữ liệu duy nhất.
- TASK-BILL-001 - vòng đời subscription/tier; `b2b_subscription` đồng bộ trạng thái thanh toán B2B với khung billing chung.
- TASK-INFRA-001 (gateway) - xác thực + rate-limit trước handler.
- Extension/lib: driver `pgx`; thư viện CSV chuẩn thư viện Go.

---

## §8 - Payload ví dụ

### Báo cáo có dữ liệu (200)

```json
{
  "scope": { "category_ids": [7], "platform_ids": [1], "from": "2026-05-22", "to": "2026-06-20" },
  "cells": [
    { "category_id": 7, "platform_id": 1, "day": "2026-06-19", "median_p": 318000, "p25_p": 249000, "p75_p": 405000, "avg_discount_pct": 13.80 },
    { "category_id": 7, "platform_id": 1, "day": "2026-06-20", "median_p": 320000, "p25_p": 250000, "p75_p": 410000, "avg_discount_pct": 14.30 }
  ],
  "cached_at": "2026-06-21T02:15:00Z",
  "source_computed_at": "2026-06-21T01:40:00Z"
}
```

### Vượt entitlement (403)

```json
{ "error": "scope_exceeded", "field": "categories", "limit": 3, "message": "Gói basic xem tối đa 3 category. Nâng cấp lên pro để xem thêm." }
```

### Rỗng-có-lý-do (200)

```json
{ "scope": { "category_ids": [991], "platform_ids": [1], "from": "2026-06-13", "to": "2026-06-20" }, "cells": [], "reason": "insufficient_data", "cached_at": "2026-06-21T02:15:00Z", "source_computed_at": null }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Báo cáo so sánh chéo category (vd điện tử vs mỹ phẩm) - thêm khi có nhu cầu khách hàng cụ thể.
- Lập lịch gửi báo cáo định kỳ qua email (digest tuần) - tái dùng TASK-NOTIF-006 ở slice sau.
- Định giá theo lượt gọi (metered) thay vì tier cố định - cân nhắc khi đã có dữ liệu sử dụng thực.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Báo cáo đọc nguồn raw | review + test grep | rò rỉ giá SKU/shop, vỡ k-anonymity | Chỉ gọi QueryCells của TASK-B2B-001 (DEC-B2B-10) |
| Phục vụ subscription hết hạn | TestReport_InactiveSubscription_402 | thất thu | Kiểm status='active' + expires_at > now() |
| Cắt scope im lặng | entitlement_test | trải nghiệm sai + mất bán nâng cấp | Trả 403 rõ ràng (DEC-B2B-13) |
| Nội suy ô suppress | TestBuild_AllSuppressed | bịa số, rủi ro tái dựng | Rỗng-có-lý-do, KHÔNG nội suy |
| Tin tier do client gửi | review handler | leo thang quyền | Entitlement nạp từ DB caller (§1 #11) |
| Export lọt khóa định danh | TestExport_NoIdentifierColumns | tái định danh | Export chỉ chỉ số tổng hợp |
| Dữ liệu cũ không rõ độ tươi | kiểm field source_computed_at | quyết định trên số cũ | Gắn cached_at + source_computed_at |
| Export cho tier không có quyền | test can_export | rò rỉ giá trị tier cao | Gating can_export=true |

---

## §11 - Ghi chú

- TASK-B2B-002 là điểm thu tiền của dòng B2B: pipeline ẩn danh (TASK-B2B-001) thành doanh thu qua báo cáo có gating.
- Quy tắc một câu: tầng báo cáo chỉ đọc ô đã phát hành của market_trend_daily, không bao giờ chạm nguồn chi tiết.
- Giới hạn entitlement được trình bày như điểm chạm bán hàng (403 + thông điệp nâng cấp), không phải lỗi câm.
- Trung thực về độ phủ dữ liệu (rỗng-có-lý-do) quan trọng hơn báo cáo trông đầy đủ - khách hàng B2B trả tiền cho số đúng.
- Độ tươi dữ liệu công khai qua source_computed_at để khách hàng biết báo cáo dựa trên job đêm nào.

---

*Hết TASK-B2B-002. Status: ready_to_implement (mục tiêu audit 10/10).*
