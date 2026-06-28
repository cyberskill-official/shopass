---
id: FR-COMPLY-002
title: "Sổ đăng ký DPIA/TIA - nộp trong 60 ngày từ khi bắt đầu xử lý, cập nhật mỗi 6 tháng; lưu đánh giá tác động + chuyển dữ liệu xuyên biên giới (Luật 91/2025, NĐ 356/2025)"
module: COMPLY
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-COMPLY-001, FR-COMPLY-003, FR-COMPLY-004, FR-COMPLY-007, FR-B2B-001]
depends_on: [FR-COMPLY-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (DPIA/TIA nộp 60 ngày, cập nhật 6 tháng, chuyển dữ liệu xuyên biên giới)"
  - "docs/... §5.5 (NĐ 356/2025 thay NĐ 13/2023), §3.1 (kiến trúc xử lý dữ liệu)"
source_decisions:
  - "DEC-COMPLY-06: mỗi hoạt động xử lý (processing_activity) sinh một DPIA; hạn nộp = ngày bắt đầu xử lý + 60 ngày"
  - "DEC-COMPLY-07: DPIA versioned; cập nhật định kỳ tối thiểu mỗi 6 tháng (review_due_at = last_reviewed + 6 tháng)"
  - "DEC-COMPLY-08: hoạt động có chuyển dữ liệu xuyên biên giới yêu cầu TIA (transfer impact) gắn kèm, ghi nước nhận + cơ chế bảo vệ"
  - "DEC-COMPLY-09: sổ đăng ký là nguồn báo cáo cho cơ quan; trạng thái filing (draft/submitted/overdue) tính từ deadline, không nhập tay"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0003_processing_activity.sql
  - services/comply/migrations/0004_dpia_register.sql
  - services/comply/internal/dpia/register.go
  - services/comply/internal/dpia/deadline.go
  - services/comply/internal/dpia/repo.go
  - services/comply/internal/dpia/repo_test.go
  - services/comply/internal/dpia/deadline_test.go
modified_files:
  - services/comply/internal/dpia/types.go      # struct ProcessingActivity, DPIA, TIA
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - nhập trạng thái filing bằng tay thay vì suy từ deadline (vi phạm DEC-COMPLY-09, che giấu quá hạn)
  - tạo hoạt động xử lý mà không sinh DPIA hạn 60 ngày (vi phạm DEC-COMPLY-06)
  - bỏ TIA cho hoạt động có chuyển dữ liệu xuyên biên giới (vi phạm DEC-COMPLY-08)

effort_hours: 6
sub_tasks:
  - "0.5h: 0003_processing_activity.sql - bảng hoạt động xử lý + started_at + cross_border"
  - "0.5h: 0004_dpia_register.sql - bảng dpia versioned + tia gắn kèm + index review_due"
  - "1.0h: types.go + register.go - tạo activity sinh DPIA, gắn TIA khi cross_border"
  - "1.0h: deadline.go - tính filing_due (start+60d) + review_due (last_reviewed+6m) + status"
  - "1.0h: repo.go - liệt kê overdue, due-soon; report cho cơ quan"
  - "1.0h: repo_test.go - tạo activity -> DPIA hạn 60 ngày; cross_border -> bắt buộc TIA"
  - "1.0h: deadline_test.go - quá 60 ngày chưa nộp -> overdue; quá 6 tháng chưa review -> review_overdue"

risk_if_skipped: "PDPL Luật 91/2025 + NĐ 356/2025 yêu cầu DPIA nộp trong 60 ngày từ khi bắt đầu xử lý và cập nhật mỗi 6 tháng; hoạt động chuyển dữ liệu xuyên biên giới cần TIA. Không có sổ đăng ký theo dõi deadline -> quá hạn âm thầm, vi phạm hạng High (§9). Chế tài tới 5% doanh thu năm trước cho vi phạm xuyên biên giới (§5.5). SănDeal xử lý dữ liệu giỏ hàng + hành vi mua + có thể chuyển dữ liệu khi mở SEA, nên DPIA/TIA là bắt buộc trước khi bật mỗi hoạt động. Đây cũng là tài liệu chứng minh trách nhiệm giải trình khi cơ quan thanh tra."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** duy trì sổ đăng ký DPIA (Data Protection Impact Assessment) và TIA (Transfer Impact Assessment): mỗi hoạt động xử lý dữ liệu cá nhân có một DPIA với hạn nộp 60 ngày và chu kỳ cập nhật 6 tháng; hoạt động chuyển dữ liệu xuyên biên giới có TIA gắn kèm. Hợp đồng:

1. **MUST** định nghĩa bảng `processing_activity (id, name, purpose_key, data_categories, started_at, cross_border, recipient_country, created_at)` - mỗi hoạt động xử lý đăng ký một dòng. `purpose_key` liên kết tập purpose của FR-COMPLY-001.
2. **MUST** định nghĩa bảng `dpia (id, activity_id, version, risk_level, mitigation_vi, status, filed_at, last_reviewed_at, created_at)` versioned: cập nhật DPIA là tạo version mới, giữ bản cũ (DEC-COMPLY-07).
3. **MUST** định nghĩa bảng `tia (id, dpia_id, recipient_country, safeguard_vi, created_at)` cho đánh giá tác động chuyển dữ liệu; bắt buộc tồn tại khi `processing_activity.cross_border = true` (DEC-COMPLY-08).
4. **MUST** tính `filing_due_at = started_at + INTERVAL '60 days'` (DEC-COMPLY-06): hạn nộp DPIA là 60 ngày kể từ khi bắt đầu xử lý.
5. **MUST** tính `review_due_at = COALESCE(last_reviewed_at, filed_at) + INTERVAL '6 months'` (DEC-COMPLY-07): DPIA phải được rà soát lại tối thiểu mỗi 6 tháng.
6. **MUST** suy ra trạng thái filing từ deadline, KHÔNG nhập tay (DEC-COMPLY-09):
    - `draft` - chưa `filed_at`, còn trong 60 ngày.
    - `submitted` - đã `filed_at` trong hạn.
    - `overdue` - chưa `filed_at` mà đã quá `filing_due_at`.
    - `review_overdue` - đã nộp nhưng quá `review_due_at`.
7. **MUST** chặn tạo `processing_activity` với `cross_border = true` mà không có TIA: hàm tạo phải yêu cầu thông tin `recipient_country` + `safeguard` và sinh TIA cùng lúc.
8. **MUST** expose hàm:
    - `RegisterActivity(ctx, a ProcessingActivity, dpia DPIAInput) (int64, error)` - tạo activity + DPIA v1 (+ TIA nếu cross_border); trả activity_id.
    - `ReviewDPIA(ctx, activityID int64, in DPIAInput) error` - tạo DPIA version mới, set `last_reviewed_at = now()`.
    - `MarkFiled(ctx, dpiaID int64) error` - set `filed_at = now()` (đã nộp cho cơ quan).
    - `Overdue(ctx) ([]ActivityStatus, error)` - liệt kê hoạt động `overdue` hoặc `review_overdue`.
9. **MUST** liệt kê được "sắp tới hạn" (due trong N ngày) để cảnh báo trước khi quá hạn.
10. **SHOULD** phát OTel gauge `dpia_overdue_total`, `dpia_review_due_soon_total`; báo cáo định kỳ qua Grafana.
11. **MUST** giữ `risk_level` trong tập `('low','medium','high')`; hoạt động `high` SHOULD kèm `mitigation_vi` không rỗng.
12. **MUST** sinh báo cáo cho cơ quan: `Report(ctx)` trả danh sách hoạt động + DPIA hiện hành + TIA + trạng thái + ngày nộp, đủ để xuất nộp.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao deadline tính từ dữ liệu, không nhập tay (DEC-COMPLY-09)?** Trạng thái "quá hạn" mà nhập tay dễ bị bỏ sót hoặc tô hồng. Tính `filing_due_at` và `review_due_at` trực tiếp từ `started_at`/`last_reviewed_at` cho ra trạng thái khách quan: hệ thống tự biết hoạt động nào quá 60 ngày chưa nộp, hoạt động nào quá 6 tháng chưa review. Đây là điểm chống "compliance hình thức".

**Vì sao DPIA versioned (DEC-COMPLY-07)?** Cập nhật mỗi 6 tháng nghĩa là DPIA tiến hóa theo thời gian (rủi ro mới, biện pháp mới). Giữ version cho phép chứng minh lịch sử rà soát: tại mỗi mốc 6 tháng, đánh giá là gì. Cơ quan có thể hỏi "lần review gần nhất ngày nào, thay đổi gì" - version trả lời được.

**Vì sao TIA bắt buộc khi cross_border (DEC-COMPLY-08)?** Chuyển dữ liệu xuyên biên giới là nhóm rủi ro PDPL chế tài nặng nhất (tới 5% doanh thu năm trước). Khi mở SEA, dữ liệu có thể rời VN. TIA ghi rõ nước nhận + cơ chế bảo vệ (safeguard). Bắt buộc tạo TIA cùng activity tránh việc bật chuyển dữ liệu mà quên đánh giá.

**Vì sao gắn `purpose_key` với FR-COMPLY-001 (§1 #1)?** DPIA và consent là hai mặt của cùng một hoạt động xử lý: consent là cơ sở pháp lý, DPIA là đánh giá rủi ro. Liên kết qua `purpose_key` giữ hai sổ đồng bộ - mỗi mục đích xử lý vừa có consent vừa có DPIA.

**Vì sao "due trong N ngày" (§1 #9)?** Phát hiện quá hạn sau khi đã quá hạn là quá muộn. Cảnh báo trước (ví dụ 14 ngày trước `filing_due_at` hoặc `review_due_at`) cho đội pháp lý thời gian chuẩn bị, biến deadline thành quy trình chủ động.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/comply/migrations/0003_processing_activity.sql
CREATE TABLE processing_activity (
  id               BIGSERIAL   PRIMARY KEY,
  name             TEXT        NOT NULL,
  purpose_key      TEXT        NOT NULL,
  data_categories  TEXT[]      NOT NULL,
  started_at       TIMESTAMPTZ NOT NULL,
  cross_border     BOOLEAN     NOT NULL DEFAULT false,
  recipient_country TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (NOT cross_border OR recipient_country IS NOT NULL)
);

-- services/comply/migrations/0004_dpia_register.sql
CREATE TABLE dpia (
  id               BIGSERIAL   PRIMARY KEY,
  activity_id      BIGINT      NOT NULL REFERENCES processing_activity(id),
  version          INTEGER     NOT NULL,
  risk_level       TEXT        NOT NULL CHECK (risk_level IN ('low','medium','high')),
  mitigation_vi    TEXT,
  status           TEXT        NOT NULL DEFAULT 'draft',
  filed_at         TIMESTAMPTZ,
  last_reviewed_at TIMESTAMPTZ,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (activity_id, version)
);

CREATE TABLE tia (
  id                BIGSERIAL   PRIMARY KEY,
  dpia_id           BIGINT      NOT NULL REFERENCES dpia(id),
  recipient_country TEXT        NOT NULL,
  safeguard_vi      TEXT        NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Truy van sap toi han / qua han.
CREATE INDEX idx_dpia_activity ON dpia (activity_id, version DESC);
```

### Deadline (Go)

```go
// services/comply/internal/dpia/deadline.go
const (
    FilingWindow = 60 * 24 * time.Hour       // 60 ngay
    ReviewCycle  = 6 * 30 * 24 * time.Hour   // ~6 thang (xap xi; doi chieu lich thuc te)
)

// Status suy tu deadline, KHONG nhap tay.
func Status(a ProcessingActivity, d DPIA, now time.Time) string {
    filingDue := a.StartedAt.Add(FilingWindow)
    if d.FiledAt == nil {
        if now.After(filingDue) {
            return "overdue"
        }
        return "draft"
    }
    base := d.FiledAt
    if d.LastReviewedAt != nil {
        base = d.LastReviewedAt
    }
    if now.After(base.Add(ReviewCycle)) {
        return "review_overdue"
    }
    return "submitted"
}
```

### Register (Go)

```go
// services/comply/internal/dpia/register.go
// RegisterActivity tao activity + DPIA v1; bat buoc TIA khi cross_border.
func (s *Service) RegisterActivity(ctx context.Context, a ProcessingActivity, in DPIAInput) (int64, error) {
    if a.CrossBorder && (in.TIA == nil || in.TIA.Safeguard == "") {
        return 0, ErrTIARequired // cross-border phai co TIA
    }
    return s.repo.createWithDPIA(ctx, a, in)
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `processing_activity`, `dpia`, `tia` tồn tại.
2. `RegisterActivity` (cross_border=false) -> tạo activity + DPIA version 1, status `draft`.
3. DPIA mới có `filing_due_at = started_at + 60 ngày` (kiểm qua `Status`).
4. Activity bắt đầu cách đây 61 ngày, chưa `filed_at` -> `Status = overdue`.
5. `MarkFiled` trong hạn -> `Status = submitted`.
6. DPIA đã nộp, `last_reviewed_at` cách đây hơn 6 tháng -> `Status = review_overdue`.
7. `RegisterActivity` (cross_border=true) thiếu TIA -> lỗi `ErrTIARequired`, không tạo gì.
8. `RegisterActivity` (cross_border=true) có TIA -> tạo activity + DPIA + 1 dòng `tia` với recipient_country + safeguard.
9. `ReviewDPIA` -> tạo DPIA version 2, `last_reviewed_at = now()`; version 1 vẫn còn.
10. INSERT `risk_level = 'critical'` -> lỗi CHECK constraint.
11. INSERT `processing_activity` cross_border=true mà `recipient_country IS NULL` -> lỗi CHECK.
12. `Overdue()` trả đúng các activity `overdue` và `review_overdue`; `Report()` xuất đủ activity + DPIA hiện hành + TIA + status.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/dpia/deadline_test.go
func TestDeadline_FilingOverdueAfter60d(t *testing.T) {
    a := ProcessingActivity{StartedAt: now.Add(-61 * 24 * time.Hour)}
    d := DPIA{FiledAt: nil}
    require.Equal(t, "overdue", Status(a, d, now))
}

func TestDeadline_DraftWithin60d(t *testing.T) {
    a := ProcessingActivity{StartedAt: now.Add(-10 * 24 * time.Hour)}
    require.Equal(t, "draft", Status(a, DPIA{FiledAt: nil}, now))
}

func TestDeadline_ReviewOverdueAfter6m(t *testing.T) {
    filed := now.Add(-200 * 24 * time.Hour) // > 6 thang
    a := ProcessingActivity{StartedAt: now.Add(-220 * 24 * time.Hour)}
    d := DPIA{FiledAt: &filed}
    require.Equal(t, "review_overdue", Status(a, d, now))
}

// services/comply/internal/dpia/repo_test.go
func TestRegister_CrossBorderRequiresTIA(t *testing.T) {
    s := setup(t)
    a := ProcessingActivity{Name: "SEA analytics", CrossBorder: true, RecipientCountry: "ID"}
    _, err := s.RegisterActivity(ctx, a, DPIAInput{RiskLevel: "high"}) // thieu TIA
    require.ErrorIs(t, err, ErrTIARequired)
}

func TestRegister_CrossBorderWithTIA_CreatesTIA(t *testing.T) {
    s := setup(t)
    a := ProcessingActivity{Name: "SEA analytics", CrossBorder: true, RecipientCountry: "ID"}
    in := DPIAInput{RiskLevel: "high", TIA: &TIAInput{RecipientCountry: "ID", Safeguard: "SCC + ma hoa"}}
    id, err := s.RegisterActivity(ctx, a, in)
    require.NoError(t, err)
    require.Equal(t, 1, countTIA(t, s, id))
}

func TestReview_CreatesNewVersion(t *testing.T) {
    s := setup(t)
    id, _ := s.RegisterActivity(ctx, basicActivity(), DPIAInput{RiskLevel: "low"})
    require.NoError(t, s.ReviewDPIA(ctx, id, DPIAInput{RiskLevel: "medium"}))
    require.Equal(t, 2, maxDPIAVersion(t, s, id))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0003 (processing_activity) -> 0004 (dpia + tia) -> types.go + deadline.go (logic deadline thuần, dễ test) -> register.go (ràng buộc TIA) -> repo.go (overdue/report) -> tests. Hằng `ReviewCycle` dùng xấp xỉ 6 tháng theo ngày; nếu cần chính xác theo lịch (tháng dương lịch), dùng `AddDate(0, 6, 0)` ở tầng query thay vì cộng Duration. Logic `Status` tách thuần hàm để test không cần DB.

---

## §7 - Phụ thuộc

- **FR-COMPLY-001** - tập purpose (`purpose_key`) là cơ sở liên kết hoạt động xử lý với consent.
- **FR-COMPLY-007 (downstream)** - adapter SEA dùng cross_border + TIA khi dữ liệu rời VN sang ID/TH.
- **FR-B2B-001 (liên quan)** - pipeline dữ liệu B2B ẩn danh là một `processing_activity` cần DPIA riêng.
- **FR-COMPLY-003 (liên quan)** - DSAR có thể tham chiếu danh mục `data_categories` của hoạt động.
- Lib: driver `pgx`.

---

## §8 - Payload ví dụ

### Đăng ký hoạt động xử lý có chuyển dữ liệu xuyên biên giới

```json
POST /v1/comply/processing-activity
{
  "name": "Bao cao xu huong gia SEA",
  "purpose_key": "analytics_b2b",
  "data_categories": ["price_behavior_anonymized"],
  "started_at": "2026-06-28T00:00:00+07:00",
  "cross_border": true,
  "recipient_country": "ID",
  "dpia": {
    "risk_level": "high",
    "mitigation_vi": "An danh hoa k-anonymity truoc khi chuyen; ma hoa khi luu/chuyen.",
    "tia": { "recipient_country": "ID", "safeguard_vi": "SCC + ma hoa AES-256" }
  }
}
```

### Báo cáo hoạt động quá hạn (cho đội pháp lý)

```sql
-- Liet ke DPIA hien hanh qua han nop (>60 ngay tu started_at, chua filed).
SELECT pa.name, pa.started_at, d.version, d.status
FROM processing_activity pa
JOIN LATERAL (
  SELECT * FROM dpia WHERE activity_id = pa.id ORDER BY version DESC LIMIT 1
) d ON true
WHERE d.filed_at IS NULL
  AND now() > pa.started_at + INTERVAL '60 days';
```

---

## §9 - Câu hỏi mở

Đã chốt khung. Hoãn:
- Tự động hóa nộp DPIA điện tử qua cổng cơ quan (nếu có API) - hiện xuất báo cáo thủ công.
- Template DPIA theo loại hoạt động (scraping vs B2B vs marketing) - bổ sung mẫu sau.
- Liên thông DPIA với risk register §9 tài liệu nguồn - đồng bộ thủ công giai đoạn đầu.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Quá 60 ngày chưa nộp DPIA | `Status=overdue` + AC #4 | vi phạm PDPL | Cảnh báo due-soon §1 #9; quy trình nộp trước hạn |
| Quá 6 tháng chưa review | `Status=review_overdue` + AC #6 | DPIA lỗi thời | Gauge review_due_soon; lịch review định kỳ |
| Cross-border thiếu TIA | `ErrTIARequired` + AC #7 | chuyển dữ liệu không đánh giá | Hàm tạo bắt buộc TIA; CHECK recipient_country |
| Nhập trạng thái filing tay | review §1 #6 | che giấu quá hạn | Status suy từ deadline, không có setter tay |
| recipient_country trống khi cross_border | DB CHECK + AC #11 | TIA thiếu đối tượng | CHECK constraint chặn |
| risk_level lạ | DB CHECK + AC #10 | phân loại sai | CHECK IN low/medium/high |
| Review ghi đè bản cũ | versioned + AC #9 | mất lịch sử rà soát | DPIA versioned, tạo dòng mới |
| Hoạt động bật mà quên đăng ký | audit coverage | xử lý không có DPIA | Quy trình: bật hoạt động phải qua RegisterActivity |

---

## §11 - Ghi chú

- DPIA/TIA là tài liệu trách nhiệm giải trình PDPL: nộp trong 60 ngày, cập nhật mỗi 6 tháng.
- Trạng thái quá hạn tính khách quan từ deadline, không phụ thuộc người nhập - chống compliance hình thức.
- TIA bắt buộc cho chuyển dữ liệu xuyên biên giới, nhóm rủi ro chế tài nặng nhất (tới 5% doanh thu năm trước).
- Liên kết `purpose_key` giữ sổ DPIA đồng bộ với sổ consent (FR-COMPLY-001): mỗi mục đích xử lý có cả hai.
- Khi mở SEA (FR-COMPLY-007), mỗi luồng chuyển dữ liệu là một activity cross_border kèm TIA ghi rõ nước nhận và cơ chế bảo vệ.
- DPIA versioned cho phép chứng minh lịch sử rà soát từng chu kỳ 6 tháng trước cơ quan.

---

*Hết FR-COMPLY-002. Status: ready_to_implement (mục tiêu audit 10/10).*
