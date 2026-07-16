---
id: TASK-AUTH-002
title: "JWT access+refresh, claims (user_id, locale, tier), JWKS endpoint, BFF verify"
module: AUTH
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-AUTH-001, TASK-INFRA-001, TASK-INFRA-003, TASK-AUTH-004, TASK-AUTH-005, TASK-WEB-001]
depends_on: [TASK-AUTH-001, TASK-INFRA-001]
blocks: [TASK-AUTH-004, TASK-EXT-005, TASK-MOBILE-001, TASK-WEB-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.1 (auth JWT tại gateway/BFF)"
  - "docs/... §3.8 (bảo mật), §3.7 (API surface cần auth)"
source_decisions:
  - "DEC-AUTH-06: cặp token - access JWT ngắn hạn (ví dụ 15 phút) + refresh token dài hạn (ví dụ 30 ngày, xoay được)"
  - "DEC-AUTH-07: access JWT ký RS256 (bất đối xứng); gateway verify bằng public key qua JWKS, KHÔNG chia sẻ khóa ký"
  - "DEC-AUTH-08: claims tối thiểu - user_id, locale, tier, plus exp/iat/iss/aud chuẩn"
  - "DEC-AUTH-09: refresh token có rotation + thu hồi (lưu hash refresh trong DB; dùng một lần, cấp refresh mới)"
  - "DEC-AUTH-10: JWKS hỗ trợ nhiều kid (xoay khóa ký không downtime); khóa ký lấy qua TASK-INFRA-003"

language: "Go 1.22 (auth-svc); thư viện JWT (golang-jwt) + JWKS; khóa RS256"
service: shopass/services/auth/
new_files:
  - services/auth/internal/auth/token.go
  - services/auth/internal/auth/jwks.go
  - services/auth/internal/auth/refresh.go
  - services/auth/internal/auth/login.go
  - services/auth/migrations/0005_refresh_token.up.sql
  - services/auth/migrations/0005_refresh_token.down.sql
  - services/auth/internal/auth/token_test.go
  - services/auth/internal/auth/refresh_test.go
modified_files:
  - services/auth/internal/auth/types.go            # struct Claims, TokenPair
allowed_tools:
  - file_read: services/auth/**
  - file_write: services/auth/**
  - bash: cd services/auth && go test ./...
disallowed_tools:
  - ký JWT bằng HS256 với secret chia sẻ giữa nhiều service (vi phạm DEC-AUTH-07; phải RS256 + JWKS)
  - lưu refresh token cleartext trong DB (phải lưu hash; vi phạm DEC-AUTH-09)
  - đặt access token TTL dài (vi phạm DEC-AUTH-06; khó thu hồi khi lộ)

effort_hours: 6
sub_tasks:
  - "1.0h: token.go - phát hành access JWT RS256 với claims user_id/locale/tier + exp/iss/aud"
  - "1.0h: jwks.go - expose /.well-known/jwks.json từ public key, hỗ trợ nhiều kid"
  - "1.0h: refresh.go - sinh refresh token, lưu hash, rotation một-lần + thu hồi"
  - "1.0h: login.go - verify mật khẩu (TASK-AUTH-001) -> cấp TokenPair; refresh -> cấp cặp mới"
  - "0.5h: 0005_refresh_token - bảng refresh_token (user_id, token_hash, expires_at, revoked_at)"
  - "1.0h: token_test.go - access verify được bằng JWKS; sai kid/hết hạn bị từ chối; claims đúng"
  - "0.5h: refresh_test.go - refresh dùng một lần (lần hai bị từ chối); rotation cấp token mới; thu hồi chặn"

risk_if_skipped: "JWT là vé thông hành cho mọi API. Dùng HS256 secret chia sẻ thì lộ một service là giả mạo được mọi phiên, và gateway phải giữ khóa ký (vỡ ranh giới TASK-INFRA-001). Access token TTL dài mà không thu hồi được thì một token lộ là phiên bị chiếm tới khi hết hạn. Refresh token lưu cleartext mà DB lộ là chiếm phiên dài hạn hàng loạt. Không rotation refresh thì token bị đánh cắp dùng lại mãi. Đây là nền xác thực mà WEB, EXT, MOBILE đều dựa vào."
---

## §1 - Mô tả (BCP-14 normative)

Service AUTH **MUST** phát hành cặp access/refresh token, expose JWKS để gateway verify, và quản refresh có rotation/thu hồi. Hợp đồng:

1. Đăng nhập thành công (mật khẩu verify qua TASK-AUTH-001) **MUST** trả một `TokenPair`: access JWT ngắn hạn + refresh token dài hạn (DEC-AUTH-06).
2. Access token **MUST** là JWT ký RS256 (DEC-AUTH-07). Khóa ký private lấy qua TASK-INFRA-003 (`auth/jwt-signing`), KHÔNG nhúng trong code/env, KHÔNG chia sẻ với service khác.
3. Access JWT **MUST** mang claims: `user_id`, `locale`, `tier`, cùng claim chuẩn `exp`, `iat`, `iss`, `aud` (DEC-AUTH-08). `aud` đặt đúng để gateway (TASK-INFRA-001 #2) verify khớp.
4. Access token **MUST** có TTL ngắn (ví dụ 15 phút, cấu hình được) (DEC-AUTH-06) để giới hạn cửa sổ rủi ro khi token lộ.
5. AUTH **MUST** expose JWKS tại `/.well-known/jwks.json` chứa public key tương ứng (DEC-AUTH-10). Mỗi khóa có `kid`; JWKS hỗ trợ nhiều `kid` đồng thời để xoay khóa ký không downtime.
6. Refresh token **MUST** dài hạn hơn (ví dụ 30 ngày) và **MUST** lưu dạng hash trong DB (DEC-AUTH-09), KHÔNG cleartext. Bảng `refresh_token` lưu `user_id`, `token_hash`, `expires_at`, `revoked_at`.
7. Refresh **MUST** rotation một-lần (DEC-AUTH-09): mỗi refresh token dùng được đúng một lần; khi đổi lấy cặp mới, token cũ bị đánh dấu đã dùng/thu hồi. Dùng lại token đã xoay -> từ chối.
8. AUTH **MUST** hỗ trợ thu hồi: token bị `revoked_at` không cấp được cặp mới (dùng cho đăng xuất và TASK-AUTH-005 suspend/delete).
9. AUTH **MUST** cung cấp `IssueTokenPair(ctx, userID) (TokenPair, error)` và `Refresh(ctx, refreshToken) (TokenPair, error)`; gateway chỉ verify access JWT (không gọi AUTH mỗi request).
10. Phát hiện tái sử dụng refresh token đã xoay **SHOULD** kích hoạt thu hồi cả họ token của phiên đó (token theft detection) - coi như dấu hiệu đánh cắp.
11. Claims `tier` **MUST** phản ánh gói người dùng (`free`/`premium`...) để gateway/route áp feature gating mà không truy DB mỗi request.
12. AUTH **MUST** ký bằng `kid` hiện hành và đưa `kid` vào header JWT, để gateway chọn đúng public key trong JWKS khi xoay khóa.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao RS256 + JWKS, không HS256 (DEC-AUTH-07)? HS256 dùng một secret đối xứng: ai verify được cũng ký được. Nếu gateway giữ secret đó để verify, một lần lộ là giả mạo mọi phiên. RS256 tách private (chỉ AUTH ký) khỏi public (gateway verify qua JWKS). Gateway không bao giờ chạm khóa ký, giữ ranh giới phát-hành vs verify (TASK-INFRA-001) sạch.

Vì sao access ngắn + refresh dài (DEC-AUTH-06)? Access token tự chứng thực (gateway verify không cần gọi AUTH), nên không thu hồi tức thời được - cách giới hạn rủi ro là cho nó sống ngắn. Refresh token sống dài để người dùng không phải đăng nhập lại liên tục, nhưng nó đi qua AUTH mỗi lần dùng nên thu hồi được. Cặp này cân bằng tiện lợi và kiểm soát.

Vì sao lưu hash refresh, không cleartext (DEC-AUTH-09)? Refresh token là credential dài hạn. Nếu DB lộ và token lưu cleartext, kẻ tấn công chiếm phiên dài hạn hàng loạt. Lưu hash (như mật khẩu) làm DB lộ không trực tiếp cho token dùng được.

Vì sao rotation một-lần (DEC-AUTH-09, §1 #7)? Refresh token tĩnh dùng mãi là rủi ro: bị đánh cắp một lần là dùng vô hạn. Rotation một-lần (mỗi refresh cấp một refresh mới và vô hiệu cái cũ) thu hẹp cửa sổ: token đánh cắp chỉ dùng được tới lần refresh hợp lệ kế tiếp của chủ thật.

Vì sao phát hiện tái sử dụng kích hoạt thu hồi họ (§1 #10)? Nếu một refresh token đã xoay bị dùng lại, hoặc chủ thật vừa dùng (và kẻ cắp đến sau) hoặc ngược lại - dù sao cũng là dấu hiệu hai bên cùng giữ token. Thu hồi cả họ token buộc đăng nhập lại, cắt kẻ cắp ra.

Vì sao nhiều kid trong JWKS (DEC-AUTH-10)? Xoay khóa ký định kỳ là thực hành tốt. Nếu JWKS chỉ có một khóa, lúc xoay sẽ có khoảng token cũ (ký khóa cũ) và token mới (ký khóa mới) cùng tồn tại. Giữ nhiều `kid` cho cả hai verify được trong giai đoạn chuyển, xoay không downtime - khớp với TASK-INFRA-001 #12.

Vì sao tier trong claims (§1 #11)? Feature gating (free vs premium) cần biết gói ở mỗi request. Nếu phải truy DB mỗi lần thì chậm và tốn. Đưa `tier` vào claims cho gateway/route quyết định ngay từ token; access TTL ngắn giữ tier không quá cũ khi người dùng nâng cấp.

---

## §3 - Hợp đồng API / DDL

### Migration refresh_token

```sql
-- services/auth/migrations/0005_refresh_token.up.sql
CREATE TABLE refresh_token (
  id          BIGSERIAL   PRIMARY KEY,
  user_id     BIGINT      NOT NULL REFERENCES app_user(id),
  token_hash  TEXT        NOT NULL,            -- hash của refresh token, KHÔNG cleartext
  family_id   UUID        NOT NULL,            -- nhóm token cùng một chuỗi rotation
  expires_at  TIMESTAMPTZ NOT NULL,
  revoked_at  TIMESTAMPTZ,                     -- NULL = còn hiệu lực
  used_at     TIMESTAMPTZ,                     -- đánh dấu đã xoay (dùng một lần)
  created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_rt_user ON refresh_token (user_id);
CREATE UNIQUE INDEX idx_rt_hash ON refresh_token (token_hash);
```

### Claims + phát hành access JWT (§1 #2, #3, #12)

```go
// services/auth/internal/auth/token.go
type Claims struct {
    UserID int64  `json:"user_id"`
    Locale string `json:"locale"`
    Tier   string `json:"tier"`
    jwt.RegisteredClaims
}

func (s *Service) issueAccess(u AppUser) (string, error) {
    now := time.Now()
    c := Claims{
        UserID: u.ID, Locale: u.Locale, Tier: u.Tier,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer: s.iss, Audience: jwt.ClaimStrings{s.aud},
            IssuedAt: jwt.NewNumericDate(now),
            ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)), // §1 #4 (ngắn)
        },
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodRS256, c)
    tok.Header["kid"] = s.currentKID // §1 #12
    return tok.SignedString(s.signingKey) // signingKey lấy qua TASK-INFRA-003
}
```

### Refresh rotation một-lần + theft detection (§1 #7, #8, #10)

```go
// services/auth/internal/auth/refresh.go
func (s *Service) Refresh(ctx context.Context, raw string) (TokenPair, error) {
    rt, err := s.repo.FindRefreshByHash(ctx, hash(raw))
    if err != nil { return TokenPair{}, ErrInvalidRefresh }
    if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
        return TokenPair{}, ErrInvalidRefresh
    }
    if rt.UsedAt != nil { // đã xoay rồi mà dùng lại = nghi đánh cắp (§1 #10)
        s.repo.RevokeFamily(ctx, rt.FamilyID)
        return TokenPair{}, ErrRefreshReuseDetected
    }
    s.repo.MarkUsed(ctx, rt.ID)                 // dùng một lần (§1 #7)
    return s.issuePairInFamily(ctx, rt.UserID, rt.FamilyID) // cấp cặp mới cùng family
}
```

---

## §4 - Acceptance criteria

1. Đăng nhập đúng mật khẩu -> trả `TokenPair` (access + refresh khác rỗng).
2. Access JWT verify được bằng public key từ `/.well-known/jwks.json` (chữ ký RS256 hợp lệ).
3. Access JWT chứa `user_id`, `locale`, `tier`, `exp`, `iss`, `aud` đúng giá trị.
4. Access JWT hết hạn (sau TTL) -> verify thất bại.
5. Access JWT ký bằng `kid` không có trong JWKS -> verify thất bại.
6. JWKS trả >=1 khóa; sau khi thêm khóa thứ hai (xoay), JWKS có 2 `kid`, token ký khóa mới verify được.
7. Refresh token lưu trong DB là hash, KHÔNG khớp giá trị thô (kiểm `token_hash` != raw).
8. `Refresh(validToken)` -> trả cặp mới; token cũ bị `used_at` set.
9. `Refresh` lại cùng token đã dùng -> `ErrRefreshReuseDetected`, và cả family bị thu hồi.
10. `Refresh(revokedToken)` -> `ErrInvalidRefresh`.
11. `Refresh(expiredToken)` -> `ErrInvalidRefresh`.
12. Sau thu hồi family (theft detection), refresh token khác cùng family cũng bị từ chối.

---

## §5 - Kiểm thử (verification)

```go
// services/auth/internal/auth/token_test.go
func TestAccess_VerifiableViaJWKS(t *testing.T) {
    s := newServiceWithKeys(t)
    pair, _ := s.IssueTokenPair(ctx, 90112)
    jwks := s.JWKS()
    claims, err := verifyWithJWKS(pair.Access, jwks)
    require.NoError(t, err)
    require.Equal(t, int64(90112), claims.UserID)
}

func TestAccess_Expired_Rejected(t *testing.T) {
    s := newServiceWithKeys(t); s.accessTTL = -time.Minute // đã hết hạn
    pair, _ := s.IssueTokenPair(ctx, 1)
    _, err := verifyWithJWKS(pair.Access, s.JWKS())
    require.Error(t, err)
}

func TestAccess_UnknownKID_Rejected(t *testing.T) {
    s := newServiceWithKeys(t)
    pair, _ := s.IssueTokenPair(ctx, 1)
    _, err := verifyWithJWKS(pair.Access, emptyJWKS()) // không có kid
    require.Error(t, err)
}

func TestJWKS_MultipleKID_AfterRotation(t *testing.T) {
    s := newServiceWithKeys(t)
    s.AddSigningKey(genKey(t), "key-2") // xoay
    require.Len(t, s.JWKS().Keys, 2)
    pair, _ := s.IssueTokenPair(ctx, 1) // ký bằng kid hiện hành
    _, err := verifyWithJWKS(pair.Access, s.JWKS())
    require.NoError(t, err)
}

// services/auth/internal/auth/refresh_test.go
func TestRefresh_OneTimeUse(t *testing.T) {
    s := newServiceWithKeys(t)
    pair, _ := s.IssueTokenPair(ctx, 1)
    p2, err := s.Refresh(ctx, pair.Refresh)
    require.NoError(t, err)
    require.NotEqual(t, pair.Refresh, p2.Refresh) // rotation cấp mới
    _, err2 := s.Refresh(ctx, pair.Refresh)       // dùng lại token cũ
    require.ErrorIs(t, err2, ErrRefreshReuseDetected)
}

func TestRefresh_StoredAsHash(t *testing.T) {
    s := newServiceWithKeys(t)
    pair, _ := s.IssueTokenPair(ctx, 1)
    row := s.repo.rawRow(t, 1)
    require.NotEqual(t, pair.Refresh, row.TokenHash) // lưu hash, không cleartext
}

func TestRefresh_ReuseRevokesFamily(t *testing.T) {
    s := newServiceWithKeys(t)
    pair, _ := s.IssueTokenPair(ctx, 1)
    p2, _ := s.Refresh(ctx, pair.Refresh)
    _, _ = s.Refresh(ctx, pair.Refresh)     // tái sử dụng → thu hồi family
    _, err := s.Refresh(ctx, p2.Refresh)     // token mới cùng family cũng bị chặn
    require.ErrorIs(t, err, ErrInvalidRefresh)
}
```

---

## §6 - Khung triển khai

Xem §3. `token.go` ký RS256 với private key lấy qua TASK-INFRA-003; header JWT mang `kid` hiện hành. `jwks.go` build JWKS từ public key(s); giữ nhiều `kid` để xoay không downtime, khớp gateway refresh JWKS (TASK-INFRA-001 #12). `refresh.go` lưu hash refresh + `family_id` cho rotation và theft detection. `login.go` nối: verify mật khẩu (TASK-AUTH-001) -> `IssueTokenPair`. Gateway chỉ verify access JWT bằng JWKS, không gọi AUTH mỗi request - đó là lý do access TTL phải ngắn.

---

## §7 - Phụ thuộc

- TASK-AUTH-001 - `Verify` mật khẩu trước khi cấp token; `app_user` cung cấp `locale`/`tier`.
- TASK-INFRA-001 - gateway verify access JWT bằng JWKS này; `aud` phải khớp.
- TASK-INFRA-003 - khóa ký RS256 lấy qua provider secret (`auth/jwt-signing`).
- TASK-AUTH-005 (downstream) - suspend/delete tài khoản dùng thu hồi refresh token.
- TASK-WEB-001 / TASK-EXT-005 / TASK-MOBILE-001 (downstream) - client dùng cặp token để gọi API.
- Thư viện: golang-jwt + JWKS.

---

## §8 - Payload ví dụ

### Phản hồi đăng nhập

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0yMDI2LTA2In0...",
  "refresh_token": "rt_9f2c...e004",
  "token_type": "Bearer",
  "expires_in": 900
}
```

### JWKS

```json
{
  "keys": [
    {"kty":"RSA","kid":"key-2026-06","use":"sig","alg":"RS256","n":"...","e":"AQAB"},
    {"kty":"RSA","kid":"key-2026-09","use":"sig","alg":"RS256","n":"...","e":"AQAB"}
  ]
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Danh sách thu hồi access token (introspection/blacklist) cho thu hồi tức thời trước khi access hết hạn - chỉ cần nếu TTL ngắn vẫn chưa đủ.
- Binding token với thiết bị (DPoP/mTLS) chống đánh cắp - tăng phòng thủ giai đoạn sau.
- Đăng nhập bằng số điện thoại + OTP (ngoài email + mật khẩu) - gắn TASK-NOTIF (SMS) khi mở.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| HS256 secret chia sẻ | review | lộ một service = giả mạo mọi phiên | RS256 + JWKS (§1 #2) |
| Access TTL dài | review | token lộ dùng tới khi hết hạn | TTL ngắn (§1 #4) |
| Refresh lưu cleartext | refresh test | DB lộ = chiếm phiên hàng loạt | Lưu hash (§1 #6) |
| Refresh dùng lại mãi | one-time test | token cắp dùng vô hạn | Rotation một-lần (§1 #7) |
| Đánh cắp refresh token | reuse test | chiếm phiên | Theft detection thu hồi family (§1 #10) |
| Xoay khóa gây downtime verify | JWKS test | token mới bị từ chối | Nhiều kid trong JWKS (§1 #5) |
| Gateway gọi AUTH mỗi request | thiết kế | chậm + AUTH thành nút cổ chai | Access tự chứng thực, TTL ngắn (§6) |
| tier cũ sau khi nâng cấp | TTL access | gating sai tạm thời | TTL ngắn làm tier mới hiệu lực nhanh |
| aud sai gateway từ chối | token test | đăng nhập được nhưng API 401 | `aud` khớp gateway (§1 #3) |

---

## §11 - Ghi chú

- RS256 + JWKS tách phát-hành (AUTH ký bằng private) khỏi verify (gateway dùng public), giữ ranh giới TASK-INFRA-001 sạch và không khóa ký nào rời AUTH.
- Access ngắn + refresh dài cân bằng: access tự chứng thực nên phải ngắn; refresh đi qua AUTH nên thu hồi được.
- Lưu hash refresh + rotation một-lần + theft detection là ba lớp bảo vệ credential dài hạn.
- Nhiều `kid` trong JWKS cho xoay khóa ký không downtime, khớp gateway refresh JWKS khi gặp `kid` lạ.
- `tier` trong claims cho feature gating không truy DB mỗi request; access TTL ngắn giữ tier không quá cũ.
- Gateway verify cục bộ bằng JWKS, không gọi AUTH mỗi request - đó là lý do access TTL ngắn là bắt buộc, không phải tùy chọn.

---

*Hết TASK-AUTH-002. Status: ready_to_implement (mục tiêu audit 10/10).*
