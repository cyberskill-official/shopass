---
id: FR-COMPLY-003
title: "Quyền chủ thể dữ liệu (DSAR) - truy cập / sửa / xóa / di chuyển qua API + quy trình; SLA phản hồi, xác minh danh tính (Luật 91/2025)"
module: COMPLY
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-COMPLY-001, FR-AUTH-005, FR-AUTH-001, FR-B2B-001]
depends_on: [FR-COMPLY-001]
blocks: [FR-AUTH-005, FR-B2B-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (quyền chủ thể dữ liệu: truy cập/sửa/xóa/di chuyển)"
  - "docs/... §3.4 (data model app_user), §5.5 (PDPL Luật 91/2025)"
source_decisions:
  - "DEC-COMPLY-10: bốn quyền DSAR (access/rectify/erase/portability) qua một endpoint thống nhất + bảng dsar_request theo dõi vòng đời"
  - "DEC-COMPLY-11: yêu cầu DSAR phải xác minh danh tính chủ thể (qua phiên đăng nhập) trước khi xử lý"
  - "DEC-COMPLY-12: xóa (erase) là soft-anonymize cho dữ liệu có ràng buộc pháp lý/kế toán + hard-delete dữ liệu cá nhân thuần; KHÔNG xóa lịch sử consent (chứng cứ)"
  - "DEC-COMPLY-13: portability xuất dữ liệu cá nhân ở định dạng máy đọc được (JSON) có cấu trúc, gồm cả lịch sử consent"
  - "DEC-COMPLY-14: mỗi DSAR có SLA phản hồi; trạng thái suy từ thời điểm tạo, không nhập tay"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0005_dsar_request.sql
  - services/comply/internal/dsar/request.go
  - services/comply/internal/dsar/export.go
  - services/comply/internal/dsar/erase.go
  - services/comply/internal/dsar/repo.go
  - services/comply/internal/dsar/request_test.go
  - services/comply/internal/dsar/erase_test.go
modified_files:
  - services/comply/internal/dsar/types.go      # struct DSARRequest, ExportBundle
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - xử lý DSAR mà không xác minh danh tính chủ thể (vi phạm DEC-COMPLY-11)
  - hard-delete xóa luôn lịch sử consent (vi phạm DEC-COMPLY-12, mất chứng cứ)
  - xuất portability ở định dạng không máy đọc được (PDF ảnh) (vi phạm DEC-COMPLY-13)

effort_hours: 8
sub_tasks:
  - "0.5h: 0005_dsar_request.sql - bảng dsar_request + kind + status + sla"
  - "1.0h: types.go + request.go - tạo DSAR, xác minh danh tính, vòng đời trạng thái"
  - "1.5h: export.go - gom dữ liệu cá nhân từ các service thành ExportBundle JSON (gồm consent history)"
  - "1.5h: erase.go - soft-anonymize dữ liệu ràng buộc + hard-delete dữ liệu thuần, giữ consent log"
  - "1.0h: repo.go - lưu/đọc request, liệt kê quá SLA"
  - "1.5h: request_test.go - 4 kind; xác minh danh tính chặn truy cập chéo user"
  - "1.0h: erase_test.go - xóa anonymize app_user nhưng giữ consent_record; dữ liệu thuần biến mất"

risk_if_skipped: "Luật 91/2025 trao chủ thể dữ liệu quyền truy cập, sửa, xóa, di chuyển dữ liệu của mình. Không có quy trình DSAR -> không đáp ứng yêu cầu hợp pháp của user và vi phạm hạng High (§9), chế tài tới 3 tỷ VND cho vi phạm nghiêm trọng (§5.5). Xóa tài khoản (FR-AUTH-005) phụ thuộc trực tiếp vào erase ở đây. Nếu xóa sai cách (xóa luôn chứng cứ consent, hoặc để rò dữ liệu sang user khác khi xuất) thì tự tạo vi phạm mới. DSAR là quyền cốt lõi của PDPL và là bề mặt cơ quan quản lý kiểm tra đầu tiên."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** cung cấp quy trình thực thi bốn quyền chủ thể dữ liệu (DSAR) theo PDPL Luật 91/2025: truy cập (access), sửa (rectify), xóa (erase), di chuyển (portability) - qua API thống nhất, có xác minh danh tính, theo dõi SLA. Hợp đồng:

1. **MUST** định nghĩa bảng `dsar_request (id, user_id, kind, status, requested_at, completed_at, sla_due_at, note)` với `kind IN ('access','rectify','erase','portability')` theo dõi vòng đời từng yêu cầu (DEC-COMPLY-10).
2. **MUST** xác minh danh tính chủ thể trước khi xử lý (DEC-COMPLY-11): chỉ chủ phiên đăng nhập tạo và nhận DSAR của chính `user_id` đó. Cấm tuyệt đối truy cập DSAR chéo user.
3. **MUST** thực thi quyền truy cập (access): trả bản tóm tắt dữ liệu cá nhân đang lưu (loại dữ liệu, nguồn, mục đích) cho user.
4. **MUST** thực thi quyền di chuyển (portability) (DEC-COMPLY-13): xuất dữ liệu cá nhân ở định dạng JSON có cấu trúc, máy đọc được, gồm hồ sơ tài khoản + sản phẩm theo dõi + lịch sử consent (FR-COMPLY-001). KHÔNG xuất dạng không máy đọc được.
5. **MUST** thực thi quyền xóa (erase) theo nguyên tắc (DEC-COMPLY-12):
    - Hard-delete dữ liệu cá nhân thuần không ràng buộc (ví dụ wishlist, alert rule cá nhân).
    - Soft-anonymize dữ liệu có ràng buộc pháp lý/kế toán (ví dụ bản ghi thanh toán phải giữ theo luật kế toán) - thay PII bằng giá trị ẩn danh, không xóa dòng.
    - KHÔNG xóa `consent_record` - đó là chứng cứ pháp lý; thay vào đó ghi nhận xóa qua một DSAR erase đã hoàn tất.
6. **MUST** thực thi quyền sửa (rectify): cho phép cập nhật trường hồ sơ cá nhân (email, phone, locale) qua luồng có kiểm tra ràng buộc (ví dụ email unique của FR-AUTH-001).
7. **MUST** tính `sla_due_at` và suy trạng thái quá hạn từ thời điểm tạo, KHÔNG nhập tay (DEC-COMPLY-14): `open` / `in_progress` / `completed` / `overdue`.
8. **MUST** expose hàm:
    - `CreateRequest(ctx, userID int64, kind string) (int64, error)` - tạo DSAR, gán `sla_due_at`.
    - `Export(ctx, userID int64) (ExportBundle, error)` - gom dữ liệu cho access/portability.
    - `Erase(ctx, userID int64) (EraseResult, error)` - hard-delete + soft-anonymize theo phân loại, giữ consent.
    - `Overdue(ctx) ([]DSARRequest, error)` - liệt kê DSAR quá SLA.
9. **MUST** đảm bảo `Export` chỉ gom dữ liệu của đúng `user_id` truyền vào; property test chống rò dữ liệu chéo user (cổng tiến phase §7 BACKLOG: rò cross-user = 0).
10. **SHOULD** phát OTel metric: `dsar_request_total{kind}`, `dsar_overdue_total`; log mọi DSAR ở mức audit.
11. **MUST** ghi `completed_at` khi DSAR hoàn tất; erase sinh một DSAR `completed` làm dấu vết (đã thực thi quyền xóa lúc nào).
12. **MUST** đảm bảo `Erase` idempotent: gọi erase hai lần cho cùng user không lỗi và không để lộ dữ liệu (lần hai là no-op trên dữ liệu đã xóa).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao xác minh danh tính bắt buộc (DEC-COMPLY-11)?** DSAR thao tác trên dữ liệu cá nhân nhạy cảm. Nếu kẻ xấu xin "xuất dữ liệu của user khác" mà không bị chặn, ta tự tạo vi phạm rò rỉ. Buộc DSAR chạy trong phiên đăng nhập của chính chủ thể, `user_id` lấy từ token chứ không từ tham số tùy ý, đóng lỗ hổng truy cập chéo.

**Vì sao xóa là hỗn hợp hard-delete + soft-anonymize (DEC-COMPLY-12)?** "Xóa hết" nghe đơn giản nhưng va vào luật khác: bản ghi thanh toán phải giữ theo luật kế toán; consent_record là chứng cứ PDPL. Giải pháp: dữ liệu cá nhân thuần thì xóa hẳn; dữ liệu ràng buộc thì ẩn danh (giữ dòng, bỏ PII); chứng cứ consent thì giữ nguyên. Cách này tôn trọng cả quyền xóa lẫn nghĩa vụ lưu trữ.

**Vì sao giữ consent_record khi erase (§1 #5)?** Nghịch lý: nếu xóa luôn lịch sử consent, ta mất bằng chứng "user từng đồng ý cái gì" - chính cái PDPL bắt phải tái lập. Nên erase ghi một DSAR `completed` đánh dấu đã xóa, nhưng consent log vẫn còn để chứng minh quy trình. User được xóa dữ liệu, ta vẫn giữ được dấu vết tuân thủ.

**Vì sao portability là JSON máy đọc được (DEC-COMPLY-13)?** Quyền di chuyển nghĩa là user mang dữ liệu sang nơi khác. Xuất PDF ảnh hay HTML rối thì không "di chuyển" được. JSON có cấu trúc cho phép nhập lại vào hệ thống khác - đúng tinh thần portability.

**Vì sao SLA suy từ thời điểm tạo (DEC-COMPLY-14)?** Giống DPIA, trạng thái quá hạn phải khách quan. Tính `sla_due_at` từ `requested_at` cho ra `overdue` tự động, không để ai "quên" một DSAR vô thời hạn. PDPL kỳ vọng phản hồi trong thời hạn hợp lý; deadline rõ ràng làm điều đó thành quy trình.

**Vì sao property test chống rò chéo user (§1 #9)?** Đây là cổng tiến phase cứng của BACKLOG (rò cross-user = 0). `Export` gom dữ liệu từ nhiều bảng; một lỗi join thiếu điều kiện `user_id` là rò rỉ. Property test sinh nhiều user, xuất từng user, khẳng định bundle chỉ chứa dữ liệu đúng chủ.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/comply/migrations/0005_dsar_request.sql
CREATE TABLE dsar_request (
  id           BIGSERIAL   PRIMARY KEY,
  user_id      BIGINT      NOT NULL REFERENCES app_user(id),
  kind         TEXT        NOT NULL CHECK (kind IN ('access','rectify','erase','portability')),
  status       TEXT        NOT NULL DEFAULT 'open',
  requested_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  sla_due_at   TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ,
  note         TEXT
);

CREATE INDEX idx_dsar_user ON dsar_request (user_id, requested_at DESC);
CREATE INDEX idx_dsar_open ON dsar_request (status, sla_due_at) WHERE status <> 'completed';
```

### Export / Erase (Go)

```go
// services/comply/internal/dsar/export.go
type ExportBundle struct {
    Account        AccountView      `json:"account"`
    TrackedProducts []ProductView   `json:"tracked_products"`
    ConsentHistory []ConsentView    `json:"consent_history"` // gom tu FR-COMPLY-001
    GeneratedAt    time.Time        `json:"generated_at"`
}

// Export gom du lieu cua DUNG user_id (chong ro cheo user).
func (s *Service) Export(ctx context.Context, userID int64) (ExportBundle, error) {
    acc, err := s.users.View(ctx, userID)
    if err != nil {
        return ExportBundle{}, err
    }
    prods, err := s.track.ByUser(ctx, userID)   // moi truy van rang buoc user_id
    if err != nil {
        return ExportBundle{}, err
    }
    consent, err := s.consent.HistoryAll(ctx, userID)
    if err != nil {
        return ExportBundle{}, err
    }
    return ExportBundle{Account: acc, TrackedProducts: prods, ConsentHistory: consent, GeneratedAt: time.Now()}, nil
}

// services/comply/internal/dsar/erase.go
// Erase: hard-delete du lieu thuan, soft-anonymize du lieu rang buoc, GIU consent log.
func (s *Service) Erase(ctx context.Context, userID int64) (EraseResult, error) {
    var r EraseResult
    r.WishlistDeleted, _ = s.track.HardDeleteByUser(ctx, userID) // du lieu thuan
    r.PaymentsAnonymized, _ = s.bill.AnonymizeByUser(ctx, userID) // rang buoc ke toan
    s.users.Anonymize(ctx, userID)   // thay PII app_user bang gia tri an danh
    // KHONG dong toi consent_record (chung cu phap ly).
    return r, nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `dsar_request` tồn tại; CHECK `kind` hoạt động.
2. `CreateRequest(user, "access")` -> một DSAR, status `open`, có `sla_due_at`.
3. `Export(user)` -> bundle JSON gồm account + tracked_products + consent_history, chỉ của đúng user.
4. `Export(userA)` KHÔNG chứa bất kỳ dữ liệu nào của `userB` (property test, rò cross-user = 0).
5. `Erase(user)` -> wishlist/alert (dữ liệu thuần) bị hard-delete; bản ghi payment bị anonymize (dòng còn, PII mất).
6. Sau `Erase`, `consent_record` của user VẪN còn (chứng cứ), nhưng app_user PII đã ẩn danh.
7. `Erase` gọi lần hai cho cùng user -> no-op, không lỗi, không lộ dữ liệu (idempotent).
8. INSERT `dsar_request` với `kind = 'unknown'` -> lỗi CHECK constraint.
9. DSAR tạo cách đây quá SLA mà chưa `completed` -> `Overdue()` trả về nó.
10. DSAR hoàn tất -> `completed_at` được set; status `completed`.
11. Cố `Export` với `user_id` không phải chủ phiên (giả lập caller chéo) -> bị tầng auth chặn (xác minh danh tính).
12. `rectify` cập nhật email trùng email user khác -> lỗi ràng buộc unique (tôn trọng FR-AUTH-001), DSAR không hoàn tất.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/dsar/request_test.go
func TestExport_OnlyOwnData(t *testing.T) {
    s := setup(t)
    a := seedUserWithData(t, s) // userA + wishlist + consent
    b := seedUserWithData(t, s) // userB

    bundle, _ := s.Export(ctx, a.ID)
    require.Equal(t, a.ID, bundle.Account.UserID)
    for _, p := range bundle.TrackedProducts {
        require.NotContains(t, idsOf(b.Products), p.ID) // khong lan du lieu userB
    }
}

func TestExport_PortabilityIsJSON(t *testing.T) {
    s := setup(t)
    u := seedUserWithData(t, s)
    bundle, _ := s.Export(ctx, u.ID)
    raw, err := json.Marshal(bundle)
    require.NoError(t, err)
    require.Contains(t, string(raw), "consent_history") // may doc duoc, co consent
}

// services/comply/internal/dsar/erase_test.go
func TestErase_KeepsConsentLog(t *testing.T) {
    s := setup(t)
    u := seedUserWithData(t, s) // co consent_record
    _, err := s.Erase(ctx, u.ID)
    require.NoError(t, err)

    require.Zero(t, countWishlist(t, s, u.ID))        // du lieu thuan bi xoa
    require.True(t, isAnonymized(t, s, u.ID))          // app_user PII an danh
    require.NotZero(t, countConsentRecord(t, s, u.ID)) // consent VAN con (chung cu)
}

func TestErase_Idempotent(t *testing.T) {
    s := setup(t)
    u := seedUserWithData(t, s)
    _, err1 := s.Erase(ctx, u.ID)
    _, err2 := s.Erase(ctx, u.ID) // lan hai
    require.NoError(t, err1)
    require.NoError(t, err2) // no-op, khong loi
}

func TestDSAR_Overdue(t *testing.T) {
    s := setup(t)
    id := seedOldOpenDSAR(t, s, "access") // qua SLA
    list, _ := s.Overdue(ctx)
    require.Contains(t, idsOf(list), id)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0005 -> types.go -> export.go (gom dữ liệu, mọi truy vấn buộc `user_id`) -> erase.go (phân loại hard/soft) -> request.go (vòng đời + SLA) -> repo.go -> tests. `Export` và `Erase` gọi sang các service khác (users, track, bill, consent) qua interface inject để test mock được. Xác minh danh tính nằm ở tầng BFF/auth: `user_id` luôn lấy từ token phiên, hàm DSAR không nhận `user_id` tùy ý từ client.

---

## §7 - Phụ thuộc

- **FR-COMPLY-001** - `Export` gom `HistoryAll` (lịch sử consent); `Erase` giữ `consent_record`.
- **FR-AUTH-001** - `app_user` (anonymize PII); `rectify` tôn trọng email unique.
- **FR-AUTH-005 (downstream)** - xóa tài khoản gọi `Erase` ở đây.
- **FR-BILL-003 (liên quan)** - bản ghi payment là dữ liệu ràng buộc kế toán -> soft-anonymize.
- **FR-TRACK-002 (liên quan)** - wishlist là dữ liệu thuần -> hard-delete.
- Lib: `encoding/json`, driver `pgx`.

---

## §8 - Payload ví dụ

### User xin xuất dữ liệu (portability) - chạy trong phiên của chính họ

```json
POST /v1/dsar
{ "kind": "portability" }
```

```json
200 OK
{
  "account": { "user_id": 90112, "email": "chi@example.com", "locale": "vi-VN" },
  "tracked_products": [ { "id": 5521, "platform": "shopee", "name": "Tai nghe ABC" } ],
  "consent_history": [
    { "purpose_key": "cart_read", "granted": true, "ts": "2026-06-28T09:12:00+07:00" }
  ],
  "generated_at": "2026-06-28T10:00:00+07:00"
}
```

### User yêu cầu xóa

```json
POST /v1/dsar
{ "kind": "erase" }
```

```json
200 OK
{
  "wishlist_deleted": 12,
  "payments_anonymized": 3,
  "consent_log_retained": true,
  "status": "completed"
}
```

---

## §9 - Câu hỏi mở

Đã chốt khung. Hoãn:
- DSAR cho người không có tài khoản (khách vãng lai có dữ liệu) - hiện gắn DSAR vào app_user.
- Hàng đợi xử lý DSAR bất đồng bộ khi dữ liệu lớn (export nền + email link tải) - tối ưu khi quy mô tăng.
- Xác minh danh tính nâng cao (re-auth/OTP) cho erase - cân nhắc khi rủi ro chiếm tài khoản tăng.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Rò dữ liệu chéo user khi export | property test §1 #9 | vi phạm rò rỉ | Mọi truy vấn buộc `user_id`; test `TestExport_OnlyOwnData` |
| Caller giả `user_id` khác | tầng auth + AC #11 | truy cập chéo | `user_id` lấy từ token, không từ tham số client |
| Erase xóa luôn consent log | review erase + AC #6 | mất chứng cứ PDPL | `Erase` không đụng `consent_record`; test giữ consent |
| Erase xóa bản ghi kế toán | review phân loại | vi phạm luật kế toán | Payment soft-anonymize, không hard-delete |
| Erase lần hai gây lỗi/lộ | idempotent + AC #7 | hỏng UX/rò | No-op trên dữ liệu đã xóa |
| DSAR quá SLA không ai biết | `Overdue()` + AC #9 | vi phạm hạn phản hồi | Status suy từ `sla_due_at`; cảnh báo |
| rectify phá ràng buộc unique | DB unique + AC #12 | DB lỗi/UX xấu | Map lỗi unique, DSAR không hoàn tất |
| Portability xuất dạng không máy đọc | review §1 #4 | quyền di chuyển vô nghĩa | Ép JSON có cấu trúc |

---

## §11 - Ghi chú

- Bốn quyền DSAR (access/rectify/erase/portability) là quyền cốt lõi PDPL và là bề mặt cơ quan kiểm tra đầu tiên.
- Xác minh danh tính + truy vấn buộc `user_id` là hai lớp chống rò cross-user (cổng tiến phase = 0 rò rỉ).
- Erase hỗn hợp: hard-delete dữ liệu thuần, soft-anonymize dữ liệu ràng buộc kế toán, giữ consent log làm chứng cứ.
- Portability JSON máy đọc được giữ đúng tinh thần "di chuyển dữ liệu" sang hệ thống khác.
- Xóa tài khoản (FR-AUTH-005) là người tiêu dùng chính của `Erase`; hai FR phải nhất quán về phạm vi xóa.
- SLA suy từ `requested_at` giữ phản hồi DSAR thành quy trình có deadline, không để treo vô thời hạn.

---

*Hết FR-COMPLY-003. Status: ready_to_implement (mục tiêu audit 10/10).*
