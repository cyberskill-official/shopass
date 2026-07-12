---
id: FR-SCRAPE-001
title: "Scraping orchestrator lõi - job scheduler + scan-frequency tiering (flash sale: phút; SKU thường: giờ/ngày) + queue + worker pool"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-002, FR-SCRAPE-003, FR-SCRAPE-004, FR-SCRAPE-005, FR-SCRAPE-006, FR-PRICE-002, FR-TRACK-001]
depends_on: [FR-INFRA-003, FR-PRICE-001]
blocks: [FR-SCRAPE-002, FR-SCRAPE-003]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.3 (backend scraping quy mô 3 sàn, scan-frequency tiering, delta-only)"
  - "docs/... §3.1 (kiến trúc scraping farm), §5.1 (cold-start 90 ngày), §4.1 (unit economics)"
source_decisions:
  - "DEC-SCRAPE-01: scheduler phân tầng tần suất quét theo độ hot của SKU (hot=phút, warm=giờ, cold=ngày)"
  - "DEC-SCRAPE-02: hàng đợi job bền (persistent queue) chịu được crash worker, retry có giới hạn + backoff"
  - "DEC-SCRAPE-03: worker pool có giới hạn concurrency per-platform để không tự đốt proxy/ban"
  - "DEC-SCRAPE-04: adapter là interface chung; mỗi sàn (Shopee/TikTok/Lazada) là một implementation cắm vào orchestrator"
  - "DEC-SCRAPE-05: orchestrator chỉ điều phối; nó KHÔNG ghi DB trực tiếp mà gọi InsertSnapshot delta-only của FR-PRICE-002"

language: "Go 1.22 (scrape-svc); hàng đợi Redis Streams; lịch trình dựa trên DB tier"
service: shopass/services/scrape/
new_files:
  - services/scrape/internal/orchestrator/scheduler.go
  - services/scrape/internal/orchestrator/queue.go
  - services/scrape/internal/orchestrator/pool.go
  - services/scrape/internal/orchestrator/tier.go
  - services/scrape/internal/orchestrator/adapter.go
  - services/scrape/internal/orchestrator/scheduler_test.go
  - services/scrape/internal/orchestrator/tier_test.go
  - services/scrape/internal/orchestrator/pool_test.go
  - services/scrape/migrations/0001_scrape_job.sql
modified_files:
  - services/scrape/internal/config/config.go            # thêm cấu hình tier + concurrency per-platform
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape && go test ./...
disallowed_tools:
  - dùng setInterval/cron cố định một-tần-suất cho mọi SKU (vi phạm DEC-SCRAPE-01, đốt proxy vô ích)
  - cho worker pool concurrency vô hạn per-platform (vi phạm DEC-SCRAPE-03, tự kích hoạt ban)
  - ghi price_snapshot trực tiếp từ orchestrator (vi phạm DEC-SCRAPE-05, bỏ qua delta-only)

effort_hours: 10
sub_tasks:
  - "1.0h: 0001_scrape_job.sql - bảng scrape_job (product_id, platform_id, tier, next_run_at, attempts, status)"
  - "1.5h: tier.go - phân loại hot/warm/cold + tính next_run_at theo tần suất tier"
  - "2.0h: scheduler.go - quét job đến hạn, đẩy vào queue, cập nhật next_run_at sau khi xong"
  - "1.5h: queue.go - Redis Streams enqueue/claim/ack + retry có giới hạn + backoff"
  - "1.5h: pool.go - worker pool giới hạn concurrency per-platform, gọi adapter, gọi InsertSnapshot"
  - "0.5h: adapter.go - interface PlatformAdapter (Fetch(ctx, job) -> PriceSnapshot)"
  - "1.0h: tier_test.go - hot->phút, warm->giờ, cold->ngày; promote/demote theo biến động giá"
  - "1.0h: pool_test.go - concurrency cap được tôn trọng; retry + backoff; ack sau thành công"

risk_if_skipped: "Không có orchestrator thì không có cơ chế quét giá có kỷ luật. Quét một-tần-suất cho mọi SKU vừa lãng phí proxy (đốt unit economics §4.1) vừa làm trễ phát hiện flash sale (SKU hot cần quét theo phút). Không giới hạn concurrency per-platform -> tự kích hoạt anti-bot và bị ban (§3.9). Không có hàng đợi bền -> crash worker làm mất job, lủng dữ liệu lịch sử, hỏng cold-start 90 ngày (§5.1). Đây là bộ não điều phối của toàn bộ scraping farm - mọi adapter sàn cắm vào đây."
---

## §1 - Mô tả (BCP-14 normative)

Service SCRAPE **MUST** cung cấp một orchestrator lõi điều phối việc quét giá 3 sàn theo tần suất phân tầng, qua hàng đợi bền và worker pool giới hạn concurrency, gọi adapter per-sàn rồi ghi delta-only. Hợp đồng:

1. **MUST** định nghĩa bảng `scrape_job (product_id, platform_id, tier, next_run_at, attempts, last_status, locked_until)` với `PRIMARY KEY (product_id)` và index trên `(next_run_at)` để scheduler quét job đến hạn nhanh.
2. **MUST** phân tầng tần suất quét (DEC-SCRAPE-01) qua enum `tier - {hot, warm, cold}`:
    - `hot` (đang flash sale hoặc SKU biến động): chu kỳ <= 5 phút.
    - `warm` (SKU được theo dõi, ổn định): chu kỳ 1-6 giờ.
    - `cold` (SKU ít quan tâm): chu kỳ 24 giờ.
   Hàm `NextRunAt(tier, now) time.Time` **MUST** trả mốc kế tiếp đúng theo tier.
3. **MUST** cho phép promote/demote tier: khi giá vừa thay đổi hoặc `flash_sale=true` -> promote lên `hot`; sau N chu kỳ không đổi -> demote dần về `warm`/`cold`. Logic này **MUST** là hàm thuần kiểm thử được `ReTier(current, changed, flashSale) tier`.
4. **MUST** dùng hàng đợi bền (DEC-SCRAPE-02): scheduler đẩy job đến hạn vào Redis Streams; worker `claim` job, xử lý, rồi `ack`. Job chưa `ack` quá thời gian `locked_until` **MUST** được re-claim (chống mất job khi worker crash).
5. **MUST** retry có giới hạn + backoff: mỗi job thất bại tăng `attempts`; backoff hàm mũ có jitter; vượt `max_attempts` (mặc định 5) -> đánh `last_status='failed'` và đẩy sang dead-letter, KHÔNG retry vô hạn.
6. **MUST** giới hạn concurrency per-platform (DEC-SCRAPE-03): worker pool có cấu hình `max_concurrency[platform_id]`; số worker đang chạy cho một sàn **MUST** không vượt ngưỡng này (tránh tự kích hoạt anti-bot §3.9).
7. **MUST** định nghĩa interface adapter chung (DEC-SCRAPE-04): `PlatformAdapter.Fetch(ctx, job ScrapeJob) (price.PriceSnapshot, error)`. Mỗi sàn là một implementation (FR-SCRAPE-002 Shopee, FR-SCRAPE-007 TikTok, FR-SCRAPE-008 Lazada) đăng ký theo `platform_id`.
8. **MUST** sau khi adapter trả snapshot, orchestrator gọi `priceRepo.InsertSnapshot` delta-only của FR-PRICE-002 (DEC-SCRAPE-05) - orchestrator KHÔNG tự viết SQL vào `price_snapshot`.
9. **MUST** cập nhật `next_run_at` của job sau mỗi lần xử lý (thành công hay thất bại) theo tier hiện tại, để vòng lặp scheduler tiếp tục.
10. **SHOULD** phát OTel metric: `scrape_job_dispatched_total{platform_id, tier}` (counter), `scrape_job_duration_ms{platform_id}` (histogram), `scrape_job_failed_total{platform_id, reason}` (counter), `scrape_worker_inflight{platform_id}` (gauge).
11. **MUST** idempotent khi re-claim: xử lý cùng một job 2 lần (do timeout re-claim) KHÔNG được tạo dữ liệu sai - delta-only của FR-PRICE-002 (`ON CONFLICT DO NOTHING`) đảm bảo điều này.
12. **MUST** đọc cấu hình bí mật (proxy creds, API base) qua FR-INFRA-003 (Vault), KHÔNG hardcode trong mã hay env cleartext.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao phân tầng tần suất (DEC-SCRAPE-01)?** Quét mọi SKU mỗi phút là bất khả thi về chi phí proxy (§4.1) và tự gọi ban. Nhưng SKU đang flash sale đổi giá theo phút, quét mỗi ngày thì bỏ lỡ. Tiering tách "SKU nóng cần nhanh" khỏi "SKU nguội quét thưa" - tiêu proxy đúng chỗ. Đây là đòn bẩy unit economics phía băng thông, song song với delta-only phía storage.

**Vì sao hàng đợi bền + re-claim (DEC-SCRAPE-02)?** Scraping farm chạy hàng nghìn job; worker sẽ crash. Nếu job sống trong bộ nhớ worker, crash là mất job -> lủng lịch sử giá, hỏng baseline 90 ngày (§5.1) vốn là điều kiện ra mắt sale ảo. Redis Streams với `claim/ack` + `locked_until` cho phép một worker khác nhặt lại job mồ côi.

**Vì sao giới hạn concurrency per-platform (DEC-SCRAPE-03)?** Anti-bot của Shopee/TikTok/Lazada (§3.9) phản ứng với mật độ request từ một nguồn. Bắn song song không giới hạn là cách nhanh nhất để bị ban. Cap concurrency per-platform giữ áp lực dưới ngưỡng phát hiện, phối hợp với pacing/jitter của FR-SCRAPE-005.

**Vì sao adapter là interface (DEC-SCRAPE-04)?** 3 sàn có cơ chế truy cập khác hẳn nhau: Shopee có internal API đọc được khi `is_login:false`; TikTok ký request nên phải đọc DOM; Lazada sau Akamai. Tách orchestrator (điều phối, tiering, queue) khỏi adapter (cách lấy dữ liệu từng sàn) cho phép thêm sàn mà không sửa bộ não.

**Vì sao orchestrator không ghi DB (DEC-SCRAPE-05)?** Delta-only là logic của PRICE (FR-PRICE-002): chỉ ghi khi giá đổi. Nếu orchestrator tự `INSERT`, nó sẽ phải nhân bản logic so sánh và dễ lệch. Gọi `InsertSnapshot` giữ một nguồn sự thật duy nhất cho quy tắc ghi.

**Vì sao re-tier theo biến động (§1 #3)?** SKU không cố định "nóng" hay "nguội". Một SKU thường có thể vào flash sale bất ngờ. Promote khi vừa thấy thay đổi/flash, demote khi yên tĩnh, giúp farm tự thích nghi mà không cần người chỉnh tay từng SKU.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/scrape/migrations/0001_scrape_job.sql
CREATE TYPE scrape_tier AS ENUM ('hot', 'warm', 'cold');

CREATE TABLE scrape_job (
  product_id    BIGINT      PRIMARY KEY REFERENCES tracked_product(id),
  platform_id   SMALLINT    NOT NULL REFERENCES platform(id),
  tier          scrape_tier NOT NULL DEFAULT 'cold',
  next_run_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  attempts      INTEGER     NOT NULL DEFAULT 0,
  last_status   TEXT        NOT NULL DEFAULT 'pending',  -- pending|ok|retry|failed
  locked_until  TIMESTAMPTZ
);

CREATE INDEX idx_job_due ON scrape_job (next_run_at)
  WHERE last_status <> 'failed';
CREATE INDEX idx_job_platform ON scrape_job (platform_id, tier);
```

### Interface adapter (Go)

```go
// services/scrape/internal/orchestrator/adapter.go
type PlatformAdapter interface {
    // Fetch lấy snapshot giá cho một job. Lỗi mạng/parse trả error;
    // orchestrator quyết định retry/backoff dựa trên error.
    Fetch(ctx context.Context, job ScrapeJob) (price.PriceSnapshot, error)
    PlatformID() int16
}

type ScrapeJob struct {
    ProductID   int64
    PlatformID  int16
    Tier        Tier
    Attempts    int
}
```

### Tiering (Go)

```go
// services/scrape/internal/orchestrator/tier.go
type Tier string

const (
    TierHot  Tier = "hot"
    TierWarm Tier = "warm"
    TierCold Tier = "cold"
)

// NextRunAt trả mốc quét kế tiếp theo tier (có jitter nhẹ để rải tải).
func NextRunAt(t Tier, now time.Time) time.Time {
    switch t {
    case TierHot:
        return now.Add(jitter(3*time.Minute, 2*time.Minute))   // ~1-5 phút
    case TierWarm:
        return now.Add(jitter(3*time.Hour, 3*time.Hour))       // ~1-6 giờ
    default:
        return now.Add(jitter(24*time.Hour, 1*time.Hour))      // ~ngày
    }
}

// ReTier quyết định tier mới dựa trên kết quả lần quét gần nhất.
func ReTier(cur Tier, changed, flashSale bool) Tier {
    if flashSale || changed {
        return TierHot
    }
    switch cur {
    case TierHot:
        return TierWarm  // hết biến động -> hạ dần
    case TierWarm:
        return TierCold
    default:
        return TierCold
    }
}
```

### Worker pool (Go)

```go
// services/scrape/internal/orchestrator/pool.go
// runOne xử lý một job: gọi adapter, ghi delta-only, re-tier, đặt next_run_at.
func (p *Pool) runOne(ctx context.Context, job ScrapeJob) error {
    a := p.adapters[job.PlatformID]
    snap, err := a.Fetch(ctx, job)
    if err != nil {
        return p.scheduleRetry(ctx, job, err)   // tăng attempts + backoff
    }
    written, err := p.price.InsertSnapshot(ctx, snap)  // delta-only (FR-PRICE-002)
    if err != nil {
        return p.scheduleRetry(ctx, job, err)
    }
    next := ReTier(job.Tier, written, snap.FlashSale)
    return p.commit(ctx, job.ProductID, next, NextRunAt(next, time.Now()))
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `scrape_job` tồn tại với enum `scrape_tier` và index `idx_job_due`.
2. `NextRunAt(TierHot, now)` trả mốc trong vòng <= 5 phút; `TierWarm` trong 1-6 giờ; `TierCold` ~24 giờ.
3. `ReTier(_, changed=true, _)` -> `hot`; `ReTier(_, _, flashSale=true)` -> `hot`.
4. `ReTier(hot, false, false)` -> `warm`; `ReTier(warm, false, false)` -> `cold`; `ReTier(cold, false, false)` -> `cold`.
5. Scheduler đẩy đúng các job có `next_run_at <= now` vào queue; bỏ qua job tương lai và job `failed`.
6. Worker `claim` job, xử lý xong gọi `ack`; job được xử lý đúng 1 lần trong điều kiện thường.
7. Job bị bỏ dở (worker "crash" mô phỏng, không `ack`) -> sau `locked_until` được worker khác re-claim.
8. Concurrency per-platform được tôn trọng: với `max_concurrency[shopee]=N`, không bao giờ có >N worker chạy job Shopee đồng thời.
9. Job thất bại liên tục -> `attempts` tăng, backoff giãn dần; vượt `max_attempts` -> `last_status='failed'`, vào dead-letter, ngừng retry.
10. Sau xử lý thành công, orchestrator gọi `InsertSnapshot` (không tự viết SQL vào `price_snapshot`).
11. Re-claim cùng job 2 lần không tạo dữ liệu sai (nhờ `ON CONFLICT DO NOTHING` của FR-PRICE-002).
12. Metric `scrape_job_dispatched_total`, `scrape_job_failed_total`, `scrape_worker_inflight` thay đổi đúng theo vòng đời job.

---

## §5 - Kiểm thử (verification)

```go
// services/scrape/internal/orchestrator/tier_test.go
func TestNextRunAt_Tiers(t *testing.T) {
    now := time.Now()
    require.LessOrEqual(t, NextRunAt(TierHot, now).Sub(now), 5*time.Minute)
    warm := NextRunAt(TierWarm, now).Sub(now)
    require.True(t, warm >= time.Hour && warm <= 6*time.Hour)
    require.GreaterOrEqual(t, NextRunAt(TierCold, now).Sub(now), 23*time.Hour)
}

func TestReTier_PromoteOnChange(t *testing.T) {
    require.Equal(t, TierHot, ReTier(TierCold, true, false))   // giá đổi -> nóng
    require.Equal(t, TierHot, ReTier(TierWarm, false, true))   // flash -> nóng
}

func TestReTier_DemoteWhenQuiet(t *testing.T) {
    require.Equal(t, TierWarm, ReTier(TierHot, false, false))
    require.Equal(t, TierCold, ReTier(TierWarm, false, false))
    require.Equal(t, TierCold, ReTier(TierCold, false, false))
}
```

```go
// services/scrape/internal/orchestrator/pool_test.go
func TestPool_ConcurrencyCapPerPlatform(t *testing.T) {
    p := newTestPool(t, map[int16]int{1: 2}) // Shopee cap = 2
    var peak int32
    p.adapters[1] = blockingAdapter(func() {
        n := atomic.AddInt32(&p.inflight[1], 1)
        if n > atomic.LoadInt32(&peak) { atomic.StoreInt32(&peak, n) }
        time.Sleep(20 * time.Millisecond)
        atomic.AddInt32(&p.inflight[1], -1)
    })
    enqueueN(t, p, 1, 10) // 10 job Shopee
    p.drain(t)
    require.LessOrEqual(t, peak, int32(2)) // không vượt cap
}

func TestPool_RetryThenFail(t *testing.T) {
    p := newTestPool(t, map[int16]int{1: 1})
    p.adapters[1] = alwaysErr()
    job := seedJob(t, p, 1)
    for i := 0; i < 6; i++ { p.runOne(ctx, job) }
    require.Equal(t, "failed", statusOf(t, p, job.ProductID))
    require.GreaterOrEqual(t, attemptsOf(t, p, job.ProductID), 5)
}

func TestPool_ReclaimOrphanJob(t *testing.T) {
    p := newTestPool(t, map[int16]int{1: 1})
    job := seedJob(t, p, 1)
    p.claim(ctx, job)                 // worker A nhận
    advanceClock(p, 2*time.Minute)    // A "crash", quá locked_until
    got, ok := p.claim(ctx, job)      // worker B re-claim
    require.True(t, ok)
    require.Equal(t, job.ProductID, got.ProductID)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0001 (scrape_job) -> tier.go (thuần, test trước) -> queue.go (Redis Streams) -> pool.go (worker, gọi adapter + InsertSnapshot) -> scheduler.go (vòng lặp quét job đến hạn) -> tests. Adapter Shopee thật đến ở FR-SCRAPE-002; ở FR này dùng adapter giả (stub) để kiểm thử orchestrator độc lập. Concurrency cap và proxy đọc từ cấu hình Vault (FR-INFRA-003).

---

## §7 - Phụ thuộc

- **FR-INFRA-003** - secrets (proxy creds, API base) qua Vault; orchestrator không hardcode.
- **FR-PRICE-001** - `scrape_job.product_id` REFERENCES `tracked_product(id)` (FK cứng); bảng này phải tồn tại trước khi migration `0001_scrape_job.sql` chạy. `platform(id)` của FR-INFRA-002 cũng phải có (FK `platform_id`).
- **FR-PRICE-002 (downstream của ghi)** - orchestrator gọi `InsertSnapshot` delta-only.
- **FR-SCRAPE-002 / 007 / 008** - các adapter sàn cắm vào interface `PlatformAdapter`.
- **FR-SCRAPE-003 / 004 / 005** - farm Playwright, proxy rotation, pacing chạy bên trong adapter/pool.
- **FR-TRACK-001** - tạo `tracked_product` (và một `scrape_job` tương ứng) khi user theo dõi SKU.
- Hạ tầng: Redis Streams (queue), Postgres (scrape_job), OTel.

---

## §8 - Payload ví dụ

### Tạo job khi user theo dõi sản phẩm (nội bộ, gọi từ FR-TRACK-001)

```go
orch.Enqueue(ctx, orchestrator.ScrapeJob{
    ProductID:  90112,
    PlatformID: 1,            // shopee
    Tier:       orchestrator.TierWarm,
})
// scheduler sẽ đặt next_run_at theo tier và đẩy vào queue khi đến hạn
```

### Cấu hình concurrency per-platform (config)

```yaml
scrape:
  max_concurrency:
    shopee: 12
    tiktok: 6     # ký request khó, đọc DOM nặng hơn -> cap thấp hơn
    lazada: 8
  max_attempts: 5
  backoff_base_ms: 500
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Phân tải scheduler đa-node (leader election) khi farm vượt một node - slice sau.
- Tier "burst" riêng cho double-date (1.1, 2.2 ... 12.12) tạm gộp vào `hot` qua re-tier; tách thành lịch sự kiện ở giai đoạn ML (FR-DEAL-004).
- Ưu tiên job (priority queue) cho SKU Premium-user - gắn khi BILL (FR-BILL-001) lên.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Worker crash giữa job | `locked_until` quá hạn | job treo | Re-claim sau timeout (§1 #4) |
| Adapter lỗi mạng/parse | error từ Fetch | quét hụt | Retry + backoff; vượt max -> dead-letter |
| Concurrency vượt cap (bug) | gauge inflight | nguy cơ ban | Pool chặn cứng tại semaphore per-platform |
| Một-tần-suất cho mọi SKU | review cấu hình | đốt proxy / trễ flash | Tiering bắt buộc (§1 #2) |
| Scheduler đẩy job tương lai | test #5 | quét sớm vô ích | Lọc `next_run_at <= now` |
| Job mồ côi re-claim -> xử lý 2 lần | property test | ghi trùng | ON CONFLICT DO NOTHING (FR-PRICE-002) |
| Backoff không jitter -> đồng pha retry | metric spike | tự dồn tải / ban | jitter trong backoff (§1 #5) |
| Proxy creds rò qua env cleartext | audit secrets | lộ bí mật | Đọc từ Vault (FR-INFRA-003) |
| Dead-letter phình (sàn đổi DOM hàng loạt) | DLQ size | quét hỏng diện rộng | Alert qua FR-SCRAPE-006, sửa adapter |
| Redis Streams mất kết nối | health check | queue đứng | Reconnect + job vẫn bền trong stream |

---

## §11 - Ghi chú

- Orchestrator là bộ não điều phối; mọi adapter sàn cắm vào interface `PlatformAdapter`, nên thêm TikTok/Lazada không đụng vào tiering/queue.
- Tiering là đòn bẩy unit economics phía băng thông (§4.1), song song với delta-only phía storage (FR-PRICE-002).
- Concurrency cap per-platform là tuyến phòng thủ anti-ban đầu tiên (§3.9), trước pacing/jitter (FR-SCRAPE-005) và proxy rotation (FR-SCRAPE-004).
- Hàng đợi bền + re-claim bảo vệ tính liên tục của dữ liệu lịch sử, điều kiện sống còn của cold-start 90 ngày (§5.1).
- Orchestrator không ghi DB trực tiếp - một nguồn sự thật duy nhất cho quy tắc ghi nằm ở delta-only của PRICE.

---

*Hết FR-SCRAPE-001. Status: ready_to_implement (mục tiêu audit 10/10).*
