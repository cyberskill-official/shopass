---
id: FR-AUTH-005
title: "Account lifecycle - reset mật khẩu, status active/suspended/deleted, xóa tài khoản thỏa DSAR PDPL"
module: AUTH
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-AUTH-001, FR-AUTH-002, FR-AUTH-003, FR-COMPLY-003, FR-NOTIF-001]
depends_on: [FR-AUTH-001, FR-COMPLY-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (PDPL: quyền chủ thể dữ liệu - xóa, DSAR)"
  - "docs/... §3.4 (app_user.status), §3.8 (bảo mật reset)"
source_decisions:
  - "DEC-AUTH-21: status app_user - {active, suspended, deleted}; chuyển trạng thái có quy tắc"
  - "DEC-AUTH-22: reset mật khẩu qua token một-lần, hết hạn ngắn, gửi qua kênh đã verify (email/phone); KHÔNG lộ tài khoản tồn tại hay không"
  - "DEC-AUTH-23: xóa tài khoản (DSAR PDPL) - xóa/ẩn danh hóa dữ liệu cá nhân; thu hồi mọi token; gỡ platform_account"
  - "DEC-AUTH-24: suspended/deleted KHÔNG đăng nhập được; refresh token bị thu hồi khi đổi trạng thái"
  - "DEC-AUTH-25: xóa là dứt khoát theo nghĩa PDPL nhưng có cửa sổ ân hạn ngắn (grace) trước khi purge cứng, chống xóa nhầm"

language: "PostgreSQL 16 + Go 1.22 (auth-svc); phối FR-COMPLY-003 (DSAR), FR-NOTIF-001 (gửi reset)"
service: shopass/services/auth/
new_files:
  - services/auth/migrations/0008_password_reset.up.sql
  - services/auth/migrations/0008_password_reset.down.sql
  - services/auth/internal/auth/reset.go
  - services/auth/internal/auth/lifecycle.go
  - services/auth/internal/auth/erasure.go
  - services/auth/internal/auth/reset_test.go
  - services/auth/internal/auth/lifecycle_test.go
  - services/auth/internal/auth/erasure_test.go
modified_files:
  - services/auth/internal/auth/login.go            # chặn đăng nhập khi status != active
allowed_tools:
  - file_read: services/auth/**
  - file_write: services/auth/**
  - bash: cd services/auth && go test ./...
disallowed_tools:
  - reset trả thông báo khác nhau cho email tồn tại vs không (vi phạm DEC-AUTH-22; user enumeration)
  - để suspended/deleted vẫn đăng nhập hay giữ refresh token còn hiệu lực (vi phạm DEC-AUTH-24)
  - xóa tài khoản mà bỏ sót dữ liệu cá nhân/không thu hồi token (vi phạm DEC-AUTH-23, PDPL)

effort_hours: 5
sub_tasks:
  - "0.5h: 0008_password_reset - bảng password_reset (user_id, token_hash, expires_at, used_at)"
  - "1.0h: reset.go - RequestReset (gửi token qua NOTIF, không lộ tồn tại) + ConfirmReset (token một-lần)"
  - "1.0h: lifecycle.go - SetStatus(active/suspended/deleted) + thu hồi refresh token khi đổi"
  - "1.0h: erasure.go - DeleteAccount (DSAR): ẩn danh hóa/xóa PII, gỡ platform_account, thu hồi token, đặt deleted"
  - "0.75h: reset_test.go - reset token một-lần; hết hạn; phản hồi đồng nhất bất kể email tồn tại"
  - "0.75h: lifecycle_test.go - suspended/deleted không đăng nhập; refresh bị thu hồi khi đổi status"
  - "1.0h: erasure_test.go - sau xóa: PII biến mất/ẩn danh; platform_account gỡ; token vô hiệu; đăng nhập chặn"

risk_if_skipped: "PDPL (§5.5) trao quyền chủ thể dữ liệu gồm quyền xóa; không có đường xóa thỏa DSAR là vi phạm trực tiếp (chế tài tới 5% doanh thu / 3 tỷ VND). Reset mật khẩu làm sai (lộ tài khoản tồn tại, token đoán được, không hết hạn) là cửa chiếm tài khoản. Không thu hồi token khi suspend/delete thì người bị khóa/đã xóa vẫn dùng được phiên cũ. Đây là phần vòng đời bắt buộc cho cả compliance lẫn bảo mật vận hành."
---

## §1 - Mô tả (BCP-14 normative)

Service AUTH **MUST** quản vòng đời tài khoản: reset mật khẩu an toàn, chuyển trạng thái có quy tắc, và xóa tài khoản thỏa DSAR PDPL, đảm bảo token bị thu hồi đúng lúc. Hợp đồng:

1. `app_user.status` **MUST** - {`active`, `suspended`, `deleted`} (DEC-AUTH-21). Chỉ `active` mới đăng nhập được; `suspended` và `deleted` bị chặn ở đường đăng nhập (sửa `login.go`).
2. Reset mật khẩu **MUST** dùng token một-lần, hết hạn ngắn (ví dụ 30 phút), gửi qua kênh đã verify (email/phone) qua FR-NOTIF-001 (DEC-AUTH-22). Token lưu dạng hash trong bảng `password_reset`.
3. `RequestReset(identifier)` **MUST** trả phản hồi ĐỒNG NHẤT bất kể tài khoản tồn tại hay không (DEC-AUTH-22): KHÔNG lộ "email này có tài khoản" (chống user enumeration).
4. `ConfirmReset(token, newPassword)` **MUST** kiểm token hợp lệ, chưa dùng, chưa hết hạn; đặt mật khẩu mới (băm argon2id qua FR-AUTH-001); đánh dấu token đã dùng; và **MUST** thu hồi mọi refresh token hiện hữu của user (đổi mật khẩu = đăng xuất mọi phiên).
5. `SetStatus(userID, status)` **MUST** chuyển trạng thái theo quy tắc và **MUST** thu hồi mọi refresh token của user khi chuyển sang `suspended` hoặc `deleted` (DEC-AUTH-24).
6. Đăng nhập **MUST** từ chối khi `status != 'active'` với lỗi xác định (ví dụ `ErrAccountNotActive`), KHÔNG tiết lộ chi tiết nội bộ.
7. `DeleteAccount(userID)` (DSAR) **MUST** (DEC-AUTH-23): xóa hoặc ẩn danh hóa dữ liệu cá nhân của `app_user` (email, phone, display_name); gỡ mọi `platform_account` (FR-AUTH-003); thu hồi mọi refresh token; đặt `status='deleted'`.
8. Xóa **MUST** dứt khoát theo nghĩa PDPL nhưng **SHOULD** có cửa sổ ân hạn ngắn (ví dụ 7-30 ngày) trước khi purge cứng, chống xóa nhầm (DEC-AUTH-25); trong ân hạn tài khoản đã ở `deleted` và không đăng nhập được.
8b. `DeleteAccount` **MUST** phối với FR-COMPLY-003 (DSAR): là điểm thực thi quyền xóa; đảm bảo dữ liệu cá nhân ở các bảng liên quan cũng được xử lý theo chính sách lưu trữ.
9. Sau xóa, dữ liệu cá nhân trực tiếp của `app_user` **MUST** không còn truy được ở dạng định danh (email/phone bị nullify hoặc thay bằng tombstone ẩn danh).
10. Reset và xóa **MUST** KHÔNG log token reset thô hay PII; thao tác nhạy cảm ghi audit (ai/khi/loại) nhưng không kèm dữ liệu nhạy cảm thô.
11. Mọi chuyển trạng thái và xóa **SHOULD** phát sự kiện/metric (`account_status_changed_total{to}`, `account_deleted_total`) cho observability và đối chiếu DSAR.
12. `ConfirmReset` và `DeleteAccount` **MUST** idempotent ở mức an toàn: token reset đã dùng không dùng lại được; xóa tài khoản đã xóa không gây lỗi nghiêm trọng.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao reset không lộ tài khoản tồn tại (DEC-AUTH-22, §1 #3)? Nếu `RequestReset("a@x.com")` trả "đã gửi" còn `RequestReset("b@x.com")` trả "không có tài khoản", kẻ tấn công dò được email nào đã đăng ký (user enumeration) - bước đầu của tấn công có mục tiêu. Phản hồi đồng nhất ("nếu tài khoản tồn tại, ta đã gửi hướng dẫn") che thông tin đó.

Vì sao token reset một-lần, hết hạn ngắn (§1 #2)? Link reset đi qua email/SMS - kênh có thể bị lưu, chuyển tiếp, hay lộ. Token sống lâu hoặc dùng nhiều lần là cửa sổ rộng để lạm dụng. Một-lần + hết hạn ngắn thu hẹp cửa sổ tới mức một lần dùng hợp lệ của chủ thật.

Vì sao đổi mật khẩu thu hồi mọi phiên (§1 #4)? Người ta đổi mật khẩu thường vì nghi bị lộ. Nếu các phiên (refresh token) cũ vẫn sống, kẻ chiếm phiên vẫn vào được dù mật khẩu đã đổi - phản tác dụng của việc reset. Thu hồi mọi refresh khi đổi mật khẩu đảm bảo reset thật sự cắt truy cập cũ.

Vì sao suspend/delete phải thu hồi token (DEC-AUTH-24)? Status là cờ ở DB; nhưng access token tự chứng thực (gateway không truy DB mỗi request). Nếu chỉ đổi status mà không thu hồi refresh, người bị khóa vẫn refresh ra access mới và dùng tiếp. Thu hồi refresh khi đổi status (cộng access TTL ngắn của FR-AUTH-002) làm việc khóa có hiệu lực thực.

Vì sao xóa có cửa sổ ân hạn (DEC-AUTH-25)? PDPL đòi quyền xóa, nhưng xóa nhầm/bị lừa xóa là rủi ro thật. Một cửa sổ ân hạn ngắn (tài khoản đã ở `deleted`, không đăng nhập được, nhưng chưa purge cứng) cân bằng: quyền xóa được tôn trọng ngay (không còn dùng được, PII đã ẩn danh ở bước đầu) mà vẫn có lối khôi phục nếu là nhầm lẫn, trước khi xóa vật lý.

Vì sao ẩn danh hóa thay vì chỉ đặt cờ (§1 #7, #9)? Đặt `status='deleted'` mà giữ nguyên email/phone không thỏa quyền xóa PDPL - dữ liệu cá nhân vẫn nằm đó. Phải nullify hoặc thay bằng tombstone ẩn danh để dữ liệu cá nhân trực tiếp thực sự không còn truy được, đúng tinh thần "xóa".

---

## §3 - Hợp đồng API / DDL

### Migration password_reset

```sql
-- services/auth/migrations/0008_password_reset.up.sql
CREATE TABLE password_reset (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  token_hash  TEXT        NOT NULL UNIQUE,   -- hash token reset, KHÔNG cleartext
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,                   -- NULL = chưa dùng
  created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_pr_user ON password_reset (user_id);
```

### Reset không lộ tồn tại + một-lần (§1 #2, #3, #4)

```go
// services/auth/internal/auth/reset.go
func (s *Service) RequestReset(ctx context.Context, identifier string) error {
    u, ok := s.repo.FindByIdentifier(ctx, identifier)
    if ok && u.Status == "active" {
        token := randToken()
        s.repo.SaveReset(ctx, u.ID, hash(token), time.Now().Add(30*time.Minute))
        s.notif.SendReset(ctx, u, token) // gửi qua kênh verified (FR-NOTIF-001)
    }
    return nil // LUÔN trả nil — phản hồi đồng nhất (§1 #3), không lộ tồn tại
}

func (s *Service) ConfirmReset(ctx context.Context, token, newPassword string) error {
    pr, ok := s.repo.FindResetByHash(ctx, hash(token))
    if !ok || pr.UsedAt != nil || time.Now().After(pr.ExpiresAt) {
        return ErrInvalidResetToken
    }
    if err := checkPasswordStrength(newPassword); err != nil { return err }
    hash, _ := Hash(newPassword, s.params)
    s.repo.UpdatePassword(ctx, pr.UserID, hash)
    s.repo.MarkResetUsed(ctx, pr.ID)           // một-lần (§1 #2)
    return s.repo.RevokeAllRefresh(ctx, pr.UserID) // đổi mật khẩu = đăng xuất mọi phiên (§1 #4)
}
```

### Lifecycle + erasure (§1 #5, #7, #9)

```go
// services/auth/internal/auth/lifecycle.go
func (s *Service) SetStatus(ctx context.Context, userID int64, status string) error {
    if !validStatus(status) { return ErrBadStatus }
    if err := s.repo.SetStatus(ctx, userID, status); err != nil { return err }
    if status == "suspended" || status == "deleted" {
        return s.repo.RevokeAllRefresh(ctx, userID) // thu hồi token (§1 #5)
    }
    return nil
}

// services/auth/internal/auth/erasure.go
func (s *Service) DeleteAccount(ctx context.Context, userID int64) error {
    // DSAR PDPL: ẩn danh hóa PII + gỡ liên kết + thu hồi token + đặt deleted (§1 #7,#9).
    if err := s.repo.AnonymizePII(ctx, userID); err != nil { return err }   // email/phone/display_name -> tombstone
    if err := s.repo.DeletePlatformAccounts(ctx, userID); err != nil { return err } // FR-AUTH-003
    if err := s.repo.RevokeAllRefresh(ctx, userID); err != nil { return err }
    return s.repo.SetStatus(ctx, userID, "deleted") // ân hạn trước purge cứng (§1 #8)
}
```

---

## §4 - Acceptance criteria

1. `RequestReset` cho email tồn tại (active) và email không tồn tại -> cùng kiểu phản hồi (không lộ tồn tại).
2. `ConfirmReset` với token hợp lệ -> đặt mật khẩu mới; đăng nhập bằng mật khẩu mới thành công.
3. `ConfirmReset` dùng lại token đã dùng -> `ErrInvalidResetToken`.
4. `ConfirmReset` với token hết hạn -> `ErrInvalidResetToken`.
5. Sau `ConfirmReset`, mọi refresh token cũ của user bị thu hồi (refresh cũ không cấp được cặp mới).
6. `SetStatus(u, "suspended")` -> đăng nhập trả `ErrAccountNotActive`; refresh token bị thu hồi.
7. `SetStatus(u, "deleted")` -> đăng nhập bị chặn; refresh thu hồi.
8. `SetStatus(u, "active")` lại -> đăng nhập được (nếu chưa purge).
9. `DeleteAccount(u)` -> `app_user.email`/`phone`/`display_name` bị nullify/tombstone (PII không còn truy ở dạng định danh).
10. Sau `DeleteAccount`, mọi `platform_account` của u bị gỡ.
11. Sau `DeleteAccount`, refresh token vô hiệu và đăng nhập bị chặn (status `deleted`).
12. `DeleteAccount` gọi lại trên tài khoản đã xóa -> không lỗi nghiêm trọng (idempotent).

---

## §5 - Kiểm thử (verification)

```go
// services/auth/internal/auth/reset_test.go
func TestRequestReset_NoEnumeration(t *testing.T) {
    s := newService(t)
    s.seedUser(t, "exists@x.com")
    err1 := s.RequestReset(ctx, "exists@x.com")
    err2 := s.RequestReset(ctx, "nope@x.com")
    require.NoError(t, err1)
    require.NoError(t, err2) // cùng phản hồi, không lộ tồn tại
}

func TestConfirmReset_OneTime(t *testing.T) {
    s := newService(t)
    tok := s.issueResetFor(t, "a@x.com")
    require.NoError(t, s.ConfirmReset(ctx, tok, "newp@ss12345"))
    require.ErrorIs(t, s.ConfirmReset(ctx, tok, "againp@ss12"), ErrInvalidResetToken)
}

func TestConfirmReset_RevokesSessions(t *testing.T) {
    s := newService(t)
    uid := s.seedUser(t, "a@x.com")
    pair, _ := s.IssueTokenPair(ctx, uid)
    tok := s.issueResetFor(t, "a@x.com")
    s.ConfirmReset(ctx, tok, "newp@ss12345")
    _, err := s.Refresh(ctx, pair.Refresh) // refresh cũ vô hiệu
    require.Error(t, err)
}

// services/auth/internal/auth/lifecycle_test.go
func TestSuspended_CannotLogin_TokensRevoked(t *testing.T) {
    s := newService(t)
    uid := s.seedActiveUser(t, "a@x.com", "p@ss12345")
    pair, _ := s.IssueTokenPair(ctx, uid)
    require.NoError(t, s.SetStatus(ctx, uid, "suspended"))
    _, err := s.Login(ctx, "a@x.com", "p@ss12345")
    require.ErrorIs(t, err, ErrAccountNotActive)
    _, rerr := s.Refresh(ctx, pair.Refresh)
    require.Error(t, rerr) // refresh thu hồi
}

// services/auth/internal/auth/erasure_test.go
func TestDelete_ErasesPII_AndLinks(t *testing.T) {
    s := newService(t)
    uid := s.seedUserWithLinks(t, "chi@x.com")
    require.NoError(t, s.DeleteAccount(ctx, uid))
    u := s.repo.rawUser(t, uid)
    require.NotEqual(t, "chi@x.com", u.Email) // email ẩn danh/tombstone
    require.Empty(t, s.repo.links(t, uid))     // platform_account gỡ
    require.Equal(t, "deleted", u.Status)
}

func TestDelete_Idempotent(t *testing.T) {
    s := newService(t)
    uid := s.seedUser(t, "a@x.com")
    require.NoError(t, s.DeleteAccount(ctx, uid))
    require.NoError(t, s.DeleteAccount(ctx, uid)) // lần hai không lỗi
}
```

---

## §6 - Khung triển khai

Xem §3. Reset token lưu hash, một-lần, TTL ngắn; gửi qua FR-NOTIF-001. `RequestReset` luôn trả nil để không lộ tồn tại. `ConfirmReset` và `SetStatus(suspended/deleted)` đều gọi `RevokeAllRefresh` - cùng access TTL ngắn của FR-AUTH-002 làm khóa/đổi mật khẩu có hiệu lực thực. `DeleteAccount` ẩn danh hóa PII (tombstone) rồi gỡ liên kết, thu hồi token, đặt `deleted`; purge cứng chạy sau cửa sổ ân hạn bằng job riêng. Phối FR-COMPLY-003 để dữ liệu liên quan ở module khác cũng theo chính sách lưu trữ. Audit log thao tác nhạy cảm không kèm token/PII thô.

---

## §7 - Phụ thuộc

- FR-AUTH-001 - `Hash` mật khẩu mới (argon2id) khi ConfirmReset; `app_user.status` cột lõi.
- FR-AUTH-002 - `RevokeAllRefresh` dùng cơ chế thu hồi refresh token; access TTL ngắn bổ trợ.
- FR-AUTH-003 - `DeleteAccount` gỡ `platform_account`.
- FR-COMPLY-003 - DSAR; FR này là điểm thực thi quyền xóa.
- FR-NOTIF-001 - gửi token reset qua kênh verified.

---

## §8 - Payload ví dụ

### Yêu cầu reset (phản hồi đồng nhất)

```http
POST /v1/auth/password/reset-request
{"identifier":"chi@example.com"}
-> 200 {"message":"Nếu tài khoản tồn tại, hướng dẫn đặt lại mật khẩu đã được gửi."}
```

### Xóa tài khoản (DSAR)

```http
DELETE /v1/account HTTP/1.1
Authorization: Bearer <access_jwt>
-> 200 {"status":"deleted","grace_until":"2026-07-27T00:00:00Z"}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Tự khôi phục trong cửa sổ ân hạn (UI "hoàn tác xóa") - trải nghiệm sau.
- Xuất dữ liệu (data portability) trước khi xóa - phối FR-COMPLY-003 ở slice DSAR đầy đủ.
- Suspend tạm có thời hạn tự hết (ví dụ khóa 24h chống abuse) - gắn FR-TRUST-004.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Reset lộ tài khoản tồn tại | no-enumeration test | user enumeration | Phản hồi đồng nhất (§1 #3) |
| Token reset dùng nhiều lần | one-time test | cửa sổ chiếm tài khoản | Một-lần + hết hạn (§1 #2) |
| Đổi mật khẩu không cắt phiên cũ | revoke test | kẻ chiếm phiên vẫn vào | Thu hồi mọi refresh (§1 #4) |
| Suspend/delete vẫn dùng phiên | lifecycle test | khóa không hiệu lực | Thu hồi refresh + access TTL ngắn (§1 #5) |
| Xóa chỉ đặt cờ, giữ PII | erasure test | không thỏa quyền xóa PDPL | Ẩn danh hóa PII (§1 #7,#9) |
| Xóa nhầm/bị lừa xóa | grace window | mất tài khoản | Cửa sổ ân hạn trước purge (§1 #8) |
| Log token reset/PII | review | rò rỉ | Cấm log nhạy cảm (§1 #10) |
| Xóa bỏ sót platform_account | erasure test | liên kết sàn còn lại | Gỡ platform_account (§1 #7) |
| Xóa lặp gây lỗi | idempotent test | UX/job gãy | Idempotent (§1 #12) |

---

## §11 - Ghi chú

- PDPL trao quyền xóa; `DeleteAccount` là điểm thực thi, ẩn danh hóa PII thật sự thay vì chỉ đặt cờ.
- Phản hồi reset đồng nhất che user enumeration - một rò rỉ tinh vi nếu thông báo khác nhau theo tồn tại.
- Đổi mật khẩu thu hồi mọi phiên là điều người dùng kỳ vọng khi nghi bị lộ; giữ phiên cũ làm reset vô nghĩa.
- Status là cờ DB nhưng access token tự chứng thực - phải thu hồi refresh khi suspend/delete để khóa có hiệu lực thực (cùng access TTL ngắn FR-AUTH-002).
- Cửa sổ ân hạn cân bằng quyền xóa PDPL với rủi ro xóa nhầm: không dùng được ngay, PII ẩn danh ngay, nhưng còn lối khôi phục trước purge cứng.
- Phối FR-COMPLY-003 để xóa lan tới dữ liệu cá nhân ở module khác theo chính sách lưu trữ, không chỉ trong `app_user`.

---

*Hết FR-AUTH-005. Status: ready_to_implement (mục tiêu audit 10/10).*
