---
id: TASK-AUTH-003
title: "`platform_account` liên kết tài khoản sàn, ext_user_ref ẩn danh, UNIQUE(user_id, platform_id), KHÔNG BAO GIỜ lưu token phiên"
module: AUTH
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-AUTH-001, TASK-INFRA-002, TASK-EXT-003, TASK-COMPLY-005, TASK-TRUST-002, TASK-TRUST-003]
depends_on: [TASK-AUTH-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (platform_account: ext_user_ref ẩn danh, KHÔNG token phiên)"
  - "docs/... §3.2 (session piggyback: token phiên KHÔNG rời client), §5.5 (PDPL no token-on-server)"
source_decisions:
  - "DEC-AUTH-11: platform_account liên kết app_user <-> sàn; lưu ext_user_ref ẩn danh, UNIQUE(user_id, platform_id)"
  - "DEC-AUTH-12: TUYỆT ĐỐI KHÔNG lưu cookie/session token/mật khẩu sàn trên server (§3.2/§5.5) - đây là cam kết niềm tin lõi"
  - "DEC-AUTH-13: ext_user_ref là định danh ẩn danh hóa (hash/pseudonym), không phải username/email sàn thật"
  - "DEC-AUTH-14: schema cấm về cấu trúc - không có cột nào cho token/cookie; CHECK độ dài ext_user_ref hợp lý"
  - "DEC-AUTH-15: liên kết khởi tạo từ extension (session piggyback) gửi tối thiểu hóa; backend chỉ nhận ext_user_ref đã ẩn danh"

language: "PostgreSQL 16 + Go 1.22 (auth-svc)"
service: shopass/services/auth/
new_files:
  - services/auth/migrations/0006_platform_account.up.sql
  - services/auth/migrations/0006_platform_account.down.sql
  - services/auth/internal/auth/linkacct.go
  - services/auth/internal/auth/linkacct_repo.go
  - services/auth/internal/auth/linkacct_test.go
modified_files:
  - services/auth/internal/auth/types.go            # struct PlatformAccount
allowed_tools:
  - file_read: services/auth/**
  - file_write: services/auth/**
  - bash: cd services/auth && go test ./...
disallowed_tools:
  - thêm cột lưu cookie/session token/mật khẩu sàn (vi phạm DEC-AUTH-12 - vi phạm cam kết niềm tin lõi + PDPL)
  - lưu username/email sàn thật thay vì ext_user_ref ẩn danh (vi phạm DEC-AUTH-13)
  - cho phép liên kết trùng (user_id, platform_id) (vi phạm DEC-AUTH-11)

effort_hours: 5
sub_tasks:
  - "0.5h: 0006_platform_account - bảng theo §3.4, UNIQUE(user_id,platform_id), KHÔNG cột token"
  - "1.0h: linkacct.go - LinkAccount(user, platform, ext_user_ref); validate ext_user_ref ẩn danh"
  - "0.5h: linkacct_repo.go - Upsert + ListByUser + Unlink"
  - "1.5h: linkacct_test.go - liên kết mới; trùng (user,platform) bị chặn; unlink; ext_user_ref rỗng/quá dài bị từ chối"
  - "1.0h: test khẳng định schema KHÔNG có cột token/cookie (introspection information_schema)"
  - "0.5h: OTel metric platform_account_linked_total / unlinked_total"

risk_if_skipped: "Đây là task mang cam kết niềm tin lõi của SănDeal hậu-Honey: KHÔNG BAO GIỜ lưu token phiên sàn trên server (§3.2/§5.5). Nếu sơ ý thêm cột token/cookie, một lần DB lộ là chiếm tài khoản sàn của hàng loạt người dùng - vừa thảm họa pháp lý PDPL (chế tài tới 5% doanh thu) vừa giết niềm tin (extension đọc cookie sàn vốn đã dễ bị nghi malware, §5.4). Lưu username sàn thật thay vì ref ẩn danh cũng làm lộ danh tính chéo sàn. Schema phải cấm token về mặt cấu trúc, không chỉ về quy ước."
---

## §1 - Mô tả (BCP-14 normative)

Service AUTH **MUST** cung cấp liên kết tài khoản sàn qua bảng `platform_account` chỉ lưu định danh ẩn danh, ràng buộc một liên kết mỗi (user, sàn), và TUYỆT ĐỐI không lưu token/cookie/mật khẩu sàn. Hợp đồng:

1. Migration `0006_platform_account` **MUST** tạo bảng theo §3.4: `id BIGSERIAL PK`, `user_id BIGINT REFERENCES app_user(id)`, `platform_id SMALLINT REFERENCES platform(id)`, `ext_user_ref TEXT`, `linked_at TIMESTAMPTZ DEFAULT now()`, với `UNIQUE(user_id, platform_id)` (DEC-AUTH-11).
2. Bảng `platform_account` **MUST NOT** có bất kỳ cột nào lưu cookie, session token, hay mật khẩu sàn (DEC-AUTH-12, DEC-AUTH-14). Đây là cấm về cấu trúc: không cột `token`, `cookie`, `session`, `password`.
3. `ext_user_ref` **MUST** là định danh ẩn danh hóa (hash/pseudonym), KHÔNG phải username/email/số điện thoại sàn thật (DEC-AUTH-13). Mục đích chỉ để nhận biết liên kết, không để định danh chéo.
4. Liên kết **MUST** ép `UNIQUE(user_id, platform_id)` (DEC-AUTH-11): một người dùng chỉ một liên kết cho mỗi sàn. Liên kết lại cùng sàn -> cập nhật (upsert), không tạo dòng trùng.
5. AUTH **MUST** expose: `LinkAccount(ctx, userID int64, platformID int16, extUserRef string) error`, `ListLinks(ctx, userID int64) ([]PlatformAccount, error)`, `Unlink(ctx, userID, platformID) error`.
6. `LinkAccount` **MUST** validate `ext_user_ref`: không rỗng, độ dài trong giới hạn hợp lý (CHECK), và (theo quy ước) là dạng ẩn danh - KHÔNG nhận giá trị trông như email/cookie/token thô.
7. Liên kết **MUST** khởi tạo từ phía extension theo nguyên tắc session piggyback (DEC-AUTH-15, §3.2): extension đọc ngữ cảnh đăng nhập của chính người dùng và gửi về backend chỉ một `ext_user_ref` đã ẩn danh - KHÔNG gửi cookie/token.
8. `Unlink` **MUST** xóa liên kết của (user, sàn); thao tác idempotent (unlink khi chưa liên kết không lỗi nghiêm trọng).
9. Hệ thống **MUST** KHÔNG log `ext_user_ref` ở mức có thể tái định danh; coi nó như dữ liệu cá nhân giả danh theo PDPL.
10. Schema + lớp truy cập **MUST** là điểm chứng minh cho audit TASK-TRUST-003 và TASK-COMPLY-005 rằng không có credential sàn nào (cookie/token/mật khẩu) tồn tại trên server.
11. `ListLinks` **MUST** chỉ trả liên kết của đúng `user_id` truyền vào (cô lập theo người dùng, không rò rỉ chéo).
12. AUTH **SHOULD** phát metric `platform_account_linked_total{platform_id}` / `platform_account_unlinked_total{platform_id}` - không kèm `ext_user_ref`.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao tuyệt đối không lưu token phiên (DEC-AUTH-12)? Đây là cam kết niềm tin trung tâm của SănDeal sau vụ Honey (§5.4). Extension đọc cookie phiên sàn của chính người dùng để lấy giỏ hàng - một hành vi vốn đã dễ bị nghi là malware. Cách duy nhất biến nghi ngờ thành niềm tin là token phiên KHÔNG bao giờ rời máy người dùng và KHÔNG bao giờ chạm server. Nếu server lưu token, một lần DB lộ là chiếm tài khoản sàn hàng loạt - vừa vi phạm PDPL nặng vừa kết liễu sản phẩm.

Vì sao cấm token về cấu trúc, không chỉ quy ước (§1 #2, DEC-AUTH-14)? Quy ước "đừng lưu token" dễ bị phá bởi một commit vô tình thêm cột tiện tay. Làm schema không có chỗ chứa token (không cột nào hợp) biến cam kết thành bất biến kiểm được: một test introspection xác nhận bảng không có cột token là bằng chứng máy kiểm, không phụ thuộc kỷ luật con người.

Vì sao ext_user_ref ẩn danh, không username sàn thật (DEC-AUTH-13)? Nếu lưu username/email sàn thật, server nắm bản đồ "người dùng SănDeal X = tài khoản Shopee Y = tài khoản Lazada Z" - một kho định danh chéo sàn nhạy cảm, hấp dẫn kẻ tấn công và rủi ro PDPL. Một pseudonym chỉ đủ để biết "có một liên kết" mà không tái dựng được danh tính thật.

Vì sao UNIQUE(user_id, platform_id) (DEC-AUTH-11)? Một người dùng SănDeal ứng với một tài khoản trên mỗi sàn trong mô hình này. Cho phép nhiều liên kết cùng sàn gây nhập nhằng (liên kết nào là chính?) và mở đường lạm dụng. Ràng buộc một-liên-kết-mỗi-sàn giữ mô hình rõ; liên kết lại là cập nhật, không nhân bản.

Vì sao liên kết khởi tạo từ extension tối thiểu hóa (DEC-AUTH-15)? Kiến trúc session piggyback (§3.2) đặt mọi tiếp xúc với phiên sàn ở phía client. Backend chỉ nhận kết quả đã ẩn danh (`ext_user_ref`), không nhận nguyên liệu nhạy cảm (cookie/token). Giữ ranh giới này là cách kỹ thuật để cam kết "token không rời client" thành thật, không chỉ là lời hứa.

Vì sao coi ext_user_ref là dữ liệu giả danh PDPL (§1 #9)? Dù đã ẩn danh, một pseudonym vẫn có thể là dữ liệu cá nhân giả danh nếu ghép thêm dữ liệu khác. Đối xử thận trọng (không log tái-định-danh được, đưa vào phạm vi DSAR của TASK-AUTH-005) giữ ta đúng tinh thần PDPL.

---

## §3 - Hợp đồng API / DDL

### Migration (cấm token về cấu trúc)

```sql
-- services/auth/migrations/0006_platform_account.up.sql
CREATE TABLE platform_account (
  id           BIGSERIAL    PRIMARY KEY,
  user_id      BIGINT       NOT NULL REFERENCES app_user(id),
  platform_id  SMALLINT     NOT NULL REFERENCES platform(id),
  ext_user_ref TEXT         NOT NULL
                 CHECK (length(ext_user_ref) BETWEEN 1 AND 128), -- §1 #6
  linked_at    TIMESTAMPTZ  DEFAULT now(),
  UNIQUE (user_id, platform_id)                                   -- §1 #4
  -- CỐ Ý KHÔNG có cột cookie/session/token/password (DEC-AUTH-12, §1 #2).
);
CREATE INDEX idx_pa_user ON platform_account (user_id);
```

### Liên kết (§1 #5, #6, #7)

```go
// services/auth/internal/auth/linkacct.go
func (s *Service) LinkAccount(ctx context.Context, userID int64, platformID int16, extUserRef string) error {
    if extUserRef == "" || len(extUserRef) > 128 {
        return ErrInvalidExtRef
    }
    if looksLikeRawCredential(extUserRef) { // chặn giá trị trông như email/cookie/token thô (§1 #6)
        return ErrExtRefNotAnonymized
    }
    // Upsert: liên kết lại cùng sàn → cập nhật, không tạo dòng trùng (§1 #4).
    return s.repo.Upsert(ctx, PlatformAccount{
        UserID: userID, PlatformID: platformID, ExtUserRef: extUserRef,
    })
}

func (s *Service) Unlink(ctx context.Context, userID int64, platformID int16) error {
    return s.repo.Delete(ctx, userID, platformID) // idempotent (§1 #8)
}
```

### Repo upsert (§1 #4, #11)

```go
// services/auth/internal/auth/linkacct_repo.go
func (r *Repo) Upsert(ctx context.Context, pa PlatformAccount) error {
    _, err := r.pool.Exec(ctx,
        `INSERT INTO platform_account (user_id, platform_id, ext_user_ref)
         VALUES ($1,$2,$3)
         ON CONFLICT (user_id, platform_id)
         DO UPDATE SET ext_user_ref = EXCLUDED.ext_user_ref, linked_at = now()`,
        pa.UserID, pa.PlatformID, pa.ExtUserRef)
    return err
}

func (r *Repo) ListByUser(ctx context.Context, userID int64) ([]PlatformAccount, error) {
    // WHERE user_id = $1 — chỉ liên kết của đúng người dùng (§1 #11).
    return r.query(ctx, `SELECT * FROM platform_account WHERE user_id=$1`, userID)
}
```

---

## §4 - Acceptance criteria

1. Migration chạy -> bảng `platform_account` tồn tại với `UNIQUE(user_id, platform_id)`.
2. Introspection `information_schema.columns` của `platform_account` KHÔNG có cột tên chứa `token`/`cookie`/`session`/`password`.
3. `LinkAccount(u1, shopee, "ref-abc")` -> tạo một liên kết.
4. `LinkAccount(u1, shopee, "ref-xyz")` lần hai (cùng sàn) -> cập nhật `ext_user_ref`, vẫn một dòng (không trùng).
5. `LinkAccount(u1, lazada, "ref-2")` -> liên kết thứ hai (sàn khác) tồn tại song song.
6. `LinkAccount(u1, shopee, "")` -> `ErrInvalidExtRef`.
7. `LinkAccount(u1, shopee, <chuỗi 200 ký tự>)` -> lỗi (vượt CHECK độ dài).
8. `LinkAccount(u1, shopee, "chi@gmail.com")` -> `ErrExtRefNotAnonymized` (trông như email thật).
9. `ListLinks(u1)` -> trả đúng các liên kết của u1; `ListLinks(u2)` không thấy liên kết của u1.
10. `Unlink(u1, shopee)` -> xóa liên kết; `ListLinks(u1)` không còn shopee.
11. `Unlink(u1, shopee)` lần hai (đã xóa) -> không lỗi nghiêm trọng (idempotent).
12. Không có log nào chứa `ext_user_ref` ở dạng tái-định-danh được (kiểm review + test output).

---

## §5 - Kiểm thử (verification)

```go
// services/auth/internal/auth/linkacct_test.go
func TestSchema_NoCredentialColumns(t *testing.T) {
    db := upToLatest(t)
    cols := columnNames(t, db, "platform_account")
    for _, c := range cols {
        require.NotRegexp(t, `(?i)token|cookie|session|password`, c) // cấm cấu trúc (§1 #2)
    }
}

func TestLink_NewAndUpsert(t *testing.T) {
    s := newService(t)
    require.NoError(t, s.LinkAccount(ctx, u1, shopee, "ref-abc"))
    require.NoError(t, s.LinkAccount(ctx, u1, shopee, "ref-xyz")) // cùng sàn → upsert
    links, _ := s.ListLinks(ctx, u1)
    require.Len(t, links, 1)
    require.Equal(t, "ref-xyz", links[0].ExtUserRef)
}

func TestLink_MultiPlatform(t *testing.T) {
    s := newService(t)
    s.LinkAccount(ctx, u1, shopee, "ref-1")
    s.LinkAccount(ctx, u1, lazada, "ref-2")
    links, _ := s.ListLinks(ctx, u1)
    require.Len(t, links, 2)
}

func TestLink_RejectsRawCredential(t *testing.T) {
    s := newService(t)
    require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, ""), ErrInvalidExtRef)
    require.ErrorIs(t, s.LinkAccount(ctx, u1, shopee, "chi@gmail.com"), ErrExtRefNotAnonymized)
}

func TestList_IsolatedPerUser(t *testing.T) {
    s := newService(t)
    s.LinkAccount(ctx, u1, shopee, "ref-1")
    out, _ := s.ListLinks(ctx, u2)
    require.Empty(t, out) // u2 không thấy liên kết của u1
}

func TestUnlink_Idempotent(t *testing.T) {
    s := newService(t)
    s.LinkAccount(ctx, u1, shopee, "ref-1")
    require.NoError(t, s.Unlink(ctx, u1, shopee))
    require.NoError(t, s.Unlink(ctx, u1, shopee)) // lần hai không lỗi
}
```

---

## §6 - Khung triển khai

Xem §3. Migration cố ý không có cột token - một test introspection (`TestSchema_NoCredentialColumns`) biến cam kết "không lưu token" thành bất biến máy kiểm. `LinkAccount` upsert theo `(user_id, platform_id)` nên liên kết lại là cập nhật. `looksLikeRawCredential` chặn giá trị trông như email/cookie/token thô để ép `ext_user_ref` đúng dạng ẩn danh. Liên kết khởi tạo từ extension (TASK-EXT-003) gửi chỉ `ext_user_ref` đã ẩn danh; backend không bao giờ nhận cookie/token. `ext_user_ref` vào phạm vi DSAR (TASK-AUTH-005) như dữ liệu giả danh.

---

## §7 - Phụ thuộc

- TASK-AUTH-001 - `app_user` (chủ thể) và TASK-INFRA-002 - `platform` (sàn) phải tồn tại cho FK.
- TASK-EXT-003 (đồng hàng/downstream) - pipeline tối thiểu hóa client tạo `ext_user_ref` ẩn danh, gửi về.
- TASK-TRUST-002 / TASK-TRUST-003 (downstream) - minh bạch + audit độc lập dùng schema này làm bằng chứng không lưu credential sàn.
- TASK-COMPLY-005 (đồng hàng) - audit no-cleartext + token-not-on-server.
- TASK-AUTH-005 (downstream) - DSAR xóa cả `platform_account`.

---

## §8 - Payload ví dụ

### Liên kết từ extension (REST qua gateway)

```http
POST /v1/account/link HTTP/1.1
Authorization: Bearer <access_jwt>
Content-Type: application/json

{"platform":"shopee","ext_user_ref":"px_9f2c8e...a14"}
// CHÚ Ý: KHÔNG có cookie/token sàn trong payload — chỉ ref ẩn danh.
```

### Liệt kê liên kết

```json
{
  "links": [
    {"platform":"shopee","linked_at":"2026-06-27T08:00:00Z"},
    {"platform":"lazada","linked_at":"2026-06-27T08:01:00Z"}
  ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cơ chế chứng thực liên kết (challenge để xác nhận người dùng thật sở hữu tài khoản sàn) mà vẫn không lưu token - nghiên cứu giai đoạn sau.
- Nhiều tài khoản cùng sàn cho một người dùng (nếu mô hình thay đổi) - hiện ràng buộc một-mỗi-sàn.
- Xoay `ext_user_ref` định kỳ để giảm khả năng tương quan dài hạn - tối ưu quyền riêng tư sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Thêm cột token/cookie | schema introspection test | thảm họa nếu DB lộ; vi phạm PDPL | Cấm cấu trúc + test (§1 #2) |
| Lưu username sàn thật | review + reject test | kho định danh chéo sàn | ext_user_ref ẩn danh (§1 #3) |
| Liên kết trùng cùng sàn | upsert + UNIQUE | nhập nhằng liên kết chính | UNIQUE(user,platform) + upsert (§1 #4) |
| ListLinks rò rỉ chéo user | isolation test | lộ liên kết người khác | WHERE user_id (§1 #11) |
| Gửi cookie/token trong payload | review + reject | token rời client (vỡ cam kết) | Chỉ nhận ext_user_ref (§1 #7) |
| Log ext_user_ref tái-định-danh | review | rò rỉ giả danh PDPL | Không log tái-định-danh (§1 #9) |
| ext_user_ref rỗng/quá dài | CHECK + validate | dữ liệu rác | Validate độ dài (§1 #6) |
| Unlink khi chưa liên kết lỗi cứng | idempotent test | UX gãy | Idempotent (§1 #8) |

---

## §11 - Ghi chú

- Đây là task mang cam kết niềm tin lõi hậu-Honey: token phiên sàn KHÔNG bao giờ rời client và KHÔNG bao giờ chạm server.
- Cấm token về cấu trúc (schema không có chỗ chứa) + test introspection biến cam kết thành bất biến máy kiểm, không phụ thuộc kỷ luật con người.
- `ext_user_ref` ẩn danh tránh biến server thành kho bản đồ định danh chéo sàn nhạy cảm.
- UNIQUE(user, platform) + upsert giữ mô hình một-liên-kết-mỗi-sàn rõ ràng; liên kết lại là cập nhật.
- Ranh giới session piggyback (extension nhận nguyên liệu nhạy cảm, backend chỉ nhận ref ẩn danh) là cách kỹ thuật để "token không rời client" thành thật.
- Schema này là bằng chứng cho audit độc lập (TASK-TRUST-003) và no-cleartext (TASK-COMPLY-005): không credential sàn nào tồn tại trên server.

---

*Hết TASK-AUTH-003. Status: ready_to_implement (mục tiêu audit 10/10).*
