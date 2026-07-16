---
id: TASK-INFRA-003
title: "Secrets management - Vault / AWS Secrets Manager, no-secret-in-env/code, rotation, lưu proxy creds + gateway tokens"
module: INFRA
priority: MUST
status: done
verify: T
phase: P0
milestone: P0 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-INFRA-001, TASK-INFRA-004, TASK-AUTH-002, TASK-SCRAPE-001, TASK-SCRAPE-004, TASK-COMPLY-005, TASK-TRUST-001]
depends_on: []
blocks: [TASK-COMPLY-005, TASK-SCRAPE-001]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.8 (bảo mật: secrets trong vault, no-cleartext)"
  - "docs/... §5.4 (trust/security), §5.5 (PDPL no-cleartext)"
source_decisions:
  - "DEC-INFRA-11: mọi secret (DB creds, JWT signing key, proxy creds, gateway tokens, payment keys) sống trong Vault/AWS Secrets Manager, KHÔNG trong env file hay code"
  - "DEC-INFRA-12: service lấy secret runtime qua một SecretProvider trừu tượng; backend Vault hoặc AWS SM hoán đổi được sau interface"
  - "DEC-INFRA-13: secret có version + hỗ trợ rotation; service cache có TTL ngắn và refresh, không pin cứng một version"
  - "DEC-INFRA-14: cấm log giá trị secret; mọi nơi in/serialize phải mask"
  - "DEC-INFRA-15: bí mật scraping (proxy creds theo nhà cung cấp) và token gateway lưu theo path có scope, least-privilege đọc"

language: "Go 1.22 (shared secrets package); HashiCorp Vault hoặc AWS Secrets Manager"
service: shopass/secrets/
new_files:
  - secrets/provider.go
  - secrets/vault.go
  - secrets/awssm.go
  - secrets/cache.go
  - secrets/mask.go
  - secrets/provider_test.go
  - secrets/cache_test.go
  - secrets/mask_test.go
modified_files: []
allowed_tools:
  - file_read: secrets/**
  - file_write: secrets/**
  - bash: cd secrets && go test ./...
disallowed_tools:
  - đặt secret trong file .env commit vào repo hay hằng số trong code (vi phạm DEC-INFRA-11)
  - log/print giá trị secret không mask (vi phạm DEC-INFRA-14)
  - pin cứng một version secret, bỏ qua rotation (vi phạm DEC-INFRA-13)

effort_hours: 5
sub_tasks:
  - "0.5h: provider.go - interface SecretProvider{Get(ctx, path) (Secret, error)} + kiểu Secret có Value/Version"
  - "1.0h: vault.go - đọc KV v2 path, parse version; awssm.go - GetSecretValue + VersionStage"
  - "1.0h: cache.go - cache TTL ngắn per-path + refresh; tránh đập backend mỗi lần đọc"
  - "0.5h: mask.go - String()/MarshalJSON mask giá trị (chỉ lộ 4 ký tự cuối / fingerprint)"
  - "0.5h: provider_test.go - Get trả version đúng; path scope đọc least-privilege"
  - "0.5h: cache_test.go - hit trong TTL không gọi backend; sau TTL refresh; rotation đổi version"
  - "0.5h: mask_test.go - secret không bao giờ xuất hiện thô qua String/JSON/log"
  - "1.0h: tài liệu path layout (db/, auth/jwt, scrape/proxy/<vendor>, gateway/tokens) + least-privilege policy"

risk_if_skipped: "Secret trong env/code là nguyên nhân rò rỉ phổ biến nhất: lộ qua repo, log, ảnh chụp màn hình, hay container layer. Với SănDeal, lộ proxy creds làm đối thủ/kẻ xấu đốt hạn mức proxy đắt tiền; lộ JWT signing key cho phép giả mạo mọi phiên; lộ DB creds là vỡ toàn bộ dữ liệu cá nhân (vi phạm PDPL, chế tài tới 5% doanh thu). Không rotation thì một lần lộ là vĩnh viễn. Đây là tuyến phòng thủ bảo mật nền tảng mà gateway, AUTH, SCRAPE đều dựa vào."
---

## §1 - Mô tả (BCP-14 normative)

Service INFRA **MUST** cung cấp một lớp quản lý secret tập trung: mọi bí mật sống trong Vault/AWS Secrets Manager, service lấy runtime qua một provider trừu tượng, hỗ trợ version/rotation, và không bao giờ log thô. Hợp đồng:

1. Mọi secret **MUST** sống trong Vault hoặc AWS Secrets Manager (DEC-INFRA-11): DB credentials, JWT signing key (cho TASK-AUTH-002), proxy credentials theo nhà cung cấp (cho TASK-SCRAPE-004), gateway tokens, payment gateway keys (cho TASK-BILL-002). KHÔNG đặt trong file `.env` commit vào repo hay hằng số trong code.
2. Service **MUST** lấy secret qua interface `SecretProvider` (DEC-INFRA-12). Backend cụ thể (Vault KV v2 hoặc AWS SM) hoán đổi được sau interface mà không sửa code gọi.
3. Kiểu trả về `Secret` **MUST** mang `Value` và `Version`; provider **MUST** đọc đúng version hiện hành (Vault: metadata version; AWS SM: `AWSCURRENT` stage).
4. Provider **MUST** hỗ trợ rotation (DEC-INFRA-13): khi secret được xoay ở backend, service đọc lần kế tiếp (sau TTL cache) nhận version mới mà không cần restart và không pin cứng một version.
5. Lớp cache **MUST** giảm tải backend: cache giá trị per-path với TTL ngắn cấu hình được (mặc định ví dụ 60s); trong TTL không gọi lại backend; hết TTL refresh.
6. Kiểu `Secret` **MUST** mask khi in/serialize (DEC-INFRA-14): `String()` và `MarshalJSON()` KHÔNG lộ giá trị thô (chỉ hiện fingerprint hoặc 4 ký tự cuối). Cấm log giá trị secret ở mọi nơi.
7. Bố cục path **MUST** có scope rõ (DEC-INFRA-15): ví dụ `db/main`, `auth/jwt-signing`, `scrape/proxy/brightdata`, `scrape/proxy/oxylabs`, `gateway/tokens`, `bill/momo`. Policy đọc **MUST** least-privilege (mỗi service chỉ đọc path nó cần).
8. Khi backend không truy cập được, provider **MUST** trả lỗi rõ ràng (KHÔNG fallback sang giá trị mặc định/cleartext); service quyết định fail-closed cho path nhạy cảm.
9. Provider **MUST** an toàn dùng đồng thời (nhiều goroutine đọc cùng path) - cache có khóa, không race.
10. Lớp secrets **SHOULD** phát OTel metric: `secret_fetch_total{path, backend, result}`, `secret_cache_hit_total`, `secret_rotation_detected_total{path}` - không bao giờ kèm giá trị secret trong nhãn.
11. Tài liệu path layout + least-privilege policy **MUST** tồn tại như nguồn sự thật để TASK-AUTH-002, TASK-SCRAPE-004, TASK-BILL-002 biết đọc path nào.
12. Lớp secrets **MUST** là điểm tham chiếu duy nhất cho audit no-cleartext của TASK-COMPLY-005 (chứng minh không có credential cleartext rời lớp này).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao secret không nằm trong env/code (DEC-INFRA-11)? Đây là nguyên nhân rò rỉ phổ biến nhất: secret lọt vào git history, container image layer, log, hay ảnh chụp màn hình. Một khi lọt vào repo, gần như không xóa sạch được khỏi lịch sử. Tập trung vào Vault/SM cho một nơi kiểm soát truy cập, audit, và xoay.

Vì sao một interface `SecretProvider` (DEC-INFRA-12)? SănDeal có thể chạy trên hạ tầng tự quản (Vault) hoặc AWS (Secrets Manager) tùy giai đoạn. Trừu tượng hóa sau một interface cho phép đổi backend mà không sửa mọi service gọi - chỉ thay implementation và policy.

Vì sao version + rotation (DEC-INFRA-13)? Bảo mật tốt đòi xoay secret định kỳ và xoay khẩn khi nghi lộ. Nếu service pin cứng một version, xoay sẽ làm hỏng (đọc version cũ đã thu hồi). Cache TTL ngắn + đọc version hiện hành cho xoay mượt: sau TTL, lần đọc kế nhận giá trị mới.

Vì sao cache TTL ngắn (§1 #5)? Gọi Vault/SM mỗi lần cần secret là chậm và tốn (rate-limit của backend). Cache giảm tải; TTL ngắn (vài chục giây) cân bằng giữa hiệu năng và độ trễ nhận giá trị xoay. Đủ ngắn để rotation lan nhanh, đủ dài để không đập backend.

Vì sao mask khi serialize (DEC-INFRA-14)? Lỗi vô tình `log.Printf("%+v", cfg)` hay trả secret trong JSON response là cách rò rỉ âm thầm. Làm `Secret` tự mask ở `String()`/`MarshalJSON()` biến an toàn thành mặc định: dù lập trình viên lỡ in ra, giá trị thô vẫn không lộ.

Vì sao path có scope + least-privilege (DEC-INFRA-15)? Nếu mọi service đọc được mọi secret, một service bị xâm nhập là lộ tất cả. Phân path theo phạm vi (proxy creds tách theo vendor, JWT key riêng) + policy chỉ-đọc-path-cần giới hạn bán kính nổ khi một service bị chiếm.

Vì sao là điểm tham chiếu cho audit PDPL (§1 #12)? TASK-COMPLY-005 phải chứng minh no-cleartext credential. Nếu mọi secret đi qua lớp này, audit chỉ cần xác nhận không có credential nào rời lớp ra env/code/log - một điểm kiểm thay vì rà toàn codebase.

---

## §3 - Hợp đồng API / DDL

### Provider interface (Go)

```go
// secrets/provider.go
type Secret struct {
    value   string // không export; truy cập qua Reveal() có kiểm soát
    Version string
}

// Reveal trả giá trị thô — chỉ gọi ở điểm dùng (mở DB, ký JWT), không log.
func (s Secret) Reveal() string { return s.value }

// String/MarshalJSON luôn mask (DEC-INFRA-14).
func (s Secret) String() string          { return mask(s.value) }
func (s Secret) MarshalJSON() ([]byte, error) { return json.Marshal(mask(s.value)) }

type SecretProvider interface {
    // Get trả secret tại path (ví dụ "auth/jwt-signing"); đọc version hiện hành.
    Get(ctx context.Context, path string) (Secret, error)
}
```

### Cache có TTL + refresh (§1 #4, #5, #9)

```go
// secrets/cache.go
type cachedProvider struct {
    inner SecretProvider
    ttl   time.Duration
    mu    sync.Mutex
    items map[string]cacheItem // path -> {secret, fetchedAt}
}

func (c *cachedProvider) Get(ctx context.Context, path string) (Secret, error) {
    c.mu.Lock(); defer c.mu.Unlock()
    if it, ok := c.items[path]; ok && time.Since(it.at) < c.ttl {
        metrics.CacheHit(path)
        return it.sec, nil
    }
    sec, err := c.inner.Get(ctx, path) // refresh: nhận version mới nếu đã xoay
    if err != nil {
        return Secret{}, err // KHÔNG fallback cleartext (§1 #8)
    }
    if old, ok := c.items[path]; ok && old.sec.Version != sec.Version {
        metrics.RotationDetected(path)
    }
    c.items[path] = cacheItem{sec: sec, at: time.Now()}
    return sec, nil
}
```

### Path layout (tài liệu, least-privilege)

```text
db/main                  -> DB credentials (đọc bởi mọi service cần DB)
auth/jwt-signing         -> RS256 signing key (đọc bởi AUTH; JWKS public expose riêng)
scrape/proxy/brightdata  -> proxy creds Bright Data  (đọc bởi SCRAPE)
scrape/proxy/oxylabs     -> proxy creds Oxylabs      (đọc bởi SCRAPE)
gateway/tokens           -> token nội bộ gateway     (đọc bởi gateway)
bill/momo, bill/zalopay  -> payment keys             (đọc bởi BILL)
```

---

## §4 - Acceptance criteria

1. `Get("auth/jwt-signing")` từ backend giả -> trả `Secret` có `Value` và `Version` đúng.
2. In `Secret` qua `fmt.Sprintf("%v")` -> ra chuỗi mask, KHÔNG lộ giá trị thô.
3. `json.Marshal(secret)` -> JSON mask, KHÔNG chứa giá trị thô.
4. Đọc cùng path hai lần trong TTL -> backend chỉ bị gọi một lần (cache hit lần hai).
5. Đọc lại sau khi TTL hết -> gọi backend lần nữa (refresh).
6. Backend xoay secret (đổi version) -> sau TTL, `Get` trả version mới; metric `secret_rotation_detected_total` tăng.
7. Backend lỗi/không truy cập được -> `Get` trả lỗi, KHÔNG trả giá trị mặc định/cleartext.
8. Đọc đồng thời cùng path từ nhiều goroutine -> không race (chạy với `-race`).
9. `Reveal()` trả giá trị thô đúng tại điểm dùng (mở DB / ký JWT).
10. Không có file `.env` chứa secret trong repo (kiểm bằng grep CI); cấu hình chỉ chứa path, không chứa giá trị.
11. Tài liệu path layout liệt kê đủ db/auth/scrape-proxy/gateway/bill với chú thích service nào đọc.
12. Provider Vault và AWS SM cùng pass bộ test hợp đồng (hoán đổi backend không đổi hành vi).

---

## §5 - Kiểm thử (verification)

```go
// secrets/mask_test.go
func TestSecret_NeverLeaksRaw(t *testing.T) {
    s := Secret{value: "super-secret-token-123", Version: "v3"}
    require.NotContains(t, fmt.Sprintf("%v", s), "super-secret-token-123")
    require.NotContains(t, fmt.Sprintf("%+v", s), "super-secret-token-123")
    b, _ := json.Marshal(s)
    require.NotContains(t, string(b), "super-secret-token-123")
    require.Equal(t, "super-secret-token-123", s.Reveal()) // chỉ Reveal() lộ
}

// secrets/cache_test.go
func TestCache_HitWithinTTL(t *testing.T) {
    backend := &countingProvider{val: "x", ver: "v1"}
    p := newCache(backend, 60*time.Second)
    _, _ = p.Get(ctx, "db/main")
    _, _ = p.Get(ctx, "db/main")
    require.Equal(t, 1, backend.calls) // lần hai từ cache
}

func TestCache_RefreshAfterTTL(t *testing.T) {
    backend := &countingProvider{val: "x", ver: "v1"}
    p := newCache(backend, 10*time.Millisecond)
    _, _ = p.Get(ctx, "db/main")
    time.Sleep(20 * time.Millisecond)
    _, _ = p.Get(ctx, "db/main")
    require.Equal(t, 2, backend.calls) // refresh sau TTL
}

func TestCache_RotationDetected(t *testing.T) {
    backend := &countingProvider{val: "old", ver: "v1"}
    p := newCache(backend, 10*time.Millisecond)
    _, _ = p.Get(ctx, "auth/jwt-signing")
    backend.val, backend.ver = "new", "v2" // xoay
    time.Sleep(20 * time.Millisecond)
    s, _ := p.Get(ctx, "auth/jwt-signing")
    require.Equal(t, "v2", s.Version)
}

func TestProvider_BackendError_NoFallback(t *testing.T) {
    p := newCache(&errProvider{}, time.Minute)
    _, err := p.Get(ctx, "db/main")
    require.Error(t, err) // KHÔNG trả cleartext mặc định
}

func TestCache_ConcurrentNoRace(t *testing.T) {
    p := newCache(&countingProvider{val: "x", ver: "v1"}, time.Minute)
    var wg sync.WaitGroup
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func() { defer wg.Done(); _, _ = p.Get(ctx, "db/main") }()
    }
    wg.Wait() // chạy với -race
}
```

---

## §6 - Khung triển khai

Xem §3. Implementation Vault dùng KV v2 (đọc `data` + `metadata.version`); AWS SM dùng `GetSecretValue` với `VersionStage=AWSCURRENT`. Cả hai bọc trong `cachedProvider` để mọi service hưởng cache + phát hiện rotation đồng nhất. `mask()` lộ tối đa 4 ký tự cuối hoặc một fingerprint hash. CI có một bước grep chặn pattern secret trong repo (key=value dạng token dài, không phải path). Least-privilege policy khai báo ở phía Vault/SM, ngoài code, nhưng tài liệu path layout sống cùng package.

---

## §7 - Phụ thuộc

- Không có task phụ thuộc trước (nền móng P0).
- TASK-INFRA-001 (downstream) - gateway lấy JWKS/khoá qua provider, không nhúng env.
- TASK-AUTH-002 (downstream) - đọc `auth/jwt-signing` để ký JWT.
- TASK-SCRAPE-004 (downstream) - đọc `scrape/proxy/<vendor>` cho proxy rotation.
- TASK-BILL-002 (downstream) - đọc payment keys.
- TASK-COMPLY-005 (downstream) - dùng lớp này làm điểm chứng minh no-cleartext.
- Hạ tầng: HashiCorp Vault hoặc AWS Secrets Manager.

---

## §8 - Payload ví dụ

### Đọc secret tại điểm dùng (nội bộ)

```go
sec, err := provider.Get(ctx, "db/main")
if err != nil { return fmt.Errorf("đọc db creds: %w", err) }
pool, err := pgxpool.New(ctx, sec.Reveal()) // Reveal() chỉ ở đây, không log
```

### Bố cục path trong Vault (KV v2)

```text
secret/data/sandeal/db/main
secret/data/sandeal/auth/jwt-signing
secret/data/sandeal/scrape/proxy/brightdata
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Dynamic DB credentials (Vault tạo creds ngắn hạn per-service) - siết hơn static; thêm khi vận hành ổn định.
- Auto-rotation lịch trình cho JWT signing key (xoay định kỳ tự động) - phối với JWKS đa-`kid` ở TASK-AUTH-002.
- Envelope encryption / KMS cho dữ liệu tại tầng ứng dụng (ngoài secret hạ tầng) - phạm vi PDPL nâng cao.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Secret lọt vào repo/env | grep CI | rò rỉ vĩnh viễn | Cấm (DEC-INFRA-11); xoay ngay nếu lộ |
| Log in secret thô | mask_test + review | rò rỉ qua log | `Secret.String()` mask mặc định (§1 #6) |
| Backend Vault/SM down | lỗi Get + metric | service không khởi động được | Fail-closed path nhạy cảm; backend HA; cache còn hạn |
| Pin cứng version cũ | rotation test | đọc creds đã thu hồi | Cache TTL + đọc version hiện hành (§1 #4) |
| Một service đọc mọi secret | review policy | bán kính nổ lớn khi bị chiếm | Least-privilege path scope (§1 #7) |
| Race khi đọc đồng thời | go test -race | giá trị/cache hỏng | Cache có mutex (§1 #9) |
| Fallback cleartext khi lỗi | test NoFallback | che lỗi + rủi ro | Cấm fallback; trả lỗi rõ (§1 #8) |
| Proxy creds lộ | giám sát hạn mức | đốt chi phí proxy đắt | Path riêng per-vendor + xoay |
| Metric kèm giá trị secret | review nhãn | rò rỉ qua observability | Nhãn chỉ path/result, không value (§1 #10) |

---

## §11 - Ghi chú

- Secret rời env/code là một trong các nguồn rò rỉ lớn nhất; tập trung vào Vault/SM cho kiểm soát, audit và xoay tại một chỗ.
- Interface `SecretProvider` cho phép đổi Vault <-> AWS SM theo giai đoạn hạ tầng mà không sửa code gọi.
- Cache TTL ngắn + đọc version hiện hành là chìa khóa để rotation mượt: xoay ở backend lan tới service sau một TTL, không cần restart.
- Mask mặc định ở `String()`/`MarshalJSON()` biến "không lộ secret" thành hành vi mặc định, chịu được lỗi in ấn vô tình.
- Path có scope + least-privilege giới hạn bán kính nổ: một service bị chiếm không lộ toàn bộ kho secret.
- Đây là điểm chứng minh no-cleartext cho audit PDPL (TASK-COMPLY-005): một điểm kiểm thay vì rà toàn codebase.

---

*Hết TASK-INFRA-003. Status: ready_to_implement (mục tiêu audit 10/10).*
