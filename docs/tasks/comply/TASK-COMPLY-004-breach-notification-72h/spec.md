---
id: TASK-COMPLY-004
title: "Quy trình thông báo vi phạm trong 72 giờ - phát hiện -> phân loại -> thông báo cơ quan + chủ thể; đồng hồ đếm ngược từ thời điểm nhận biết (Luật 91/2025)"
module: COMPLY
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-COMPLY-001, TASK-INFRA-004, TASK-COMPLY-003, TASK-NOTIF-001]
depends_on: [TASK-COMPLY-001, TASK-INFRA-004]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (thông báo vi phạm trong 72 giờ)"
  - "docs/... §3.8 (observability spine - nguồn tín hiệu phát hiện), §5.5 (PDPL Luật 91/2025)"
source_decisions:
  - "DEC-COMPLY-15: đồng hồ 72 giờ đếm từ acknowledged_at (thời điểm tổ chức nhận biết), không phải từ occurred_at"
  - "DEC-COMPLY-16: vi phạm chạy qua state machine detected -> triaged -> notified_authority -> notified_subjects -> closed; mỗi bước có dấu thời gian"
  - "DEC-COMPLY-17: phân loại mức nghiêm trọng quyết định có phải thông báo chủ thể hay không; quá hạn 72h là cờ đỏ tự suy ra"
  - "DEC-COMPLY-18: tín hiệu phát hiện đến từ observability spine (TASK-INFRA-004); breach record liên kết trace/alert nguồn"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0006_breach_incident.sql
  - services/comply/internal/breach/incident.go
  - services/comply/internal/breach/clock.go
  - services/comply/internal/breach/repo.go
  - services/comply/internal/breach/incident_test.go
  - services/comply/internal/breach/clock_test.go
modified_files:
  - services/comply/internal/breach/types.go      # struct BreachIncident, Severity
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - đếm 72 giờ từ occurred_at thay vì acknowledged_at (vi phạm DEC-COMPLY-15, sai mốc luật)
  - cho phép nhảy trạng thái bỏ qua bước (notified mà chưa triaged) (vi phạm DEC-COMPLY-16)
  - đóng vi phạm nghiêm trọng mà chưa thông báo chủ thể (vi phạm DEC-COMPLY-17)

effort_hours: 5
sub_tasks:
  - "0.5h: 0006_breach_incident.sql - bảng breach_incident + severity + các mốc thời gian + source_ref"
  - "1.0h: types.go + clock.go - đồng hồ 72h từ acknowledged_at + cờ deadline"
  - "1.5h: incident.go - state machine + ràng buộc chuyển trạng thái hợp lệ + phân loại"
  - "0.5h: repo.go - lưu/đọc, liệt kê vi phạm sắp/đã quá 72h chưa thông báo cơ quan"
  - "1.0h: incident_test.go - chuyển trạng thái hợp lệ; cấm nhảy bước; nghiêm trọng phải thông báo chủ thể"
  - "0.5h: clock_test.go - quá 72h chưa thông báo -> breach_overdue; còn hạn -> within_window"

risk_if_skipped: "PDPL Luật 91/2025 yêu cầu thông báo vi phạm dữ liệu trong 72 giờ. Không có quy trình theo dõi đồng hồ 72h + state machine -> bỏ lỡ hạn thông báo, biến một sự cố kỹ thuật thành vi phạm pháp lý hạng High (§9), chế tài tới 3 tỷ VND cho vi phạm nghiêm trọng (§5.5). Khi sự cố xảy ra, áp lực thời gian cao và dễ rối; cần state machine rõ ràng (phát hiện -> phân loại -> thông báo cơ quan -> thông báo chủ thể) với deadline tự suy để đội ứng cứu không bỏ sót bước. Nguồn tín hiệu phát hiện là observability spine (TASK-INFRA-004)."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** cung cấp quy trình quản lý sự cố vi phạm dữ liệu với đồng hồ 72 giờ và state machine, theo PDPL Luật 91/2025: từ phát hiện qua phân loại đến thông báo cơ quan và (khi cần) thông báo chủ thể dữ liệu. Hợp đồng:

1. **MUST** định nghĩa bảng `breach_incident (id, summary, severity, status, occurred_at, acknowledged_at, triaged_at, notified_authority_at, notified_subjects_at, closed_at, source_ref, created_at)` theo dõi toàn bộ vòng đời sự cố.
2. **MUST** đếm đồng hồ 72 giờ từ `acknowledged_at` (thời điểm tổ chức nhận biết), KHÔNG từ `occurred_at` (DEC-COMPLY-15): `authority_due_at = acknowledged_at + INTERVAL '72 hours'`.
3. **MUST** vận hành state machine (DEC-COMPLY-16): `detected -> triaged -> notified_authority -> notified_subjects -> closed`. Chuyển trạng thái phải tuần tự; cấm nhảy bước (ví dụ `detected` sang `notified_authority` mà chưa `triaged`).
4. **MUST** đặt dấu thời gian cho từng bước khi chuyển: `triaged_at`, `notified_authority_at`, `notified_subjects_at`, `closed_at` được set đúng lúc bước đó hoàn tất.
5. **MUST** phân loại mức nghiêm trọng `severity IN ('low','medium','high','critical')`; phân loại quyết định nghĩa vụ thông báo chủ thể (DEC-COMPLY-17): `high`/`critical` BẮT BUỘC qua `notified_subjects` trước khi `closed`.
6. **MUST** suy cờ deadline từ đồng hồ, KHÔNG nhập tay (DEC-COMPLY-17):
- `within_window` - chưa thông báo cơ quan, còn trong 72h.
- `breach_overdue` - chưa `notified_authority_at` mà đã quá `authority_due_at`.
- `notified` - đã thông báo cơ quan đúng hạn.
7. **MUST** liên kết tín hiệu nguồn (DEC-COMPLY-18): `source_ref` trỏ tới alert/trace của observability spine (TASK-INFRA-004) đã kích hoạt phát hiện, phục vụ điều tra.
8. **MUST** expose hàm:
- `Open(ctx, in BreachInput) (int64, error)` - tạo sự cố ở `detected`, set `acknowledged_at = now()` (bắt đầu đồng hồ).
- `Advance(ctx, id int64, to Status) error` - chuyển trạng thái hợp lệ tuần tự; set dấu thời gian tương ứng.
- `Close(ctx, id int64) error` - đóng; từ chối đóng `high`/`critical` chưa `notified_subjects`.
- `Overdue(ctx) ([]BreachIncident, error)` - liệt kê sự cố `breach_overdue` (cần báo động ngay).
9. **MUST** đảm bảo `Advance` từ chối chuyển không hợp lệ (lùi trạng thái, nhảy bước) bằng lỗi xác định `ErrInvalidTransition`.
10. **SHOULD** phát OTel metric: `breach_incident_total{severity}`, `breach_overdue_total` (cờ đỏ); báo động khi `breach_overdue_total > 0`.
11. **MUST** chặn `Close` khi severity `high`/`critical` mà `notified_subjects_at IS NULL` bằng lỗi `ErrSubjectsNotNotified`.
12. **MUST** ghi log audit mọi chuyển trạng thái (ai, lúc nào, từ đâu đến đâu) để dựng lại dòng thời gian sự cố khi báo cáo cơ quan.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao đồng hồ từ acknowledged_at, không từ occurred_at (DEC-COMPLY-15)?** Vi phạm có thể đã xảy ra âm thầm từ lâu trước khi ta biết. Luật tính 72 giờ từ thời điểm tổ chức "nhận biết" (became aware), không từ thời điểm sự cố vật lý xảy ra. Đếm từ `acknowledged_at` mới đúng mốc luật. Lưu cả `occurred_at` để điều tra nhưng đồng hồ pháp lý chạy từ lúc nhận biết.

**Vì sao state machine tuần tự (DEC-COMPLY-16)?** Lúc sự cố, áp lực cao, dễ làm tắt hoặc bỏ bước. Ép `detected -> triaged -> notified_authority -> notified_subjects -> closed` đảm bảo không ai "đóng" sự cố mà chưa phân loại hay chưa thông báo. Mỗi bước là một cổng có dấu thời gian, dựng lại được dòng thời gian khi cơ quan hỏi.

**Vì sao phân loại quyết định thông báo chủ thể (DEC-COMPLY-17)?** Không phải vi phạm nào cũng phải báo cho từng user; nhưng vi phạm nghiêm trọng (lộ dữ liệu nhạy cảm) thì có. Gắn nghĩa vụ thông báo chủ thể vào `severity`: `high`/`critical` không được `closed` nếu chưa `notified_subjects`. Đây là rào chặn quên thông báo người bị ảnh hưởng.

**Vì sao cờ deadline tự suy (§1 #6)?** Giống DPIA và DSAR, "quá hạn" phải khách quan. Tính `authority_due_at` từ `acknowledged_at` cho ra `breach_overdue` tự động. Khi `breach_overdue_total > 0`, báo động nổ - không để một sự cố trôi qua mốc 72h vì con người quên.

**Vì sao liên kết observability spine (DEC-COMPLY-18)?** Phát hiện vi phạm thường bắt đầu từ một alert (truy cập bất thường, rò log). `source_ref` trỏ về alert/trace nguồn nối sự cố pháp lý với bằng chứng kỹ thuật, giúp điều tra nhanh và báo cáo có căn cứ. TASK-INFRA-004 là nơi sinh tín hiệu đó.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/comply/migrations/0006_breach_incident.sql
CREATE TABLE breach_incident (
  id                    BIGSERIAL   PRIMARY KEY,
  summary               TEXT        NOT NULL,
  severity              TEXT        NOT NULL CHECK (severity IN ('low','medium','high','critical')),
  status                TEXT        NOT NULL DEFAULT 'detected'
                          CHECK (status IN ('detected','triaged','notified_authority','notified_subjects','closed')),
  occurred_at           TIMESTAMPTZ,
  acknowledged_at       TIMESTAMPTZ NOT NULL,   -- dong ho 72h dem tu day
  triaged_at            TIMESTAMPTZ,
  notified_authority_at TIMESTAMPTZ,
  notified_subjects_at  TIMESTAMPTZ,
  closed_at             TIMESTAMPTZ,
  source_ref            TEXT,                   -- alert/trace cua TASK-INFRA-004
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_breach_open ON breach_incident (status, acknowledged_at)
  WHERE status <> 'closed';
```

### Clock + state machine (Go)

```go
// services/comply/internal/breach/clock.go
const AuthorityWindow = 72 * time.Hour

// DeadlineFlag suy tu dong ho, KHONG nhap tay.
func DeadlineFlag(b BreachIncident, now time.Time) string {
    due := b.AcknowledgedAt.Add(AuthorityWindow)
    if b.NotifiedAuthorityAt == nil {
        if now.After(due) {
            return "breach_overdue"
        }
        return "within_window"
    }
    return "notified"
}

// services/comply/internal/breach/incident.go
var order = map[Status]int{
    "detected": 0, "triaged": 1, "notified_authority": 2,
    "notified_subjects": 3, "closed": 4,
}

// Advance chi cho phep tien dung mot buoc tuan tu.
func (s *Service) Advance(ctx context.Context, id int64, to Status) error {
    b, err := s.repo.get(ctx, id)
    if err != nil {
        return err
    }
    if order[to] != order[Status(b.Status)]+1 {
        return ErrInvalidTransition // cam lui hoac nhay buoc
    }
    return s.repo.transition(ctx, id, to, time.Now())
}

// Close tu choi dong neu nghiem trong ma chua thong bao chu the.
func (s *Service) Close(ctx context.Context, id int64) error {
    b, _ := s.repo.get(ctx, id)
    if (b.Severity == "high" || b.Severity == "critical") && b.NotifiedSubjectsAt == nil {
        return ErrSubjectsNotNotified
    }
    return s.repo.transition(ctx, id, "closed", time.Now())
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `breach_incident` tồn tại; CHECK severity + status hoạt động.
2. `Open` -> sự cố status `detected`, `acknowledged_at = now()` (đồng hồ bắt đầu).
3. `Advance(detected->triaged)` -> status `triaged`, `triaged_at` được set.
4. `Advance(detected->notified_authority)` (nhảy bước) -> lỗi `ErrInvalidTransition`.
5. `Advance` lùi trạng thái (triaged->detected) -> lỗi `ErrInvalidTransition`.
6. Sự cố `acknowledged_at` cách đây 73 giờ, chưa `notified_authority_at` -> `DeadlineFlag = breach_overdue`.
7. Sự cố trong 72h chưa thông báo -> `DeadlineFlag = within_window`.
8. `Close` sự cố `critical` chưa `notified_subjects` -> lỗi `ErrSubjectsNotNotified`.
9. `Close` sự cố `low` (không cần thông báo chủ thể) -> đóng được, `closed_at` set.
10. Tiến đủ chuỗi `high` qua `notified_subjects` rồi `Close` -> đóng được.
11. INSERT `severity = 'urgent'` -> lỗi CHECK constraint.
12. `Overdue()` trả đúng các sự cố `breach_overdue`; metric `breach_overdue_total` phản ánh.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/breach/clock_test.go
func TestClock_OverdueAfter72h(t *testing.T) {
    b := BreachIncident{AcknowledgedAt: now.Add(-73 * time.Hour), NotifiedAuthorityAt: nil}
    require.Equal(t, "breach_overdue", DeadlineFlag(b, now))
}

func TestClock_WithinWindow(t *testing.T) {
    b := BreachIncident{AcknowledgedAt: now.Add(-10 * time.Hour), NotifiedAuthorityAt: nil}
    require.Equal(t, "within_window", DeadlineFlag(b, now))
}

// services/comply/internal/breach/incident_test.go
func TestAdvance_SequentialOnly(t *testing.T) {
    s := setup(t)
    id, _ := s.Open(ctx, BreachInput{Summary: "log leak", Severity: "high"})
    require.NoError(t, s.Advance(ctx, id, "triaged"))
    err := s.Advance(ctx, id, "notified_subjects") // nhay buoc (chua notified_authority)
    require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestAdvance_NoBackward(t *testing.T) {
    s := setup(t)
    id, _ := s.Open(ctx, BreachInput{Summary: "x", Severity: "low"})
    s.Advance(ctx, id, "triaged")
    err := s.Advance(ctx, id, "detected") // lui
    require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestClose_CriticalNeedsSubjectNotice(t *testing.T) {
    s := setup(t)
    id, _ := s.Open(ctx, BreachInput{Summary: "PII leak", Severity: "critical"})
    s.Advance(ctx, id, "triaged")
    s.Advance(ctx, id, "notified_authority")
    err := s.Close(ctx, id) // chua notified_subjects
    require.ErrorIs(t, err, ErrSubjectsNotNotified)
}

func TestClose_LowSeverityNoSubjectNotice(t *testing.T) {
    s := setup(t)
    id, _ := s.Open(ctx, BreachInput{Summary: "minor", Severity: "low"})
    s.Advance(ctx, id, "triaged")
    s.Advance(ctx, id, "notified_authority")
    s.Advance(ctx, id, "notified_subjects")
    require.NoError(t, s.Close(ctx, id))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0006 -> types.go -> clock.go (đồng hồ thuần hàm) -> incident.go (state machine + ràng buộc Close) -> repo.go -> tests. `DeadlineFlag` và logic chuyển trạng thái tách thuần hàm để test không cần DB. Khi tích hợp, `Open` được gọi từ runbook ứng cứu sự cố (thủ công hoặc tự động từ alert TASK-INFRA-004); `source_ref` mang ID alert/trace nguồn. Báo động `breach_overdue_total > 0` nối vào kênh on-call.

---

## §7 - Phụ thuộc

- **TASK-INFRA-004** - observability spine sinh alert/trace; `source_ref` trỏ về đó.
- **TASK-COMPLY-001 (liên quan)** - phân loại dữ liệu bị ảnh hưởng tham chiếu danh mục purpose.
- **TASK-COMPLY-003 (liên quan)** - sự cố ảnh hưởng quyền chủ thể; thông báo chủ thể có thể trùng tập user của DSAR.
- **TASK-NOTIF-001 (liên quan)** - kênh thông báo chủ thể dữ liệu có thể tái dùng hạ tầng notification.
- Lib: driver `pgx`.

---

## §8 - Payload ví dụ

### Mở sự cố từ runbook (kèm tham chiếu alert observability)

```json
POST /v1/comply/breach
{
  "summary": "Truy cap bat thuong vao bang price_snapshot tu IP la",
  "severity": "high",
  "occurred_at": "2026-06-28T01:10:00+07:00",
  "source_ref": "otel-trace:7f3a... / grafana-alert:1180"
}
```

```json
201 Created
{
  "id": 42,
  "status": "detected",
  "acknowledged_at": "2026-06-28T03:00:00+07:00",
  "authority_due_at": "2026-07-01T03:00:00+07:00",
  "deadline_flag": "within_window"
}
```

### Báo cáo sự cố sắp quá 72h chưa thông báo cơ quan

```sql
SELECT id, summary, severity, acknowledged_at
FROM breach_incident
WHERE notified_authority_at IS NULL
  AND now() > acknowledged_at + INTERVAL '72 hours'
ORDER BY acknowledged_at;
```

---

## §9 - Câu hỏi mở

Đã chốt khung. Hoãn:
- Tự động hóa soạn nội dung thông báo cơ quan/chủ thể (mẫu) - hiện thao tác thủ công theo runbook.
- Tích hợp gửi thông báo chủ thể hàng loạt qua TASK-NOTIF (khi danh sách ảnh hưởng lớn) - nối sau.
- Phân loại tự động severity từ đặc trưng sự cố - hiện do người ứng cứu phân loại.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Đếm 72h sai mốc (từ occurred_at) | review §1 #2 | sai hạn pháp lý | Đồng hồ từ `acknowledged_at`; test clock |
| Nhảy bước state machine | `ErrInvalidTransition` + AC #4 | bỏ phân loại/thông báo | `Advance` tuần tự một bước |
| Lùi trạng thái | `ErrInvalidTransition` + AC #5 | dòng thời gian sai | Cấm lùi qua `order` map |
| Đóng sự cố nghiêm trọng chưa báo chủ thể | `ErrSubjectsNotNotified` + AC #8 | bỏ thông báo người ảnh hưởng | `Close` chặn high/critical chưa notified_subjects |
| Quá 72h chưa báo cơ quan, không ai biết | `breach_overdue` + AC #6 | vi phạm hạn | Cờ tự suy + báo động on-call |
| severity lạ | DB CHECK + AC #11 | phân loại sai | CHECK IN low/medium/high/critical |
| Mất liên kết bằng chứng nguồn | `source_ref` §1 #7 | điều tra khó | Trỏ alert/trace TASK-INFRA-004 |
| Không dựng lại được dòng thời gian | log audit §1 #12 | báo cáo cơ quan thiếu | Log mọi transition + dấu thời gian từng bước |

---

## §11 - Ghi chú

- Đồng hồ 72 giờ đếm từ `acknowledged_at` (thời điểm nhận biết) - đúng mốc PDPL, không phải từ lúc sự cố vật lý xảy ra.
- State machine tuần tự + dấu thời gian từng bước giúp dựng lại dòng thời gian sự cố khi báo cáo cơ quan.
- Phân loại severity quyết định nghĩa vụ thông báo chủ thể: high/critical không đóng được nếu chưa thông báo người ảnh hưởng.
- Cờ `breach_overdue` tự suy từ đồng hồ; báo động khi `breach_overdue_total > 0` để không bỏ lỡ mốc 72h.
- `source_ref` nối sự cố pháp lý với bằng chứng kỹ thuật từ observability spine (TASK-INFRA-004).
- Khi danh sách chủ thể ảnh hưởng lớn, bước `notified_subjects` tái dùng hạ tầng TASK-NOTIF (giai đoạn sau).

---

*Hết TASK-COMPLY-004. Status: ready_to_implement (mục tiêu audit 10/10).*
