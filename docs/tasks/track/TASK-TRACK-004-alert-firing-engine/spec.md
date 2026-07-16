---
id: TASK-TRACK-004
title: "Engine kích hoạt alert - đánh giá alert_rule trên price_snapshot sau mỗi lần ghi giá mới, dedup rising-edge, tạo dòng alert, bàn giao notification (TASK-NOTIF-001) - không tự gửi"
module: TRACK
priority: MUST
status: done
verify: T
phase: P1
milestone: P1 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-TRACK-003, TASK-PRICE-002, TASK-NOTIF-001, TASK-DEAL-001, TASK-SCRAPE-005]
depends_on: [TASK-TRACK-003, TASK-PRICE-002, TASK-NOTIF-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.4 (alert_rule -> alert; notification handoff)"
  - "docs/... §3.5 (phát hiện sale ảo real_sale), §3.6 (notification fan-out là tầng gửi riêng)"
source_decisions:
  - "DEC-TRACK-30: engine kích hoạt theo sự kiện - sau mỗi InsertSnapshot có written=true (giá thật đổi, TASK-PRICE-002), đánh giá các alert_rule active của đúng product_id đó"
  - "DEC-TRACK-31: bốn nhánh đánh giá: price_below (price<=threshold), drop_pct (giảm >=threshold% so median7), real_sale (TASK-DEAL-001 trả SALE_XIN), bottom_predicted (đã do TASK-DEAL-006 lo, engine này bỏ qua)"
  - "DEC-TRACK-32: dedup rising-edge - chỉ bắn khi điều kiện chuyển từ false sang true; còn thỏa liên tục thì KHÔNG bắn lại (tránh spam mỗi snapshot)"
  - "DEC-TRACK-33: engine TẠO dòng alert (status pending) rồi bàn giao cho notification (TASK-NOTIF-001) - KHÔNG tự gọi FCM/email/sms; tách lo gửi sang NOTIF"
  - "DEC-TRACK-34: drop_pct mốc tham chiếu là median giá 7 ngày (đọc price_daily của TASK-PRICE-002) - ổn định hơn giá-hôm-qua, tránh nhiễu flash sale ngắn"

language: "Go 1.22 (track-svc engine); đọc price_snapshot + price_daily (TASK-PRICE-002), gọi TASK-DEAL-001, enqueue notification (TASK-NOTIF-001)"
service: shopass/services/track/
new_files:
  - services/track/migrations/0004_alert_fired_state.sql
  - services/track/internal/engine/evaluate.go
  - services/track/internal/engine/dedup.go
  - services/track/internal/engine/handoff.go
  - services/track/internal/engine/evaluate_test.go
  - services/track/internal/engine/dedup_test.go
modified_files:
  - services/track/cmd/tracksvc/main.go            # đăng ký consumer sự kiện price-written
allowed_tools:
  - file_read: services/track/**
  - file_read: services/price/**
  - file_write: services/track/**
  - bash: cd services/track && go test ./...
disallowed_tools:
  - tự gọi FCM/email/sms trong engine (vi phạm DEC-TRACK-33, lấn việc của NOTIF, không có rate-limit/fan-out)
  - bắn lại cảnh báo mỗi snapshot khi điều kiện vẫn thỏa (vi phạm DEC-TRACK-32, spam người dùng)
  - quét toàn bộ alert_rule mỗi lần (phải lọc theo product_id vừa đổi + active, dùng idx_ar_eval)

effort_hours: 6
sub_tasks:
  - "0.5h: 0004_alert_fired_state.sql - bảng alert_fired_state (rule_id PK, last_fired_at, last_condition_met) cho dedup rising-edge"
  - "1.5h: evaluate.go - bốn nhánh rule_type: price_below, drop_pct (median7 từ price_daily), real_sale (gọi TASK-DEAL-001), bỏ bottom_predicted"
  - "1.0h: dedup.go - rising-edge: chỉ bắn khi met chuyển false->true; cập nhật alert_fired_state"
  - "1.0h: handoff.go - tạo dòng alert (status pending) + enqueue notification (TASK-NOTIF-001), KHÔNG tự gửi"
  - "0.5h: main.go - đăng ký consumer sự kiện price-written (gọi sau InsertSnapshot written=true)"
  - "1.5h: evaluate_test.go + dedup_test.go - 8 test (4 nhánh, rising-edge không spam, handoff tạo alert + enqueue, không tự gửi)"

risk_if_skipped: "Engine kích hoạt là mắt xích biến lịch sử giá (TASK-PRICE-002) + luật người dùng (TASK-TRACK-003) thành cảnh báo thật. Không có nó thì luật người dùng đặt nằm im, giá đổi mà không ai được báo - lời hứa cốt lõi 'theo dõi rồi nhắc đúng lúc' của SănDeal không thành. Nếu engine tự gọi FCM/email/sms thay vì bàn giao NOTIF thì nó vượt qua tầng fan-out + rate-limit + flatten-the-curve (§3.6), bắn thẳng vào FCM lúc đỉnh 00:00 và ăn 429, đồng thời nhân đôi logic gửi ở hai nơi. Nếu thiếu dedup rising-edge thì mỗi snapshot giá thấp lại bắn một cảnh báo - người dùng nhận hàng chục push cho cùng một deal trong vài phút flash sale, tắt thông báo, mất kênh. Quét toàn bộ alert_rule mỗi lần giá đổi thay vì lọc theo product_id active làm engine không co giãn được khi số luật tăng."
---

## §1 - Mô tả (BCP-14 normative)

Service TRACK **MUST** chạy engine đánh giá `alert_rule` mỗi khi `price_snapshot` có dòng mới thật (TASK-PRICE-002 `written=true`), áp dedup rising-edge, tạo dòng `alert` (status `pending`), rồi bàn giao cho notification (TASK-NOTIF-001). Engine KHÔNG tự gửi. Hợp đồng:

1. **MUST** kích hoạt theo sự kiện (DEC-TRACK-30): sau mỗi `InsertSnapshot` trả `written=true` cho một `product_id`, engine đánh giá các `alert_rule` có `product_id` đó và `active = true` (đọc qua `idx_ar_eval` của TASK-TRACK-003). KHÔNG đánh giá khi `written=false` (giá không đổi).
2. **MUST** lọc luật theo `(product_id, active=true)` (DEC-TRACK-30) - KHÔNG quét toàn bảng `alert_rule` mỗi lần. Một SKU không có luật active nào -> engine không làm gì cho SKU đó.
3. **MUST** đánh giá `rule_type = price_below`: điều kiện met khi `current_price <= threshold` (so sánh BIGINT VND, đồng nhất DEC-PRICE-05).
4. **MUST** đánh giá `rule_type = drop_pct`: tính `ref = median giá 7 ngày` đọc từ `price_daily` (TASK-PRICE-002, DEC-TRACK-34); điều kiện met khi `current_price <= ref * (100 - threshold) / 100` (giảm >= threshold% so mốc tham chiếu ổn định).
5. **MUST** đánh giá `rule_type = real_sale`: gọi `TASK-DEAL-001 DetectFakeSale(product_id, current_price, list_price)`; điều kiện met khi kết quả là `SALE_XIN` (sale thật). `SALE_AO`/`TAM_DUOC`/`UNKNOWN` không met.
6. **MUST** bỏ qua `rule_type = bottom_predicted` trong engine này (DEC-TRACK-31): loại đó do batch đêm TASK-DEAL-006 đánh giá và bắn, không gắn vào sự kiện giá tức thời. Engine này KHÔNG xử lý bottom_predicted để tránh bắn trùng.
7. **MUST** áp dedup rising-edge (DEC-TRACK-32): chỉ tạo `alert` khi điều kiện met chuyển từ `false` (lần đánh giá trước) sang `true`. Nếu điều kiện đã met ở lần trước và vẫn met, **MUST NOT** bắn lại. Trạng thái lưu ở bảng `alert_fired_state (alert_rule_id PK, last_condition_met BOOLEAN, last_fired_at TIMESTAMPTZ)`.
8. **MUST** reset cạnh khi điều kiện rời `true` về `false` (giá tăng lại trên ngưỡng): cập nhật `last_condition_met=false` để lần met tiếp theo lại là một cạnh lên mới (cho phép cảnh báo lần sau khi deal quay lại).
9. **MUST** với mỗi luật vừa met (qua dedup), tạo một dòng `alert (alert_rule_id, fired_at=now(), payload, status='pending')` (DEC-TRACK-33); `payload` JSONB gồm `{price, ref_price, rule_type, ...}` đủ để NOTIF dựng nội dung.
10. **MUST** bàn giao mỗi `alert` vừa tạo cho notification (TASK-NOTIF-001) bằng cách enqueue một `notification` request gồm `user_id` (từ luật), `channel[]` (từ luật), và `alert_id`. Engine **MUST NOT** tự gọi FCM/APNs/email/SMS (DEC-TRACK-33) - tầng NOTIF lo chọn kênh, rate-limit, fan-out.
11. **MUST** bền vững với lỗi cục bộ: một luật lỗi khi đánh giá (vd thiếu dữ liệu median7) **MUST NOT** làm hỏng việc đánh giá các luật khác của cùng SKU; ghi log và tiếp tục. Nếu enqueue notification lỗi, giữ `alert.status='pending'` để retry, không mất dòng alert.
12. **SHOULD** phát OTel: `alert_evaluated_total{rule_type}` (counter), `alert_fired_total{rule_type}` (counter), `alert_dedup_skipped_total{rule_type}` (counter), `alert_eval_duration_ms` (histogram).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao kích hoạt theo sự kiện written=true (DEC-TRACK-30)?** Giá chỉ đổi ở những thời điểm rời rạc (delta-only của TASK-PRICE-002 chỉ ghi khi giá thật đổi). Đánh giá luật đúng vào lúc có dòng `price_snapshot` mới là tự nhiên và tiết kiệm: không cần cron quét định kỳ toàn bộ luật. `written=false` nghĩa giá không đổi -> không có gì để đánh giá lại. Lọc theo `product_id` + `active` (partial index của TASK-TRACK-003) giữ mỗi sự kiện chỉ chạm đúng các luật liên quan.

**Vì sao drop_pct lấy median 7 ngày làm mốc (DEC-TRACK-34)?** "Giảm 20%" cần một mốc gốc. Lấy "giá hôm qua" thì nhiễu: một flash sale ngắn hôm qua làm mốc thấp giả, hôm nay giá thường lại trông như "tăng". Median 7 ngày (đọc từ `price_daily`) là mốc ổn định phản ánh mặt bằng giá gần đây, nên "giảm X%" có nghĩa thật. Đọc từ continuous aggregate cũng nhanh (bảng tổng hợp nhỏ).

**Vì sao real_sale gọi TASK-DEAL-001 thay vì tự tính (DEC-TRACK-31)?** Phân biệt sale thật/ảo là thuật toán thống kê riêng (median90, p10, trailing_min - §3.5) đã đặc tả ở TASK-DEAL-001. Engine alert không nên nhân bản logic đó; nó gọi `DetectFakeSale` và chỉ bắn khi `SALE_XIN`. Giữ một nguồn sự thật cho định nghĩa "sale thật".

**Vì sao dedup rising-edge (DEC-TRACK-32, §1 #7, #8)?** Trong một đợt flash sale, giá có thể nằm dưới ngưỡng suốt 30 phút với nhiều snapshot. Bắn một cảnh báo mỗi snapshot là tra tấn người dùng - họ nhận chục push cho cùng một deal rồi tắt thông báo. Rising-edge chỉ bắn ở khoảnh khắc điều kiện chuyển thành đúng (lần đầu giá chạm ngưỡng). Khi giá tăng lại rồi giảm lần nữa, đó là cạnh lên mới, đáng một cảnh báo mới - nên ta reset trạng thái khi điều kiện rời true.

**Vì sao tạo alert rồi bàn giao NOTIF, không tự gửi (DEC-TRACK-33)?** Gửi thông báo ở quy mô là bài toán riêng: fan-out qua Kafka/Redis Streams, rate-limit FCM 600k/phút, flatten-the-curve đỉnh 00:00, dead-letter, backoff (§3.6 - cả module NOTIF). Nếu engine tự gọi FCM, nó vượt qua mọi cơ chế đó, bắn thẳng lúc cao điểm và ăn 429. Engine chỉ làm phần nó biết (quyết định "có nên cảnh báo"), ghi `alert` làm bằng chứng, rồi đẩy cho NOTIF lo phần gửi. Hai trách nhiệm tách rời, mỗi tầng test được độc lập.

---

## §3 - Hợp đồng API / DDL

### Migration (trạng thái dedup)

```sql
-- services/track/migrations/0004_alert_fired_state.sql
CREATE TABLE alert_fired_state (
  alert_rule_id      BIGINT      PRIMARY KEY REFERENCES alert_rule(id) ON DELETE CASCADE,
  last_condition_met BOOLEAN     NOT NULL DEFAULT false,
  last_fired_at      TIMESTAMPTZ
);
```

### Đánh giá + dedup (Go)

```go
// services/track/internal/engine/evaluate.go

type Snapshot struct {
    ProductID int64
    Price     int64  // VND
    ListPrice *int64
}

// EvaluateForProduct đánh giá mọi luật active của một SKU sau khi giá đổi (DEC-TRACK-30).
func (e *Engine) EvaluateForProduct(ctx context.Context, snap Snapshot) error {
    rules, err := e.rules.ActiveByProduct(ctx, snap.ProductID) // dùng idx_ar_eval
    if err != nil {
        return err
    }
    for _, r := range rules {
        met, payload, err := e.conditionMet(ctx, r, snap)
        if err != nil {
            log.Warn("eval rule lỗi", "rule_id", r.ID, "err", err) // §1 #11: không làm hỏng luật khác
            continue
        }
        if err := e.fireIfRisingEdge(ctx, r, met, payload); err != nil {
            log.Warn("fire lỗi", "rule_id", r.ID, "err", err)
        }
    }
    return nil
}

func (e *Engine) conditionMet(ctx context.Context, r AlertRule, s Snapshot) (bool, map[string]any, error) {
    switch r.RuleType {
    case "price_below":
        met := s.Price <= *r.Threshold
        return met, map[string]any{"price": s.Price, "threshold": *r.Threshold}, nil
    case "drop_pct":
        ref, err := e.price.Median7d(ctx, s.ProductID) // price_daily (DEC-TRACK-34)
        if err != nil {
            return false, nil, err
        }
        limit := ref * (100 - *r.Threshold) / 100
        return s.Price <= limit, map[string]any{"price": s.Price, "ref_price": ref}, nil
    case "real_sale":
        verdict, err := e.deal.DetectFakeSale(ctx, s.ProductID, s.Price, s.ListPrice) // TASK-DEAL-001
        if err != nil {
            return false, nil, err
        }
        return verdict == deal.SaleXin, map[string]any{"price": s.Price, "verdict": verdict}, nil
    case "bottom_predicted":
        return false, nil, nil // DEC-TRACK-31: do TASK-DEAL-006 lo, engine này bỏ qua
    default:
        return false, nil, fmt.Errorf("rule_type lạ: %s", r.RuleType)
    }
}
```

```go
// services/track/internal/engine/dedup.go

// fireIfRisingEdge chỉ bắn khi điều kiện chuyển false->true (DEC-TRACK-32).
func (e *Engine) fireIfRisingEdge(ctx context.Context, r AlertRule, met bool, payload map[string]any) error {
    prev, err := e.state.LastConditionMet(ctx, r.ID) // mặc định false nếu chưa có
    if err != nil {
        return err
    }
    if met && !prev {
        alertID, err := e.handoff.CreateAndEnqueue(ctx, r, payload) // tạo alert pending + enqueue NOTIF
        if err != nil {
            return err
        }
        _ = alertID
        return e.state.Set(ctx, r.ID, true) // ghi cạnh đã bắn
    }
    if !met && prev {
        return e.state.Set(ctx, r.ID, false) // reset cạnh (§1 #8) - cho phép cảnh báo lần sau
    }
    return nil // met&&prev (vẫn thỏa) hoặc !met&&!prev: không bắn
}
```

---

## §4 - Acceptance criteria

1. `InsertSnapshot` trả `written=true` cho SKU -> engine đánh giá các luật active của SKU đó; `written=false` -> engine không chạy.
2. Engine chỉ lọc luật `(product_id, active=true)` (qua `idx_ar_eval`), không quét toàn bảng.
3. `price_below threshold=89000`, giá mới `79000` -> điều kiện met, tạo một `alert`.
4. `drop_pct threshold=20`, median7 = `100000`, giá mới `75000` (giảm 25% >= 20%) -> met; giá `85000` (giảm 15%) -> không met.
5. `real_sale`: khi `DetectFakeSale` trả `SALE_XIN` -> met; trả `SALE_AO`/`TAM_DUOC`/`UNKNOWN` -> không met.
6. `bottom_predicted` -> engine bỏ qua, không tạo alert (TASK-DEAL-006 lo).
7. Rising-edge: giá nằm dưới ngưỡng qua 3 snapshot liên tiếp -> chỉ một `alert` (cạnh lên đầu tiên), không ba.
8. Reset cạnh: giá lên trên ngưỡng (không met) rồi lại xuống dưới (met) -> cảnh báo thứ hai được bắn (cạnh lên mới).
9. Mỗi `alert` tạo có `status='pending'` và `payload` JSONB chứa `price` + mốc tham chiếu phù hợp loại.
10. Engine bàn giao mỗi alert cho NOTIF (enqueue `notification` với `user_id` + `channel[]` + `alert_id`); engine KHÔNG tự gọi FCM/email/SMS (kiểm qua mock dispatcher không bị chạm).
11. Một luật lỗi khi đánh giá (median7 thiếu) không làm hỏng đánh giá luật khác cùng SKU; log được ghi.
12. Metric `alert_fired_total` + `alert_dedup_skipped_total` tăng đúng theo cạnh lên vs lần bị dedup.

---

## §5 - Kiểm thử (verification)

```go
// services/track/internal/engine/evaluate_test.go
func TestPriceBelow_Met(t *testing.T) {
    e, h := setupEngine(t)
    rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
    require.NoError(t, e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}))
    require.Equal(t, 1, h.AlertCount(rid)) // tạo một alert
    require.Equal(t, 1, h.EnqueueCount())  // bàn giao NOTIF
    require.Equal(t, 0, h.DirectSendCount()) // KHÔNG tự gửi (DEC-TRACK-33)
}

func TestDropPct_Median7Reference(t *testing.T) {
    e, _ := setupEngineWithMedian(t, 100_000) // median7 = 100k
    seedRule(t, e, "drop_pct", ptr(int64(20)))
    metLow := e.evalOnce(t, Snapshot{ProductID: pid, Price: 75_000})  // -25%
    metHigh := e.evalOnce(t, Snapshot{ProductID: pid, Price: 85_000}) // -15%
    require.True(t, metLow)
    require.False(t, metHigh)
}

func TestRealSale_OnlyOnSaleXin(t *testing.T) {
    e, _ := setupEngine(t)
    seedRule(t, e, "real_sale", nil)
    e.deal.SetVerdict(deal.SaleXin)
    require.True(t, e.evalOnce(t, Snapshot{ProductID: pid, Price: 79_000}))
    e.deal.SetVerdict(deal.SaleAo)
    require.False(t, e.evalOnce(t, Snapshot{ProductID: pid, Price: 79_000}))
}

func TestBottomPredicted_Skipped(t *testing.T) {
    e, h := setupEngine(t)
    rid := seedRule(t, e, "bottom_predicted", nil)
    e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 1})
    require.Equal(t, 0, h.AlertCount(rid)) // engine này bỏ qua (TASK-DEAL-006 lo)
}

// services/track/internal/engine/dedup_test.go
func TestRisingEdge_NoSpam(t *testing.T) {
    e, h := setupEngine(t)
    rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
    for i := 0; i < 3; i++ { // ba snapshot đều dưới ngưỡng
        e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000})
    }
    require.Equal(t, 1, h.AlertCount(rid)) // chỉ một alert, không ba
}

func TestRisingEdge_ResetAllowsRefire(t *testing.T) {
    e, h := setupEngine(t)
    rid := seedRule(t, e, "price_below", ptr(int64(89_000)))
    e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}) // met -> bắn 1
    e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 99_000}) // lên trên -> reset
    e.EvaluateForProduct(ctx, Snapshot{ProductID: pid, Price: 79_000}) // met lại -> bắn 2
    require.Equal(t, 2, h.AlertCount(rid))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0004_alert_fired_state.sql` -> `evaluate.go` (bốn nhánh rule_type) -> `dedup.go` (rising-edge + reset) -> `handoff.go` (tạo alert pending + enqueue NOTIF) -> đăng ký consumer sự kiện price-written trong `main.go` -> tests. Engine chạy như consumer của sự kiện do TASK-PRICE-002 phát sau `InsertSnapshot` (in-process hook hoặc qua hàng đợi nội bộ). `Median7d` đọc `price_daily`. `DetectFakeSale` gọi package `deal` (TASK-DEAL-001). `CreateAndEnqueue` ghi `alert` rồi đẩy `notification` request - không có lời gọi FCM/SMS nào trong package `engine`.

---

## §7 - Phụ thuộc

- **TASK-TRACK-003** - cung cấp `alert_rule` (đọc qua `idx_ar_eval`) và bảng `alert` (engine ghi dòng `status=pending`).
- **TASK-PRICE-002** - cung cấp sự kiện `written=true`, `price_snapshot` (giá hiện tại), `price_daily` (median 7 ngày cho drop_pct).
- **TASK-NOTIF-001** - nhận bàn giao: engine enqueue `notification` request; NOTIF lo chọn kênh + fan-out + rate-limit.
- **TASK-DEAL-001** - `DetectFakeSale` cho nhánh `real_sale`.
- **TASK-DEAL-006 (ranh giới)** - lo `bottom_predicted` ở batch đêm; engine này cố ý bỏ qua loại đó để tránh bắn trùng.
- Lib: `pgx`, `encoding/json` (payload JSONB).

---

## §8 - Payload ví dụ

### notification request engine đẩy cho TASK-NOTIF-001 (nội bộ)

```json
{
  "user_id": 4021,
  "channel": ["push", "email"],
  "alert_id": 88123,
  "template": "price_below",
  "data": { "price": 79000, "threshold": 89000, "product_id": 90112 }
}
```

### Dòng alert engine tạo

```sql
SELECT id, alert_rule_id, fired_at, status, payload
FROM alert WHERE alert_rule_id = 5012 ORDER BY fired_at DESC LIMIT 1;
-- 88123 | 5012 | 2026-06-27 12:00:05+07 | pending | {"price":79000,"threshold":89000}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Cửa sổ cooldown thời gian (vd không bắn lại trong 6 giờ kể cả khi reset cạnh) - thêm nếu rising-edge vẫn quá ồn với SKU dao động mạnh; đo trước.
- Đánh giá batch khi nhiều SKU đổi giá cùng lúc (gom theo product_id) - tối ưu throughput giai đoạn sau.
- Mốc tham chiếu drop_pct cấu hình được (7d/14d/30d) - giữ 7d cho slice này.
- Gộp nhiều alert của một user trong khoảng ngắn thành một digest - thuộc NOTIF, không phải engine.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Đánh giá khi giá không đổi | chỉ chạy khi written=true | bỏ qua | Hook theo sự kiện delta-only (DEC-TRACK-30) |
| Quét toàn bảng alert_rule | EXPLAIN | chậm khi nhiều luật | Lọc product_id + active qua idx_ar_eval |
| Spam mỗi snapshot | rising-edge | một cảnh báo/cạnh lên | last_condition_met (DEC-TRACK-32) |
| Không cảnh báo lại sau khi deal quay lại | reset cạnh | cạnh lên mới bắn được | Set false khi điều kiện rời true (§1 #8) |
| Engine tự gọi FCM | code review + mock | vượt rate-limit, trùng logic | Chỉ enqueue NOTIF (DEC-TRACK-33) |
| median7 thiếu (SKU mới) | nhánh trả lỗi | luật drop_pct skip, log | Không làm hỏng luật khác (§1 #11) |
| bottom_predicted bắn trùng | engine bỏ qua | tránh trùng TASK-DEAL-006 | DEC-TRACK-31 |
| Enqueue NOTIF lỗi | alert.status=pending | giữ để retry | Không mất dòng alert (§1 #11) |
| Race hai snapshot cùng SKU | đánh giá tuần tự theo rule | trạng thái nhất quán | Khóa trên alert_fired_state nếu cần |

---

## §11 - Ghi chú

- Engine là mắt xích "quyết định có nên cảnh báo"; việc gửi tách hẳn sang NOTIF (TASK-NOTIF-001..007).
- Kích hoạt theo sự kiện `written=true` ăn khớp với delta-only của TASK-PRICE-002: chỉ đánh giá đúng lúc giá thật đổi, không cron quét.
- Dedup rising-edge là rào chống spam quan trọng nhất - bắn ở khoảnh khắc chạm ngưỡng, không bắn lại suốt thời gian còn thỏa; reset khi rời ngưỡng cho phép cảnh báo lần sau.
- drop_pct lấy median 7 ngày (price_daily) làm mốc để "giảm X%" có nghĩa ổn định, tránh nhiễu flash sale ngắn.
- real_sale tái dùng `DetectFakeSale` của TASK-DEAL-001 - một nguồn sự thật cho định nghĩa sale thật, engine không nhân bản.
- Tạo `alert` (bằng chứng, status pending) rồi bàn giao NOTIF giữ ranh giới rõ: engine test được độc lập với tầng gửi.

---

*Hết TASK-TRACK-004. Status: ready_to_implement (mục tiêu audit 10/10).*
