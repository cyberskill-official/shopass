---
id: TASK-COMPLY-008
title: "Tuân thủ luật TMĐT VN - Nghị định 52/2013 + 85/2021 (đăng ký MOIT, trách nhiệm sàn, ngưỡng >100.000 giao dịch/năm cho foreign platform); dự thảo Luật TMĐT 2025 (livestream + affiliate)"
module: COMPLY
priority: SHOULD
status: done
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-COMPLY-001, TASK-AFFIL-002, TASK-AFFIL-004, TASK-COMPLY-006]
depends_on: [TASK-COMPLY-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (TMĐT VN: NĐ 52/2013 + 85/2021, đăng ký MOIT, trách nhiệm sàn, ngưỡng >100.000 giao dịch/năm foreign platform; dự thảo Luật TMĐT 2025 quản livestream + affiliate)"
  - "docs/... §5.6 (cạnh tranh, mô hình sàn/extension), §4.2 (affiliate compliant)"
source_decisions:
  - "DEC-COMPLY-31: sổ đăng ký nghĩa vụ TMĐT (ecommerce_obligation) theo dõi trạng thái đăng ký MOIT + ngưỡng giao dịch + mốc dự thảo 2025"
  - "DEC-COMPLY-32: bộ đếm giao dịch/năm so ngưỡng 100.000 (foreign platform); vượt -> cờ phải đăng ký, suy tự động không nhập tay"
  - "DEC-COMPLY-33: disclosure affiliate/livestream tuân dự thảo Luật TMĐT 2025; checklist nghĩa vụ versioned khi luật chốt"
  - "DEC-COMPLY-34: SHOULD - khung sẵn sàng; giá trị/ngưỡng chính xác chốt khi dự thảo 2025 ban hành (§10 tài liệu nguồn)"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0008_ecommerce_obligation.sql
  - services/comply/internal/ecom/obligation.go
  - services/comply/internal/ecom/threshold.go
  - services/comply/internal/ecom/repo.go
  - services/comply/internal/ecom/obligation_test.go
  - services/comply/internal/ecom/threshold_test.go
modified_files:
  - services/comply/internal/ecom/types.go      # struct EcommerceObligation, ThresholdState
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - nhập trạng thái "đã vượt ngưỡng" bằng tay thay vì suy từ bộ đếm (vi phạm DEC-COMPLY-32)
  - bỏ disclosure affiliate/livestream khi luật yêu cầu (vi phạm DEC-COMPLY-33)
  - hardcode ngưỡng 100.000 rải rác thay vì cấu hình versioned (vi phạm DEC-COMPLY-34)

effort_hours: 6
sub_tasks:
  - "0.5h: 0008_ecommerce_obligation.sql - bảng nghĩa vụ + trạng thái đăng ký + ngưỡng"
  - "1.0h: types.go + obligation.go - model nghĩa vụ, checklist versioned"
  - "1.0h: threshold.go - bộ đếm giao dịch/năm so ngưỡng, suy cờ phải đăng ký"
  - "1.0h: repo.go - lưu/đọc, liệt kê nghĩa vụ quá hạn/chưa hoàn tất"
  - "1.5h: obligation_test.go - đăng ký MOIT trạng thái; disclosure bắt buộc khi áp dụng"
  - "1.0h: threshold_test.go - dưới ngưỡng -> không cờ; vượt 100.000 -> cờ phải đăng ký"

risk_if_skipped: "Luật TMĐT VN (NĐ 52/2013 + 85/2021) yêu cầu đăng ký MOIT, quy định trách nhiệm sàn, và đặt ngưỡng >100.000 giao dịch/năm cho foreign platform (§5.5). Dự thảo Luật TMĐT 2025 lần đầu quản livestream và affiliate marketing - đúng hai hoạt động SănDeal chạm tới. Không theo dõi nghĩa vụ đăng ký + ngưỡng giao dịch -> có thể vận hành mà thiếu đăng ký bắt buộc, rủi ro pháp lý. SHOULD vì phần lớn nghĩa vụ kích hoạt theo quy mô (ngưỡng) và theo thời điểm luật chốt; khung sẵn sàng để bật đúng lúc. Phối hợp với TASK-AFFIL-002/004 (disclosure affiliate) để đáp ứng dự thảo 2025."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **SHOULD** duy trì sổ theo dõi nghĩa vụ luật TMĐT VN: trạng thái đăng ký MOIT, bộ đếm giao dịch so ngưỡng foreign-platform, và checklist disclosure affiliate/livestream theo dự thảo Luật TMĐT 2025. Hợp đồng:

1. **SHOULD** định nghĩa bảng `ecommerce_obligation (id, obligation_key, description_vi, status, due_at, completed_at, source_law, version, created_at)` theo dõi từng nghĩa vụ (đăng ký MOIT, disclosure affiliate, disclosure livestream...).
2. **SHOULD** định nghĩa bảng/bộ đếm giao dịch theo năm để so ngưỡng (DEC-COMPLY-32): `yearly_transaction_count (year, count)` hoặc nguồn tổng hợp; ngưỡng `100000` lưu dạng cấu hình versioned, KHÔNG hardcode rải rác.
3. **SHOULD** suy cờ `must_register_threshold` tự động từ bộ đếm (DEC-COMPLY-32): khi `count(year) > threshold` -> cờ bật; KHÔNG nhập tay.
4. **SHOULD** theo dõi trạng thái đăng ký MOIT qua `obligation_key = 'moit_registration'` với `status IN ('not_started','submitted','approved')`.
5. **SHOULD** quản checklist disclosure theo dự thảo Luật TMĐT 2025 (DEC-COMPLY-33): nghĩa vụ `affiliate_disclosure` và `livestream_disclosure` với trạng thái và nguồn luật; phối hợp TASK-AFFIL-002/004.
6. **SHOULD** lưu `source_law` cho mỗi nghĩa vụ (NĐ 52/2013, NĐ 85/2021, hoặc dự thảo 2025) để truy về căn cứ pháp lý.
7. **SHOULD** versioned checklist (DEC-COMPLY-33): khi dự thảo 2025 ban hành chính thức, tạo version mới của nghĩa vụ với giá trị/ngưỡng đã chốt; bản cũ giữ lịch sử.
8. **SHOULD** expose hàm:
    - `Threshold(ctx, year int) (ThresholdState, error)` - trả count + ngưỡng + cờ `must_register`.
    - `Obligations(ctx) ([]EcommerceObligation, error)` - danh sách nghĩa vụ hiện hành + trạng thái.
    - `MarkObligation(ctx, key string, status string) error` - cập nhật trạng thái (ví dụ MOIT approved).
    - `Outstanding(ctx) ([]EcommerceObligation, error)` - nghĩa vụ chưa hoàn tất/quá hạn.
9. **SHOULD** validate `status` và `obligation_key` thuộc tập đã biết; giá trị lạ -> lỗi xác định.
10. **MAY** phát OTel metric `ecom_obligation_outstanding_total`, `ecom_threshold_exceeded` (cờ) để báo động khi vượt ngưỡng hoặc nghĩa vụ quá hạn.
11. **SHOULD** đảm bảo ngưỡng 100.000 đến từ cấu hình versioned (một chỗ): đổi ngưỡng (nếu luật đổi) tạo version mới, không sửa hằng số rải rác.
12. **SHOULD** đánh dấu rõ giá trị tạm thời cần xác minh: phần thuộc dự thảo 2025 ghi chú "chờ luật chốt" cho tới khi ban hành (DEC-COMPLY-34, §10 tài liệu nguồn).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao cờ ngưỡng suy tự động (DEC-COMPLY-32)?** Ngưỡng 100.000 giao dịch/năm cho foreign platform là một con số khách quan. Nếu trạng thái "đã vượt ngưỡng" nhập tay, dễ quên hoặc nhập trễ. Tính cờ `must_register` trực tiếp từ bộ đếm giao dịch năm cho ra cảnh báo đúng lúc vượt - biến nghĩa vụ đăng ký thành tín hiệu tự động thay vì việc phải nhớ.

**Vì sao checklist versioned (DEC-COMPLY-33)?** Dự thảo Luật TMĐT 2025 chưa ban hành chính thức; nội dung và ngưỡng có thể đổi khi chốt. Versioned cho phép giữ phiên bản hiện tại (theo NĐ 52/85 đang hiệu lực) và thêm phiên bản mới khi luật 2025 ban hành, không phá lịch sử. Đây cũng phản ánh đúng trạng thái "đang theo dõi dự thảo" của tài liệu nguồn.

**Vì sao tách disclosure affiliate/livestream thành nghĩa vụ riêng (§1 #5)?** Dự thảo 2025 lần đầu quản hai hoạt động này - đúng hai thứ SănDeal chạm tới (affiliate user-initiated, và nếu có nội dung livestream/KOL). Tách thành `obligation_key` riêng cho phép theo dõi từng nghĩa vụ độc lập và phối hợp với TASK-AFFIL-002/004 (nơi thực thi disclosure kỹ thuật).

**Vì sao lưu source_law (§1 #6)?** Nghĩa vụ đến từ nhiều văn bản (NĐ 52/2013, NĐ 85/2021, dự thảo 2025). Khi rà soát, phải biết nghĩa vụ này căn cứ luật nào để kiểm tra còn hiệu lực không, đã sửa đổi chưa. `source_law` là điểm truy về văn bản gốc.

**Vì sao SHOULD chứ không MUST, đánh dấu giá trị tạm thời (DEC-COMPLY-34)?** Phần lớn nghĩa vụ kích hoạt theo quy mô (ngưỡng giao dịch) hoặc theo thời điểm luật 2025 chốt. Trước khi đạt quy mô hoặc luật ban hành, đây là khung chuẩn bị, không chặn vận hành P1/P2. Đánh dấu giá trị tạm thời ("chờ luật chốt") trung thực với §10 tài liệu nguồn - khung sẵn sàng nhận con số chính xác khi có.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/comply/migrations/0008_ecommerce_obligation.sql
CREATE TABLE ecommerce_obligation (
  id             BIGSERIAL   PRIMARY KEY,
  obligation_key TEXT        NOT NULL,
  description_vi TEXT        NOT NULL,
  status         TEXT        NOT NULL DEFAULT 'not_started'
                   CHECK (status IN ('not_started','submitted','approved','done','n_a')),
  due_at         TIMESTAMPTZ,
  completed_at   TIMESTAMPTZ,
  source_law     TEXT        NOT NULL,   -- 'ND_52_2013' | 'ND_85_2021' | 'DRAFT_2025'
  version        INTEGER     NOT NULL DEFAULT 1,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (obligation_key, version)
);

CREATE TABLE yearly_transaction_count (
  year   INTEGER PRIMARY KEY,
  count  BIGINT  NOT NULL DEFAULT 0
);

-- Nguong cau hinh versioned (mot cho), khong hardcode rai rac.
CREATE TABLE compliance_threshold (
  key     TEXT    NOT NULL,
  value   BIGINT  NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  UNIQUE (key, version)
);
INSERT INTO compliance_threshold (key, value) VALUES
  ('foreign_platform_yearly_tx', 100000); -- nguong NĐ 85/2021

INSERT INTO ecommerce_obligation (obligation_key, description_vi, source_law) VALUES
  ('moit_registration', 'Dang ky/thong bao website TMĐT voi Bo Cong Thuong', 'ND_52_2013'),
  ('affiliate_disclosure', 'Cong bo quan he affiliate (du thao Luat TMĐT 2025 - cho luat chot)', 'DRAFT_2025'),
  ('livestream_disclosure', 'Cong bo noi dung livestream thuong mai (du thao 2025 - cho luat chot)', 'DRAFT_2025');
```

### Threshold (Go)

```go
// services/comply/internal/ecom/threshold.go
type ThresholdState struct {
    Year         int
    Count        int64
    Threshold    int64
    MustRegister bool
}

// Threshold suy co tu bo dem, KHONG nhap tay.
func (s *Service) Threshold(ctx context.Context, year int) (ThresholdState, error) {
    cnt, err := s.repo.txCount(ctx, year)
    if err != nil {
        return ThresholdState{}, err
    }
    th, err := s.repo.threshold(ctx, "foreign_platform_yearly_tx")
    if err != nil {
        return ThresholdState{}, err
    }
    return ThresholdState{
        Year: year, Count: cnt, Threshold: th,
        MustRegister: cnt > th, // vuot nguong -> phai dang ky
    }, nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `ecommerce_obligation`, `yearly_transaction_count`, `compliance_threshold` tồn tại; ngưỡng 100.000 được seed.
2. `Threshold(2026)` với count = 50.000 -> `MustRegister = false`.
3. `Threshold(2026)` với count = 120.000 -> `MustRegister = true` (vượt 100.000).
4. Cờ `must_register` suy từ bộ đếm, không có setter tay (review API).
5. `MarkObligation("moit_registration", "approved")` -> trạng thái cập nhật, `completed_at` set.
6. Nghĩa vụ `affiliate_disclosure` tồn tại với `source_law = 'DRAFT_2025'` và mô tả ghi "chờ luật chốt".
7. `Obligations()` trả 3 nghĩa vụ seed với source_law đúng (NĐ 52, dự thảo 2025).
8. INSERT `status = 'rejected'` -> lỗi CHECK constraint.
9. Đổi ngưỡng (tạo `compliance_threshold` version 2 = 80.000) -> `Threshold` dùng giá trị mới sau reload; bản cũ vẫn lưu.
10. `Outstanding()` trả nghĩa vụ chưa `done/approved` (ví dụ MOIT `not_started`).
11. Nghĩa vụ thuộc dự thảo 2025 đánh dấu rõ tạm thời (mô tả + source_law DRAFT_2025).
12. Metric `ecom_threshold_exceeded` bật khi vượt ngưỡng; `ecom_obligation_outstanding_total` phản ánh số chưa hoàn tất.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/ecom/threshold_test.go
func TestThreshold_BelowDoesNotFlag(t *testing.T) {
    s := setupSeeded(t)
    setTxCount(t, s, 2026, 50_000)
    st, _ := s.Threshold(ctx, 2026)
    require.False(t, st.MustRegister)
}

func TestThreshold_AboveFlagsRegister(t *testing.T) {
    s := setupSeeded(t)
    setTxCount(t, s, 2026, 120_000)
    st, _ := s.Threshold(ctx, 2026)
    require.True(t, st.MustRegister) // vuot 100.000
}

func TestThreshold_FromVersionedConfig(t *testing.T) {
    s := setupSeeded(t)
    seedThresholdVersion(t, s, "foreign_platform_yearly_tx", 2, 80_000)
    s.Reload(ctx)
    setTxCount(t, s, 2026, 90_000)
    st, _ := s.Threshold(ctx, 2026)
    require.True(t, st.MustRegister) // nguong moi 80.000
}

// services/comply/internal/ecom/obligation_test.go
func TestObligation_DraftMarkedProvisional(t *testing.T) {
    s := setupSeeded(t)
    obs, _ := s.Obligations(ctx)
    aff := findKey(obs, "affiliate_disclosure")
    require.Equal(t, "DRAFT_2025", aff.SourceLaw)
    require.Contains(t, aff.DescriptionVi, "cho luat chot")
}

func TestObligation_MarkApproved(t *testing.T) {
    s := setupSeeded(t)
    require.NoError(t, s.MarkObligation(ctx, "moit_registration", "approved"))
    obs, _ := s.Obligations(ctx)
    require.Equal(t, "approved", findKey(obs, "moit_registration").Status)
}

func TestObligation_InvalidStatusRejected(t *testing.T) {
    s := setupSeeded(t)
    err := s.MarkObligation(ctx, "moit_registration", "rejected")
    require.Error(t, err) // CHECK status
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0008 (3 bảng + seed ngưỡng + 3 nghĩa vụ) -> types.go -> threshold.go (suy cờ từ bộ đếm) -> obligation.go (checklist + trạng thái) -> repo.go -> tests. Bộ đếm `yearly_transaction_count` được cập nhật từ nguồn giao dịch thật (BILL/AFFIL tổng hợp định kỳ); task này chỉ so ngưỡng và suy cờ, không tự đếm giao dịch. Ngưỡng và checklist là cấu hình versioned ở một chỗ; khi dự thảo 2025 chốt, thêm version mới với giá trị đã xác minh. Disclosure kỹ thuật thực thi ở TASK-AFFIL-002/004; task này theo dõi trạng thái nghĩa vụ.

---

## §7 - Phụ thuộc

- **TASK-COMPLY-001 (liên quan)** - khung consent/disclosure nền; disclosure affiliate phối hợp.
- **TASK-AFFIL-002** - deep-link user-initiated kèm disclosure; nghĩa vụ `affiliate_disclosure` theo dõi trạng thái.
- **TASK-AFFIL-004 (liên quan)** - guardrails né Honey + disclosure; đáp ứng dự thảo 2025.
- **TASK-COMPLY-006 (liên quan)** - per-country gating; nghĩa vụ TMĐT là phần VN của ma trận luật.
- Nguồn bộ đếm: BILL/AFFIL tổng hợp giao dịch năm (ngoài phạm vi task này).
- Lib: driver `pgx`.

---

## §8 - Payload ví dụ

### Kiểm ngưỡng giao dịch năm (cờ phải đăng ký)

```json
GET /v1/comply/ecommerce/threshold?year=2026
```

```json
200 OK
{
  "year": 2026,
  "count": 120000,
  "threshold": 100000,
  "must_register": true,
  "source_law": "ND_85_2021"
}
```

### Danh sách nghĩa vụ TMĐT hiện hành

```json
GET /v1/comply/ecommerce/obligations
```

```json
200 OK
[
  { "obligation_key": "moit_registration", "status": "submitted", "source_law": "ND_52_2013" },
  { "obligation_key": "affiliate_disclosure", "status": "not_started", "source_law": "DRAFT_2025",
    "description_vi": "Cong bo quan he affiliate (du thao Luat TMĐT 2025 - cho luat chot)" }
]
```

---

## §9 - Câu hỏi mở

Khung sẵn sàng; giá trị chính xác chờ xác minh (§10 tài liệu nguồn):
- Nội dung và ngưỡng chính thức của Luật TMĐT 2025 khi ban hành (livestream + affiliate) - cập nhật version checklist khi chốt.
- Quy trình đăng ký MOIT cụ thể cho mô hình extension + web (sàn hay không phải sàn) - xác minh với tư vấn pháp lý.
- Cách tổng hợp "giao dịch/năm" đúng định nghĩa luật (đơn hàng affiliate confirmed vs click) - thống nhất với BILL/AFFIL.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Nhập trạng thái vượt ngưỡng tay | review §1 #3 | quên/trễ đăng ký | Cờ suy từ bộ đếm, không setter tay |
| Hardcode ngưỡng 100.000 rải rác | review §1 #11 | đổi ngưỡng sót chỗ | Ngưỡng cấu hình versioned một chỗ |
| Bỏ disclosure affiliate/livestream | obligation key + AC #6 | vi phạm dự thảo 2025 | Nghĩa vụ riêng + phối hợp TASK-AFFIL |
| Coi giá trị dự thảo là chính thức | source_law DRAFT_2025 + AC #11 | tuân thủ theo luật chưa chốt | Đánh dấu tạm thời "chờ luật chốt" |
| status lạ | DB CHECK + AC #8 | dữ liệu bẩn | CHECK IN tập trạng thái |
| Đổi ngưỡng phá lịch sử | versioned + AC #9 | mất căn cứ cũ | compliance_threshold versioned |
| Đếm giao dịch sai định nghĩa | §9 + phối hợp BILL | cờ sai | Thống nhất định nghĩa với BILL/AFFIL |
| Mất truy vết căn cứ luật | source_law §1 #6 | rà soát khó | Mỗi nghĩa vụ ghi văn bản gốc |

---

## §11 - Ghi chú

- Theo dõi nghĩa vụ TMĐT VN: đăng ký MOIT (NĐ 52/2013), trách nhiệm sàn + ngưỡng 100.000 giao dịch/năm foreign platform (NĐ 85/2021), disclosure livestream + affiliate (dự thảo Luật TMĐT 2025).
- Cờ vượt ngưỡng suy tự động từ bộ đếm giao dịch năm; báo động khi vượt thay vì phải nhớ.
- Ngưỡng và checklist là cấu hình versioned ở một chỗ; khi dự thảo 2025 chốt, thêm version với giá trị đã xác minh.
- Nghĩa vụ dự thảo 2025 đánh dấu rõ tạm thời ("chờ luật chốt"), trung thực với §10 tài liệu nguồn.
- Disclosure kỹ thuật thực thi ở TASK-AFFIL-002/004; task này theo dõi trạng thái và căn cứ pháp lý.
- SHOULD vì nghĩa vụ kích hoạt theo quy mô (ngưỡng) hoặc thời điểm luật chốt; khung sẵn sàng để bật đúng lúc, không chặn P1/P2.

---

*Hết TASK-COMPLY-008. Status: ready_to_implement (mục tiêu audit 10/10).*
