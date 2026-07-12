---
id: FR-DEAL-006
title: "Batch chấm điểm đáy giá hằng đêm + cảnh báo khi P(bottom within 14d) > 0.7 - model-agnostic đọc price_forecast, idempotent 1 alert/SKU/ngày + cooldown chống spam, khớp alert_rule 'bottom_predicted' rồi enqueue notification"
module: DEAL
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-DEAL-004, FR-DEAL-005, FR-TRACK-003, FR-NOTIF-003, FR-DEAL-002]
depends_on: [FR-DEAL-004, FR-TRACK-003]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5 (serving: batch nightly score, alert nếu P(bottom within 14d) > 0.7)"
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.6 (notification fan-out)"
source_decisions:
  - "DEC-DEAL-50: ngưỡng cảnh báo P(bottom within 14d) > 0.7 chốt từ §3.5 - hằng load-bearing, so sánh strict greater-than"
  - "DEC-DEAL-51: chạy batch chấm điểm hằng đêm off-peak (cron 02:00 Asia/Ho_Chi_Minh) - dự báo đổi chậm nên 1 lần/ngày là đủ và rẻ"
  - "DEC-DEAL-52: model-agnostic - đọc cột p_bottom_14d chung của bảng price_forecast do Prophet (FR-DEAL-004) hoặc LightGBM (FR-DEAL-005) sinh, batch không biết và không quan tâm model nào"
  - "DEC-DEAL-53: idempotent tối đa 1 alert/(user, product)/ngày + cooldown rising-edge - không bắn lại đêm này qua đêm khác khi P vẫn > 0.7, tránh huấn luyện người dùng phớt lờ cảnh báo"
  - "DEC-DEAL-54: chỉ chấm điểm SKU đã qua maturity gate MATURE (FR-DEAL-002 IsFeatureReady) và chỉ khớp alert_rule type 'bottom_predicted' đang bật (FR-TRACK-003) rồi enqueue vào notification fan-out (FR-NOTIF-003)"

language: "Go 1.22 (deal-svc batch orchestrator); reads price_forecast (Postgres, do Python sinh), enqueues to notification fan-out (FR-NOTIF-003)"
service: shopass/services/deal/
new_files:
  - services/deal/internal/batch/nightly_score.go
  - services/deal/internal/batch/alert_match.go
  - services/deal/internal/batch/dedupe.go
  - services/deal/internal/batch/nightly_score_test.go
  - services/deal/migrations/0002_bottom_alert_log.sql
modified_files:
  - services/deal/cmd/dealsvc/main.go            # đăng ký cron job nightly score off-peak
allowed_tools:
  - file_read: services/deal/**
  - file_write: services/deal/**
  - bash: cd services/deal && go test ./...
disallowed_tools:
  - hạ ngưỡng dưới 0.7 hoặc đổi sang so sánh >= (vi phạm DEC-DEAL-50, đổi định nghĩa cảnh báo, spam)
  - bắn lại cảnh báo đáy mỗi đêm khi P vẫn > 0.7 cho cùng (user, product) (vi phạm DEC-DEAL-53, huấn luyện người dùng phớt lờ)
  - chấm điểm và cảnh báo SKU chưa MATURE (vi phạm DEC-DEAL-54, dự báo trên dữ liệu mỏng)
  - gắn cứng batch vào một model (Prophet hoặc LightGBM) thay vì đọc p_bottom_14d chung (vi phạm DEC-DEAL-52)

effort_hours: 6
sub_tasks:
  - "0.5h: 0002_bottom_alert_log.sql - bảng log alert đã bắn + UNIQUE(user_id, product_id, fired_on)"
  - "1.0h: nightly_score.go - vòng batch: query price_forecast của SKU MATURE, lọc p_bottom_14d > 0.7"
  - "1.0h: alert_match.go - join sang alert_rule type 'bottom_predicted' đang bật, gom theo user"
  - "1.0h: dedupe.go - kiểm tra bottom_alert_log (idempotent 1/ngày) + cooldown rising-edge"
  - "0.5h: enqueue vào notification fan-out (FR-NOTIF-003) + ghi bottom_alert_log"
  - "0.5h: main.go - đăng ký cron 02:00 Asia/Ho_Chi_Minh (off-peak), khóa chống chạy chồng"
  - "1.0h: nightly_score_test.go - 6 test (biên 0.71/0.69, maturity gate, dedupe ngày, cooldown, đúng type, enqueue)"
  - "0.5h: OTel metric deal_nightly_scored_total + deal_bottom_alert_fired_total + deal_alert_dedupe_skipped_total"

risk_if_skipped: "Dự đoán đáy giá (FR-DEAL-004/005) chỉ có giá trị khi tới được người dùng đúng lúc - một xác suất chạm đáy nằm trong bảng mà không ai thấy thì vô dụng. FR-DEAL-006 là cầu serving biến tín hiệu P(bottom within 14d) thành cảnh báo hành động được, đúng như §3.5 mô tả. Thiếu nó, người dùng theo dõi sản phẩm không bao giờ biết món hàng sắp chạm đáy và bỏ lỡ thời điểm mua, làm rỗng giá trị cốt lõi của module DEAL. Sai ngưỡng 0,7 (hạ thấp hoặc đổi sang so sánh không strict) làm tràn cảnh báo nhiễu, kéo precision xuống và phá niềm tin (§5.4). Thiếu idempotent + cooldown thì người dùng bị bắn lại cùng một cảnh báo mỗi đêm suốt thời gian P còn cao, dẫn tới mệt mỏi cảnh báo và tắt thông báo - mất luôn kênh tới người dùng. Bỏ maturity gate thì batch chấm điểm cả SKU dữ liệu mỏng, lan truyền dự báo không đáng tin thành cảnh báo sai."
---

## §1 - Mô tả (BCP-14 normative)

Service DEAL **MUST** chạy một batch hằng đêm chấm điểm mọi SKU đủ điều kiện bằng tín hiệu đáy giá chung `p_bottom_14d` (đọc từ bảng `price_forecast` do Python sinh), và bắn cảnh báo khi `p_bottom_14d > 0.7`, khớp vào `alert_rule` type `'bottom_predicted'` (FR-TRACK-003) rồi enqueue vào notification fan-out (FR-NOTIF-003). Batch là model-agnostic và idempotent. Hợp đồng:

1. **MUST** đăng ký một cron job chạy hằng đêm vào khung off-peak `02:00 Asia/Ho_Chi_Minh` (DEC-DEAL-51); job dùng khóa (advisory lock) để không chạy chồng nếu lần trước chưa xong.
2. **MUST** chỉ chấm điểm SKU đã đạt `MATURE` qua maturity gate `IsFeatureReady` của FR-DEAL-002 (DEC-DEAL-54): query `price_forecast` JOIN điều kiện trưởng thành, loại SKU `NEW`/`WARMING` khỏi vòng chấm điểm.
3. **MUST** đọc cột `p_bottom_14d` chung của `price_forecast` (DEC-DEAL-52): batch KHÔNG biết và KHÔNG quan tâm dòng đó do Prophet (FR-DEAL-004) hay LightGBM (FR-DEAL-005) sinh; hai model hoán đổi được qua cùng một hợp đồng bảng.
4. **MUST** chỉ lấy bản dự báo còn tươi: dòng `price_forecast` có `scored_at >= now() - INTERVAL '36 hours'` (bao một nhịp batch dự báo lỡ); dự báo cũ hơn coi là stale và bị loại khỏi chấm điểm đêm nay.
5. **MUST** áp ngưỡng `p_bottom_14d > 0.7` đúng nguyên văn §3.5 (DEC-DEAL-50): so sánh strict greater-than, `0.7` là hằng load-bearing; `p_bottom_14d == 0.7` KHÔNG bắn cảnh báo.
6. **MUST** với mỗi SKU vượt ngưỡng, join sang `alert_rule` type `'bottom_predicted'` đang bật (`active = true`) để lấy danh sách user đang theo dõi SKU đó muốn nhận cảnh báo đáy (FR-TRACK-003); SKU không có luật khớp thì bỏ qua, không enqueue gì.
7. **MUST** idempotent ở mức (user, product, ngày): tối đa một cảnh báo đáy cho mỗi cặp `(user_id, product_id)` trong một ngày lịch (`fired_on`), được bảo đảm bởi `UNIQUE (user_id, product_id, fired_on)` trên `bottom_alert_log` (DEC-DEAL-53). Retry batch trong cùng ngày KHÔNG sinh cảnh báo trùng.
8. **MUST** chống spam liên ngày bằng cooldown rising-edge (DEC-DEAL-53): KHÔNG bắn lại cảnh báo cho cùng `(user_id, product_id)` nếu đã có alert trong vòng `cooldown` ngày gần nhất (mặc định 7 ngày) mà `p_bottom_14d` vẫn liên tục `> 0.7`. Chỉ bắn trên cạnh lên (P vượt ngưỡng sau khi đã rơi xuống dưới, hoặc đã hết cooldown), không bắn đêm-này-qua-đêm-khác khi P dính ngưỡng cao.
9. **MUST** enqueue mỗi cảnh báo đã qua dedupe vào notification fan-out (FR-NOTIF-003) với payload gồm `user_id`, `product_id`, `p_bottom_14d`, và lý do `reason = "bottom_predicted"`; fan-out lo chuyện chọn kênh và rate-limit phía người dùng.
10. **MUST** ghi mọi cảnh báo đã bắn vào `bottom_alert_log (user_id, product_id, fired_on, p_bottom)` ngay khi enqueue thành công, để dedupe ngày và cooldown đọc lại ở đêm sau.
11. **MUST** xử lý lỗi enqueue cục bộ: nếu fan-out từ chối một mục, KHÔNG ghi `bottom_alert_log` cho mục đó (để đêm sau thử lại), ghi log lỗi và tiếp tục các mục còn lại - một SKU lỗi không làm hỏng cả batch.
12. **SHOULD** phát OTel metric: `deal_nightly_scored_total` (counter số SKU đã chấm), `deal_bottom_alert_fired_total` (counter cảnh báo đã enqueue), `deal_alert_dedupe_skipped_total{reason}` (counter bỏ qua vì trùng-ngày hoặc cooldown), `deal_nightly_batch_duration_ms` (histogram).
13. **MUST** chạy toàn batch trong một transaction-per-user-batch hợp lý (đọc forecast và alert_rule nhất quán trong một ảnh chụp), nhưng commit `bottom_alert_log` theo từng mục đã enqueue để lỗi giữa chừng không mất dấu các cảnh báo đã gửi.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao batch hằng đêm off-peak chứ không realtime (DEC-DEAL-51)?** Dự báo đáy giá đổi chậm: `p_bottom_14d` là xác suất chạm đáy trong cửa sổ 14 ngày, không nhảy theo từng phút. Chấm lại mỗi đêm một lần là đủ độ tươi mà rẻ hơn nhiều so với streaming. Chạy lúc 02:00 giờ Việt Nam né giờ cao điểm đọc giá ban ngày, để tải nặng (quét toàn bộ SKU MATURE) không tranh tài nguyên với truy vấn người dùng.

**Vì sao ngưỡng đúng 0,7, so sánh strict (DEC-DEAL-50, §3.5)?** §3.5 chốt nguyên văn "alert nếu P(bottom within 14d) > 0.7". Đây là điểm đánh đổi precision-recall: 0,7 đủ cao để chỉ bắn khi mô hình khá chắc, ưu tiên precision hơn recall. Một cảnh báo "sắp chạm đáy" mà sai thì người dùng mua hớ và mất tin (§5.4); thà bỏ lỡ vài đáy còn hơn kêu sói nhầm. Strict greater-than giữ đúng ranh giới: tại đúng 0,7 mô hình mới ở mức cân bằng, chưa đủ chắc để làm phiền người dùng.

**Vì sao model-agnostic qua bảng chung (DEC-DEAL-52)?** Prophet (FR-DEAL-004) và LightGBM (FR-DEAL-005) là hai cách tính ra cùng một đại lượng: xác suất chạm đáy 14 ngày. Nếu batch serving gắn cứng vào một model, đổi model là phải sửa batch. Đọc cột `p_bottom_14d` chung của `price_forecast` tách hẳn tầng serving (Go) khỏi tầng sinh dự báo (Python): nâng cấp từ Prophet sang LightGBM, hay chạy A/B hai model, không đụng gì tới code cảnh báo. Batch chỉ tin một hợp đồng: bảng có cột xác suất tươi cho SKU MATURE.

**Vì sao idempotent ngày + cooldown rising-edge (DEC-DEAL-53)?** Một cảnh báo lặp lại mỗi đêm trong khi `p_bottom_14d` cứ dính trên 0,7 là cách nhanh nhất huấn luyện người dùng phớt lờ thông báo của SănDeal. Khi đã phớt lờ, họ tắt push, và ta mất kênh tới người dùng vĩnh viễn. `UNIQUE (user_id, product_id, fired_on)` chặn trùng trong một ngày (kể cả khi batch retry). Cooldown rising-edge chặn trùng liên ngày: một xác suất đáy cao kéo dài một tuần chỉ đáng một cảnh báo, không phải bảy. Người dùng nhận đúng một tín hiệu "giờ là lúc mua" thay vì bảy tiếng ồn.

**Vì sao chỉ chấm SKU MATURE (DEC-DEAL-54)?** Dự báo đáy trên SKU chưa đủ 90 ngày lịch sử là dự báo trên cát: cửa sổ quá ngắn để mô hình mùa vụ có nghĩa. FR-DEAL-002 đã chốt cổng `IsFeatureReady` cho mọi tính năng công khai; cảnh báo đáy đi qua cùng cổng đó để không bao giờ bắn dựa trên dự báo từ dữ liệu mỏng. Giữ nhất quán một định nghĩa "đủ chín" toàn module.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/deal/migrations/0002_bottom_alert_log.sql
-- Nhật ký cảnh báo đáy đã bắn. Dùng cho dedupe ngày (UNIQUE) và cooldown liên ngày.
CREATE TABLE bottom_alert_log (
  user_id    BIGINT       NOT NULL,
  product_id BIGINT       NOT NULL REFERENCES tracked_product(id),
  fired_on   DATE         NOT NULL,                 -- ngày lịch theo Asia/Ho_Chi_Minh
  p_bottom   DOUBLE PRECISION NOT NULL CHECK (p_bottom > 0.7),
  fired_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (user_id, product_id, fired_on)            -- idempotent 1 alert/cặp/ngày (§1 #7)
);

-- Tra cooldown liên ngày: alert gần nhất của 1 cặp (user, product).
CREATE INDEX idx_bal_cooldown ON bottom_alert_log (user_id, product_id, fired_on DESC);
```

### Vòng batch (Go)

```go
// services/deal/internal/batch/nightly_score.go
package batch

// RunNightlyScore chấm điểm mọi SKU MATURE, bắn cảnh báo đáy khi p_bottom_14d > 0.7.
// Model-agnostic: chỉ đọc price_forecast.p_bottom_14d (Prophet hoặc LightGBM).
func (b *Batch) RunNightlyScore(ctx context.Context, today time.Time) error {
    // §1 #2 #3 #4 #5: chỉ SKU MATURE, dự báo còn tươi, p_bottom_14d > 0.7 (strict).
    rows, err := b.pool.Query(ctx, `
        SELECT f.product_id, f.p_bottom_14d
        FROM price_forecast f
        JOIN tracked_product tp ON tp.id = f.product_id
        WHERE f.p_bottom_14d > 0.7                              -- DEC-DEAL-50, strict
          AND f.scored_at  >= now() - INTERVAL '36 hours'       -- §1 #4, dự báo tươi
          AND now() - tp.first_seen >= INTERVAL '90 days'`)     // §1 #2, maturity gate
    if err != nil {
        return err
    }
    defer rows.Close()

    var scored, fired, skipped int
    for rows.Next() {
        var productID int64
        var pBottom float64
        if err := rows.Scan(&productID, &pBottom); err != nil {
            return err
        }
        scored++
        // §1 #6: gom user từ alert_rule type 'bottom_predicted' đang bật.
        users, err := b.matchBottomRules(ctx, productID)
        if err != nil {
            b.log.Warn("match rules failed", "product", productID, "err", err)
            continue // §1 #11: một SKU lỗi không làm hỏng batch
        }
        for _, userID := range users {
            if b.shouldSkip(ctx, userID, productID, today) { // §1 #7 #8 dedupe + cooldown
                skipped++
                continue
            }
            if err := b.enqueueAndLog(ctx, userID, productID, pBottom, today); err != nil {
                b.log.Warn("enqueue failed", "user", userID, "product", productID, "err", err)
                continue // §1 #11: không ghi log -> đêm sau thử lại
            }
            fired++
        }
    }
    metrics.NightlyScored(scored)
    metrics.BottomAlertFired(fired)
    metrics.DedupeSkipped(skipped)
    return rows.Err()
}
```

### Khớp luật + dedupe + enqueue (Go)

```go
// services/deal/internal/batch/alert_match.go
// matchBottomRules trả các user_id có alert_rule type 'bottom_predicted' đang bật cho SKU.
func (b *Batch) matchBottomRules(ctx context.Context, productID int64) ([]int64, error) {
    rows, err := b.pool.Query(ctx, `
        SELECT user_id FROM alert_rule
        WHERE product_id = $1 AND rule_type = 'bottom_predicted' AND active = true`,
        productID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var users []int64
    for rows.Next() {
        var u int64
        if err := rows.Scan(&u); err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}
```

```go
// services/deal/internal/batch/dedupe.go
const cooldownDays = 7 // cooldown rising-edge liên ngày (DEC-DEAL-53)

// shouldSkip: true nếu (user, product) đã có alert hôm nay (dedupe ngày)
// hoặc còn trong cooldown rising-edge (đã bắn trong cooldownDays ngày gần nhất).
func (b *Batch) shouldSkip(ctx context.Context, userID, productID int64, today time.Time) bool {
    var lastFired *time.Time
    err := b.pool.QueryRow(ctx, `
        SELECT max(fired_on) FROM bottom_alert_log
        WHERE user_id = $1 AND product_id = $2`, userID, productID).Scan(&lastFired)
    if err != nil || lastFired == nil {
        return false // chưa từng bắn -> cạnh lên, cho phép
    }
    days := int(today.Sub(*lastFired).Hours() / 24)
    return days < cooldownDays // trong cooldown (gồm cả cùng ngày = 0) -> bỏ qua
}

// enqueueAndLog enqueue vào notification fan-out rồi ghi bottom_alert_log.
// Ghi log CHỈ khi enqueue thành công (§1 #10 #11).
func (b *Batch) enqueueAndLog(ctx context.Context, userID, productID int64, p float64, today time.Time) error {
    if err := b.notif.Enqueue(ctx, notif.Item{
        UserID:    userID,
        ProductID: productID,
        Reason:    "bottom_predicted",
        Payload:   map[string]any{"p_bottom_14d": p},
    }); err != nil {
        return err
    }
    _, err := b.pool.Exec(ctx, `
        INSERT INTO bottom_alert_log (user_id, product_id, fired_on, p_bottom)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, product_id, fired_on) DO NOTHING`, // idempotent retry
        userID, productID, today, p)
    return err
}
```

---

## §4 - Acceptance criteria

1. Cron đăng ký chạy `02:00 Asia/Ho_Chi_Minh`; advisory lock chặn chạy chồng khi lần trước chưa xong.
2. Vòng chấm điểm chỉ lấy SKU `MATURE` (`first_seen >= 90 ngày`); SKU `NEW`/`WARMING` không vào kết quả query.
3. Batch đọc `p_bottom_14d` từ `price_forecast` không phụ thuộc model; dòng do Prophet hay LightGBM sinh đều được chấm như nhau.
4. Dòng `price_forecast` có `scored_at` cũ hơn 36 giờ bị loại; chỉ dự báo tươi vào chấm điểm.
5. `p_bottom_14d = 0.71` -> vượt ngưỡng (ứng viên bắn); `p_bottom_14d = 0.69` -> không vượt; `p_bottom_14d = 0.70` -> KHÔNG bắn (strict greater-than).
6. SKU vượt ngưỡng join đúng `alert_rule` type `'bottom_predicted'` đang bật; SKU không có luật khớp -> không enqueue.
7. Hai lần chạy batch trong cùng một ngày cho cùng `(user, product)` -> chỉ một dòng `bottom_alert_log`, chỉ một enqueue (idempotent ngày).
8. `(user, product)` đã có alert trong vòng 7 ngày, `p_bottom_14d` vẫn `> 0.7` ở đêm sau -> bị cooldown bỏ qua, không enqueue lại.
9. Mỗi cảnh báo qua dedupe enqueue vào fan-out với payload `user_id`, `product_id`, `p_bottom_14d`, `reason = "bottom_predicted"`.
10. `bottom_alert_log` được ghi ngay sau khi enqueue thành công, với `p_bottom` đúng giá trị đã bắn.
11. Khi fan-out từ chối một mục, mục đó KHÔNG được ghi `bottom_alert_log` và các mục khác vẫn chạy tiếp (một lỗi không làm hỏng batch).
12. Metric `deal_nightly_scored_total`, `deal_bottom_alert_fired_total`, `deal_alert_dedupe_skipped_total{reason}` tăng đúng theo từng nhánh.
13. `CHECK (p_bottom > 0.7)` trên `bottom_alert_log` từ chối ghi giá trị tại hoặc dưới ngưỡng (bảo hiểm bất biến ngưỡng ở tầng DB).

---

## §5 - Kiểm thử (verification)

```go
// services/deal/internal/batch/nightly_score_test.go
package batch

func TestNightly_FiresAboveThreshold(t *testing.T) {
    b, deps := setupBatch(t)
    seedMatureForecast(t, deps, productID, 0.71) // trên ngưỡng
    seedBottomRule(t, deps, userID, productID)   // có luật khớp
    require.NoError(t, b.RunNightlyScore(ctx, today))
    require.Equal(t, 1, deps.notif.Count()) // 0.71 -> bắn

    // 0.69 không bắn; 0.70 (đúng biên) cũng không bắn (strict >).
    for _, p := range []float64{0.69, 0.70} {
        b2, d2 := setupBatch(t)
        seedMatureForecast(t, d2, productID, p)
        seedBottomRule(t, d2, userID, productID)
        require.NoError(t, b2.RunNightlyScore(ctx, today))
        require.Equal(t, 0, d2.notif.Count(), "p=%.2f không được bắn", p)
    }
}

func TestNightly_RespectsMaturityGate(t *testing.T) {
    b, deps := setupBatch(t)
    seedImmatureForecast(t, deps, productID, 0.95) // SKU 30 ngày, P rất cao
    seedBottomRule(t, deps, userID, productID)
    require.NoError(t, b.RunNightlyScore(ctx, today))
    require.Equal(t, 0, deps.notif.Count()) // chưa MATURE -> không chấm, không bắn
}

func TestNightly_DedupePerDay(t *testing.T) {
    b, deps := setupBatch(t)
    seedMatureForecast(t, deps, productID, 0.80)
    seedBottomRule(t, deps, userID, productID)
    require.NoError(t, b.RunNightlyScore(ctx, today))
    require.NoError(t, b.RunNightlyScore(ctx, today)) // chạy lại cùng ngày
    require.Equal(t, 1, deps.notif.Count())            // chỉ 1 alert
    require.Equal(t, 1, countAlertLog(t, deps, userID, productID))
}

func TestNightly_Cooldown_NoRepeatWhileHigh(t *testing.T) {
    b, deps := setupBatch(t)
    seedMatureForecast(t, deps, productID, 0.85)
    seedBottomRule(t, deps, userID, productID)
    require.NoError(t, b.RunNightlyScore(ctx, today))            // ngày 0: bắn
    require.NoError(t, b.RunNightlyScore(ctx, today.AddDate(0, 0, 3))) // ngày 3, P vẫn cao
    require.Equal(t, 1, deps.notif.Count()) // trong cooldown 7 ngày -> không bắn lại
    // sau cooldown (ngày 8) thì cạnh lên lại -> bắn.
    require.NoError(t, b.RunNightlyScore(ctx, today.AddDate(0, 0, 8)))
    require.Equal(t, 2, deps.notif.Count())
}

func TestNightly_MatchesAlertRuleType(t *testing.T) {
    b, deps := setupBatch(t)
    seedMatureForecast(t, deps, productID, 0.90)
    seedRule(t, deps, userID, productID, "price_drop") // sai type
    require.NoError(t, b.RunNightlyScore(ctx, today))
    require.Equal(t, 0, deps.notif.Count()) // chỉ 'bottom_predicted' mới khớp
}

func TestNightly_EnqueuesNotification(t *testing.T) {
    b, deps := setupBatch(t)
    seedMatureForecast(t, deps, productID, 0.78)
    seedBottomRule(t, deps, userID, productID)
    require.NoError(t, b.RunNightlyScore(ctx, today))
    item := deps.notif.Last()
    require.Equal(t, "bottom_predicted", item.Reason)
    require.Equal(t, userID, item.UserID)
    require.InDelta(t, 0.78, item.Payload["p_bottom_14d"], 1e-9)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0002_bottom_alert_log.sql` (bảng + UNIQUE + index cooldown) -> `dedupe.go` (shouldSkip + enqueueAndLog) -> `alert_match.go` (matchBottomRules) -> `nightly_score.go` (RunNightlyScore) -> đăng ký cron trong `main.go` -> tests. Cron dùng scheduler sẵn của deal-svc (cùng cơ chế refresh `category_prior` của FR-DEAL-002), đặt khung `02:00 Asia/Ho_Chi_Minh` và bọc advisory lock `pg_try_advisory_lock` để một instance chạy tại một thời điểm. Driver dùng `pgx`. Notification fan-out là interface `notif.Enqueuer` (FR-NOTIF-003) inject vào `Batch`, nên test thay bằng fake đếm mục, không cần hệ thống thật. Bảng `price_forecast` do tầng Python (FR-DEAL-004/005) sinh và là hợp đồng đọc-chỉ của batch này.

---

## §7 - Phụ thuộc

- **FR-DEAL-004** (depends_on) - Prophet baseline sinh `price_forecast.p_bottom_14d`; là nguồn dự báo chính cho batch.
- **FR-DEAL-005** (related) - LightGBM thay thế/song song, ghi cùng cột `p_bottom_14d` - đây là điều làm batch model-agnostic.
- **FR-TRACK-003** (depends_on) - định nghĩa `alert_rule` type `'bottom_predicted'`; batch join vào để biết user nào muốn cảnh báo đáy cho SKU nào.
- **FR-NOTIF-003** (consumer) - notification fan-out nhận mục enqueue, lo chọn kênh và rate-limit phía người dùng.
- **FR-DEAL-002** (related) - maturity gate `IsFeatureReady`/`MATURE` quyết định SKU nào đủ chín để chấm điểm.
- Extension/lib: PostgreSQL advisory lock; driver `pgx`; scheduler nội bộ deal-svc.

---

## §8 - Payload ví dụ

### Mục enqueue vào notification fan-out (SKU sắp chạm đáy, P = 0,78)

```json
{
  "user_id": 50231,
  "product_id": 90112,
  "reason": "bottom_predicted",
  "payload": {
    "p_bottom_14d": 0.78
  }
}
```

### Dòng log bỏ qua vì cooldown (đã bắn 3 ngày trước, P vẫn cao)

```
level=info msg="bottom alert skipped" user=50231 product=90112 reason=cooldown last_fired=2026-06-24 today=2026-06-27 p_bottom_14d=0.81
```

### Dòng ghi vào bottom_alert_log sau khi enqueue thành công

```sql
INSERT INTO bottom_alert_log (user_id, product_id, fired_on, p_bottom)
VALUES (50231, 90112, '2026-06-27', 0.78)
ON CONFLICT (user_id, product_id, fired_on) DO NOTHING;
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- `cooldownDays` cố định 7 hay co giãn theo độ biến động giá của SKU - tinh chỉnh khi có dữ liệu phản hồi thực.
- Có nên kèm "khoảng cách tới đáy ước lượng" (giá đáy dự báo) vào payload ngoài xác suất - chờ FR-DEAL-004/005 expose thêm cột rồi mở rộng hợp đồng.
- Gộp nhiều SKU vượt ngưỡng của cùng một user thành một thông báo tóm tắt (digest) thay vì từng cái - cân nhắc ở tầng fan-out (FR-NOTIF-003), không ở batch này.
- Ngưỡng `0.7` có nên khác theo ngành hàng (thời trang biến động khác điện máy) - giữ một hằng toàn cục theo §3.5 ở v1, để ML FR-DEAL-004/005 hấp thụ khác biệt ngành.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| `price_forecast` stale (Python không chạy) | điều kiện `scored_at >= now()-36h` | bỏ qua dòng cũ, ít/không cảnh báo đêm đó | Loại dự báo cũ; cảnh báo lỗi pipeline dự báo qua metric |
| Batch retry trong ngày bắn trùng | UNIQUE(user, product, fired_on) | enqueue/log trùng | `ON CONFLICT DO NOTHING` + dedupe ngày (§1 #7) |
| Bắn lại mỗi đêm khi P dính ngưỡng cao | test Cooldown | mệt mỏi cảnh báo, người dùng tắt push | Cooldown rising-edge 7 ngày (§1 #8) |
| Enqueue fan-out thất bại | lỗi từ `Enqueue` | mục đó không tới người dùng | Không ghi log -> đêm sau thử lại; một lỗi không hỏng batch (§1 #11) |
| Hạ ngưỡng dưới 0,7 hoặc đổi sang `>=` | code review + AC #5 + CHECK | tràn cảnh báo nhiễu, mất precision | Hằng load-bearing 0.7 strict; CHECK ở DB chặn ghi <= 0.7 |
| Chấm điểm SKU chưa MATURE | test MaturityGate | cảnh báo từ dự báo dữ liệu mỏng | JOIN điều kiện `first_seen >= 90d` (§1 #2) |
| Gắn cứng batch vào một model | code review | đổi model phải sửa serving | Đọc cột `p_bottom_14d` chung, không tham chiếu model (§1 #3) |
| Cron chạy chồng (lần trước treo) | advisory lock fail-fast | tải đôi, chấm trùng | `pg_try_advisory_lock`, một instance một lúc (§1 #1) |
| alert_rule bị tắt (`active=false`) lọt qua | điều kiện `active=true` | cảnh báo cho user đã tắt luật | Lọc `active=true` trong join (§1 #6) |

---

## §11 - Ghi chú

- FR-DEAL-006 là tầng serving của dự đoán đáy: nó biến một cột xác suất trong bảng thành cảnh báo hành động được, đúng như §3.5 mô tả "batch nightly score -> alert".
- Ngưỡng `0.7` là load-bearing và chốt từ §3.5: đổi nó là đổi cán cân precision-recall của cảnh báo đáy, phải sửa kèm đặc tả và audit.
- Model-agnostic qua `price_forecast.p_bottom_14d` là điểm tách tầng quan trọng: Prophet và LightGBM hoán đổi được mà không đụng code cảnh báo.
- Idempotent ngày (UNIQUE) chống trùng khi retry; cooldown rising-edge chống spam liên ngày - hai cơ chế khác nhau cho hai loại lặp khác nhau.
- Maturity gate dùng lại đúng `IsFeatureReady` của FR-DEAL-002, giữ một định nghĩa "đủ chín" thống nhất toàn module DEAL.
- Ghi `bottom_alert_log` chỉ sau khi enqueue thành công là cố ý: nó cho phép đêm sau thử lại mục lỗi mà không mất dấu mục đã gửi.
- Batch đọc-chỉ với `price_forecast` (Python sinh) và `alert_rule` (FR-TRACK-003), chỉ ghi vào `bottom_alert_log` và hàng đợi fan-out - ranh giới quyền rõ ràng.

---

*Hết FR-DEAL-006. Status: ready_to_implement (mục tiêu audit 10/10).*
