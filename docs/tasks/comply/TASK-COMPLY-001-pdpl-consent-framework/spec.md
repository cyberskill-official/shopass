---
id: TASK-COMPLY-001
title: "Khung consent PDPL (Luật 91/2025/QH15, NĐ 356/2025) - đồng thuận tự nguyện/cụ thể/đơn mục đích/tái lập, im lặng != đồng thuận, bản ghi consent versioned"
module: COMPLY
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-INFRA-002, TASK-COMPLY-002, TASK-COMPLY-003, TASK-COMPLY-004, TASK-COMPLY-005, TASK-EXT-006, TASK-AUTH-005]
depends_on: [TASK-INFRA-002]
blocks: [TASK-COMPLY-002, TASK-COMPLY-003, TASK-COMPLY-004, TASK-COMPLY-007, TASK-COMPLY-008, TASK-EXT-006]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (PDPL Luật 91/2025, NĐ 356/2025, đồng thuận, chế tài)"
  - "docs/... §3.4 (data model app_user), §1.1 (kiến trúc trụ cột C - tuân PDPL)"
source_decisions:
  - "DEC-COMPLY-01: cơ sở pháp lý lưu xử lý dữ liệu là consent versioned theo từng mục đích (purpose); im lặng/checkbox tích sẵn KHÔNG phải đồng thuận"
  - "DEC-COMPLY-02: mỗi consent gắn đúng một purpose_key (đơn mục đích); gộp nhiều mục đích vào một lần bấm là vi phạm tính cụ thể"
  - "DEC-COMPLY-03: consent_policy versioned (immutable); thu hồi (withdraw) ghi dòng mới chứ KHÔNG xóa lịch sử (tái lập được)"
  - "DEC-COMPLY-04: bản ghi consent lưu purpose_key + policy_version + granted + source + ts + ip/user_agent để chứng minh trước cơ quan"
  - "DEC-COMPLY-05: PDPL hiệu lực 01/01/2026 (KHÔNG phải 01/07/2026, đã đính chính §10); NĐ 356/2025/NĐ-CP thay NĐ 13/2023"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0001_consent_policy.sql
  - services/comply/migrations/0002_consent_record.sql
  - services/comply/internal/consent/policy.go
  - services/comply/internal/consent/record.go
  - services/comply/internal/consent/repo.go
  - services/comply/internal/consent/service.go
  - services/comply/internal/consent/repo_test.go
  - services/comply/internal/consent/service_test.go
modified_files:
  - services/comply/internal/consent/types.go      # struct ConsentPolicy, ConsentRecord, Purpose
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - checkbox tích sẵn (pre-checked) hoặc opt-out mặc định coi là đồng thuận (vi phạm DEC-COMPLY-01)
  - gộp nhiều purpose vào một bản ghi consent duy nhất (vi phạm DEC-COMPLY-02 đơn mục đích)
  - xóa cứng (hard delete) bản ghi consent khi user thu hồi (vi phạm DEC-COMPLY-03 tái lập)

effort_hours: 8
sub_tasks:
  - "0.5h: 0001_consent_policy.sql - bảng consent_policy versioned (immutable) + seed purpose lõi"
  - "0.5h: 0002_consent_record.sql - bảng consent_record append-only + index theo (user, purpose)"
  - "1.0h: types.go + policy.go - enum Purpose, load policy version đang hiệu lực"
  - "1.5h: record.go + repo.go - Grant/Withdraw append-only, đọc trạng thái hiện tại"
  - "1.0h: service.go - IsAllowed(user, purpose) cho các service khác gọi như cổng pháp lý"
  - "1.0h: validate đầu vào - từ chối purpose lạ, từ chối policy_version cũ hơn bản hiệu lực"
  - "1.5h: repo_test.go - grant, withdraw ghi dòng mới, lịch sử giữ nguyên, đơn mục đích"
  - "1.0h: service_test.go - im lặng != đồng thuận, withdraw -> IsAllowed=false"

risk_if_skipped: "Không có khung consent đúng PDPL là rủi ro pháp lý hạng High (§9). Luật 91/2025 yêu cầu đồng thuận tự nguyện, cụ thể, đơn mục đích, tái lập được; im lặng không phải đồng thuận. Chế tài tới 5% doanh thu năm trước cho vi phạm xuyên biên giới, tới 3 tỷ VND cho vi phạm nghiêm trọng (§5.5). Mọi bề mặt xử lý dữ liệu cá nhân (extension đọc giỏ hàng, đăng ký, alert) phải tham chiếu bản ghi consent này làm cơ sở pháp lý. Đây là task nền của toàn bộ module COMPLY: DPIA (TASK-COMPLY-002), DSAR (TASK-COMPLY-003), breach (TASK-COMPLY-004), consent extension (TASK-EXT-006) đều dựng trên nó."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** cung cấp khung quản lý đồng thuận (consent) tuân thủ PDPL Luật 91/2025/QH15 và Nghị định 356/2025/NĐ-CP: mỗi mục đích xử lý dữ liệu cá nhân có một bản ghi consent versioned, tái lập được, và mọi service khác **MUST** kiểm tra consent này trước khi xử lý. Hợp đồng:

1. **MUST** định nghĩa bảng `consent_policy (id, purpose_key, version, title_vi, body_vi, effective_from, created_at)` - mỗi (purpose_key, version) là một bản chính sách bất biến (immutable). Sửa nội dung là tạo version mới, KHÔNG update tại chỗ (DEC-COMPLY-03).
2. **MUST** định nghĩa bảng `consent_record (id, user_id, purpose_key, policy_version, granted, source, ts, ip, user_agent)` append-only: mỗi lần grant hoặc withdraw là một dòng MỚI; KHÔNG xóa, KHÔNG update dòng cũ (DEC-COMPLY-03, tái lập được).
3. **MUST** ràng buộc đơn mục đích (DEC-COMPLY-02): một `consent_record` gắn đúng một `purpose_key`. Cấm gộp nhiều mục đích vào một lần bấm "Đồng ý tất cả". Mỗi mục đích cần một lần đồng thuận riêng, cụ thể.
4. **MUST** coi đồng thuận là tự nguyện và chủ động (DEC-COMPLY-01): chỉ `granted = true` khi user CHỦ ĐỘNG bật. Checkbox tích sẵn, opt-out mặc định, hoặc im lặng KHÔNG được coi là đồng thuận. Giá trị mặc định của mọi purpose là chưa đồng thuận.
5. **MUST** lưu đủ trường chứng minh (DEC-COMPLY-04): `purpose_key`, `policy_version` (đúng bản user đã thấy), `source` (web/extension/mobile), `ts`, `ip`, `user_agent`. Đây là bằng chứng trước cơ quan quản lý rằng đồng thuận hợp lệ.
6. **MUST** expose hàm service cho các service khác gọi như cổng pháp lý:
- `Grant(ctx, userID int64, purpose string, src string, meta ReqMeta) error` - ghi dòng granted=true với policy_version đang hiệu lực.
- `Withdraw(ctx, userID int64, purpose string, src string, meta ReqMeta) error` - ghi dòng granted=false (thu hồi), KHÔNG xóa lịch sử.
- `IsAllowed(ctx, userID int64, purpose string) (bool, error)` - trả trạng thái hiệu lực hiện tại (dòng mới nhất của (user, purpose)); chưa có bản ghi -> false.
- `History(ctx, userID int64, purpose string) ([]ConsentRecord, error)` - toàn bộ lịch sử cho DSAR (TASK-COMPLY-003) và kiểm toán.
7. **MUST** từ chối `purpose_key` không nằm trong tập đã đăng ký (enum `Purpose`); từ chối `policy_version` cũ hơn bản đang hiệu lực (user phải đồng thuận trên bản mới nhất).
8. **MUST** seed tập purpose lõi tối thiểu: `cart_read` (extension đọc giỏ hàng/voucher), `price_tracking` (theo dõi giá theo tài khoản), `marketing_notification` (gửi alert/khuyến mãi), `analytics_b2b` (đóng góp dữ liệu xu hướng ẩn danh). Mỗi purpose là một bản `consent_policy` riêng.
9. **MUST** đảm bảo idempotent với double-submit: grant hai lần liên tiếp cùng trạng thái KHÔNG tạo ra mâu thuẫn ngữ nghĩa (dòng mới nhất vẫn phản ánh đúng); trạng thái suy ra từ dòng mới nhất theo `ts`.
10. **SHOULD** phát OTel metric: `consent_granted_total{purpose}`, `consent_withdrawn_total{purpose}` (counter); log mọi grant/withdraw ở mức audit (KHÔNG log nội dung dữ liệu cá nhân kèm theo).
11. **MUST** bảo đảm `IsAllowed` là nguồn sự thật duy nhất cho cơ sở pháp lý: các service như EXT (TASK-EXT-006), TRACK, NOTIF gọi `IsAllowed` trước khi xử lý dữ liệu của purpose tương ứng; thiếu consent -> từ chối xử lý.
12. **MUST** ghi `policy_version` đúng bản user đã xem lúc bấm; nếu chính sách đổi sau đó, consent cũ vẫn gắn version cũ (không tự "nâng cấp" sang version mới - đó là tái lập được).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao consent versioned bất biến (DEC-COMPLY-03)?** PDPL yêu cầu đồng thuận "tái lập được": khi cơ quan quản lý hỏi "user đồng ý cái gì, ngày nào, trên bản chính sách nào", phải trả lời chính xác. Nếu update nội dung chính sách tại chỗ, ta mất bằng chứng user đã thấy gì lúc bấm. Tách `consent_policy` thành bản version bất biến + `consent_record` trỏ tới đúng version giải đúng yêu cầu này.

**Vì sao append-only cho consent_record (§1 #2)?** Thu hồi (withdraw) không được xóa dấu vết "đã từng đồng ý". Ghi dòng mới `granted=false` giữ trọn lịch sử: cấp lúc nào, thu hồi lúc nào. Đây vừa là yêu cầu tái lập của PDPL, vừa là dữ liệu cho DSAR (user có quyền xem lịch sử đồng thuận của mình).

**Vì sao đơn mục đích (DEC-COMPLY-02)?** Luật 91/2025 đòi đồng thuận "cụ thể" và "đơn mục đích". Gộp "đọc giỏ hàng + gửi marketing + bán dữ liệu B2B" vào một nút "Đồng ý" là đồng thuận mơ hồ, không hợp lệ. Tách purpose để user đồng ý chọn lọc: có thể bật theo dõi giá nhưng tắt marketing.

**Vì sao mặc định là chưa đồng thuận, im lặng != đồng thuận (DEC-COMPLY-01)?** Đây là điểm cốt lõi PDPL khác với "implied consent". Checkbox tích sẵn hay "tiếp tục dùng tức là đồng ý" KHÔNG hợp lệ. Hệ thống phải bắt đầu từ trạng thái false và chỉ chuyển true khi có hành động chủ động. `IsAllowed` trả false khi chưa có bản ghi - không cho phép xử lý theo mặc định.

**Vì sao IsAllowed là cổng pháp lý dùng chung (§1 #11)?** Tập trung cơ sở pháp lý vào một hàm tránh việc mỗi service tự diễn giải consent một kiểu. EXT đọc giỏ hàng phải hỏi `IsAllowed(cart_read)`; NOTIF gửi marketing phải hỏi `IsAllowed(marketing_notification)`. Một nguồn sự thật giảm rủi ro lọt bề mặt xử lý không có cơ sở pháp lý (cổng tiến phase §7 BACKLOG: coverage consent = 100%).

**Vì sao lưu ip/user_agent/source (DEC-COMPLY-04)?** Bằng chứng đồng thuận cần ngữ cảnh: ai bấm, từ kênh nào, lúc nào. Khi tranh chấp, bộ trường này chứng minh đồng thuận là hành động thật của chủ thể dữ liệu, không phải hệ thống tự gán.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/comply/migrations/0001_consent_policy.sql
CREATE TABLE consent_policy (
  id             BIGSERIAL   PRIMARY KEY,
  purpose_key    TEXT        NOT NULL,
  version        INTEGER     NOT NULL,
  title_vi       TEXT        NOT NULL,
  body_vi        TEXT        NOT NULL,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (purpose_key, version)
);

-- Seed purpose loi (moi purpose mot ban chinh sach v1).
INSERT INTO consent_policy (purpose_key, version, title_vi, body_vi) VALUES
  ('cart_read', 1, 'Doc gio hang va voucher cua ban',
   'SanDeal doc gio hang/voucher tren trinh duyet cua ban de toi uu. KHONG gui cookie/token.'),
  ('price_tracking', 1, 'Theo doi gia theo tai khoan',
   'Luu san pham ban theo doi de canh bao khi gia thay doi.'),
  ('marketing_notification', 1, 'Nhan thong bao khuyen mai',
   'Gui alert sale va goi y san pham. Co the thu hoi bat ky luc nao.'),
  ('analytics_b2b', 1, 'Dong gop du lieu xu huong an danh',
   'Du lieu gia da an danh hoa (k-anonymity) phuc vu bao cao thi truong.');

-- services/comply/migrations/0002_consent_record.sql
CREATE TABLE consent_record (
  id             BIGSERIAL   PRIMARY KEY,
  user_id        BIGINT      NOT NULL REFERENCES app_user(id),
  purpose_key    TEXT        NOT NULL,
  policy_version INTEGER     NOT NULL,
  granted        BOOLEAN     NOT NULL,
  source         TEXT        NOT NULL CHECK (source IN ('web','extension','mobile')),
  ts             TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip             INET,
  user_agent     TEXT,
  FOREIGN KEY (purpose_key, policy_version)
    REFERENCES consent_policy(purpose_key, version)
);

-- Tra trang thai hieu luc hien tai = dong moi nhat theo ts cho moi (user, purpose).
CREATE INDEX idx_consent_latest ON consent_record (user_id, purpose_key, ts DESC);
```

### Types (Go)

```go
// services/comply/internal/consent/types.go
type Purpose string

const (
    PurposeCartRead    Purpose = "cart_read"
    PurposeTracking    Purpose = "price_tracking"
    PurposeMarketing   Purpose = "marketing_notification"
    PurposeAnalyticsB2B Purpose = "analytics_b2b"
)

type ConsentRecord struct {
    ID            int64     `db:"id"`
    UserID        int64     `db:"user_id"`
    PurposeKey    string    `db:"purpose_key"`
    PolicyVersion int32     `db:"policy_version"`
    Granted       bool      `db:"granted"`
    Source        string    `db:"source"`
    TS            time.Time `db:"ts"`
    IP            *netip.Addr `db:"ip"`
    UserAgent     *string   `db:"user_agent"`
}

type ReqMeta struct {
    IP        *netip.Addr
    UserAgent *string
}
```

### Service (cổng pháp lý)

```go
// services/comply/internal/consent/service.go
// IsAllowed la nguon su that duy nhat cho co so phap ly cua mot purpose.
func (s *Service) IsAllowed(ctx context.Context, userID int64, p Purpose) (bool, error) {
    if !validPurpose(p) {
        return false, ErrUnknownPurpose
    }
    rec, err := s.repo.latest(ctx, userID, string(p))
    if errors.Is(err, pgx.ErrNoRows) {
        return false, nil // chua co ban ghi -> chua dong thuan (im lang != dong thuan)
    }
    if err != nil {
        return false, err
    }
    return rec.Granted, nil
}

// Grant ghi dong moi granted=true voi policy_version dang hieu luc.
func (s *Service) Grant(ctx context.Context, userID int64, p Purpose, src string, m ReqMeta) error {
    if !validPurpose(p) {
        return ErrUnknownPurpose
    }
    ver, err := s.repo.effectiveVersion(ctx, string(p))
    if err != nil {
        return err
    }
    return s.repo.append(ctx, ConsentRecord{
        UserID: userID, PurposeKey: string(p), PolicyVersion: ver,
        Granted: true, Source: src, IP: m.IP, UserAgent: m.UserAgent,
    })
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `consent_policy` và `consent_record` tồn tại; 4 purpose lõi được seed (version 1).
2. `Grant(user, cart_read)` -> một dòng `consent_record` granted=true, policy_version = bản hiệu lực, đủ source/ts.
3. `IsAllowed(user, cart_read)` sau grant -> true.
4. `IsAllowed(user, marketing_notification)` khi chưa có bản ghi -> false (im lặng != đồng thuận).
5. `Withdraw(user, cart_read)` -> ghi dòng MỚI granted=false; dòng granted=true cũ vẫn còn (append-only).
6. `IsAllowed(user, cart_read)` sau withdraw -> false.
7. `History(user, cart_read)` trả cả hai dòng (grant rồi withdraw) theo thứ tự thời gian.
8. `Grant(user, "bad_purpose")` -> lỗi `ErrUnknownPurpose`, không ghi dòng nào.
9. Cố INSERT `consent_record` với `policy_version` không tồn tại -> lỗi FK.
10. Cố INSERT `source = 'sms'` -> lỗi CHECK constraint.
11. Sau khi tạo `consent_policy (cart_read, version 2)`, `Grant` ghi version 2; consent cũ vẫn giữ version 1 (không tự nâng cấp).
12. Metric `consent_granted_total{purpose}` tăng khi grant; `consent_withdrawn_total{purpose}` tăng khi withdraw.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/consent/service_test.go
func TestConsent_SilenceIsNotConsent(t *testing.T) {
    s, uid := setupWithUser(t)
    ok, err := s.IsAllowed(ctx, uid, PurposeMarketing)
    require.NoError(t, err)
    require.False(t, ok) // chua co ban ghi -> false, khong mac dinh dong thuan
}

func TestConsent_GrantThenAllowed(t *testing.T) {
    s, uid := setupWithUser(t)
    require.NoError(t, s.Grant(ctx, uid, PurposeCartRead, "extension", ReqMeta{}))
    ok, _ := s.IsAllowed(ctx, uid, PurposeCartRead)
    require.True(t, ok)
}

func TestConsent_WithdrawAppendsRow_KeepsHistory(t *testing.T) {
    s, uid := setupWithUser(t)
    s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{})
    require.NoError(t, s.Withdraw(ctx, uid, PurposeCartRead, "web", ReqMeta{}))

    ok, _ := s.IsAllowed(ctx, uid, PurposeCartRead)
    require.False(t, ok) // dong moi nhat granted=false

    h, _ := s.History(ctx, uid, PurposeCartRead)
    require.Len(t, h, 2) // grant + withdraw, khong xoa lich su
}

func TestConsent_UnknownPurpose_Rejected(t *testing.T) {
    s, uid := setupWithUser(t)
    err := s.Grant(ctx, uid, Purpose("bad_purpose"), "web", ReqMeta{})
    require.ErrorIs(t, err, ErrUnknownPurpose)
}

func TestConsent_OldConsentKeepsOldVersion(t *testing.T) {
    s, uid := setupWithUser(t)
    s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{}) // version 1
    seedPolicyVersion(t, s, "cart_read", 2)
    s.Grant(ctx, uid, PurposeCartRead, "web", ReqMeta{}) // version 2
    h, _ := s.History(ctx, uid, PurposeCartRead)
    require.Equal(t, int32(1), h[0].PolicyVersion)
    require.Equal(t, int32(2), h[1].PolicyVersion) // khong tu nang cap dong cu
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0001 (consent_policy + seed) -> 0002 (consent_record) -> types.go + policy.go -> record.go + repo.go (append-only) -> service.go (cổng IsAllowed) -> tests. `consent_record` là append-only ở tầng ứng dụng: repo chỉ có `append`, KHÔNG có update/delete. Trạng thái hiện tại suy ra qua truy vấn dòng mới nhất (index `idx_consent_latest`). Các service khác inject `consent.Service` và gọi `IsAllowed` trước mọi xử lý dữ liệu cá nhân.

---

## §7 - Phụ thuộc

- **TASK-INFRA-002** - `app_user` phải tồn tại trước (FK `user_id`).
- **TASK-EXT-006 (downstream)** - UI consent lúc cài extension gọi `Grant(cart_read)`; đọc giỏ hàng kiểm `IsAllowed(cart_read)`.
- **TASK-COMPLY-003 (downstream)** - DSAR xuất `History` của user; xóa tài khoản tham chiếu trạng thái consent.
- **TASK-AUTH-005 (downstream)** - vòng đời tài khoản (xóa) phối hợp thu hồi consent.
- **TASK-COMPLY-002 (downstream)** - DPIA tham chiếu danh mục purpose làm cơ sở đánh giá tác động.
- Lib: driver `pgx`; không phụ thuộc dịch vụ ngoài.

---

## §8 - Payload ví dụ

### Extension xin consent đọc giỏ hàng lúc cài (qua BFF)

```json
POST /v1/consent/grant
{
  "purpose": "cart_read",
  "source": "extension"
}
```

```json
200 OK
{
  "purpose": "cart_read",
  "policy_version": 1,
  "granted": true,
  "ts": "2026-06-28T09:12:00+07:00"
}
```

### Service nội bộ kiểm cổng pháp lý trước khi xử lý

```go
ok, err := complySvc.IsAllowed(ctx, userID, consent.PurposeCartRead)
if err != nil || !ok {
    return ErrNoLegalBasis // tu choi xu ly khi thieu dong thuan
}
// ... tiep tuc doc gio hang
```

---

## §9 - Câu hỏi mở

Đã chốt khung lõi. Hoãn:
- Consent cho người chưa thành niên (xác minh tuổi/giám hộ) - slice sau khi có luồng định danh tuổi.
- Consent banner đa ngôn ngữ (EN/ID/TH) - gắn vào TASK-COMPLY-007 khi mở SEA.
- Đồng bộ thu hồi consent xuyên thiết bị realtime (push withdraw tới extension đang mở) - tối ưu UX giai đoạn sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Checkbox tích sẵn lọt vào UI | review + test client | đồng thuận vô hiệu, vi phạm PDPL | UI bắt đầu false; `Grant` chỉ gọi khi user chủ động bấm |
| Gộp nhiều purpose một nút | review §1 #3 | đồng thuận không cụ thể | Một purpose một bản ghi; tách nút trong TASK-EXT-006 |
| Service quên gọi IsAllowed | audit coverage §7 BACKLOG | xử lý không cơ sở pháp lý | Cổng tiến phase yêu cầu coverage 100%; review từng bề mặt |
| Xóa cứng consent khi withdraw | review repo (không có delete) | mất tái lập | Repo chỉ `append`; withdraw ghi dòng false |
| policy_version không khớp bản user thấy | FK + AC #11 | bằng chứng sai lệch | Ghi đúng version hiệu lực lúc grant; FK chặn version lạ |
| purpose_key gõ sai ở caller | `ErrUnknownPurpose` + AC #8 | từ chối ghi | Dùng hằng `Purpose`, không truyền chuỗi tự do |
| Double-submit grant | AC, idempotent §1 #9 | dòng trùng vô hại | Trạng thái = dòng mới nhất; trùng không đổi ngữ nghĩa |
| Log lộ dữ liệu cá nhân kèm consent | review log §1 #10 | rò rỉ PII | Chỉ log purpose/ts/granted, không log nội dung dữ liệu |

---

## §11 - Ghi chú

- Đây là task nền pháp lý của SănDeal: mọi bề mặt xử lý dữ liệu cá nhân lấy cơ sở pháp lý từ `IsAllowed`.
- PDPL hiệu lực 01/01/2026 (đã đính chính so với đề bài ghi 01/07/2026, §10 tài liệu nguồn); NĐ 356/2025/NĐ-CP thay NĐ 13/2023.
- Bốn nguyên tắc PDPL được mã hóa trực tiếp: tự nguyện (chỉ grant khi chủ động), cụ thể + đơn mục đích (một purpose một bản ghi), tái lập được (append-only + policy versioned).
- Im lặng không phải đồng thuận: mặc định false, `IsAllowed` trả false khi chưa có bản ghi. Đây là khác biệt cốt lõi so với implied consent.
- Tập purpose mở rộng được: thêm purpose mới = thêm bản `consent_policy`, không đổi schema.
- Khi mở SEA, adapter TASK-COMPLY-007 tái dùng khung này, chỉ thay nội dung chính sách và bổ sung yêu cầu địa phương (Indonesia PDP, Thailand PDPA).

---

*Hết TASK-COMPLY-001. Status: ready_to_implement (mục tiêu audit 10/10).*
