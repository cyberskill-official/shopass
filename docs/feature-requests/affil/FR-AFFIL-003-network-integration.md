---
id: FR-AFFIL-003
title: "Tích hợp affiliate network (Involve Asia / Accesstrade) - bảng cấu hình network template + sub_id tracking + last-click attribution + webhook postback ghi conversion (idempotent, có ký xác thực)"
module: AFFIL
priority: MUST
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-AFFIL-001, FR-AFFIL-002, FR-AFFIL-004, FR-AFFIL-005, FR-INFRA-003, FR-TRUST-005, NFR-AFFIL-001]
depends_on: [FR-AFFIL-002]
blocks: [FR-AFFIL-005, FR-TRUST-005]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §4.2 (affiliate network, last-click, sub_id, postback)"
  - "docs/... §8 (tích hợp Involve Asia/Accesstrade compliant), §3.7 (POST /v1/affiliate/link)"
source_decisions:
  - "DEC-AFFIL-11: cấu hình network lưu ở bảng affiliate_network (code, base_url, target_param, sub_id_param, postback_secret_ref) - secret tham chiếu Vault (FR-INFRA-003), KHÔNG lưu cleartext"
  - "DEC-AFFIL-12: deep link build từ template network; sub_id nhúng qua sub_id_param để postback đối soát ngược (last-click)"
  - "DEC-AFFIL-13: postback là webhook network -> SănDeal mang (sub_id, order_value, commission, status); xác thực bằng chữ ký/secret trước khi ghi"
  - "DEC-AFFIL-14: postback ánh xạ sub_id -> click_id (FR-AFFIL-001), ghi conversion 'pending'; status network 'approved' -> ConfirmConversion, 'rejected' -> RejectConversion"
  - "DEC-AFFIL-15: postback idempotent (network retry) + ghi raw payload vào affiliate_postback_log để truy vết tranh chấp"

language: "Go 1.22 (affil-svc); PostgreSQL 16; webhook nhận qua API Gateway (FR-INFRA-001)"
service: shopass/services/affil/
new_files:
  - services/affil/migrations/0003_affiliate_network.sql
  - services/affil/migrations/0004_affiliate_postback_log.sql
  - services/affil/internal/affil/network.go
  - services/affil/internal/api/postback.go
  - services/affil/internal/affil/postback_verify.go
  - services/affil/internal/api/postback_test.go
  - services/affil/internal/affil/postback_verify_test.go
modified_files:
  - services/affil/internal/api/router.go            # đăng ký POST /v1/affiliate/postback/{network}
allowed_tools:
  - file_read: services/affil/**
  - file_write: services/affil/**
  - bash: cd services/affil && go test ./...
disallowed_tools:
  - ghi conversion từ postback chưa xác thực chữ ký (vi phạm DEC-AFFIL-13 - giả mạo conversion)
  - lưu postback_secret cleartext trong DB (vi phạm DEC-AFFIL-11, PDPL no-cleartext)
  - xác nhận conversion 'confirmed' khi network chưa approve (vi phạm DEC-AFFIL-14, DEC-AFFIL-05)
  - bỏ ghi raw postback (vi phạm DEC-AFFIL-15 - mất bằng chứng tranh chấp)

effort_hours: 8
sub_tasks:
  - "1.0h: 0003_affiliate_network.sql - bảng affiliate_network (code, base_url, target_param, sub_id_param, postback_secret_ref, platform map)"
  - "0.5h: 0004_affiliate_postback_log.sql - bảng log raw payload + received_at + verified"
  - "1.0h: network.go - TemplateFor(platform_id) -> NetworkTemplate (dùng bởi FR-AFFIL-002); seed Involve Asia + Accesstrade"
  - "1.5h: postback_verify.go - xác thực chữ ký HMAC/secret từ Vault (FR-INFRA-003); reject nếu sai/thiếu"
  - "1.5h: postback.go - handler webhook: verify -> log raw -> RecordConversion(sub_id) -> map status network -> Confirm/Reject; idempotent"
  - "0.5h: router.go - đăng ký POST /v1/affiliate/postback/{network}"
  - "1.0h: postback_test.go - chữ ký sai -> 401 không ghi; hợp lệ -> conversion pending; approved -> confirmed; retry -> không trùng"
  - "1.0h: postback_verify_test.go + OTel metric affiliate_postback_total{network,result}"

risk_if_skipped: "FR-AFFIL-003 đóng vòng tiền của affiliate: nó biến một click (FR-AFFIL-002) thành conversion được network xác nhận, rồi thành cashback (FR-AFFIL-005). Không có nó thì deep link tạo ra không có template network đúng (FR-AFFIL-002 trả 503) và conversion không bao giờ được ghi/đối soát - dòng doanh thu affiliate đứng. Nếu ghi conversion từ postback chưa xác thực chữ ký thì bất kỳ ai cũng giả mạo được conversion để rút cashback (mất tiền + gian lận). Nếu lưu postback_secret cleartext thì lộ secret là lộ khả năng giả mạo toàn bộ postback (vi phạm PDPL no-cleartext §5.5). Nếu xác nhận conversion confirmed khi network chưa approve thì trả cashback cho đơn có thể bị đảo - lỗ. Nếu không ghi raw postback thì khi network tranh chấp số liệu, SănDeal không có bằng chứng. Đây là mắt xích đối soát tiền của toàn hệ affiliate."
---

## §1 - Mô tả (BCP-14 normative)

Service AFFIL **MUST** tích hợp affiliate network (Involve Asia, Accesstrade): lưu cấu hình network (template build link + tham chiếu secret), cung cấp template cho deep link (FR-AFFIL-002), và nhận webhook postback để ghi/đối soát conversion theo last-click, có xác thực chữ ký và idempotent. Hợp đồng:

1. **MUST** định nghĩa bảng `affiliate_network (id, code, base_url, target_param, sub_id_param, platform_id, postback_secret_ref, active)`: `code` - {`'involve_asia'`,`'accesstrade'`} (CHECK), `platform_id` REFERENCES `platform(id)` (network nào phục vụ sàn nào).
2. `postback_secret_ref` **MUST** là một tham chiếu tới secret trong Vault (FR-INFRA-003), KHÔNG phải secret cleartext (DEC-AFFIL-11). Cột này chứa key path/handle; secret thật đọc từ Vault lúc verify.
3. **MUST** expose `TemplateFor(platformID) (network string, tmpl NetworkTemplate, ok bool)` (dùng bởi FR-AFFIL-002): trả `base_url`, `target_param`, `sub_id_param` của network `active` phục vụ sàn đó. Không có network active -> `ok=false` (FR-AFFIL-002 trả 503).
4. **MUST** phục vụ webhook `POST /v1/affiliate/postback/{network}` nhận payload network mang tối thiểu `{sub_id, order_value, commission, status}` (`status` là trạng thái phía network, ví dụ `approved`/`pending`/`rejected`).
5. **MUST** xác thực mọi postback trước khi ghi (DEC-AFFIL-13): kiểm chữ ký HMAC (hoặc secret token) bằng `postback_secret` đọc từ Vault cho `{network}`. Chữ ký sai/thiếu -> `401`, KHÔNG ghi conversion, KHÔNG đổi sổ cái.
6. **MUST** ghi raw payload + header chữ ký + `received_at` + cờ `verified` vào `affiliate_postback_log` cho mọi postback (kể cả bị từ chối) (DEC-AFFIL-15) - bằng chứng truy vết tranh chấp.
7. Sau khi xác thực, **MUST** ánh xạ `sub_id` -> `click_id` (FR-AFFIL-001) và ghi conversion qua `RecordConversion`. `sub_id` không khớp click nào -> `404`/`ErrUnknownSubID`, không tạo conversion mồ côi (đồng nhất FR-AFFIL-001 #8).
8. **MUST** map trạng thái network sang vòng đời conversion (DEC-AFFIL-14): network `approved` -> `ConfirmConversion`; `rejected` -> `RejectConversion`; `pending` -> giữ conversion ở `pending`. KHÔNG đặt `confirmed` khi network chưa `approved` (DEC-AFFIL-05).
9. **MUST** idempotent với postback lặp (network retry): cùng `sub_id` postback nhiều lần **MUST NOT** tạo conversion trùng (dựa UNIQUE click_id ở FR-AFFIL-001); postback `approved` lặp trên conversion đã `confirmed` là no-op.
10. **MUST** thực thi last-click attribution (DEC-AFFIL-12): conversion quy về đúng click đã sinh `sub_id` trong postback - không đoán, không gán theo thời gian.
11. **MUST** đọc `postback_secret` từ Vault ở thời điểm verify, KHÔNG cache cleartext lâu trong tiến trình ngoài vòng đời request (giảm bề mặt lộ secret).
12. **SHOULD** phát OTel: `affiliate_postback_total{network, result}` (`result - {confirmed, pending, rejected, unauthorized, unknown_subid}`), `affiliate_conversion_confirmed_value_vnd` (histogram commission đã confirm).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao secret tham chiếu Vault, không cleartext (DEC-AFFIL-11, §1 #2)?** `postback_secret` là khóa để xác thực postback. Ai có nó giả mạo được conversion và rút cashback. Lưu cleartext trong DB nghĩa là một lần lộ DB là lộ khả năng giả mạo toàn bộ tiền affiliate, đồng thời vi phạm no-cleartext của PDPL (§5.5). Lưu một tham chiếu và đọc secret từ Vault lúc cần thu hẹp bề mặt lộ về đúng Vault.

**Vì sao xác thực chữ ký trước khi ghi (DEC-AFFIL-13)?** Webhook postback là endpoint công khai: bất kỳ ai biết URL đều POST được. Nếu ghi conversion từ payload chưa xác thực, kẻ xấu bịa `{sub_id, commission}` để bơm conversion giả rồi rút cashback. HMAC bằng secret chung với network đảm bảo chỉ network thật mới ghi được - đây là rào chống gian lận tài chính cốt lõi.

**Vì sao ghi raw postback dù bị từ chối (DEC-AFFIL-15)?** Đối soát affiliate hay có tranh chấp số liệu (network báo X conversion, ta thấy Y). Có raw payload + thời điểm + kết quả verify cho mỗi postback là bằng chứng để hòa giải. Ghi cả postback bị từ chối còn giúp phát hiện tấn công giả mạo (nhiều postback chữ ký sai = ai đó đang dò).

**Vì sao map trạng thái network sang vòng đời, không confirm sớm (DEC-AFFIL-14, §1 #8)?** Network có vòng đời riêng: đơn mới là `pending`, được duyệt sau khi qua cửa sổ trả hàng thành `approved`, hoặc bị `rejected`. SănDeal phải phản chiếu đúng: chỉ `approved` mới `ConfirmConversion` để mở khóa cashback. Confirm sớm khi network còn `pending` là trả tiền cho đơn có thể đảo - lỗ.

**Vì sao idempotent postback (§1 #9)?** Network gửi lại postback khi không nhận được 200 (mạng chập chờn). Nếu mỗi lần tạo một conversion, sổ cái phình và hoa hồng đếm trùng. Dựa UNIQUE click_id (FR-AFFIL-001) làm postback idempotent: lần đầu tạo, lần sau cập nhật trạng thái hoặc no-op.

**Vì sao đọc secret từ Vault mỗi lần verify (§1 #11)?** Cache secret cleartext lâu trong tiến trình là một bản sao secret nằm trong RAM dài hạn - bề mặt lộ thêm (core dump, debugger). Đọc đúng lúc verify, dùng xong bỏ, giữ secret ở Vault là nguồn sự thật. Có thể cache ngắn với TTL nếu cần hiệu năng, nhưng mặc định là đọc theo nhu cầu.

---

## §3 - Hợp đồng API / DDL

### Migrations

```sql
-- services/affil/migrations/0003_affiliate_network.sql
CREATE TABLE affiliate_network (
  id                  SMALLSERIAL PRIMARY KEY,
  code                TEXT     NOT NULL UNIQUE
                        CHECK (code IN ('involve_asia','accesstrade')),
  platform_id         SMALLINT NOT NULL REFERENCES platform(id),
  base_url            TEXT     NOT NULL,   -- gốc deep link network
  target_param        TEXT     NOT NULL,   -- tên tham số bọc URL đích
  sub_id_param        TEXT     NOT NULL,   -- tên tham số mang sub_id (last-click)
  postback_secret_ref TEXT     NOT NULL,   -- Vault key path; KHÔNG cleartext (§1 #2)
  active              BOOLEAN  NOT NULL DEFAULT true
);

-- services/affil/migrations/0004_affiliate_postback_log.sql
CREATE TABLE affiliate_postback_log (
  id          BIGSERIAL   PRIMARY KEY,
  network     TEXT        NOT NULL,
  raw_payload JSONB       NOT NULL,        -- bằng chứng tranh chấp (§1 #6)
  signature   TEXT,
  verified    BOOLEAN     NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_postback_received ON affiliate_postback_log (received_at DESC);
```

### Postback verify (Go)

```go
// services/affil/internal/affil/postback_verify.go

// VerifyPostback kiểm HMAC-SHA256 của body bằng secret đọc từ Vault cho network.
// Trả false nếu thiếu/sai chữ ký -> handler trả 401, KHÔNG ghi conversion (§1 #5).
func VerifyPostback(ctx context.Context, secrets SecretReader, network string, body []byte, gotSig string) (bool, error) {
    secret, err := secrets.Get(ctx, secretRefFor(network)) // đọc từ Vault (FR-INFRA-003)
    if err != nil {
        return false, err
    }
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    want := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(want), []byte(gotSig)), nil // so sánh hằng thời gian
}
```

### Postback handler (Go)

```go
// services/affil/internal/api/postback.go

func (h *Handler) HandlePostback(w http.ResponseWriter, req *http.Request) {
    network := req.PathValue("network")
    body, _ := io.ReadAll(req.Body)
    sig := req.Header.Get("X-Signature")

    ok, err := affil.VerifyPostback(req.Context(), h.secrets, network, body, sig)
    h.repo.LogPostback(req.Context(), network, body, sig, ok) // luôn ghi raw (§1 #6)
    if err != nil { writeErr(w, 500, "verify error"); return }
    if !ok {
        metrics.Postback(network, "unauthorized")
        writeErr(w, 401, "invalid signature"); return // KHÔNG ghi conversion (§1 #5)
    }

    var p PostbackPayload
    if err := json.Unmarshal(body, &p); err != nil { writeErr(w, 400, "bad payload"); return }

    cid, err := h.repo.RecordConversion(req.Context(), p.SubID, p.OrderValue, p.Commission, network)
    switch {
    case errors.Is(err, affil.ErrUnknownSubID):
        metrics.Postback(network, "unknown_subid")
        writeErr(w, 404, "unknown sub_id"); return       // không conversion mồ côi (§1 #7)
    case errors.Is(err, affil.ErrConversionExists):
        cid = h.repo.ConversionIDBySubID(req.Context(), p.SubID) // idempotent (§1 #9)
    case err != nil:
        writeErr(w, 500, "internal error"); return
    }

    switch p.Status { // map trạng thái network -> vòng đời (§1 #8)
    case "approved":
        _ = h.repo.ConfirmConversion(req.Context(), cid); metrics.Postback(network, "confirmed")
    case "rejected":
        _ = h.repo.RejectConversion(req.Context(), cid, "network rejected"); metrics.Postback(network, "rejected")
    default:
        metrics.Postback(network, "pending") // giữ pending
    }
    w.WriteHeader(200)
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `affiliate_network` và `affiliate_postback_log` tồn tại; `affiliate_network` có CHECK `code`.
2. Seed Involve Asia + Accesstrade -> `TemplateFor(platform_id)` trả `base_url/target_param/sub_id_param` đúng cho network active.
3. `TemplateFor` cho sàn không có network active -> `ok=false` (FR-AFFIL-002 sẽ trả 503).
4. `postback_secret_ref` lưu là tham chiếu Vault (không phải secret thô); review xác nhận không có cột secret cleartext.
5. Postback chữ ký sai -> `401`, KHÔNG tạo conversion, nhưng VẪN có dòng `affiliate_postback_log` với `verified=false`.
6. Postback hợp lệ với `sub_id` đã có click + `status='pending'` -> conversion `pending`, một dòng.
7. Postback hợp lệ với `status='approved'` -> conversion `confirmed`, `confirmed_at` được set.
8. Postback hợp lệ với `status='rejected'` -> conversion `rejected`.
9. Postback với `sub_id` không khớp click -> `404`, KHÔNG tạo conversion mồ côi.
10. Postback lặp (cùng `sub_id`, network retry) -> không tạo conversion thứ hai; postback `approved` lặp trên conversion đã `confirmed` là no-op (vẫn `200`).
11. Mọi postback (kể cả bị từ chối) đều có dòng raw trong `affiliate_postback_log`.
12. Metric `affiliate_postback_total{network,result}` tăng đúng nhãn theo từng nhánh.

---

## §5 - Kiểm thử (verification)

```go
// services/affil/internal/api/postback_test.go
func TestPostback_BadSignature_401_NoConversion(t *testing.T) {
    h, repo := setupPostback(t)
    rec := doSignedPOST(t, h, "involve_asia", `{"sub_id":"sd_x","order_value":250000,"commission":12000,"status":"approved"}`, "WRONGSIG")
    require.Equal(t, 401, rec.Code)
    require.Equal(t, 0, repo.ConversionCount())   // không ghi conversion
    require.Equal(t, 1, repo.PostbackLogCount())  // vẫn log raw (verified=false)
}

func TestPostback_Valid_Pending(t *testing.T) {
    h, repo := setupPostbackWithClick(t, "sd_x")
    body := `{"sub_id":"sd_x","order_value":250000,"commission":12000,"status":"pending"}`
    rec := doSignedPOST(t, h, "involve_asia", body, sign(body))
    require.Equal(t, 200, rec.Code)
    require.Equal(t, "pending", repo.StatusBySubID("sd_x"))
}

func TestPostback_Approved_Confirms(t *testing.T) {
    h, repo := setupPostbackWithClick(t, "sd_y")
    body := `{"sub_id":"sd_y","order_value":250000,"commission":12000,"status":"approved"}`
    doSignedPOST(t, h, "involve_asia", body, sign(body))
    require.Equal(t, "confirmed", repo.StatusBySubID("sd_y"))
}

func TestPostback_UnknownSubID_404_NoOrphan(t *testing.T) {
    h, repo := setupPostback(t)
    body := `{"sub_id":"sd_none","order_value":1,"commission":0,"status":"approved"}`
    rec := doSignedPOST(t, h, "involve_asia", body, sign(body))
    require.Equal(t, 404, rec.Code)
    require.Equal(t, 0, repo.ConversionCount())
}

func TestPostback_RetryIdempotent(t *testing.T) {
    h, repo := setupPostbackWithClick(t, "sd_z")
    body := `{"sub_id":"sd_z","order_value":250000,"commission":12000,"status":"approved"}`
    doSignedPOST(t, h, "involve_asia", body, sign(body))
    rec := doSignedPOST(t, h, "involve_asia", body, sign(body)) // network retry
    require.Equal(t, 200, rec.Code)
    require.Equal(t, 1, repo.ConversionCount()) // không trùng
    require.Equal(t, "confirmed", repo.StatusBySubID("sd_z"))
}
```

```go
// services/affil/internal/affil/postback_verify_test.go
func TestVerify_GoodSignature(t *testing.T) {
    secrets := fakeVault{"affil/involve_asia": "shh-secret"}
    body := []byte(`{"sub_id":"sd_x"}`)
    sig := hmacHex("shh-secret", body)
    ok, err := VerifyPostback(ctx, secrets, "involve_asia", body, sig)
    require.NoError(t, err); require.True(t, ok)
}

func TestVerify_TamperedBody(t *testing.T) {
    secrets := fakeVault{"affil/involve_asia": "shh-secret"}
    sig := hmacHex("shh-secret", []byte(`{"commission":1000}`))
    ok, _ := VerifyPostback(ctx, secrets, "involve_asia", []byte(`{"commission":9999999}`), sig)
    require.False(t, ok) // body đổi -> chữ ký không khớp
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0003_affiliate_network.sql` -> `0004_affiliate_postback_log.sql` -> `network.go` (`TemplateFor` + seed Involve Asia/Accesstrade với `postback_secret_ref` trỏ Vault) -> `postback_verify.go` (HMAC-SHA256, `hmac.Equal` so sánh hằng thời gian, secret từ FR-INFRA-003) -> `postback.go` (verify -> log raw -> RecordConversion -> map status -> Confirm/Reject, idempotent) -> đăng ký route `POST /v1/affiliate/postback/{network}` trong `router.go` (`http.ServeMux` Go 1.22 với `PathValue`) -> tests. Webhook KHÔNG đi qua JWT user (network không có JWT); thay vào đó xác thực bằng chữ ký HMAC. Secret đọc qua client Vault của FR-INFRA-003.

---

## §7 - Phụ thuộc

- **FR-AFFIL-002** - tạo deep link + sinh `sub_id` mà postback dùng để đối soát; `TemplateFor` của FR này phục vụ FR-AFFIL-002.
- **FR-AFFIL-001** - cung cấp `RecordConversion`/`ConfirmConversion`/`RejectConversion` + bảng `affiliate_click`/`affiliate_conversion`.
- **FR-INFRA-003 (Vault)** - đọc `postback_secret` để xác thực postback; secret không nằm trong DB.
- **FR-INFRA-001 (gateway)** - định tuyến webhook; route postback bỏ qua JWT user, dùng chữ ký HMAC.
- **FR-AFFIL-005 (downstream, P3)** - cashback đọc conversion `confirmed` (chỉ chi sau khi network approve).
- **FR-TRUST-005 (downstream)** - phát hiện gaming attribution dùng log postback + tỷ lệ confirm/reject.
- **NFR-AFFIL-001** - compliance: postback đối soát đúng last-click do click user-initiated sinh ra.
- Lib: `crypto/hmac`, `crypto/sha256`, `encoding/hex`, `net/http`, `encoding/json`.

---

## §8 - Payload ví dụ

### Webhook postback từ network (Involve Asia)

```http
POST /v1/affiliate/postback/involve_asia HTTP/1.1
Content-Type: application/json
X-Signature: 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08

{
  "sub_id": "sd_ab12cd34ef56",
  "order_value": 250000,
  "commission": 12000,
  "status": "approved"
}
```

### Cấu hình network (seed)

```sql
INSERT INTO affiliate_network (code, platform_id, base_url, target_param, sub_id_param, postback_secret_ref)
VALUES
  ('involve_asia', 1, 'https://go.involve.asia/aff', 'url', 'sub_id', 'affil/involve_asia/postback_secret'),
  ('accesstrade',  1, 'https://go.isclix.com/deep_link', 'url_enc', 'utm_content', 'affil/accesstrade/postback_secret');
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cookie window per-network (7 ngày vs 30/7 - cần xác minh §10 tài liệu nguồn với Involve Asia/Accesstrade) - thêm cột `cookie_window_days` vào `affiliate_network` khi xác minh xong.
- Chọn network trả hoa hồng cao nhất khi một sàn có nhiều network active - thêm logic chọn ở `TemplateFor`; hiện dùng network active đầu tiên.
- Đối soát định kỳ kéo report từ API network (ngoài postback) để bắt conversion bị bỏ lỡ - thêm job nền sau.
- Reject conversion khi đơn bị trả hàng sau khi đã confirm (chargeback) - cần postback `reversed`; map thêm khi network hỗ trợ.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Ghi conversion từ postback chưa xác thực | bad-signature test | giả mạo conversion, rút cashback | Verify HMAC trước khi ghi (§1 #5) |
| Secret cleartext trong DB | review (§4 #4) | lộ DB = giả mạo toàn bộ postback | Tham chiếu Vault (DEC-AFFIL-11) |
| Confirm sớm khi network pending | status-map test | trả cashback đơn có thể đảo | Chỉ approved -> confirm (§1 #8) |
| Postback lặp tạo conversion trùng | retry test | đếm hoa hồng trùng | Idempotent qua UNIQUE click_id (§1 #9) |
| sub_id không khớp click | 404 test | conversion mồ côi | Từ chối ghi (§1 #7) |
| Không ghi raw postback | log-count test | mất bằng chứng tranh chấp | Luôn LogPostback (§1 #6) |
| Body bị sửa giữa đường | tampered-body test | số tiền giả | HMAC trên body, hmac.Equal (§3) |
| Thiếu network template | TemplateFor ok=false | FR-AFFIL-002 trả 503 | Seed network; lộ lỗi cấu hình (§1 #3) |
| Secret cache cleartext lâu | review | bề mặt lộ thêm | Đọc Vault lúc verify (§1 #11) |

---

## §11 - Ghi chú

- FR-AFFIL-003 đóng vòng tiền: click user-initiated (FR-AFFIL-002) -> conversion (postback) -> confirm khi network approve -> cashback (FR-AFFIL-005).
- Xác thực chữ ký HMAC trước khi ghi là rào chống giả mạo conversion - webhook là endpoint công khai, không xác thực là mời gian lận tài chính.
- Secret postback tham chiếu Vault, không cleartext: thu hẹp bề mặt lộ và tuân no-cleartext PDPL (§5.5).
- Map trạng thái network sang vòng đời conversion giữ SănDeal chỉ trả cashback cho đơn đã approve, không đơn còn pending (có thể đảo).
- Idempotent postback (UNIQUE click_id) chịu được network retry mà không đếm trùng hoa hồng.
- Ghi raw mọi postback (cả bị từ chối) là bằng chứng đối soát và tín hiệu phát hiện tấn công giả mạo.
- Last-click thực thi qua sub_id: conversion quy về đúng click do user bấm sinh ra, không đoán.

---

*Hết FR-AFFIL-003. Status: ready_to_implement (mục tiêu audit 10/10).*
