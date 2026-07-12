---
id: FR-SCRAPE-006
title: "Giám sát DOM/selector drift (Shopee A/B test) + adapter health + alert khi parse-failure tăng đột biến"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-001, FR-SCRAPE-002, FR-SCRAPE-003, FR-SCRAPE-005, FR-SCRAPE-007, FR-SCRAPE-008, FR-INFRA-004]
depends_on: [FR-SCRAPE-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.2 (DOM giỏ hàng Shopee thay đổi theo A/B test; cấu trúc per-sàn khác nhau)"
  - "docs/... §5.2 (rủi ro phụ thuộc nền tảng: sàn đổi DOM/API; kiến trúc đọc DOM resilient + giám sát thay đổi DOM)"
source_decisions:
  - "DEC-SCRAPE-24: theo dõi tỷ lệ parse-failure per (platform, adapter version) theo cửa sổ trượt; baseline động"
  - "DEC-SCRAPE-25: alert khi parse-failure vượt ngưỡng đột biến so baseline (selector drift / A/B test / sàn đổi schema)"
  - "DEC-SCRAPE-26: adapter health state machine (healthy -> degraded -> broken) điều khiển hành vi orchestrator"
  - "DEC-SCRAPE-27: khi adapter broken, hạ tải target đó (giảm tần suất quét) để không đốt proxy vào parse hỏng hàng loạt"

language: "Go 1.22 (scrape-svc); rolling-window stats + alert qua observability (FR-INFRA-004)"
service: shopass/services/scrape/
new_files:
  - services/scrape/internal/health/monitor.go
  - services/scrape/internal/health/window.go
  - services/scrape/internal/health/state.go
  - services/scrape/internal/health/alert.go
  - services/scrape/internal/health/window_test.go
  - services/scrape/internal/health/state_test.go
  - services/scrape/internal/health/monitor_test.go
modified_files:
  - services/scrape/internal/orchestrator/pool.go         # báo cáo outcome parse vào monitor + đọc health state
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape && go test ./...
disallowed_tools:
  - dùng ngưỡng parse-failure tuyệt đối cố định bất kể mùa vụ (vi phạm DEC-SCRAPE-24, baseline phải động)
  - giữ nguyên tần suất quét khi adapter broken (vi phạm DEC-SCRAPE-27, đốt proxy vào parse hỏng)
  - nuốt lỗi parse im lặng không alert (vi phạm DEC-SCRAPE-25, mù trước sàn đổi DOM)

effort_hours: 6
sub_tasks:
  - "1.0h: window.go - rolling window đếm (success, parse_fail) per (platform, adapter_version)"
  - "1.0h: state.go - state machine healthy->degraded->broken theo tỷ lệ fail + hysteresis"
  - "1.5h: monitor.go - nhận outcome từ pool, cập nhật window, suy ra health state, baseline động"
  - "1.0h: alert.go - phát alert khi vượt ngưỡng đột biến; dedup + cooldown để không spam"
  - "0.5h: tích hợp pool.go - báo outcome + đọc state để hạ tải khi broken"
  - "1.0h: window_test.go + state_test.go - đột biến fail -> degraded -> broken; hồi phục -> healthy"

risk_if_skipped: "Rủi ro phụ thuộc nền tảng là existential (§5.2): sàn đổi DOM/API bất cứ lúc nào, Shopee còn A/B test DOM liên tục (§3.2). Không giám sát thì adapter hỏng âm thầm - dữ liệu giá ngừng cập nhật mà không ai biết cho tới khi user phàn nàn, sale ảo (FR-DEAL-001) và biểu đồ (FR-DEAL-003) chạy trên dữ liệu chết. Tệ hơn: parse hỏng hàng loạt vẫn quét full tần suất sẽ đốt proxy vào các request vô dụng (vỡ §4.1). Đây là hệ thần kinh cảm giác của farm - phát hiện sớm khi sàn đổi để con người sửa adapter trước khi dữ liệu thối."
---

## §1 - Mô tả (BCP-14 normative)

Module health **MUST** theo dõi tỷ lệ parse-failure theo cửa sổ trượt, suy ra trạng thái adapter, alert khi đột biến, và điều khiển orchestrator hạ tải khi broken. Hợp đồng:

1. **MUST** đếm outcome mỗi lần quét theo cửa sổ trượt per `(platform_id, adapter_version)` (DEC-SCRAPE-24): phân loại `success | parse_fail | challenge | network_err`. `parse_fail` là trọng tâm (selector drift / schema đổi).
2. **MUST** tính baseline động: tỷ lệ `parse_fail` "bình thường" được ước lượng từ lịch sử gần (cửa sổ dài), KHÔNG dùng hằng số tuyệt đối cố định (mùa sale, traffic biến động làm nền thay đổi).
3. **MUST** suy ra health state qua state machine (DEC-SCRAPE-26): `healthy -> degraded -> broken`:
    - `healthy`: tỷ lệ parse_fail gần baseline.
    - `degraded`: parse_fail vượt baseline đáng kể nhưng chưa toàn diện.
    - `broken`: phần lớn request parse_fail (sàn đã đổi DOM/schema).
   Chuyển trạng thái **MUST** có hysteresis (ngưỡng lên khác ngưỡng xuống) để không nhấp nháy.
4. **MUST** phát alert khi vượt ngưỡng đột biến (DEC-SCRAPE-25): khi state chuyển sang `degraded` hoặc `broken`, gửi alert qua observability (FR-INFRA-004) kèm `(platform, adapter_version, fail_rate, sample_count)`.
5. **MUST** dedup + cooldown alert: một adapter chuyển `broken` chỉ alert một lần trong cửa sổ cooldown, KHÔNG spam mỗi request hỏng.
6. **MUST** điều khiển orchestrator hạ tải khi `broken` (DEC-SCRAPE-27): expose `ShouldThrottle(platformID, adapterVersion) (bool, factor)`; khi broken, orchestrator giảm tần suất quét target đó (tăng `next_run_at`) để không đốt proxy vào parse hỏng hàng loạt.
7. **MUST** yêu cầu số mẫu tối thiểu trước khi đổi state: không tuyên bố `broken` chỉ vì 2-3 request đầu fail (tránh báo động giả từ mẫu nhỏ).
8. **MUST** tự hồi phục: khi parse_fail trở lại gần baseline (sau khi adapter được sửa/sàn revert), state quay về `healthy` và bỏ throttle.
9. **MUST** phân biệt `parse_fail` (lỗi của ta - selector drift, đáng sửa adapter) với `challenge`/`network_err` (lỗi môi trường - CAPTCHA, proxy) để alert đúng nguyên nhân; chỉ `parse_fail` cao mới báo "sàn đổi DOM".
10. **SHOULD** phát OTel metric: `adapter_parse_fail_rate{platform, version}` (gauge), `adapter_health_state{platform, version}` (gauge enum), `adapter_alert_total{platform, transition}` (counter).
11. **MUST** an toàn đồng thời: nhiều worker báo outcome song song vào cùng window không gây race (đếm có khóa/atomic).
12. **SHOULD** giữ lịch sử chuyển trạng thái (state transition log) đủ để điều tra "adapter Shopee hỏng lúc nào, tự hồi phục hay người sửa".

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao baseline động (DEC-SCRAPE-24)?** Tỷ lệ fail "bình thường" không cố định. Mùa sale traffic tăng, một số SKU bị gỡ, sàn đôi khi chậm - nền nhiễu thay đổi. Ngưỡng tuyệt đối cố định sẽ hoặc quá nhạy (báo động giả mùa cao điểm) hoặc quá điếc (bỏ lỡ drift khi nền thấp). Baseline ước lượng từ lịch sử gần cho ngưỡng tự điều chỉnh.

**Vì sao state machine có hysteresis (§1 #3)?** Tỷ lệ fail dao động quanh ngưỡng sẽ làm state nhấp nháy healthy/degraded liên tục nếu dùng một ngưỡng. Hysteresis (lên ở ngưỡng cao, xuống ở ngưỡng thấp hơn) cho trạng thái ổn định - chỉ chuyển khi tín hiệu rõ ràng, không phản ứng với nhiễu.

**Vì sao hạ tải khi broken (DEC-SCRAPE-27)?** Khi Shopee đổi DOM, mọi request tới adapter cũ đều parse_fail. Quét full tần suất nghĩa là đốt proxy (§4.1) vào hàng nghìn request chắc chắn hỏng cho tới khi người sửa adapter. Hạ tải target broken giữ một nhịp tối thiểu (để biết khi nào sàn revert) nhưng không lãng phí băng thông vào thất bại đã biết.

**Vì sao số mẫu tối thiểu (§1 #7)?** 2-3 request đầu fail có thể là ngẫu nhiên (một SKU gỡ, một timeout). Tuyên bố `broken` từ mẫu nhỏ là báo động giả, kéo theo throttle nhầm. Đợi đủ mẫu cho tín hiệu đáng tin trước khi hành động.

**Vì sao phân biệt parse_fail với challenge/network (§1 #9)?** Ba loại lỗi cần ba phản ứng khác nhau. `challenge` cao -> việc của FR-SCRAPE-005 (CAPTCHA) và FR-SCRAPE-004 (proxy). `network_err` cao -> hạ tầng. Chỉ `parse_fail` cao mới nghĩa "sàn đổi DOM, cần người sửa adapter". Gộp chung làm alert sai địa chỉ, người trực sửa nhầm chỗ.

**Vì sao tự hồi phục (§1 #8)?** Sàn đôi khi revert A/B test, hoặc người vừa deploy adapter mới. State phải tự thấy parse_fail trở lại bình thường và bỏ throttle - không bắt người gỡ cờ thủ công, tránh quên để target chạy nửa công suất mãi.

---

## §3 - Hợp đồng API / DDL

### Rolling window (Go)

```go
// services/scrape/internal/health/window.go
type Outcome int
const (
    OutcomeSuccess Outcome = iota
    OutcomeParseFail
    OutcomeChallenge
    OutcomeNetworkErr
)

type Window struct {
    mu      sync.Mutex
    samples []Outcome   // ring buffer cửa sổ trượt
}

// Record thêm outcome; thread-safe (nhiều worker song song).
func (w *Window) Record(o Outcome) { /* khóa + ring buffer */ }

// ParseFailRate trả tỷ lệ parse_fail hiện tại và số mẫu.
func (w *Window) ParseFailRate() (rate float64, n int)
```

### Health state machine (Go)

```go
// services/scrape/internal/health/state.go
type Health int
const (
    Healthy Health = iota
    Degraded
    Broken
)

const minSamples = 30 // số mẫu tối thiểu trước khi đổi state (§1 #7)

// Next suy ra state mới từ tỷ lệ fail, baseline, và state hiện tại (có hysteresis).
func Next(cur Health, failRate, baseline float64, n int) Health {
    if n < minSamples {
        return cur // chưa đủ mẫu, giữ nguyên
    }
    up := baseline + 0.25   // ngưỡng lên
    upHard := 0.70          // phần lớn fail -> broken
    down := baseline + 0.10 // ngưỡng xuống (hysteresis)
    switch {
    case failRate >= upHard:
        return Broken
    case failRate >= up:
        return Degraded
    case failRate <= down:
        return Healthy
    default:
        return cur // vùng hysteresis -> giữ nguyên
    }
}
```

### Throttle control (Go)

```go
// services/scrape/internal/health/monitor.go
// ShouldThrottle: khi adapter broken, orchestrator giảm tần suất quét target.
func (m *Monitor) ShouldThrottle(platformID int16, version string) (bool, float64) {
    switch m.stateOf(platformID, version) {
    case Broken:
        return true, 0.1   // còn ~10% tần suất để dò sàn revert
    case Degraded:
        return true, 0.5
    default:
        return false, 1.0
    }
}
```

---

## §4 - Acceptance criteria

1. `Window.Record` đếm đúng outcome; `ParseFailRate` trả tỷ lệ và số mẫu chính xác.
2. Baseline được tính động từ cửa sổ dài, không phải hằng số (đổi khi nền fail đổi).
3. `Next` giữ state khi `n < minSamples` (không đổi vì mẫu nhỏ).
4. Tỷ lệ parse_fail vượt `baseline + 0.25` (đủ mẫu) -> `Degraded`; vượt 0.70 -> `Broken`.
5. Hysteresis: từ `Degraded`, tỷ lệ trong vùng giữa `down` và `up` -> giữ `Degraded` (không nhấp nháy).
6. Tỷ lệ fail trở lại <= `baseline + 0.10` -> `Healthy` (tự hồi phục).
7. Chuyển sang `Degraded`/`Broken` -> phát đúng một alert qua observability với `(platform, version, fail_rate, n)`.
8. Cùng adapter ở `Broken` không alert lại trong cooldown (dedup).
9. `ShouldThrottle` trả `(true, 0.1)` khi `Broken`; `(false, 1.0)` khi `Healthy`.
10. `parse_fail` cao -> alert "sàn đổi DOM"; `challenge` cao (không phải parse_fail) -> KHÔNG alert DOM (phân loại đúng).
11. Nhiều worker `Record` song song không gây race (chạy với `-race`).
12. Metric `adapter_parse_fail_rate`, `adapter_health_state`, `adapter_alert_total` thay đổi đúng theo chuyển trạng thái.

---

## §5 - Kiểm thử (verification)

```go
// services/scrape/internal/health/window_test.go
func TestWindow_RateAndCount(t *testing.T) {
    w := newWindow(100)
    for i := 0; i < 30; i++ { w.Record(OutcomeSuccess) }
    for i := 0; i < 10; i++ { w.Record(OutcomeParseFail) }
    rate, n := w.ParseFailRate()
    require.Equal(t, 40, n)
    require.InDelta(t, 0.25, rate, 0.01)
}

func TestWindow_RaceSafe(t *testing.T) {
    w := newWindow(1000)
    var wg sync.WaitGroup
    for g := 0; g < 8; g++ {
        wg.Add(1)
        go func() { defer wg.Done(); for i := 0; i < 200; i++ { w.Record(OutcomeSuccess) } }()
    }
    wg.Wait()
    _, n := w.ParseFailRate()
    require.Equal(t, 1000, n) // không mất bản ghi do race
}
```

```go
// services/scrape/internal/health/state_test.go
func TestNext_SpikeToBroken(t *testing.T) {
    require.Equal(t, Broken, Next(Healthy, 0.85, 0.05, 100))
}

func TestNext_MinSamplesGuard(t *testing.T) {
    require.Equal(t, Healthy, Next(Healthy, 0.90, 0.05, 5)) // mẫu nhỏ -> giữ nguyên
}

func TestNext_Hysteresis_NoFlicker(t *testing.T) {
    // vùng giữa down(0.15) và up(0.30) với baseline 0.05 -> giữ Degraded
    require.Equal(t, Degraded, Next(Degraded, 0.20, 0.05, 100))
}

func TestNext_Recovers(t *testing.T) {
    require.Equal(t, Healthy, Next(Broken, 0.08, 0.05, 100)) // trở lại <= baseline+0.10
}
```

```go
// services/scrape/internal/health/monitor_test.go
func TestMonitor_AlertDedup(t *testing.T) {
    m, sink := newMonitor(t)
    feedFails(t, m, 1, "v1", 100) // đẩy lên broken
    feedFails(t, m, 1, "v1", 100) // vẫn broken
    require.Equal(t, 1, sink.alertCount) // chỉ alert một lần trong cooldown
}

func TestMonitor_ParseFailNotChallenge(t *testing.T) {
    m, sink := newMonitor(t)
    feedOutcome(t, m, 1, "v1", OutcomeChallenge, 100) // challenge cao, parse_fail thấp
    require.Zero(t, sink.domAlerts) // không báo "sàn đổi DOM"
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: window.go (ring buffer thread-safe, test trước) -> state.go (Next thuần với hysteresis) -> monitor.go (ghép window + state + baseline động) -> alert.go (dedup + cooldown) -> tích hợp pool.go (báo outcome + đọc ShouldThrottle) -> tests. Alert đi qua observability của FR-INFRA-004. Orchestrator (FR-SCRAPE-001) đọc `ShouldThrottle` để nhân `next_run_at` khi target broken.

---

## §7 - Phụ thuộc

- **FR-SCRAPE-002 / 007 / 008** - adapter báo outcome (success/parse_fail) vào monitor.
- **FR-SCRAPE-001** - orchestrator đọc `ShouldThrottle` để hạ tải target broken.
- **FR-SCRAPE-005** - `challenge` outcome đến từ CAPTCHA detect; monitor phân biệt với parse_fail.
- **FR-INFRA-004** - alert đi qua observability spine (Prometheus/Grafana/log).
- §5.2 - giám sát DOM là mitigation cho rủi ro phụ thuộc nền tảng existential.

---

## §8 - Payload ví dụ

### Adapter báo outcome (nội bộ)

```go
// trong adapter sau khi parse
if err != nil {
    monitor.Report(job.PlatformID, adapterVersion, health.OutcomeParseFail)
} else {
    monitor.Report(job.PlatformID, adapterVersion, health.OutcomeSuccess)
}
```

### Alert khi Shopee đổi DOM (gửi qua observability)

```json
{
  "alert": "adapter_broken",
  "platform": "shopee",
  "adapter_version": "v1.4.2",
  "parse_fail_rate": 0.91,
  "sample_count": 240,
  "baseline": 0.04,
  "action": "throttled_to_10pct",
  "hint": "Shopee có thể đã đổi DOM/JSON schema (A/B test §3.2) - kiểm tra selectors/struct"
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Tự động fallback sang adapter version cũ/mới khi broken - hiện chỉ alert + throttle, người sửa; tự chuyển để sau.
- Học baseline theo mùa (double-date có nền fail cao hơn) - bắt đầu baseline trượt đơn giản, thêm seasonality sau.
- Phát hiện drift ở cấp trường (giá vẫn parse được nhưng stock thì không) - hiện coi parse_fail nhị phân, tinh chỉnh sau.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Sàn đổi DOM (Shopee A/B) | parse_fail spike | dữ liệu giá chết | Alert + throttle (§1 #4,#6) |
| Ngưỡng tuyệt đối cố định | báo động giả mùa sale | nhiễu/điếc | Baseline động (§1 #2) |
| State nhấp nháy | log transition | alert spam | Hysteresis (§1 #3) |
| Báo broken từ mẫu nhỏ | `MinSamplesGuard` test | throttle nhầm | minSamples=30 (§1 #7) |
| Quét full khi broken | `proxy_cost` không giảm | đốt proxy | ShouldThrottle 0.1 (§1 #6) |
| Alert spam mỗi request | dedup test | nhiễu cho người trực | Cooldown + dedup (§1 #5) |
| challenge gây alert DOM sai | `ParseFailNotChallenge` | sửa nhầm chỗ | Phân loại outcome (§1 #9) |
| Race khi nhiều worker Record | `-race` test | mất bản ghi | Khóa/atomic (§1 #11) |
| Không tự hồi phục | giữ throttle mãi | nửa công suất | Tự về Healthy (§1 #8) |
| Monitor không nhận outcome | metric phẳng | mù | Pool báo mọi outcome (§3 §8) |

---

## §11 - Ghi chú

- Đây là hệ thần kinh cảm giác của farm: phát hiện sớm khi sàn đổi DOM/schema để người sửa adapter trước khi dữ liệu thối.
- Rủi ro phụ thuộc nền tảng là existential (§5.2); giám sát DOM + đa sàn là hai mitigation chính, FR này lo phần giám sát.
- Phân biệt parse_fail (lỗi của ta, sửa adapter) với challenge/network (lỗi môi trường) để alert đúng địa chỉ - khác biệt quan trọng cho người trực.
- Hạ tải target broken bảo vệ unit economics (§4.1): không đốt proxy vào hàng nghìn request chắc chắn parse_fail.
- Baseline động + hysteresis + số mẫu tối thiểu là ba lớp chống báo động giả; tự hồi phục tránh để target chạy nửa công suất vì quên gỡ cờ.

---

*Hết FR-SCRAPE-006. Status: ready_to_implement (mục tiêu audit 10/10).*
