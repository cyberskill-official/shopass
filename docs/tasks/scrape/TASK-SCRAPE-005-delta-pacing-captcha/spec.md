---
id: TASK-SCRAPE-005
title: "Pacing ngẫu nhiên + jitter + CAPTCHA handling (slider/puzzle) + gọi delta-only InsertSnapshot của TASK-PRICE-002"
module: SCRAPE
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [TASK-SCRAPE-001, TASK-SCRAPE-002, TASK-SCRAPE-003, TASK-SCRAPE-004, TASK-SCRAPE-006, TASK-PRICE-002]
depends_on: [TASK-SCRAPE-002, TASK-PRICE-002]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.3 (CAPTCHA, request pacing, delta-only writes; puzzle slider khi nghi ngờ)"
  - "docs/... §3.4 (delta-only price_snapshot), §3.9 (CAPTCHA slider/puzzle per-sàn)"
source_decisions:
  - "DEC-SCRAPE-20: pacing ngẫu nhiên + jitter giữa các request để tránh rate-limit và dấu vết đều-đặn-bot"
  - "DEC-SCRAPE-21: phát hiện CAPTCHA (slider/puzzle/verify) -> giải qua dịch vụ hoặc farm hành vi người, có giới hạn"
  - "DEC-SCRAPE-22: ghi giá qua InsertSnapshot delta-only của TASK-PRICE-002 - không tự viết SQL; chỉ ghi khi giá đổi"
  - "DEC-SCRAPE-23: ngân sách CAPTCHA per ngày; vượt -> tạm lùi target thay vì giải vô hạn (chi phí + rủi ro)"

language: "Go 1.22 (scrape-svc); pacing limiter + CAPTCHA solver client; gọi price.Repo.InsertSnapshot"
service: shopass/services/scrape/
new_files:
  - services/scrape/internal/pacing/limiter.go
  - services/scrape/internal/captcha/detect.go
  - services/scrape/internal/captcha/solver.go
  - services/scrape/internal/sink/sink.go
  - services/scrape/internal/pacing/limiter_test.go
  - services/scrape/internal/captcha/detect_test.go
  - services/scrape/internal/sink/sink_test.go
modified_files:
  - services/scrape/internal/orchestrator/pool.go         # chèn pacing trước Fetch + sink sau Fetch
allowed_tools:
  - file_read: services/scrape/**
  - file_write: services/scrape/**
  - bash: cd services/scrape && go test ./...
disallowed_tools:
  - request đều đặn không jitter (vi phạm DEC-SCRAPE-20, dấu vết bot)
  - tự INSERT vào price_snapshot thay vì gọi InsertSnapshot (vi phạm DEC-SCRAPE-22, lặp logic delta-only)
  - giải CAPTCHA không giới hạn ngân sách (vi phạm DEC-SCRAPE-23, đốt tiền + leo thang rủi ro)

effort_hours: 6
sub_tasks:
  - "1.0h: limiter.go - pacing per-platform với delay ngẫu nhiên [min,max] + jitter, token bucket mềm"
  - "1.0h: detect.go - nhận diện CAPTCHA/slider/puzzle từ response (HTML/markers) -> CaptchaKind"
  - "1.0h: solver.go - gọi dịch vụ giải CAPTCHA hoặc farm-behavior, có ngân sách + giới hạn lần"
  - "1.0h: sink.go - nhận PriceSnapshot từ adapter/farm -> gọi price.Repo.InsertSnapshot (delta-only)"
  - "0.5h: limiter_test.go - khoảng cách request có jitter, không đều; tôn trọng min delay"
  - "1.0h: detect_test.go - fixture slider/puzzle -> đúng CaptchaKind; sink_test.go - giá đổi->written, không đổi->skip"
  - "0.5h: OTel metric scrape_pacing_wait_ms + captcha_seen_total{kind} + captcha_solved_total + snapshot_skipped_total"

risk_if_skipped: "Request đều đặn không jitter là dấu vết bot rõ nhất - bị rate-limit và ban (§3.3). Không phát hiện/giải CAPTCHA thì mỗi lần sàn nghi ngờ là một SKU mất dữ liệu cho tới khi người can thiệp. Tự INSERT vào price_snapshot thay vì gọi InsertSnapshot làm lặp logic delta-only và dễ lệch - hỏng tính toàn vẹn chuỗi giá mà sale ảo (TASK-DEAL-001) dựa vào. Giải CAPTCHA không trần ngân sách đốt tiền và leo thang đối đầu với sàn. Đây là lớp khép kín giữa scraping và ghi dữ liệu, và là điểm nối tới delta-only của PRICE."
---

## §1 - Mô tả (BCP-14 normative)

Module pacing/CAPTCHA/sink **MUST** điều tiết nhịp request, xử lý CAPTCHA có giới hạn, và ghi giá qua delta-only của TASK-PRICE-002. Hợp đồng:

1. **MUST** áp pacing per-platform trước mỗi request (DEC-SCRAPE-20): chèn delay ngẫu nhiên trong khoảng `[min, max]` cấu hình per-sàn, cộng jitter, để khoảng cách giữa các request KHÔNG đều đặn.
2. **MUST** tôn trọng `min_delay` cứng per-platform: không request nào được bắn nhanh hơn ngưỡng tối thiểu của sàn đó (bảo vệ trước rate-limit).
3. **MUST** phối hợp pacing với concurrency cap của TASK-SCRAPE-001 (#6) - hai lớp độc lập: cap giới hạn số request song song, pacing giới hạn nhịp request tuần tự.
4. **MUST** phát hiện CAPTCHA (DEC-SCRAPE-21) qua hàm `Detect(resp) CaptchaKind` nhận diện slider/puzzle/verify-page từ markers trong response; trả `CaptchaNone` khi không có.
5. **MUST** xử lý CAPTCHA có giới hạn: khi phát hiện, gọi solver (dịch vụ giải hoặc farm-behavior); solver **MUST** chịu ngân sách per ngày và giới hạn số lần per target (DEC-SCRAPE-23). Vượt ngưỡng -> trả `ErrCaptchaBudget`, orchestrator lùi target thay vì giải vô hạn.
6. **MUST** ghi giá qua `price.Repo.InsertSnapshot` delta-only của TASK-PRICE-002 (DEC-SCRAPE-22) - sink KHÔNG tự viết SQL `INSERT INTO price_snapshot`. Quy tắc "chỉ ghi khi giá đổi" thuộc về PRICE, không nhân bản ở đây.
7. **MUST** xử lý kết quả `written bool` từ `InsertSnapshot`: khi `written=false` (giá không đổi, delta-only bỏ qua) -> tăng metric `snapshot_skipped_total`, không coi là lỗi.
8. **MUST** truyền tín hiệu re-tier cho orchestrator: sink trả `(written, flashSale)` để TASK-SCRAPE-001 (#3) quyết định promote/demote tier - giá đổi hoặc flash -> hot.
9. **MUST** không chặn (block) toàn pool khi một target gặp CAPTCHA: chỉ target đó lùi; các sàn/SKU khác tiếp tục.
10. **SHOULD** phát OTel metric: `scrape_pacing_wait_ms{platform}` (histogram), `captcha_seen_total{platform, kind}` (counter), `captcha_solved_total{platform}` (counter), `captcha_budget_exhausted_total{platform}` (counter), `snapshot_written_total` vs `snapshot_skipped_total`.
11. **MUST** đảm bảo pacing không làm vỡ SLA tier `hot`: với flash sale cần quét theo phút, `min_delay` của tier hot phải đủ nhỏ để vẫn quét kịp trong khi vẫn có jitter.
12. **MUST** idempotent với retry: gọi `InsertSnapshot` cùng `(product_id, ts)` 2 lần (do retry) an toàn nhờ `ON CONFLICT DO NOTHING` của TASK-PRICE-002.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao pacing + jitter (DEC-SCRAPE-20)?** Bot lộ rõ nhất qua tính đều đặn: request cách nhau đúng 1000ms là chữ ký máy. Người dùng thật có nhịp ngẫu nhiên. Delay ngẫu nhiên `[min,max]` cộng jitter phá tính đều đặn, phối hợp với hành vi người-thật của farm (TASK-SCRAPE-003) để giảm điểm nghi ngờ.

**Vì sao pacing tách khỏi concurrency cap (§1 #3)?** Cap (TASK-SCRAPE-001) trả lời "bao nhiêu request song song"; pacing trả lời "cách nhau bao lâu". Một sàn có thể chịu 6 luồng song song nhưng không chịu 6 request/giây trên một luồng. Hai lớp giải hai bài toán khác nhau, cần cả hai.

**Vì sao CAPTCHA có ngân sách (DEC-SCRAPE-23)?** Giải CAPTCHA tốn tiền (dịch vụ giải) và là tín hiệu sàn đang nghi ngờ. Giải vô hạn vừa đốt ngân sách vừa leo thang đối đầu - sàn thấy "thực thể này cứ vượt CAPTCHA mãi" sẽ siết mạnh hơn. Đặt trần: giải tới một mức rồi lùi, để dấu vết tự nguội.

**Vì sao gọi InsertSnapshot thay vì tự ghi (DEC-SCRAPE-22)?** Delta-only là một quyết định thiết kế của PRICE (TASK-PRICE-002 #4): chỉ ghi khi `(price, list_price, stock, flash_sale)` đổi. Nếu sink tự `INSERT`, nó phải nhân bản hàm `changed()` và sẽ lệch khi PRICE đổi quy tắc. Một nguồn sự thật duy nhất cho quy tắc ghi - sink chỉ là người gọi.

**Vì sao xử lý written=false không phải lỗi (§1 #7)?** Phần lớn lần quét giá không đổi - đó là điều mong đợi, là cốt lõi của delta-only tiết kiệm storage. `written=false` nghĩa là "đã quét, giá y nguyên", một kết quả thành công. Đếm `snapshot_skipped_total` để đo hiệu quả delta-only, không báo động.

**Vì sao truyền tín hiệu re-tier (§1 #8)?** Sink là nơi biết "giá vừa đổi chưa" và "có đang flash sale không" - đúng thông tin orchestrator cần để promote SKU lên hot. Trả `(written, flashSale)` khép vòng phản hồi: phát hiện biến động -> tăng tần suất quét đúng SKU đó.

---

## §3 - Hợp đồng API / DDL

### Pacing limiter (Go)

```go
// services/scrape/internal/pacing/limiter.go
type Limiter struct {
    minDelay map[int16]time.Duration  // per platform_id
    maxDelay map[int16]time.Duration
    last     map[int16]time.Time
    mu       sync.Mutex
}

// Wait chặn tới khi đủ pacing cho platform: delay ngẫu nhiên [min,max] + jitter.
func (l *Limiter) Wait(ctx context.Context, platformID int16) error {
    l.mu.Lock()
    elapsed := time.Since(l.last[platformID])
    target := randDuration(l.minDelay[platformID], l.maxDelay[platformID]) // có jitter
    l.last[platformID] = time.Now().Add(maxDur(0, target-elapsed))
    l.mu.Unlock()
    if wait := target - elapsed; wait > 0 {
        return sleepCtx(ctx, wait)
    }
    return nil
}
```

### CAPTCHA detect (Go)

```go
// services/scrape/internal/captcha/detect.go
type CaptchaKind int
const (
    CaptchaNone CaptchaKind = iota
    CaptchaSlider                // Shopee slider/puzzle
    CaptchaPuzzle
    CaptchaVerifyPage
)

// Detect nhận diện CAPTCHA từ response (markers HTML / status / body).
func Detect(status int, contentType string, body []byte) CaptchaKind {
    switch {
    case bytes.Contains(body, []byte("slider")) && bytes.Contains(body, []byte("captcha")):
        return CaptchaSlider
    case bytes.Contains(body, []byte("puzzle-verify")):
        return CaptchaPuzzle
    case status == 403 && bytes.Contains(body, []byte("verify")):
        return CaptchaVerifyPage
    default:
        return CaptchaNone
    }
}
```

### Sink -> delta-only (Go)

```go
// services/scrape/internal/sink/sink.go
// Write đẩy snapshot vào PRICE qua delta-only; trả tín hiệu re-tier cho orchestrator.
func (s *Sink) Write(ctx context.Context, snap price.PriceSnapshot) (written, flashSale bool, err error) {
    written, err = s.price.InsertSnapshot(ctx, snap) // TASK-PRICE-002 delta-only, ON CONFLICT DO NOTHING
    if err != nil {
        return false, snap.FlashSale, err
    }
    if written {
        metrics.SnapshotWritten(snap.ProductID)
    } else {
        metrics.SnapshotSkipped(snap.ProductID) // giá không đổi -> bỏ qua, không phải lỗi
    }
    return written, snap.FlashSale, nil
}
```

---

## §4 - Acceptance criteria

1. `Limiter.Wait` áp delay trong `[min,max]` per-platform; hai lần Wait liên tiếp cách nhau >= `min_delay`.
2. Khoảng cách giữa các request có jitter (không hằng số) qua nhiều lần đo.
3. `min_delay` của tier hot đủ nhỏ để quét theo phút vẫn khả thi (không vỡ SLA hot, §1 #11).
4. `Detect` trả `CaptchaSlider` cho fixture slider; `CaptchaPuzzle` cho puzzle; `CaptchaVerifyPage` cho 403+verify.
5. `Detect` trả `CaptchaNone` cho response JSON giá hợp lệ.
6. Solver vượt ngân sách per ngày hoặc quá số lần per target -> trả `ErrCaptchaBudget` (không giải tiếp).
7. CAPTCHA tại một target -> chỉ target đó lùi; pool vẫn xử lý SKU/sàn khác (không block toàn cục).
8. `Sink.Write` gọi `InsertSnapshot` (không có câu `INSERT INTO price_snapshot` nào trong sink).
9. Giá đổi -> `Write` trả `written=true`; giá không đổi -> `written=false` và `snapshot_skipped_total` tăng (không lỗi).
10. `Write` trả `flashSale` đúng theo snapshot để orchestrator re-tier.
11. Gọi `Write` cùng `(product_id, ts)` 2 lần an toàn (ON CONFLICT DO NOTHING của TASK-PRICE-002), không nhân đôi dòng.
12. Metric `scrape_pacing_wait_ms`, `captcha_seen_total`, `captcha_budget_exhausted_total`, `snapshot_written_total`/`snapshot_skipped_total` thay đổi đúng.

---

## §5 - Kiểm thử (verification)

```go
// services/scrape/internal/pacing/limiter_test.go
func TestLimiter_RespectsMinDelay(t *testing.T) {
    l := newLimiter(map[int16]time.Duration{1: 50 * time.Millisecond}, map[int16]time.Duration{1: 120 * time.Millisecond})
    t0 := time.Now()
    l.Wait(ctx, 1)
    l.Wait(ctx, 1)
    require.GreaterOrEqual(t, time.Since(t0), 50*time.Millisecond) // không nhanh hơn min
}

func TestLimiter_HasJitter(t *testing.T) {
    l := newLimiter(map[int16]time.Duration{1: 20 * time.Millisecond}, map[int16]time.Duration{1: 80 * time.Millisecond})
    gaps := measureGaps(t, l, 1, 8)
    require.Greater(t, distinct(gaps), 1) // khoảng cách không hằng số
}
```

```go
// services/scrape/internal/captcha/detect_test.go
func TestDetect_Kinds(t *testing.T) {
    require.Equal(t, CaptchaSlider, Detect(200, "text/html", readFixture(t, "slider.html")))
    require.Equal(t, CaptchaPuzzle, Detect(200, "text/html", []byte(`<div id="puzzle-verify">`)))
    require.Equal(t, CaptchaVerifyPage, Detect(403, "text/html", []byte("please verify")))
    require.Equal(t, CaptchaNone, Detect(200, "application/json", []byte(`{"error":0}`)))
}
```

```go
// services/scrape/internal/sink/sink_test.go
func TestSink_DelegatesToInsertSnapshot(t *testing.T) {
    spy := &spyRepo{}
    s := &Sink{price: spy}
    s.Write(ctx, price.PriceSnapshot{ProductID: 1, TS: t0, Price: 89_000})
    require.Equal(t, 1, spy.insertCalls) // gọi InsertSnapshot, không tự ghi SQL
}

func TestSink_NoChangeSkips(t *testing.T) {
    s := &Sink{price: repoReturning(false, nil)} // delta-only báo skip
    written, _, err := s.Write(ctx, price.PriceSnapshot{ProductID: 1, TS: t0, Price: 89_000})
    require.NoError(t, err)
    require.False(t, written) // written=false không phải lỗi
}

func TestSink_ReturnsFlashForRetier(t *testing.T) {
    s := &Sink{price: repoReturning(true, nil)}
    _, flash, _ := s.Write(ctx, price.PriceSnapshot{ProductID: 1, TS: t0, Price: 89_000, FlashSale: true})
    require.True(t, flash) // tín hiệu re-tier
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: limiter.go (pacing thuần, test trước) -> detect.go (nhận diện CAPTCHA từ fixture) -> solver.go (giải có ngân sách) -> sink.go (gọi InsertSnapshot) -> chèn vào `orchestrator/pool.go` (Wait trước Fetch, Write sau Fetch) -> tests. Sink phụ thuộc `price.Repo` của TASK-PRICE-002; ngân sách CAPTCHA và pacing delays ở config per-platform. Tier hot có `min_delay` riêng nhỏ để không vỡ SLA quét-theo-phút.

---

## §7 - Phụ thuộc

- **TASK-SCRAPE-002** - adapter cung cấp snapshot và response để Detect; pacing chèn trước Fetch.
- **TASK-PRICE-002** - `Repo.InsertSnapshot` delta-only + `ON CONFLICT DO NOTHING`; sink chỉ gọi, không tự ghi.
- **TASK-SCRAPE-001** - orchestrator nhận `(written, flashSale)` để re-tier; pacing phối hợp concurrency cap.
- **TASK-SCRAPE-003** - farm-behavior là một đường giải CAPTCHA (slider) thay dịch vụ ngoài.
- **TASK-SCRAPE-006 (giám sát)** - tỷ lệ CAPTCHA tăng đột biến là tín hiệu adapter-health.

---

## §8 - Payload ví dụ

### Đường xử lý trong pool (nội bộ, rút gọn)

```go
// trong orchestrator/pool.go::runOne
if err := pacing.Wait(ctx, job.PlatformID); err != nil { return err } // jitter
snap, err := adapter.Fetch(ctx, job)                                   // TASK-SCRAPE-002
if kind := captcha.Detect(resp.Status, resp.CT, resp.Body); kind != captcha.CaptchaNone {
    if err := solver.Solve(ctx, job, kind); errors.Is(err, captcha.ErrCaptchaBudget) {
        return orchestrator.ErrBackoffTarget // lùi target, không block pool
    }
}
written, flash, err := sink.Write(ctx, snap)                          // TASK-PRICE-002 delta-only
next := orchestrator.ReTier(job.Tier, written, flash)                 // re-tier
```

### Cấu hình pacing + ngân sách CAPTCHA (config)

```yaml
pacing:
  shopee: { min_ms: 800,  max_ms: 2500 }
  tiktok: { min_ms: 1500, max_ms: 4000 }
  lazada: { min_ms: 1200, max_ms: 3500 }
  hot_min_ms: 400          # tier hot: nhỏ hơn để quét theo phút
captcha:
  daily_budget_solves: 2000
  max_solves_per_target: 3
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Dịch vụ giải CAPTCHA cụ thể vs farm-behavior tự giải slider - đánh giá theo chi phí và tỷ lệ thành công.
- Pacing thích nghi (tăng delay khi thấy CAPTCHA, giảm khi yên) - bắt đầu tĩnh, thêm vòng phản hồi sau.
- Phân biệt CAPTCHA "soft" (qua được bằng hành vi) vs "hard" (buộc dịch vụ) - tinh chỉnh khi có dữ liệu thực.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Request đều đặn không jitter | `HasJitter` test | dấu vết bot | Delay ngẫu nhiên + jitter (§1 #1) |
| Bắn nhanh hơn min_delay | `RespectsMinDelay` | rate-limit | min_delay cứng per-platform (§1 #2) |
| CAPTCHA không phát hiện | `captcha_seen_total`=0 bất thường | mất dữ liệu SKU | `Detect` markers (§1 #4) |
| Giải CAPTCHA vô hạn | `captcha_budget_exhausted` | đốt tiền + leo thang | Ngân sách + giới hạn lần (§1 #5) |
| Sink tự INSERT | grep code | lặp logic delta-only | Gọi InsertSnapshot (§1 #6) |
| written=false coi là lỗi | review | báo động giả | Đếm skip, không lỗi (§1 #7) |
| CAPTCHA block toàn pool | test #7 | dừng cả farm | Chỉ lùi target đó (§1 #9) |
| Pacing vỡ SLA hot | đo độ trễ flash | trễ flash sale | hot_min_ms nhỏ riêng (§1 #11) |
| Retry ghi trùng | property test | nhân đôi dòng | ON CONFLICT DO NOTHING (§1 #12) |
| Tín hiệu re-tier sai | sink_test | tần suất quét lệch | Trả (written, flashSale) đúng (§1 #8) |

---

## §11 - Ghi chú

- Đây là lớp khép kín giữa scraping (lấy dữ liệu) và PRICE (ghi dữ liệu); điểm nối cứng là `InsertSnapshot` delta-only của TASK-PRICE-002.
- Pacing/jitter và hành vi người-thật (TASK-SCRAPE-003) là hai lớp chống dấu vết bot bổ trợ: một về nhịp, một về tương tác.
- CAPTCHA có trần ngân sách là kỷ luật chi phí và là cách tránh leo thang đối đầu với sàn - giải tới mức rồi để dấu vết nguội.
- `written=false` là kết quả mong đợi (phần lớn lần quét giá không đổi) - cốt lõi của delta-only; đếm skip để đo hiệu quả, không báo động.
- Sink trả `(written, flashSale)` khép vòng re-tier: biến động giá tự kéo SKU lên tần suất quét cao hơn (TASK-SCRAPE-001).

---

*Hết TASK-SCRAPE-005. Status: ready_to_implement (mục tiêu audit 10/10).*
