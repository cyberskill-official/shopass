---
id: FR-COMPLY-005
title: "Cưỡng chế no-cleartext + token-not-on-server - argon2id cho mật khẩu, KHÔNG lưu token phiên sàn trên server, secrets trong Vault, audit gate CI"
module: COMPLY
priority: MUST
status: ready_to_implement
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-INFRA-003, FR-AUTH-001, FR-AUTH-003, FR-EXT-003, FR-TRUST-003]
depends_on: [FR-INFRA-003]
blocks: [FR-TRUST-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (no-cleartext credential, token không rời client, argon2id, Vault)"
  - "docs/... §5.5 (PDPL: cookie/session token lưu trữ phải tuân PDPL; chủ trương KHÔNG lưu token phiên trên server)"
source_decisions:
  - "DEC-COMPLY-19: audit gate trong CI quét cấm pattern lưu cleartext/token sàn; fail build nếu phát hiện"
  - "DEC-COMPLY-20: token phiên sàn (Shopee/TikTok/Lazada) KHÔNG bao giờ rời client; backend KHÔNG có cột/log/biến nào chứa token sàn"
  - "DEC-COMPLY-21: mật khẩu chỉ tồn tại dạng argon2id PHC (tái dùng FR-AUTH-001); secrets ứng dụng chỉ trong Vault (tái dùng FR-INFRA-003)"
  - "DEC-COMPLY-22: cưỡng chế là kiểm tra tĩnh (CI gate) + kiểm tra động (test khẳng định payload từ extension không chứa cookie/token)"

language: "Go 1.22 (comply-svc audit lib) + shell (CI gate) + quét repo"
service: shopass/services/comply/
new_files:
  - services/comply/internal/audit/scan.go
  - services/comply/internal/audit/rules.go
  - services/comply/internal/audit/scan_test.go
  - services/comply/scripts/no_cleartext_gate.sh
  - services/comply/internal/audit/payload_guard.go
  - services/comply/internal/audit/payload_guard_test.go
modified_files:
  - .github/workflows/ci.yml      # thêm bước no-cleartext audit gate (chặn merge)
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - thêm cột/biến/log lưu token phiên sàn ở backend (vi phạm DEC-COMPLY-20)
  - lưu mật khẩu dạng cleartext hoặc băm yếu (md5/sha1) (vi phạm DEC-COMPLY-21)
  - đọc secret ứng dụng từ file .env commit vào repo thay vì Vault (vi phạm DEC-COMPLY-21)

effort_hours: 5
sub_tasks:
  - "1.0h: rules.go - tập pattern cấm (cleartext password, token sàn, secret hardcode)"
  - "1.0h: scan.go - quét cây mã/migration, trả vi phạm có vị trí dòng"
  - "0.5h: no_cleartext_gate.sh - gọi scan, exit !=0 khi có vi phạm (CI gate)"
  - "1.0h: payload_guard.go - khẳng định payload nhận từ extension KHÔNG chứa cookie/token field"
  - "1.0h: scan_test.go - fixture vi phạm bị bắt; mã sạch pass"
  - "0.5h: payload_guard_test.go - payload có cookie/token -> reject; payload tối thiểu -> pass"

risk_if_skipped: "Đây là điểm chứng minh kỹ thuật cho lời hứa niềm tin cốt lõi của SănDeal (hậu-Honey): KHÔNG gửi cookie/mật khẩu, token không rời client (§5.4). ~45% người tiêu dùng VN lo ngại lừa đảo/lộ dữ liệu (Ken Research) - nếu lộ ra backend có lưu token phiên sàn thì sụp đổ niềm tin và vi phạm PDPL hạng High (§9). FR-AUTH-001 đặt argon2id, FR-INFRA-003 đặt Vault, nhưng không có cưỡng chế tự động thì quy tắc bị phá âm thầm qua một commit sơ ý. CI gate + payload guard biến lời hứa thành ràng buộc máy kiểm. FR-TRUST-003 (audit độc lập) dựa trên cưỡng chế này."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** cưỡng chế hai bất biến bảo mật của SănDeal bằng kiểm tra tự động: (a) no-cleartext - mật khẩu chỉ tồn tại dạng argon2id, secrets chỉ trong Vault; (b) token-not-on-server - token phiên sàn không bao giờ rời client và không có mặt ở backend. Hợp đồng:

1. **MUST** cung cấp bộ quét tĩnh `audit.Scan(root string) ([]Finding, error)` quét cây mã + migration tìm pattern cấm, trả vi phạm kèm `file` + `line` + `rule`.
2. **MUST** định nghĩa tập quy tắc cấm (DEC-COMPLY-19) tối thiểu:
    - Cột/biến tên gợi ý lưu cleartext mật khẩu (ví dụ `password TEXT` không phải `pwd_hash`, `plain_password`).
    - Cột/biến/log lưu token phiên sàn (ví dụ `shopee_token`, `session_cookie`, `platform_access_token` ở backend).
    - Secret hardcode (khóa API/DB password chuỗi literal) thay vì đọc từ Vault (FR-INFRA-003).
    - Băm mật khẩu yếu (`md5`, `sha1`) cho credential.
3. **MUST** cung cấp CI gate `no_cleartext_gate.sh` gọi `Scan` và thoát mã khác 0 khi có vi phạm (DEC-COMPLY-22), chặn merge. Tích hợp vào `.github/workflows/ci.yml`.
4. **MUST** cưỡng chế token-not-on-server (DEC-COMPLY-20): backend KHÔNG có cột, biến, hoặc dòng log nào chứa token phiên sàn. Quy tắc quét bắt mọi tên trường khớp danh sách token sàn.
5. **MUST** cung cấp payload guard động `audit.GuardPayload(p map[string]any) error` (DEC-COMPLY-22): khẳng định payload nhận từ extension chỉ chứa trường cho phép (productId/price/qty...) và KHÔNG chứa `cookie`, `token`, `session`, `authorization` (tôn trọng FR-EXT-003 tối thiểu hóa).
6. **MUST** tái dùng argon2id PHC của FR-AUTH-001 cho mật khẩu (DEC-COMPLY-21): quy tắc quét xác nhận không có đường ghi mật khẩu nào ngoài hàm hash chuẩn.
7. **MUST** tái dùng Vault/Secrets Manager của FR-INFRA-003 cho secrets (DEC-COMPLY-21): quy tắc quét cấm secret literal trong mã và file `.env` commit.
8. **MUST** trả `Finding` xác định (ổn định thứ tự) để CI log rõ ràng và để test so khớp được.
9. **MUST** phân biệt true-positive với mã hợp lệ (ví dụ biến tên `token` của JWT nội bộ KHÁC token phiên sàn): quy tắc nhắm danh sách token SÀN cụ thể, có cơ chế allowlist dòng (`// audit:allow <lý do>`) cho ngoại lệ có kiểm soát.
10. **SHOULD** phát báo cáo tóm tắt số vi phạm theo rule khi chạy gate; log ở CI.
11. **MUST** đảm bảo allowlist không thành lỗ hổng: mỗi `// audit:allow` phải kèm lý do; quét đếm và liệt kê allowlist để review định kỳ.
12. **MUST** cung cấp test fixture cả hai chiều: mã vi phạm bị bắt (no false-negative trên tập đã biết), mã sạch không báo nhầm (kiểm soát false-positive).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao cưỡng chế tự động, không chỉ quy ước (DEC-COMPLY-22)?** FR-AUTH-001 nói "dùng argon2id" và FR-INFRA-003 nói "secrets trong Vault", nhưng quy ước do con người giữ sẽ bị phá qua một commit vội. CI gate biến quy tắc thành ràng buộc máy kiểm: build đỏ nếu ai đó thêm cột `password TEXT` hay log token sàn. Đây là khác biệt giữa "chính sách trên giấy" và "chính sách thực thi".

**Vì sao token-not-on-server là bất biến tuyệt đối (DEC-COMPLY-20)?** Toàn bộ định vị niềm tin hậu-Honey của SănDeal đặt trên việc backend KHÔNG bao giờ thấy token phiên sàn của user. Extension đọc giỏ hàng bằng phiên của chính user (session piggyback) và chỉ gửi về dữ liệu tối thiểu. Nếu một token sàn lọt vào backend, lời hứa sụp đổ. Quét tĩnh bắt mọi tên trường token sàn ở backend là rào cứng.

**Vì sao payload guard động bên cạnh quét tĩnh (§1 #5)?** Quét tĩnh bắt mã; payload guard bắt dữ liệu chạy thực. Ngay cả khi mã sạch, một payload bất ngờ từ extension (do bug client) có thể mang theo cookie. Guard từ chối tại biên: backend chỉ nhận trường cho phép, drop mọi thứ khác. Hai lớp (tĩnh + động) phủ cả "mã sai" lẫn "dữ liệu sai".

**Vì sao allowlist có lý do bắt buộc (§1 #9, #11)?** Quét tĩnh sẽ có false-positive (biến `token` của JWT nội bộ hợp lệ). Cấm tuyệt đối thì gate vô dụng vì luôn đỏ; cho qua tùy tiện thì gate vô nghĩa. Giải pháp: allowlist từng dòng kèm lý do, đếm và liệt kê để review. Ngoại lệ tồn tại nhưng có dấu vết và bị soi.

**Vì sao đây là nền cho FR-TRUST-003 (§7)?** Audit bảo mật độc lập (FR-TRUST-003) cần chứng minh "không gửi cookie/mật khẩu". Bộ quét + payload guard ở đây cung cấp bằng chứng tự động, lặp lại được mà auditor có thể chạy. Cưỡng chế nội bộ làm audit độc lập đáng tin hơn.

---

## §3 - Hợp đồng API / DDL

### Quy tắc + quét (Go)

```go
// services/comply/internal/audit/rules.go
type Rule struct {
    Name    string
    Pattern *regexp.Regexp
    Hint    string
}

// Token phien SAN cu the (KHAC token JWT noi bo).
var bannedRules = []Rule{
    {"cleartext_password",
        regexp.MustCompile(`(?i)\b(plain_?password|password)\s+TEXT\b`),
        "Mat khau phai la pwd_hash argon2id (FR-AUTH-001)"},
    {"platform_session_token",
        regexp.MustCompile(`(?i)(shopee|tiktok|lazada)_?(token|cookie|session)|platform_(access_)?token`),
        "Token phien san KHONG duoc co o backend (token-not-on-server)"},
    {"hardcoded_secret",
        regexp.MustCompile(`(?i)(api_?key|db_?password)\s*[:=]\s*["'][A-Za-z0-9/+]{12,}["']`),
        "Secret phai doc tu Vault (FR-INFRA-003)"},
    {"weak_password_hash",
        regexp.MustCompile(`(?i)\b(md5|sha1)\s*\(`),
        "Cam bam yeu cho credential; dung argon2id"},
}

// services/comply/internal/audit/scan.go
type Finding struct {
    File string
    Line int
    Rule string
    Hint string
}

// Scan quet cay ma, bo qua dong co // audit:allow.
func Scan(root string) ([]Finding, error) {
    var out []Finding
    // ... walk files, match rules, skip allowlisted lines, sort on dinh ...
    return out, nil
}
```

### Payload guard động + CI gate

```go
// services/comply/internal/audit/payload_guard.go
var forbiddenKeys = []string{"cookie", "token", "session", "authorization", "set-cookie"}

// GuardPayload tu choi payload chua truong nhay cam tu extension.
func GuardPayload(p map[string]any) error {
    for k := range p {
        lk := strings.ToLower(k)
        for _, f := range forbiddenKeys {
            if strings.Contains(lk, f) {
                return fmt.Errorf("%w: %s", ErrForbiddenField, k)
            }
        }
    }
    return nil
}
```

```sh
# services/comply/scripts/no_cleartext_gate.sh
#!/usr/bin/env bash
set -euo pipefail
n=$(go run ./services/comply/cmd/auditscan --root . --count)
if [ "$n" -ne 0 ]; then
  echo "no-cleartext gate FAILED: $n vi pham (xem log tren)"
  exit 1
fi
echo "no-cleartext gate PASSED"
```

---

## §4 - Acceptance criteria

1. `Scan` trên fixture chứa `password TEXT` -> trả Finding rule `cleartext_password` với đúng file/line.
2. `Scan` trên fixture chứa `shopee_token` -> trả Finding rule `platform_session_token`.
3. `Scan` trên fixture chứa `api_key = "AKIA..."` -> trả Finding rule `hardcoded_secret`.
4. `Scan` trên fixture dùng `md5(` cho mật khẩu -> trả Finding rule `weak_password_hash`.
5. `Scan` trên mã sạch (chỉ `pwd_hash`, đọc secret từ Vault) -> 0 Finding.
6. Dòng có `// audit:allow lý do` -> bị bỏ qua, không tính vi phạm.
7. `no_cleartext_gate.sh` thoát mã != 0 khi có >= 1 vi phạm; = 0 khi sạch.
8. `GuardPayload` với payload chứa `cookie` -> lỗi `ErrForbiddenField`.
9. `GuardPayload` với payload chứa `authorization` (hoa/thường) -> lỗi.
10. `GuardPayload` với payload tối thiểu (`productId`, `price`, `qty`) -> nil (pass).
11. Thứ tự `Finding` ổn định giữa hai lần chạy cùng input (xác định).
12. Báo cáo gate liệt kê số vi phạm theo rule và danh sách allowlist hiện có.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/audit/scan_test.go
func TestScan_CatchesPlatformToken(t *testing.T) {
    dir := writeFixture(t, "repo.go", `var shopeeToken = readCookie()`)
    f, _ := Scan(dir)
    require.True(t, hasRule(f, "platform_session_token"))
}

func TestScan_CatchesCleartextPassword(t *testing.T) {
    dir := writeFixture(t, "schema.sql", `CREATE TABLE u (password TEXT);`)
    f, _ := Scan(dir)
    require.True(t, hasRule(f, "cleartext_password"))
}

func TestScan_CleanCodePasses(t *testing.T) {
    dir := writeFixture(t, "ok.sql", `CREATE TABLE u (pwd_hash TEXT NOT NULL);`)
    f, _ := Scan(dir)
    require.Empty(t, f)
}

func TestScan_AllowlistSkipped(t *testing.T) {
    dir := writeFixture(t, "x.go", `var token = jwtInternal() // audit:allow JWT noi bo, khong phai token san`)
    f, _ := Scan(dir)
    require.Empty(t, f)
}

// services/comply/internal/audit/payload_guard_test.go
func TestGuard_RejectsCookie(t *testing.T) {
    err := GuardPayload(map[string]any{"productId": 1, "cookie": "abc"})
    require.ErrorIs(t, err, ErrForbiddenField)
}

func TestGuard_RejectsAuthorizationAnyCase(t *testing.T) {
    err := GuardPayload(map[string]any{"Authorization": "Bearer x"})
    require.ErrorIs(t, err, ErrForbiddenField)
}

func TestGuard_MinimalPayloadPasses(t *testing.T) {
    err := GuardPayload(map[string]any{"productId": 1, "price": 89000, "qty": 2})
    require.NoError(t, err)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: rules.go (tập quy tắc) -> scan.go (walk + match + allowlist + sort) -> scan_test.go -> payload_guard.go -> payload_guard_test.go -> no_cleartext_gate.sh -> nối vào ci.yml. Gate chạy ở mọi PR, chặn merge nếu đỏ. `GuardPayload` được gọi tại biên BFF/EXT-ingest (FR-EXT-003) trên mọi payload nhận từ extension trước khi xử lý. Quy tắc nhắm danh sách token SÀN cụ thể để hạn chế false-positive; allowlist dòng cho ngoại lệ hiếm và hợp lệ.

---

## §7 - Phụ thuộc

- **FR-INFRA-003** - Vault/Secrets Manager; quy tắc `hardcoded_secret` cưỡng chế đọc secret từ đây.
- **FR-AUTH-001** - argon2id PHC; quy tắc `cleartext_password`/`weak_password_hash` cưỡng chế.
- **FR-AUTH-003 (liên quan)** - liên kết tài khoản sàn KHÔNG lưu token; quy tắc `platform_session_token` bảo vệ.
- **FR-EXT-003 (liên quan)** - tối thiểu hóa dữ liệu client; `GuardPayload` là lớp biên backend tương ứng.
- **FR-TRUST-003 (downstream)** - audit độc lập dùng bộ quét + guard làm bằng chứng tự động.
- Lib: `regexp`, `strings`; CI: GitHub Actions.

---

## §8 - Payload ví dụ

### Payload hợp lệ từ extension (tối thiểu hóa, qua được guard)

```json
{
  "platform": "shopee",
  "items": [ { "productId": "90112", "price": 89000, "qty": 2 } ]
}
```

### Payload bị chặn (chứa trường cấm) - guard trả lỗi

```json
{
  "platform": "shopee",
  "items": [ { "productId": "90112", "price": 89000 } ],
  "cookie": "SPC_SI=...; SPC_EC=..."
}
```

```
400 Bad Request
{ "error": "forbidden_field", "field": "cookie" }
```

---

## §9 - Câu hỏi mở

Đã chốt khung. Hoãn:
- Mở rộng quy tắc sang quét bí mật trong lịch sử git (secret scanning sâu) - bổ sung công cụ chuyên dụng sau.
- Quét nhị phân/artifact build cho secret nhúng - giai đoạn CI nâng cao.
- Tự động phân loại false-positive bằng heuristic ngữ cảnh - hiện dựa allowlist thủ công có lý do.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Thêm cột lưu token sàn | `Scan` rule platform_session_token + AC #2 | vi phạm bất biến niềm tin | CI gate chặn merge |
| Mật khẩu cleartext/băm yếu | `Scan` rule cleartext/weak + AC #1,#4 | rủi ro lộ DB + PDPL | Gate đỏ; ép argon2id |
| Secret hardcode trong mã | `Scan` rule hardcoded_secret + AC #3 | rò khóa | Ép đọc từ Vault (FR-INFRA-003) |
| Payload extension lọt cookie | `GuardPayload` + AC #8 | token rời client lên server | Guard reject tại biên |
| False-positive làm gate luôn đỏ | allowlist §1 #9 | gate bị bỏ qua | `// audit:allow` kèm lý do, có review |
| Allowlist lạm dụng thành lỗ hổng | đếm/liệt kê §1 #11 | quy tắc vô nghĩa | Review định kỳ danh sách allowlist |
| Finding thứ tự ngẫu nhiên | sort §1 #8 + AC #11 | CI log khó đọc, test giòn | Sắp xếp xác định |
| Gate không chạy ở PR | ci.yml §1 #3 | quy tắc không cưỡng chế | Bước gate bắt buộc, chặn merge |

---

## §11 - Ghi chú

- Đây là điểm chứng minh kỹ thuật cho lời hứa niềm tin hậu-Honey: no-cleartext + token-not-on-server.
- Hai lớp cưỡng chế: quét tĩnh (bắt mã sai) + payload guard động (bắt dữ liệu sai tại biên).
- Token phiên sàn là bất biến tuyệt đối: backend không có cột/biến/log nào chứa nó; quy tắc nhắm danh sách token sàn cụ thể.
- Tái dùng argon2id (FR-AUTH-001) và Vault (FR-INFRA-003): FR này cưỡng chế hai quy tắc đó tự động, không phát minh lại.
- Allowlist có lý do bắt buộc cân bằng giữa gate cứng và false-positive; mọi ngoại lệ có dấu vết và bị review.
- FR-TRUST-003 (audit độc lập) dựa trên bộ quét + guard này làm bằng chứng tự động lặp lại được.

---

*Hết FR-COMPLY-005. Status: ready_to_implement (mục tiêu audit 10/10).*
