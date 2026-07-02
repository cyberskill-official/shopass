---
id: FR-AUTH-004
title: "Social login - OAuth Google/Facebook/Zalo + account linking"
module: AUTH
priority: SHOULD
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-AUTH-001, FR-AUTH-002, FR-INFRA-003, FR-WEB-001]
depends_on: [FR-AUTH-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.1 (auth), §4.3 (payment rails VN - Zalo phổ biến)"
  - "docs/... §3.8 (bảo mật), persona Chi/Huy/Linh (giảm ma sát đăng ký)"
source_decisions:
  - "DEC-AUTH-16: hỗ trợ OAuth 2.0 / OIDC với Google, Facebook, Zalo (Zalo phổ biến ở VN)"
  - "DEC-AUTH-17: dùng Authorization Code + PKCE; state chống CSRF; nonce chống replay (OIDC)"
  - "DEC-AUTH-18: liên kết identity provider vào app_user qua bảng social_identity (provider, subject), UNIQUE(provider, subject)"
  - "DEC-AUTH-19: khớp theo email đã verify để liên kết identity mới vào tài khoản hiện hữu; nếu email chưa verify thì KHÔNG auto-merge"
  - "DEC-AUTH-20: sau khi xác thực OAuth, phát hành cùng TokenPair của FR-AUTH-002 (một đường token thống nhất)"

language: "Go 1.22 (auth-svc); OAuth2/OIDC client; PKCE"
service: shopass/services/auth/
new_files:
  - services/auth/migrations/0007_social_identity.up.sql
  - services/auth/migrations/0007_social_identity.down.sql
  - services/auth/internal/auth/oauth.go
  - services/auth/internal/auth/oauth_google.go
  - services/auth/internal/auth/oauth_callback.go
  - services/auth/internal/auth/oauth_test.go
modified_files:
  - services/auth/internal/auth/types.go            # struct SocialIdentity
allowed_tools:
  - file_read: services/auth/**
  - file_write: services/auth/**
  - bash: cd services/auth && go test ./...
disallowed_tools:
  - bỏ qua state/PKCE/nonce (vi phạm DEC-AUTH-17; mở CSRF/replay)
  - auto-merge tài khoản theo email chưa verify (vi phạm DEC-AUTH-19; account takeover)
  - lưu OAuth client secret trong code/env (phải qua FR-INFRA-003)

effort_hours: 6
sub_tasks:
  - "1.0h: 0007_social_identity - bảng (user_id, provider, subject, UNIQUE(provider,subject))"
  - "1.5h: oauth.go - khởi tạo flow Authorization Code + PKCE + state; verify callback"
  - "1.0h: oauth_google.go - adapter Google OIDC (verify id_token, lấy sub/email_verified)"
  - "1.0h: oauth_callback.go - khớp/liên kết identity -> app_user; phát TokenPair (FR-AUTH-002)"
  - "1.0h: oauth_test.go - state sai bị từ chối; email verified -> liên kết; email chưa verify -> không auto-merge"
  - "0.5h: stub adapter Facebook/Zalo (cùng interface), test interface chung"

risk_if_skipped: "Social login là tính năng SHOULD giảm ma sát đăng ký cho persona VN (nhiều người ngại tạo mật khẩu mới; Zalo phổ biến). Nếu làm sai bảo mật, nó thành lỗ hổng: thiếu state là CSRF đăng nhập; thiếu PKCE/nonce là dễ replay; auto-merge theo email chưa verify là account takeover kinh điển (kẻ tấn công đăng ký provider với email nạn nhân rồi chiếm tài khoản). Vì là SHOULD, nó không chặn release, nhưng nếu ship thì phải ship đúng chuẩn OAuth/OIDC, không nửa vời."
---

## §1 - Mô tả (BCP-14 normative)

Service AUTH **SHOULD** hỗ trợ đăng nhập qua OAuth/OIDC với Google, Facebook, Zalo, liên kết identity vào `app_user` an toàn, và phát hành cùng TokenPair của FR-AUTH-002. Hợp đồng:

1. AUTH **SHOULD** hỗ trợ ít nhất Google (OIDC) ở slice này, với kiến trúc adapter cho Facebook và Zalo theo cùng interface (DEC-AUTH-16).
2. Flow **MUST** dùng Authorization Code + PKCE (DEC-AUTH-17): sinh `code_verifier`/`code_challenge`; tham số `state` ngẫu nhiên chống CSRF; với OIDC dùng `nonce` chống replay id_token.
3. Callback **MUST** kiểm `state` khớp giá trị đã phát (chống CSRF); `state` không khớp/thiếu -> từ chối.
4. Với OIDC (Google), AUTH **MUST** verify `id_token` (chữ ký qua JWKS của provider, `aud` = client id, `iss` đúng, `nonce` khớp) trước khi tin claims.
5. Migration `0007_social_identity` **MUST** tạo bảng `social_identity (id, user_id, provider, subject, created_at)` với `UNIQUE(provider, subject)` (DEC-AUTH-18): một identity provider chỉ gắn một `app_user`.
6. Liên kết identity vào tài khoản hiện hữu **MUST** chỉ khi email từ provider đã được provider đánh dấu verify (`email_verified = true`) và khớp `app_user.email` (DEC-AUTH-19). Email CHƯA verify -> KHÔNG auto-merge; tạo tài khoản mới hoặc yêu cầu bước xác minh.
7. Nếu chưa có `app_user` khớp, AUTH **MUST** tạo `app_user` mới (qua đường FR-AUTH-001, `pwd_hash` có thể NULL cho tài khoản chỉ-social) và gắn `social_identity`.
8. Sau xác thực thành công, AUTH **MUST** phát hành cùng `TokenPair` của FR-AUTH-002 (DEC-AUTH-20) - một đường token thống nhất, không token riêng cho social.
9. OAuth client secret và cấu hình provider **MUST** lấy qua FR-INFRA-003, KHÔNG nhúng trong code/env.
10. AUTH **MUST** xử lý trường hợp một `app_user` liên kết nhiều provider (Google + Facebook cùng một người), miễn mỗi `(provider, subject)` là duy nhất.
11. `state` và `code_verifier` **MUST** lưu tạm có thời hạn ngắn (ví dụ trong Redis với TTL vài phút) và dùng một lần, để callback chậm/lặp không bị lạm dụng.
12. AUTH **SHOULD** phát metric `social_login_total{provider, result}` (counter) - không kèm email/subject thô.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao Authorization Code + PKCE, không Implicit (DEC-AUTH-17)? Implicit flow trả token thẳng trên URL, dễ lộ qua lịch sử/referrer và đã bị khuyến nghị bỏ. Authorization Code đổi code lấy token ở backend; PKCE buộc bên đổi code phải biết `code_verifier` gốc, chống chặn-code. Đây là chuẩn hiện hành cho cả web và app.

Vì sao state + nonce (§1 #2, #3, #4)? `state` chống CSRF: không có nó, kẻ tấn công ép nạn nhân hoàn tất một flow đăng nhập của kẻ tấn công (login CSRF). `nonce` (OIDC) chống replay id_token: token cũ phát lại sẽ lệch nonce. Hai tham số này nhỏ nhưng là khác biệt giữa OAuth đúng và OAuth hổng.

Vì sao verify id_token thay vì tin thẳng (§1 #4)? id_token đến từ redirect do trình duyệt trung chuyển; tin claims mà không verify chữ ký là tin dữ liệu chưa xác thực. Verify qua JWKS của provider + kiểm `aud`/`iss`/`nonce` đảm bảo token đúng do provider phát cho đúng client này, không bị giả/chỉnh.

Vì sao chỉ merge khi email verified (DEC-AUTH-19)? Đây là lỗ hổng account takeover kinh điển: kẻ tấn công tạo tài khoản provider khai email của nạn nhân (provider chưa verify), rồi đăng nhập social và được merge vào tài khoản nạn nhân. Chỉ merge khi `email_verified = true` cắt đường này: provider đã chứng minh chủ thể sở hữu email đó.

Vì sao một đường token thống nhất (DEC-AUTH-20)? Sau khi xác thực (mật khẩu hay social), phần còn lại của hệ thống chỉ nên thấy một loại phiên. Phát cùng `TokenPair` của FR-AUTH-002 cho social tránh nhân đôi logic phiên/refresh và giữ gateway/route đồng nhất bất kể cách đăng nhập.

Vì sao state/verifier lưu tạm dùng một lần (§1 #11)? Nếu `state`/`code_verifier` sống lâu hoặc tái dùng được, một callback bị chặn/phát lại có thể bị lạm dụng. TTL ngắn + một-lần thu hẹp cửa sổ và khớp tính nhất thời của flow đăng nhập.

---

## §3 - Hợp đồng API / DDL

### Migration social_identity

```sql
-- services/auth/migrations/0007_social_identity.up.sql
CREATE TABLE social_identity (
  id         BIGSERIAL   PRIMARY KEY,
  user_id    BIGINT      NOT NULL REFERENCES app_user(id),
  provider   TEXT        NOT NULL CHECK (provider IN ('google','facebook','zalo')),
  subject    TEXT        NOT NULL,          -- 'sub' từ provider (định danh ổn định)
  created_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE (provider, subject)                 -- §1 #5
);
CREATE INDEX idx_si_user ON social_identity (user_id);
```

### OAuth flow start + callback (§1 #2, #3, #4, #6)

```go
// services/auth/internal/auth/oauth.go
func (s *Service) StartOAuth(ctx context.Context, provider string) (authURL string, err error) {
    verifier := pkce.NewVerifier()
    state := randToken()
    s.tmp.Put(ctx, "oauth:"+state, oauthTmp{Verifier: verifier, Provider: provider}, 5*time.Minute) // §1 #11
    p := s.providers[provider]
    return p.AuthCodeURL(state, pkce.Challenge(verifier), oidcNonce(state)), nil
}

// services/auth/internal/auth/oauth_callback.go
func (s *Service) OAuthCallback(ctx context.Context, provider, code, state string) (TokenPair, error) {
    tmp, ok := s.tmp.Take(ctx, "oauth:"+state) // dùng một lần (§1 #11)
    if !ok { return TokenPair{}, ErrBadState } // state sai/thiếu (§1 #3)
    claims, err := s.providers[provider].ExchangeAndVerify(ctx, code, tmp.Verifier) // verify id_token (§1 #4)
    if err != nil { return TokenPair{}, err }

    uid, err := s.resolveUser(ctx, provider, claims) // §1 #6,#7
    if err != nil { return TokenPair{}, err }
    return s.IssueTokenPair(ctx, uid) // cùng đường token (§1 #8)
}

func (s *Service) resolveUser(ctx context.Context, provider string, c OIDCClaims) (int64, error) {
    if uid, ok := s.repo.FindBySocial(ctx, provider, c.Subject); ok { return uid, nil }
    if c.EmailVerified { // chỉ merge khi email verified (§1 #6)
        if u, ok := s.repo.FindByEmail(ctx, c.Email); ok {
            s.repo.LinkSocial(ctx, u.ID, provider, c.Subject)
            return u.ID, nil
        }
    }
    return s.createUserWithSocial(ctx, provider, c) // tạo mới (§1 #7)
}
```

---

## §4 - Acceptance criteria

1. Migration -> bảng `social_identity` với `UNIQUE(provider, subject)`.
2. `StartOAuth("google")` -> trả authURL chứa `state`, `code_challenge`, `nonce`.
3. Callback với `state` không khớp/thiếu -> `ErrBadState`.
4. Callback với `id_token` chữ ký sai (không verify được qua JWKS provider) -> từ chối.
5. Callback với `id_token` `nonce` không khớp -> từ chối.
6. Identity mới (chưa có app_user khớp) -> tạo `app_user` mới + `social_identity`; phát TokenPair.
7. Identity có email `email_verified=true` khớp `app_user` hiện hữu -> liên kết vào tài khoản đó (không tạo trùng).
8. Identity có email khớp nhưng `email_verified=false` -> KHÔNG auto-merge (tạo mới hoặc yêu cầu xác minh).
9. Đăng nhập lại cùng provider+subject -> tìm `app_user` cũ qua `FindBySocial`, không tạo trùng.
10. Một `app_user` liên kết Google rồi Facebook (subject khác) -> hai dòng `social_identity`, cùng `user_id`.
11. OAuth client secret đọc qua provider secret (FR-INFRA-003), không có trong code/env (grep).
12. Sau xác thực social -> trả cùng cấu trúc `TokenPair` như đăng nhập mật khẩu (FR-AUTH-002).

---

## §5 - Kiểm thử (verification)

```go
// services/auth/internal/auth/oauth_test.go
func TestCallback_BadState_Rejected(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    _, err := s.OAuthCallback(ctx, "google", "code", "unknown-state")
    require.ErrorIs(t, err, ErrBadState)
}

func TestCallback_NonceMismatch_Rejected(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    url, _ := s.StartOAuth(ctx, "google")
    state := extractState(url)
    s.fakeProvider.SetIDToken(idTokenWithNonce("wrong")) // nonce sai
    _, err := s.OAuthCallback(ctx, "google", "code", state)
    require.Error(t, err)
}

func TestResolve_VerifiedEmail_LinksExisting(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    existing := s.seedUser(t, "chi@x.com")
    uid, _ := s.resolveUser(ctx, "google", OIDCClaims{Subject: "g-1", Email: "chi@x.com", EmailVerified: true})
    require.Equal(t, existing, uid) // liên kết vào tài khoản cũ
}

func TestResolve_UnverifiedEmail_NoMerge(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    existing := s.seedUser(t, "chi@x.com")
    uid, _ := s.resolveUser(ctx, "google", OIDCClaims{Subject: "g-2", Email: "chi@x.com", EmailVerified: false})
    require.NotEqual(t, existing, uid) // KHÔNG merge — tài khoản mới
}

func TestResolve_ReturningUser_NoDuplicate(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    u1, _ := s.resolveUser(ctx, "google", OIDCClaims{Subject: "g-3", Email: "a@x.com", EmailVerified: true})
    u2, _ := s.resolveUser(ctx, "google", OIDCClaims{Subject: "g-3", Email: "a@x.com", EmailVerified: true})
    require.Equal(t, u1, u2) // cùng subject → cùng user
}

func TestResolve_MultiProvider_OneUser(t *testing.T) {
    s := newServiceWithFakeProvider(t)
    uid, _ := s.resolveUser(ctx, "google", OIDCClaims{Subject: "g-4", Email: "b@x.com", EmailVerified: true})
    s.resolveUser(ctx, "facebook", OIDCClaims{Subject: "f-4", Email: "b@x.com", EmailVerified: true})
    links := s.repo.SocialLinks(t, uid)
    require.Len(t, links, 2)
}
```

---

## §6 - Khung triển khai

Xem §3. Adapter provider theo một interface chung (`AuthCodeURL`, `ExchangeAndVerify`) - Google triển khai trước, Facebook/Zalo stub cùng interface để mở rộng không sửa lõi. `state`/`code_verifier` lưu Redis TTL ngắn, dùng một lần. id_token Google verify qua JWKS của Google. Merge chỉ khi `email_verified`. Mọi đường kết thúc ở `IssueTokenPair` (FR-AUTH-002). Client secret qua FR-INFRA-003. Là FR SHOULD: nếu cắt slice, không chặn release, nhưng nếu làm thì theo đúng chuẩn ở §1.

---

## §7 - Phụ thuộc

- FR-AUTH-002 - phát hành `TokenPair` sau xác thực social (đường token thống nhất).
- FR-AUTH-001 - tạo `app_user` mới cho tài khoản chỉ-social (pwd_hash NULL hợp lệ).
- FR-INFRA-003 - OAuth client secret/cấu hình provider qua provider secret.
- FR-WEB-001 (downstream) - UI nút "Đăng nhập với Google/Facebook/Zalo".
- Thư viện: OAuth2/OIDC client + PKCE; hạ tầng Redis cho state tạm.

---

## §8 - Payload ví dụ

### Bắt đầu flow

```http
GET /v1/auth/oauth/google/start HTTP/1.1
-> 302 Location: https://accounts.google.com/o/oauth2/v2/auth?client_id=...&code_challenge=...&state=...&nonce=...&response_type=code
```

### Callback -> token (giống đăng nhập mật khẩu)

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiI...",
  "refresh_token": "rt_...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Liên kết/gỡ provider từ trang cài đặt (quản lý nhiều identity sau khi đăng nhập) - UI giai đoạn sau.
- Đăng nhập Apple (nếu mở mobile iOS) - thêm adapter khi cần.
- Hợp nhất tài khoản trùng (một người lỡ tạo hai tài khoản qua hai đường) - quy trình merge có xác minh sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Thiếu state | callback test | login CSRF | Bắt buộc state khớp (§1 #3) |
| Thiếu PKCE | review | chặn-code đổi token | Authorization Code + PKCE (§1 #2) |
| Không verify id_token | review | tin claims giả | Verify JWKS provider + aud/iss/nonce (§1 #4) |
| Auto-merge email chưa verify | resolve test | account takeover | Chỉ merge khi email_verified (§1 #6) |
| state/verifier tái dùng | take-once | lạm dụng callback lặp | TTL ngắn + một lần (§1 #11) |
| Client secret trong env | grep CI | rò rỉ | Qua FR-INFRA-003 (§1 #9) |
| (provider,subject) trùng | UNIQUE | hai user một identity | UNIQUE(provider,subject) (§1 #5) |
| Token riêng cho social | thiết kế | nhân đôi logic phiên | Cùng TokenPair (§1 #8) |
| Email/subject lọt vào metric | review nhãn | rò rỉ | Nhãn chỉ provider/result (§1 #12) |

---

## §11 - Ghi chú

- Authorization Code + PKCE là chuẩn hiện hành; Implicit flow đã bị khuyến nghị bỏ vì lộ token qua URL.
- `state` (chống CSRF) và `nonce` (chống replay) nhỏ nhưng là khác biệt giữa OAuth đúng và OAuth hổng.
- Merge chỉ khi `email_verified` cắt account takeover kinh điển (kẻ tấn công khai email nạn nhân ở provider chưa verify).
- Một đường token thống nhất (cùng TokenPair FR-AUTH-002) giữ phần còn lại của hệ thống chỉ thấy một loại phiên.
- Là FR SHOULD: giảm ma sát đăng ký cho persona VN (Zalo phổ biến) nhưng không chặn release; nếu ship thì ship đúng chuẩn.
- Adapter interface chung cho Google/Facebook/Zalo để thêm provider không sửa lõi.

---

*Hết FR-AUTH-004. Status: ready_to_implement (mục tiêu audit 10/10).*
