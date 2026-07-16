---
id: TASK-AUTH-001
title: "Schema `app_user` (argon2id pwd_hash, CITEXT email unique, phone, locale 'vi-VN', status, referral_code_id) + đăng ký; no cleartext"
module: AUTH
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-INFRA-002, TASK-AUTH-002, TASK-AUTH-003, TASK-AUTH-005, TASK-COMPLY-005, TASK-BILL-004]
depends_on: [TASK-INFRA-002]
blocks: [TASK-AUTH-002, TASK-AUTH-003, TASK-AUTH-005, TASK-BILL-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (data model app_user, argon2id, no cleartext)"
  - "docs/... §3.8 (bảo mật argon2id), §5.5 (PDPL no-cleartext)"
source_decisions:
  - "DEC-AUTH-01: mật khẩu băm bằng argon2id (memory-hard); KHÔNG bao giờ lưu cleartext (DEC chung §3.8/§5.5)"
  - "DEC-AUTH-02: app_user mở rộng từ cột lõi của TASK-INFRA-002 qua migration mới; thêm pwd_hash, referral_code_id"
  - "DEC-AUTH-03: email là CITEXT unique (đã đặt ở INFRA-002); đăng ký chuẩn hóa email (trim) trước khi lưu"
  - "DEC-AUTH-04: tham số argon2id chuẩn hóa (time, memory, parallelism) lưu kèm hash (định dạng PHC) để verify bền vững khi đổi tham số"
  - "DEC-AUTH-05: đăng ký yêu cầu email HOẶC phone (ít nhất một định danh); mật khẩu tối thiểu độ dài/độ mạnh"

language: "PostgreSQL 16 + Go 1.22 (auth-svc); thư viện argon2id (golang.org/x/crypto/argon2)"
service: shopass/services/auth/
new_files:
  - services/auth/migrations/0004_app_user_secrets.up.sql
  - services/auth/migrations/0004_app_user_secrets.down.sql
  - services/auth/internal/auth/password.go
  - services/auth/internal/auth/register.go
  - services/auth/internal/auth/repo.go
  - services/auth/internal/auth/password_test.go
  - services/auth/internal/auth/register_test.go
modified_files:
  - services/auth/internal/auth/types.go            # struct AppUser
allowed_tools:
  - file_read: services/auth/**
  - file_write: services/auth/**
  - bash: cd services/auth && go test ./...
disallowed_tools:
  - lưu mật khẩu cleartext hay băm yếu (md5/sha1) (vi phạm DEC-AUTH-01, PDPL §5.5)
  - log mật khẩu hay hash đầy đủ (rò rỉ credential)
  - bỏ qua chuẩn hóa email cho phép trùng qua biến thể hoa thường (vi phạm DEC-AUTH-03)

effort_hours: 6
sub_tasks:
  - "1.0h: 0004_app_user_secrets - thêm cột pwd_hash TEXT, referral_code_id BIGINT vào app_user"
  - "1.5h: password.go - Hash/Verify argon2id định dạng PHC (lưu tham số kèm hash)"
  - "1.0h: register.go - validate email/phone + độ mạnh mật khẩu, chuẩn hóa email, insert"
  - "0.5h: repo.go - InsertUser + FindByEmail (CITEXT)"
  - "1.0h: password_test.go - hash khác nhau mỗi lần (salt); Verify đúng/sai; nâng tham số vẫn verify hash cũ"
  - "1.0h: register_test.go - đăng ký mới; trùng email hoa thường bị từ chối; thiếu cả email lẫn phone bị từ chối; mật khẩu yếu bị từ chối"

risk_if_skipped: "app_user là chủ thể trung tâm: AUTH, BILL, TRACK, AFFIL đều gắn vào. Lưu mật khẩu cleartext hay băm yếu là vi phạm PDPL nghiêm trọng (chế tài tới 3 tỷ VND/§5.5) và là thảm họa niềm tin nếu DB lộ. Không chuẩn hóa email cho phép một người tạo nhiều tài khoản qua biến thể hoa thường, làm rối billing/referral. Không lưu tham số argon2id kèm hash thì khi nâng cấu hình bảo mật, hash cũ không verify được nữa. Đây là nền của toàn bộ định danh người dùng."
---

## §1 - Mô tả (BCP-14 normative)

Service AUTH **MUST** mở rộng `app_user` với cột bảo mật và cung cấp đăng ký an toàn: mật khẩu băm argon2id, email CITEXT unique đã chuẩn hóa, không bao giờ lưu cleartext. Hợp đồng:

1. Migration `0004_app_user_secrets` **MUST** thêm vào `app_user` (đã có cột lõi từ TASK-INFRA-002): `pwd_hash TEXT` và `referral_code_id BIGINT` (DEC-AUTH-02). KHÔNG định nghĩa lại cột lõi.
2. Mật khẩu **MUST** băm bằng argon2id (DEC-AUTH-01). KHÔNG lưu cleartext; KHÔNG dùng md5/sha1/bcrypt-yếu. `pwd_hash` lưu định dạng PHC chứa thuật toán, tham số (time, memory, parallelism), salt và hash (DEC-AUTH-04).
3. Mỗi lần băm **MUST** dùng salt ngẫu nhiên riêng -> hai người dùng cùng mật khẩu có `pwd_hash` khác nhau.
4. `Verify(password, phcHash)` **MUST** đọc tham số từ chính chuỗi PHC để xác thực, nên hash băm bằng tham số cũ vẫn verify được sau khi nâng tham số mặc định (DEC-AUTH-04).
5. Đăng ký **MUST** yêu cầu ít nhất một định danh: email HOẶC phone (DEC-AUTH-05). Thiếu cả hai -> từ chối.
6. Đăng ký **MUST** chuẩn hóa email trước khi lưu (trim khoảng trắng) (DEC-AUTH-03); nhờ CITEXT unique (TASK-INFRA-002), `Chi@Mail.com` và `chi@mail.com` xung đột - từ chối tạo trùng.
7. Đăng ký **MUST** ép độ mạnh mật khẩu tối thiểu (DEC-AUTH-05): độ dài >= 8 ký tự (ngưỡng cấu hình được). Mật khẩu yếu -> từ chối với thông báo rõ.
8. Repo **MUST** expose: `InsertUser(ctx, u AppUser) (int64, error)` và `FindByEmail(ctx, email string) (AppUser, error)` (so sánh CITEXT, case-insensitive).
9. Hệ thống **MUST** KHÔNG log mật khẩu thô hay `pwd_hash` đầy đủ ở bất kỳ đâu (vi phạm là rò rỉ credential).
10. Đăng ký trùng email **MUST** trả lỗi xác định (ví dụ `ErrEmailTaken`) để tầng API ánh xạ sang `409 Conflict`, KHÔNG để lỗi UNIQUE thô lọt ra ngoài.
11. `app_user.status` (cột lõi) **MUST** mặc định `'active'` khi đăng ký; TASK-AUTH-005 quản vòng đời status về sau.
12. Tham số argon2id mặc định **SHOULD** theo khuyến nghị hiện hành (ví dụ memory >= 64 MiB, time >= 3, parallelism phù hợp CPU), cấu hình được để nâng theo phần cứng.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao argon2id, không bcrypt/sha (DEC-AUTH-01)? argon2id là hàm băm mật khẩu memory-hard, thắng Password Hashing Competition, kháng tốt cả tấn công GPU lẫn side-channel. md5/sha1 không phải hàm băm mật khẩu (quá nhanh, dễ brute-force). Với một sản phẩm giữ định danh người dùng VN và chịu PDPL, dùng đúng hàm băm mật khẩu là yêu cầu cơ bản, không phải tùy chọn.

Vì sao salt ngẫu nhiên mỗi lần (§1 #3)? Không có salt riêng, hai người dùng cùng mật khẩu sẽ có cùng hash, và kẻ tấn công dùng rainbow table phá hàng loạt. Salt riêng làm mỗi hash độc nhất, vô hiệu rainbow table và buộc tấn công từng tài khoản một.

Vì sao lưu tham số kèm hash dạng PHC (DEC-AUTH-04)? Khuyến nghị bảo mật tăng theo thời gian (memory/time cao hơn khi phần cứng mạnh hơn). Nếu hash chỉ lưu digest mà không lưu tham số, ta không verify được hash cũ sau khi đổi mặc định. Định dạng PHC nhúng tham số vào chính chuỗi hash, nên `Verify` luôn dùng đúng tham số của thời điểm băm - nâng mặc định không phá tài khoản cũ.

Vì sao email HOẶC phone (DEC-AUTH-05)? Người dùng VN nhiều người quen đăng nhập bằng số điện thoại hơn email. Bắt buộc cả hai là rào cản; cho phép ít nhất một định danh giữ trải nghiệm mở mà vẫn có một khóa định danh để liên hệ và khôi phục.

Vì sao chuẩn hóa email + CITEXT (DEC-AUTH-03, §1 #6)? Người dùng gõ email tùy hứng hoa thường và có thể thừa khoảng trắng. Trim + CITEXT unique chặn một người tạo nhiều tài khoản qua `Chi@x.com` / `chi@x.com` / ` chi@x.com `, vốn làm rối billing và referral (TASK-BILL-004).

Vì sao lỗi đăng ký trùng phải xác định (§1 #10)? Để lỗi UNIQUE thô của Postgres lọt ra API là vừa xấu vừa lộ chi tiết schema. Trả `ErrEmailTaken` cho tầng API ánh xạ sạch sang 409 và thông báo thân thiện, không rò rỉ nội bộ.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/auth/migrations/0004_app_user_secrets.up.sql
ALTER TABLE app_user
  ADD COLUMN pwd_hash         TEXT,          -- argon2id định dạng PHC; KHÔNG cleartext
  ADD COLUMN referral_code_id BIGINT;
-- FK referral_code_id -> referral_code(id) thêm khi TASK-BILL-004 tạo bảng referral_code.
```

### Password hash argon2id (PHC) (§1 #2, #3, #4)

```go
// services/auth/internal/auth/password.go
type Argon2Params struct{ Time, Memory uint32; Parallelism uint8; SaltLen, KeyLen uint32 }

// Hash trả chuỗi PHC: $argon2id$v=19$m=65536,t=3,p=2$<salt_b64>$<hash_b64>
func Hash(password string, p Argon2Params) (string, error) {
    salt := make([]byte, p.SaltLen)
    if _, err := rand.Read(salt); err != nil { return "", err } // salt ngẫu nhiên (§1 #3)
    key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, p.KeyLen)
    return encodePHC(p, salt, key), nil
}

// Verify đọc tham số từ chính chuỗi PHC → hash cũ vẫn verify sau khi nâng mặc định (§1 #4).
func Verify(password, phc string) (bool, error) {
    p, salt, want, err := decodePHC(phc)
    if err != nil { return false, err }
    got := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, uint32(len(want)))
    return subtle.ConstantTimeCompare(got, want) == 1, nil // so sánh hằng thời gian
}
```

### Đăng ký (§1 #5, #6, #7, #10)

```go
// services/auth/internal/auth/register.go
func (s *Service) Register(ctx context.Context, in RegisterInput) (int64, error) {
    if in.Email == "" && in.Phone == "" { return 0, ErrNoIdentifier }      // §1 #5
    if err := checkPasswordStrength(in.Password); err != nil { return 0, err } // §1 #7
    email := normalizeEmail(in.Email) // trim; CITEXT lo case (§1 #6)
    hash, err := Hash(in.Password, s.params)
    if err != nil { return 0, err }
    id, err := s.repo.InsertUser(ctx, AppUser{Email: email, Phone: in.Phone, PwdHash: hash})
    if isUniqueViolation(err, "app_user_email_key") { return 0, ErrEmailTaken } // §1 #10
    return id, err
}
```

---

## §4 - Acceptance criteria

1. Migration chạy -> `app_user` có cột `pwd_hash` và `referral_code_id`.
2. `Hash("p@ss12345")` hai lần -> hai chuỗi PHC KHÁC nhau (salt ngẫu nhiên).
3. `Verify("p@ss12345", hash)` với hash đúng -> `true`; với mật khẩu sai -> `false`.
4. Hash băm bằng tham số cũ (memory thấp) vẫn `Verify` đúng sau khi đổi mặc định sang tham số cao.
5. `pwd_hash` lưu định dạng PHC chứa `argon2id` + tham số (kiểm chuỗi bắt đầu `$argon2id$`).
6. Đăng ký không email và không phone -> `ErrNoIdentifier`.
7. Đăng ký với mật khẩu < 8 ký tự -> lỗi độ mạnh.
8. Đăng ký email `Chi@Mail.com` rồi `chi@mail.com` -> lần hai trả `ErrEmailTaken` (CITEXT).
9. Đăng ký email có khoảng trắng thừa `" a@x.com "` -> lưu chuẩn hóa `a@x.com`.
10. `FindByEmail("A@X.COM")` tìm được user đăng ký bằng `a@x.com` (case-insensitive).
11. Không có log nào chứa mật khẩu thô hay `pwd_hash` đầy đủ (kiểm test output + review).
12. Đăng ký thành công -> user có `status='active'` (mặc định cột lõi).

---

## §5 - Kiểm thử (verification)

```go
// services/auth/internal/auth/password_test.go
func TestHash_DifferentEachTime(t *testing.T) {
    h1, _ := Hash("p@ss12345", defaultParams)
    h2, _ := Hash("p@ss12345", defaultParams)
    require.NotEqual(t, h1, h2) // salt ngẫu nhiên
    require.True(t, strings.HasPrefix(h1, "$argon2id$"))
}

func TestVerify_CorrectAndWrong(t *testing.T) {
    h, _ := Hash("p@ss12345", defaultParams)
    ok, _ := Verify("p@ss12345", h); require.True(t, ok)
    bad, _ := Verify("wrong", h);    require.False(t, bad)
}

func TestVerify_OldParamsStillWork(t *testing.T) {
    weak := Argon2Params{Time: 1, Memory: 8 * 1024, Parallelism: 1, SaltLen: 16, KeyLen: 32}
    h, _ := Hash("p@ss12345", weak)          // băm bằng tham số cũ/yếu
    ok, _ := Verify("p@ss12345", h)          // verify dùng tham số đọc từ PHC
    require.True(t, ok)
}

// services/auth/internal/auth/register_test.go
func TestRegister_NoIdentifier(t *testing.T) {
    s := newService(t)
    _, err := s.Register(ctx, RegisterInput{Password: "p@ss12345"})
    require.ErrorIs(t, err, ErrNoIdentifier)
}

func TestRegister_WeakPassword(t *testing.T) {
    s := newService(t)
    _, err := s.Register(ctx, RegisterInput{Email: "a@x.com", Password: "123"})
    require.Error(t, err)
}

func TestRegister_DuplicateEmail_CaseInsensitive(t *testing.T) {
    s := newService(t)
    _, err1 := s.Register(ctx, RegisterInput{Email: "Chi@Mail.com", Password: "p@ss12345"})
    require.NoError(t, err1)
    _, err2 := s.Register(ctx, RegisterInput{Email: "chi@mail.com", Password: "p@ss12345"})
    require.ErrorIs(t, err2, ErrEmailTaken)
}

func TestRegister_TrimsEmail(t *testing.T) {
    s := newService(t)
    id, _ := s.Register(ctx, RegisterInput{Email: " a@x.com ", Password: "p@ss12345"})
    u, _ := s.repo.FindByEmail(ctx, "a@x.com")
    require.Equal(t, id, u.ID)
}
```

---

## §6 - Khung triển khai

Xem §3. Migration 0004 thêm cột bảo mật vào `app_user` (cột lõi do TASK-INFRA-002). `password.go` dùng `golang.org/x/crypto/argon2` với định dạng PHC tự mã hóa/giải mã tham số. `Verify` so sánh hằng thời gian (`subtle.ConstantTimeCompare`) chống timing attack. Tham số mặc định lấy từ cấu hình (cho phép nâng theo phần cứng) - có thể đọc qua TASK-INFRA-003 nếu cần điều chỉnh tập trung. FK `referral_code_id -> referral_code(id)` hoãn tới khi TASK-BILL-004 tạo bảng đó.

---

## §7 - Phụ thuộc

- TASK-INFRA-002 - `app_user` cột lõi (id, email CITEXT unique, locale, status) phải tồn tại trước.
- TASK-AUTH-002 (downstream) - phát hành JWT sau khi xác thực mật khẩu qua `Verify`.
- TASK-AUTH-003 / TASK-AUTH-005 (downstream) - liên kết sàn và vòng đời tài khoản gắn vào `app_user`.
- TASK-BILL-004 (downstream) - tạo bảng `referral_code`; khi đó thêm FK cho `referral_code_id`.
- TASK-COMPLY-005 (đồng hàng) - audit no-cleartext xác nhận chỉ có `pwd_hash` argon2id, không cleartext.
- Thư viện: `golang.org/x/crypto/argon2`.

---

## §8 - Payload ví dụ

### Đăng ký (REST qua gateway)

```http
POST /v1/auth/register HTTP/1.1
Content-Type: application/json

{"email":"chi@example.com","phone":"+84906878091","password":"san-deal-2026!"}
```

### pwd_hash lưu trong DB (định dạng PHC)

```text
$argon2id$v=19$m=65536,t=3,p=2$Y2hpLXNhbHQtMTYtYnl0ZXM$3sJ...hash_b64...
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Pepper toàn cục (khóa bí mật trộn vào mọi hash, lưu trong Vault) ngoài salt per-hash - tăng phòng thủ khi DB lộ; gắn TASK-INFRA-003.
- Rehash trong suốt khi đăng nhập (nâng tham số hash cũ lúc user login thành công) - tối ưu bảo mật dần.
- Kiểm mật khẩu lộ (HaveIBeenPwned k-anonymity) khi đăng ký - thêm ở slice trải nghiệm sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Lưu mật khẩu cleartext/băm yếu | review + COMPLY-005 audit | thảm họa nếu DB lộ; vi phạm PDPL | argon2id PHC bắt buộc (§1 #2) |
| Cùng mật khẩu -> cùng hash | password test | rainbow table phá hàng loạt | Salt ngẫu nhiên mỗi lần (§1 #3) |
| Nâng tham số phá hash cũ | test OldParams | user cũ không đăng nhập được | PHC lưu tham số kèm hash (§1 #4) |
| Trùng tài khoản qua hoa thường | register test | rối billing/referral | Chuẩn hóa + CITEXT (§1 #6) |
| Lỗi UNIQUE thô lọt ra API | register test | xấu UX + lộ schema | Ánh xạ `ErrEmailTaken`->409 (§1 #10) |
| Log mật khẩu/hash | test output + review | rò rỉ credential | Cấm log (§1 #9) |
| Timing attack khi Verify | review | dò mật khẩu qua thời gian | ConstantTimeCompare (§6) |
| Mật khẩu quá yếu | register test | dễ brute-force | Ép độ mạnh tối thiểu (§1 #7) |
| Thiếu cả email lẫn phone | register test | không có khóa định danh | Bắt buộc >=1 định danh (§1 #5) |

---

## §11 - Ghi chú

- `app_user` là chủ thể trung tâm: AUTH/BILL/TRACK/AFFIL gắn vào; làm đúng phần bảo mật ở đây bảo vệ toàn hệ.
- argon2id là hàm băm mật khẩu đúng nghĩa (memory-hard), khác hẳn md5/sha vốn quá nhanh và dễ brute-force.
- Định dạng PHC nhúng tham số vào hash là chìa khóa để nâng cấu hình bảo mật theo thời gian mà không phá tài khoản cũ.
- Chuẩn hóa email + CITEXT chặn tạo nhiều tài khoản qua biến thể hoa thường, vốn làm rối billing/referral.
- Trả lỗi đăng ký xác định (`ErrEmailTaken`) giữ API sạch và không lộ chi tiết schema nội bộ.
- Đây là điểm chứng minh no-cleartext cho TASK-COMPLY-005: chỉ `pwd_hash` argon2id rời lớp này, không có mật khẩu thô.

---

*Hết TASK-AUTH-001. Status: ready_to_implement (mục tiêu audit 10/10).*
