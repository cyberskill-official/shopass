---
id: TASK-SCRAPE-004
title: "Residential proxy rotation + tiering + cost-guard - Bright Data/Oxylabs (enterprise) -> Decodo/SOAX/NetNut (mid) -> IPRoyal (budget); datacenter vô dụng với Cloudflare/Akamai"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-SCRAPE-001, TASK-SCRAPE-003, TASK-SCRAPE-005, TASK-SCRAPE-007, TASK-SCRAPE-008, TASK-INFRA-003]
depends_on: [TASK-SCRAPE-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.3 (proxy residential xoay vòng, phân tầng giá 2026, datacenter vô dụng với Cloudflare/Akamai)"
  - "docs/... §4.1 (unit economics, proxy/scraping ~0,1-0,2 USD/user/tháng)"
source_decisions:
  - "DEC-SCRAPE-15: bắt buộc residential proxy cho scraping nghiêm túc; datacenter chỉ dùng cho target không có Cloudflare/Akamai"
  - "DEC-SCRAPE-16: phân tầng nhà cung cấp - enterprise (Bright Data/Oxylabs $8,5-12/GB) -> mid (Decodo/SOAX/NetNut $3-6/GB) -> budget (IPRoyal $1,75/GB)"
  - "DEC-SCRAPE-17: chọn tier proxy theo độ khó target (Akamai/ByteDance -> enterprise; Shopee JSON dễ -> budget/mid)"
  - "DEC-SCRAPE-18: cost-guard theo dõi GB tiêu thụ + ngân sách/ngày; vượt ngưỡng -> hạ tier hoặc tạm dừng quét cold-tier"
  - "DEC-SCRAPE-19: proxy theo nước (geo-targeting) phải khớp profile fingerprint của TASK-SCRAPE-003 (cùng nước)"

language: "Go 1.22 (scrape-svc); proxy pool + cost accounting; creds từ Vault (TASK-INFRA-003)"
service: shopass/services/scrape/
new_files:
  - services/scrape/internal/proxy/pool.go
  - services/scrape/internal/proxy/tier.go
  - services/scrape/internal/proxy/costguard.go
  - services/scrape/internal/proxy/session.go
  - services/scrape/internal/proxy/tier_test.go
  - services/scrape/internal/proxy/costguard_test.go
  - services/scrape/internal/proxy/pool_test.go
  - services/scrape/migrations/0002_proxy_usage.sql
modified_files:
  - services/scrape/internal/config/config.go            # thêm provider tiers + ngân sách/ngày
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape && go test ./...
disallowed_tools:
  - dùng datacenter proxy cho target có Cloudflare/Akamai (vi phạm DEC-SCRAPE-15, ban chắc)
  - dùng tier enterprise cho mọi request bất kể độ khó (vi phạm DEC-SCRAPE-16/17, đốt tiền vô ích)
  - bỏ cost-guard, quét không trần ngân sách (vi phạm DEC-SCRAPE-18, vỡ unit economics §4.1)
  - cấp proxy nước khác với fingerprint profile (vi phạm DEC-SCRAPE-19, mâu thuẫn geo)

effort_hours: 8
sub_tasks:
  - "1.0h: 0002_proxy_usage.sql - bảng proxy_usage (provider, tier, gb_used, cost_usd, day) đo tiêu thụ"
  - "1.5h: tier.go - bảng tier nhà cung cấp + chọn tier theo độ khó target (TargetDifficulty -> Tier)"
  - "1.5h: pool.go - cấp session proxy theo (tier, country), xoay vòng IP, đánh dấu IP bị ban"
  - "1.5h: costguard.go - cộng dồn GB/ngày, so ngân sách, quyết định hạ tier / chặn cold-tier"
  - "1.0h: session.go - gắn proxy session với fingerprint profile (cùng nước, TASK-SCRAPE-003)"
  - "0.5h: tier_test.go - Akamai->enterprise, Shopee-JSON->budget/mid"
  - "1.0h: costguard_test.go - vượt ngân sách -> downgrade/stop; pool_test.go - geo khớp + xoay IP + creds từ Vault"

risk_if_skipped: "Datacenter proxy gần như vô dụng với Cloudflare/Akamai (§3.3) -> không có residential rotation thì Lazada (Akamai) và TikTok (ByteDance) bị chặn ngay, farm Playwright (TASK-SCRAPE-003) vô dụng vì IP lộ. Dùng tier enterprise cho mọi request đốt tiền gấp 5-7 lần (8,5-12 vs 1,75 USD/GB) phá unit economics (§4.1 ~0,1-0,2 USD/user). Không cost-guard thì một SKU bị loop retry có thể đốt cả ngân sách ngày. Cấp proxy nước khác với fingerprint tạo mâu thuẫn geo - cờ đỏ phát hiện bot. Đây là biến phí lớn nhất phía băng thông và phải được quản chặt."
---

## §1 - Mô tả (BCP-14 normative)

Module proxy **MUST** cấp session residential xoay vòng theo tier nhà cung cấp và nước, khớp fingerprint, dưới một cost-guard ngân sách. Hợp đồng:

1. **MUST** mặc định dùng residential proxy cho mọi target có WAF/anti-bot (DEC-SCRAPE-15). Datacenter chỉ được dùng cho target xác nhận KHÔNG có Cloudflare/Akamai; với Shopee/TikTok/Lazada luôn residential.
2. **MUST** phân tầng nhà cung cấp (DEC-SCRAPE-16) qua enum `tier - {enterprise, mid, budget}`:
    - `enterprise`: Bright Data, Oxylabs (~8,5-12 USD/GB) - độ tin cậy cao nhất, tỷ lệ success ~99,95%.
    - `mid`: Decodo, SOAX, NetNut (~3-6 USD/GB).
    - `budget`: IPRoyal (~1,75 USD/GB) - rẻ nhất.
3. **MUST** chọn tier theo độ khó target (DEC-SCRAPE-17) qua hàm thuần `SelectTier(d TargetDifficulty) Tier`:
    - Akamai (Lazada), ByteDance attestation (TikTok) -> `enterprise`.
    - Shopee qua internal JSON (dễ hơn) -> `budget` hoặc `mid`.
    - Mặc định khi không rõ -> `mid` (cân bằng).
4. **MUST** cấp session proxy theo cặp `(tier, country)` và xoay vòng IP: `Acquire(ctx, tier, country) (ProxySession, error)`; mỗi session có `url/user/pass`, và pool **MUST** tránh tái dùng cùng một IP liên tục cho cùng một sàn trong cửa sổ ngắn.
5. **MUST** đánh dấu IP bị ban/challenge: khi adapter/farm báo IP bị chặn, pool **MUST** ghi nhận và không cấp lại IP đó trong khoảng cooldown.
6. **MUST** gắn proxy session với fingerprint profile cùng nước (DEC-SCRAPE-19): proxy VN chỉ ghép profile VN của TASK-SCRAPE-003; ghép sai nước bị từ chối.
7. **MUST** cost-guard (DEC-SCRAPE-18): theo dõi GB tiêu thụ và chi phí ước tính per provider per ngày vào `proxy_usage`; khi chi phí ngày vượt ngân sách cấu hình, cost-guard **MUST**:
    - hạ tier (enterprise -> mid -> budget) cho request không tới hạn, và/hoặc
    - tạm dừng quét tier `cold` (TASK-SCRAPE-001), giữ tier `hot` (flash sale) chạy.
8. **MUST** expose `CanProceed(tier, country) (bool, reason)` để orchestrator hỏi trước khi cấp session khi gần trần ngân sách.
9. **MUST** quy đổi chi phí bằng số nguyên (USD micro hoặc cent), KHÔNG dùng float tích lũy cho tiền (đồng bộ nguyên tắc tiền tệ của hệ thống).
10. **SHOULD** phát OTel metric: `proxy_gb_used_total{provider, tier}` (counter), `proxy_cost_usd_total{provider}` (counter), `proxy_ip_banned_total{provider}` (counter), `proxy_acquire_duration_ms{tier}` (histogram).
11. **MUST** đọc credentials proxy từ Vault (TASK-INFRA-003), KHÔNG hardcode/cleartext.
12. **MUST** ghi `proxy_usage` per ngày để báo cáo unit economics đối chiếu ngưỡng §4.1 (~0,1-0,2 USD/user/tháng).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao bắt buộc residential (DEC-SCRAPE-15)?** Tài liệu nói thẳng: datacenter proxy gần như vô dụng với Cloudflare/Akamai (§3.3). IP datacenter nằm trong dải ASN dễ nhận diện, bị chặn theo lô. Lazada (Akamai) và TikTok (ByteDance) sẽ chặn ngay. Residential mang IP của hộ gia đình thật, khó phân biệt với người dùng thường - đó là lý do nó đắt và là lý do nó cần thiết.

**Vì sao phân tầng nhà cung cấp (DEC-SCRAPE-16/17)?** Chênh lệch giá 5-7 lần giữa IPRoyal (1,75) và Bright Data (12). Dùng enterprise cho mọi thứ là đốt tiền; dùng budget cho Akamai là bị chặn. Chọn tier theo độ khó target: trả tiền cao chỉ ở nơi cần độ tin cậy cao (Akamai/ByteDance), dùng budget/mid ở nơi dễ (Shopee JSON). Đây là cách giữ proxy trong ngân sách §4.1.

**Vì sao cost-guard chủ động hạ tier/dừng cold (DEC-SCRAPE-18)?** Proxy là biến phí lớn nhất phía băng thông. Một bug retry-loop hay một đợt sàn siết có thể đốt cả ngân sách ngày trong vài giờ. Cost-guard đặt trần cứng: khi gần trần, ưu tiên giữ tier `hot` (flash sale - giá trị cao nhất) và hy sinh tier `cold` (SKU ít quan tâm). Tiền được tiêu vào nơi tạo giá trị.

**Vì sao gắn proxy với fingerprint cùng nước (DEC-SCRAPE-19)?** Như đã nêu ở TASK-SCRAPE-003: IP mang geolocation. Profile VN (timezone, ngôn ngữ VN) qua IP nước khác là mâu thuẫn. Buộc cặp (proxy <-> profile cùng nước) ngay tại tầng cấp session giữ câu chuyện nhất quán, không để adapter lỡ tay ghép sai.

**Vì sao xoay IP + cooldown ban (§1 #4, #5)?** Tái dùng một IP dày đặc cho cùng một sàn làm IP đó "nóng" và bị ban, kéo theo các request sau cũng hỏng. Xoay vòng phân tán dấu vết; đánh dấu IP đã bị chặn và cho nghỉ tránh cấp lại IP đang bị sàn theo dõi.

**Vì sao chi phí số nguyên (§1 #9)?** Cộng dồn GB và tiền bằng float tích lũy sai số qua hàng triệu request. Dùng cent/micro-USD nguyên giữ sổ chi phí chính xác để báo cáo unit economics đáng tin.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/scrape/migrations/0002_proxy_usage.sql
CREATE TABLE proxy_usage (
  day          DATE        NOT NULL,
  provider     TEXT        NOT NULL,           -- 'brightdata','oxylabs','decodo','iproyal',...
  tier         TEXT        NOT NULL,           -- 'enterprise'|'mid'|'budget'
  country      TEXT        NOT NULL,
  bytes_used   BIGINT      NOT NULL DEFAULT 0,
  cost_micro_usd BIGINT    NOT NULL DEFAULT 0, -- chi phí ước tính, micro-USD (số nguyên)
  PRIMARY KEY (day, provider, tier, country)
);
```

### Tier selection (Go)

```go
// services/scrape/internal/proxy/tier.go
type Tier string
const (
    TierEnterprise Tier = "enterprise"  // Bright Data, Oxylabs (~8,5-12 USD/GB)
    TierMid        Tier = "mid"         // Decodo, SOAX, NetNut (~3-6 USD/GB)
    TierBudget     Tier = "budget"      // IPRoyal (~1,75 USD/GB)
)

type TargetDifficulty int
const (
    DiffShopeeJSON TargetDifficulty = iota // dễ: internal JSON
    DiffAkamai                              // Lazada
    DiffByteDance                           // TikTok attestation
    DiffUnknown
)

// SelectTier chọn tier proxy theo độ khó target (DEC-SCRAPE-17).
func SelectTier(d TargetDifficulty) Tier {
    switch d {
    case DiffAkamai, DiffByteDance:
        return TierEnterprise   // độ tin cậy cao nhất cho WAF mạnh
    case DiffShopeeJSON:
        return TierBudget       // JSON dễ -> tiết kiệm
    default:
        return TierMid
    }
}
```

### Cost-guard (Go)

```go
// services/scrape/internal/proxy/costguard.go
type Decision int
const (
    Proceed     Decision = iota // dưới ngân sách
    DowngradeTier                // hạ tier cho request không tới hạn
    BlockCold                    // dừng tier cold, giữ hot
)

// Evaluate quyết định dựa trên chi phí ngày so ngân sách (micro-USD số nguyên).
func (g *CostGuard) Evaluate(ctx context.Context, day time.Time) (Decision, error) {
    spent, err := g.repo.SpentMicroUSD(ctx, day)
    if err != nil {
        return Proceed, err
    }
    switch {
    case spent >= g.dailyBudgetMicro:
        return BlockCold, nil          // chạm trần -> chỉ giữ hot
    case spent >= g.dailyBudgetMicro*8/10:
        return DowngradeTier, nil      // 80% -> hạ tier
    default:
        return Proceed, nil
    }
}
```

### Pool (Go)

```go
// services/scrape/internal/proxy/pool.go
func (p *Pool) Acquire(ctx context.Context, tier Tier, country string) (ProxySession, error) {
    if ok, reason := p.guard.CanProceed(tier, country); !ok {
        return ProxySession{}, fmt.Errorf("cost-guard: %s", reason)
    }
    ip := p.rotate(tier, country)         // tránh IP nóng + IP đang cooldown ban
    creds := p.vault.ProxyCreds(tier)     // TASK-INFRA-003, không cleartext
    return ProxySession{URL: ip.URL, User: creds.User, Pass: creds.Pass, Country: country}, nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `proxy_usage` tồn tại với khóa `(day, provider, tier, country)` và `cost_micro_usd` BIGINT.
2. `SelectTier(DiffAkamai)` = `enterprise`; `SelectTier(DiffByteDance)` = `enterprise`.
3. `SelectTier(DiffShopeeJSON)` = `budget`; `SelectTier(DiffUnknown)` = `mid`.
4. `Acquire(tier, "VN")` trả session có `Country="VN"` và creds lấy từ Vault (không cleartext trong code).
5. Pool không cấp lại cùng một IP cho cùng sàn trong cửa sổ ngắn (xoay vòng).
6. IP bị báo ban -> không được cấp lại trong cooldown (kiểm qua `MarkBanned` + `Acquire`).
7. Proxy session VN chỉ ghép profile VN; yêu cầu ghép proxy VN với profile ID -> bị từ chối.
8. `CostGuard.Evaluate` dưới 80% ngân sách -> `Proceed`; >=80% -> `DowngradeTier`; chạm trần -> `BlockCold`.
9. Khi `BlockCold`, orchestrator vẫn cấp được proxy cho job tier `hot`, nhưng job `cold` bị chặn.
10. Chi phí cộng dồn dùng số nguyên (micro-USD); với 1,75 USD/GB x 2GB -> `cost_micro_usd` đúng, không sai số float.
11. `proxy_usage` ghi đúng `bytes_used` và `cost_micro_usd` per ngày/provider để báo cáo §4.1.
12. Metric `proxy_gb_used_total`, `proxy_cost_usd_total`, `proxy_ip_banned_total` thay đổi đúng.

---

## §5 - Kiểm thử (verification)

```go
// services/scrape/internal/proxy/tier_test.go
func TestSelectTier_ByDifficulty(t *testing.T) {
    require.Equal(t, TierEnterprise, SelectTier(DiffAkamai))    // Lazada
    require.Equal(t, TierEnterprise, SelectTier(DiffByteDance)) // TikTok
    require.Equal(t, TierBudget, SelectTier(DiffShopeeJSON))    // Shopee JSON
    require.Equal(t, TierMid, SelectTier(DiffUnknown))
}
```

```go
// services/scrape/internal/proxy/costguard_test.go
func TestCostGuard_DowngradeThenBlock(t *testing.T) {
    g := newGuard(t, /*dailyBudgetMicro=*/10_000_000) // 10 USD/ngày
    setSpent(t, g, 5_000_000)
    d, _ := g.Evaluate(ctx, today)
    require.Equal(t, Proceed, d)

    setSpent(t, g, 8_500_000) // 85%
    d, _ = g.Evaluate(ctx, today)
    require.Equal(t, DowngradeTier, d)

    setSpent(t, g, 10_000_000) // chạm trần
    d, _ = g.Evaluate(ctx, today)
    require.Equal(t, BlockCold, d)
}

func TestCost_IntegerMicroUSD(t *testing.T) {
    // 1,75 USD/GB x 2GB = 3,50 USD = 3_500_000 micro-USD, chính xác
    got := costMicroUSD(/*usdPerGBMicro=*/1_750_000, /*bytes=*/2<<30)
    require.Equal(t, int64(3_500_000), got)
}
```

```go
// services/scrape/internal/proxy/pool_test.go
func TestAcquire_GeoMatchesProfile(t *testing.T) {
    p := newPool(t)
    _, err := p.BindProfile(proxyVN(), profileID("VN"))
    require.NoError(t, err)
    _, err = p.BindProfile(proxyVN(), profileID("US"))
    require.Error(t, err) // proxy VN + profile US bị từ chối
}

func TestAcquire_BannedIPCooldown(t *testing.T) {
    p := newPool(t)
    s, _ := p.Acquire(ctx, TierBudget, "VN")
    p.MarkBanned(s.IP)
    for i := 0; i < 20; i++ {
        s2, _ := p.Acquire(ctx, TierBudget, "VN")
        require.NotEqual(t, s.IP, s2.IP) // IP bị ban không cấp lại
    }
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0002 (proxy_usage) -> tier.go (SelectTier thuần, test trước) -> costguard.go (Evaluate + repo SpentMicroUSD) -> pool.go (Acquire + rotate + MarkBanned) -> session.go (bind profile cùng nước) -> tests. Provider tiers và ngân sách/ngày ở config.go (scrape-svc, Go); creds đọc từ Vault (TASK-INFRA-003). Cost-guard chạy đồng bộ trong đường `Acquire` để chặn trước khi tiêu, không phải cảnh báo sau.

---

## §7 - Phụ thuộc

- **TASK-SCRAPE-003** - fingerprint profile ghép proxy cùng nước; farm dùng proxy khi render.
- **TASK-SCRAPE-001** - orchestrator hỏi `CanProceed` trước khi cấp session; cost-guard chặn tier cold.
- **TASK-SCRAPE-002 / 007 / 008** - adapter lấy proxy session để gọi endpoint / render.
- **TASK-INFRA-003** - credentials proxy từ Vault.
- §4.1 - `proxy_usage` cung cấp số liệu báo cáo unit economics.

---

## §8 - Payload ví dụ

### Cấp session cho job Lazada (Akamai -> enterprise)

```go
tier := proxy.SelectTier(proxy.DiffAkamai)        // enterprise
sess, err := proxyPool.Acquire(ctx, tier, "VN")
// sess.URL trỏ residential VN tier enterprise; ghép profile VN của farm
```

### Cấu hình provider tiers + ngân sách (config)

```yaml
proxy:
  daily_budget_usd: 40
  providers:
    enterprise: [brightdata, oxylabs]   # ~8,5-12 USD/GB
    mid:        [decodo, soax, netnut]  # ~3-6 USD/GB
    budget:     [iproyal]               # ~1,75 USD/GB
  ban_cooldown_min: 30
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Giá GB chính xác từng provider (thay đổi theo hợp đồng/khối lượng) - đọc từ config, cập nhật khi ký hợp đồng thật.
- Tự động chuyển provider khi một nhà cung cấp tụt success-rate - bắt đầu chọn tier tĩnh, thêm chọn động theo metric sau.
- Mobile proxy (4G) cho target khó nhất - cân nhắc nếu enterprise residential vẫn bị chặn.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Datacenter cho Akamai/Cloudflare | tỷ lệ block cao | ban chắc | Bắt buộc residential (§1 #1) |
| Enterprise cho mọi request | `proxy_cost_usd_total` | đốt tiền 5-7x | SelectTier theo độ khó (§1 #3) |
| Retry-loop đốt ngân sách | cost-guard trần | vỡ §4.1 | BlockCold + DowngradeTier (§1 #7) |
| Proxy nước khác + fingerprint | guard BindProfile | mâu thuẫn geo | Cặp (proxy <-> profile cùng nước) (§1 #6) |
| Tái dùng IP nóng | `proxy_ip_banned_total` | ban chùm | Xoay vòng + cooldown (§1 #4,#5) |
| Float tích lũy chi phí | cost test | sổ sai | micro-USD số nguyên (§1 #9) |
| Creds proxy cleartext | audit secrets | lộ bí mật | Đọc từ Vault (§1 #11) |
| Chạm trần làm tắt cả hot | test BlockCold | mất flash sale | BlockCold chỉ chặn cold, giữ hot (§1 #7) |
| proxy_usage không ghi | báo cáo trống | không đối chiếu §4.1 được | Ghi per ngày/provider (§1 #12) |
| Một provider sập | success-rate sụt | quét hụt | Nhiều provider/tier; chuyển tay (hoãn auto) |

---

## §11 - Ghi chú

- Proxy là biến phí lớn nhất phía băng thông; tiering nhà cung cấp + cost-guard là hai đòn bẩy giữ §4.1 (~0,1-0,2 USD/user).
- Residential bắt buộc vì datacenter vô dụng với Cloudflare/Akamai - không phải lựa chọn tiết kiệm mà là điều kiện qua được Lazada/TikTok.
- Cost-guard ưu tiên giá trị: chạm trần thì hy sinh `cold` (SKU ít quan tâm), giữ `hot` (flash sale) - tiền vào nơi tạo giá trị nhất.
- Cặp (proxy <-> fingerprint cùng nước) là một nửa của câu chuyện nhất quán geo; nửa kia là profile của TASK-SCRAPE-003.
- Sổ chi phí số nguyên (micro-USD) cho phép báo cáo unit economics đáng tin để đối chiếu ngưỡng §4.1 hằng tháng.

---

*Hết TASK-SCRAPE-004. Status: ready_to_implement (mục tiêu audit 10/10).*
