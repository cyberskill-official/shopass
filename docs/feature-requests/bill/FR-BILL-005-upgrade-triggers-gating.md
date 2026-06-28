---
id: FR-BILL-005
title: "Trigger upgrade free->Premium (gamified) + feature gating theo tier - cổng kiểm tier ở backend (không tin client), điểm chạm gợi ý nâng cấp đúng lúc, không khóa cứng tính năng miễn phí lõi"
module: BILL
priority: SHOULD
status: ready_to_implement
verify: T
phase: P2
milestone: P2 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-BILL-001, FR-BILL-003, FR-INFRA-001, FR-TRACK-002, FR-DEAL-004]
depends_on: [FR-BILL-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §4.3 (free-tier mạnh, Premium nhẹ + gamified upgrade trigger; WTP thấp ở VN)"
  - "docs/... §4.1 (free-tier tài trợ affiliate; conversion free->paid 2-5%), §6 (tính năng Premium đề xuất)"
source_decisions:
  - "DEC-BILL-21: feature gating thực thi ở BACKEND theo subscription tier (FR-BILL-001 GetActive); KHÔNG tin cờ tier do client gửi"
  - "DEC-BILL-22: bảng plan_feature ánh xạ (tier -> feature_key -> limit) - nguồn sự thật cho gating; đổi quyền lợi là cập nhật bảng, không sửa code rải rác"
  - "DEC-BILL-23: tính năng miễn phí lõi (theo dõi giá, sale ảo, biểu đồ) KHÔNG bị khóa cứng - free-tier mạnh là chiến lược (§4.3); gating chỉ áp giới hạn (số wishlist, tần suất, tính năng nâng cao)"
  - "DEC-BILL-24: trigger upgrade là sự kiện gợi ý đúng lúc (chạm giới hạn, dùng tính năng Premium) -> trả tín hiệu cho client hiển thị CTA; KHÔNG ép buộc, KHÔNG dark pattern"
  - "DEC-BILL-25: gating fail-safe: lỗi đọc tier -> coi như free (an toàn, không cấp nhầm Premium); vượt giới hạn trả mã lỗi xác định cho client xử lý"

language: "Go 1.22 (bill-svc, shared gating middleware); PostgreSQL 16"
service: shopass/services/bill/
new_files:
  - services/bill/migrations/0005_plan_feature.sql
  - services/bill/internal/gating/gate.go
  - services/bill/internal/gating/triggers.go
  - services/bill/internal/gating/repo.go
  - services/bill/internal/gating/gate_test.go
  - services/bill/internal/gating/triggers_test.go
modified_files: []
allowed_tools:
  - file_read: services/bill/**
  - file_write: services/bill/**
  - bash: cd services/bill && go test ./...
disallowed_tools:
  - tin cờ tier do client gửi để mở Premium (vi phạm DEC-BILL-21 - bỏ qua gating phía client)
  - khóa cứng tính năng miễn phí lõi (theo dõi giá/sale ảo/biểu đồ) (vi phạm DEC-BILL-23, phá free-tier)
  - hardcode giới hạn per-tier rải rác thay vì plan_feature (vi phạm DEC-BILL-22)
  - dùng dark pattern ép upgrade (vi phạm DEC-BILL-24)

effort_hours: 6
sub_tasks:
  - "0.5h: 0005_plan_feature.sql - bảng plan_feature (tier, feature_key, limit_value) + seed quyền lợi free/Premium"
  - "1.0h: repo.go - LimitFor(tier, feature_key) + CountUsage hook (vd số wishlist hiện có)"
  - "1.5h: gate.go - Allow(ctx, userID, feature_key) -> đọc tier (FR-BILL-001) + plan_feature -> so giới hạn; fail-safe free"
  - "1.0h: triggers.go - đánh giá điểm chạm upgrade (chạm limit, dùng tính năng Premium) -> tín hiệu CTA, không ép"
  - "1.0h: gate_test.go - free chạm giới hạn wishlist bị chặn; Premium không; client gửi tier giả bị bỏ qua; lỗi tier -> free"
  - "0.5h: triggers_test.go - chạm limit phát tín hiệu upgrade; tính năng lõi không bao giờ bị chặn"
  - "0.5h: OTel metric feature_gate_denied_total{feature} + upgrade_trigger_shown_total{trigger}"

risk_if_skipped: "Mô hình kinh doanh là free-tier mạnh tài trợ affiliate + Premium nhẹ gamified (§4.3); gating + upgrade trigger là cơ chế chuyển free->paid (conversion 2-5%, §4.1). Không có gating ở backend thì hoặc mọi người dùng được mọi tính năng (không ai trả Premium, dòng doanh thu Premium sụp), hoặc gating chỉ ở client (user sửa cờ để mở Premium miễn phí - thất thoát). Nếu khóa cứng tính năng miễn phí lõi (theo dõi giá/sale ảo/biểu đồ) thì phá chính chiến lược free-tier mạnh vốn là điểm thu hút người dùng VN (WTP thấp) - đuổi người dùng đi. Nếu hardcode giới hạn rải rác thì đổi quyền lợi phải sửa nhiều nơi, dễ lệch. Nếu dùng dark pattern ép upgrade thì tổn hại niềm tin (moat của SănDeal). Làm sai sẽ hoặc giết doanh thu Premium hoặc giết tăng trưởng free."
---

## §1 - Mô tả (BCP-14 normative)

Service BILL **MUST** thực thi feature gating ở backend theo subscription tier (không tin client), với quyền lợi tier ở bảng `plan_feature`, giữ tính năng miễn phí lõi không bị khóa cứng, và phát tín hiệu trigger upgrade đúng lúc mà không ép buộc. Hợp đồng:

1. **MUST** định nghĩa bảng `plan_feature (id, tier, feature_key, limit_value)`: ánh xạ mỗi `tier` (FR-BILL-001) tới giới hạn của từng `feature_key` (DEC-BILL-22). `limit_value = -1` nghĩa không giới hạn (unlimited); `0` nghĩa không có quyền.
2. **MUST** thực thi gating ở BACKEND (DEC-BILL-21): hàm `Allow(ctx, userID, feature_key)` đọc tier hiện tại qua FR-BILL-001 `GetActive` rồi tra `plan_feature`. KHÔNG tin bất kỳ cờ tier/quyền nào do client gửi.
3. **MUST** giữ tính năng miễn phí lõi KHÔNG bị khóa cứng (DEC-BILL-23): theo dõi giá, phát hiện sale ảo (FR-DEAL-001), biểu đồ giá (FR-DEAL-003) phải khả dụng cho tier `free`. Gating chỉ áp GIỚI HẠN (số lượng/tần suất) và tính năng NÂNG CAO, không chặn giá trị lõi.
4. **MUST** so sánh mức dùng hiện tại với `limit_value` cho feature có đếm (ví dụ số wishlist, số sản phẩm theo dõi, dự đoán đáy nâng cao): dùng `< limit_value` -> cho phép; `>= limit_value` (và `limit_value >= 0`) -> từ chối với mã xác định `ErrLimitReached`. Khi đọc tier/`plan_feature` lỗi, `Allow` **MUST** fail-safe coi user như `free` (DEC-BILL-25): an toàn nghiêng về "ít quyền hơn", KHÔNG cấp nhầm Premium.
5. **MUST** đặt quyền lợi tier ở `plan_feature` làm nguồn sự thật (DEC-BILL-22): KHÔNG hardcode giới hạn per-tier rải rác trong code. Đổi quyền lợi là cập nhật `plan_feature`.
6. **MUST** trả mã lỗi xác định khi vượt giới hạn (`ErrLimitReached`) để tầng API ánh xạ sang một phản hồi rõ (ví dụ `402 Payment Required` hoặc `403` kèm gợi ý nâng cấp), KHÔNG để lỗi thô lọt ra.
7. **MUST** đánh giá trigger upgrade là tín hiệu gợi ý (DEC-BILL-24): khi user chạm giới hạn hoặc thử dùng tính năng Premium, `triggers.go` trả một tín hiệu `{trigger, suggested_tier}` cho client hiển thị CTA nâng cấp - tín hiệu, KHÔNG tự thu tiền, KHÔNG ép.
8. Trigger upgrade **MUST NOT** dùng dark pattern (DEC-BILL-24): không che nút đóng, không hù dọa, không ép quyết định ngay. CTA là một gợi ý người dùng có thể bỏ qua và tiếp tục dùng free.
9. **MUST** seed `plan_feature` idempotent với quyền lợi free + 3 bậc Premium qua `ON CONFLICT (tier, feature_key) DO NOTHING`; tính năng lõi để `free` ở mức khả dụng (limit hợp lý hoặc unlimited).
10. **MUST** expose:
    - `Allow(ctx, userID int64, featureKey string) (bool, error)` - quyết định gating.
    - `LimitFor(ctx, tier, featureKey string) (int64, error)` - đọc giới hạn.
    - `EvaluateTriggers(ctx, userID int64, event UsageEvent) (*UpgradeSignal, error)` - tín hiệu CTA.
11. Gating **MUST** tách bạch "có quyền không" (boolean/limit) khỏi "thanh toán" (FR-BILL-002/003): `Allow` chỉ đọc tier hiện tại; nâng cấp thật đi qua checkout. `Allow` không tự kích hoạt subscription.
12. **SHOULD** phát OTel: `feature_gate_denied_total{feature_key}` (counter), `upgrade_trigger_shown_total{trigger}` (counter), `feature_gate_eval_total{feature_key, decision}` (counter).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao gating ở backend, không tin client (DEC-BILL-21)?** Nếu client tự quyết "user này là Premium" qua một cờ, ai cũng sửa được cờ đó để mở Premium miễn phí - thất thoát doanh thu. Quyền lợi phải được quyết ở nơi không sửa được từ ngoài: backend đọc subscription thật (FR-BILL-001) rồi tra quyền. Client chỉ hiển thị theo những gì backend cho phép, không tự cấp quyền.

**Vì sao không khóa cứng tính năng lõi (DEC-BILL-23)?** Chiến lược của SănDeal ở thị trường WTP thấp là free-tier mạnh thu hút đông người dùng, tài trợ bằng affiliate, rồi một phần nhỏ (2-5%) lên Premium (§4.3). Nếu khóa theo dõi giá/sale ảo/biểu đồ sau paywall, ta phá chính sức hút đó - người dùng bỏ đi, mất cả base lẫn affiliate. Gating đúng là áp giới hạn (số wishlist, tần suất alert) và tính năng nâng cao, để giá trị lõi luôn miễn phí.

**Vì sao quyền lợi ở plan_feature, không hardcode (DEC-BILL-22)?** Quyền lợi từng tier sẽ được tinh chỉnh nhiều lần khi tìm điểm cân bằng free/paid. Rải `if tier == "premium_basic" { limit = 50 }` khắp code làm mỗi lần điều chỉnh phải sửa nhiều nơi và dễ lệch. Một bảng ánh xạ (tier, feature_key, limit) là nguồn sự thật; thử nghiệm quyền lợi là cập nhật dữ liệu.

**Vì sao fail-safe về free (DEC-BILL-25)?** Khi đọc tier lỗi (DB chập), có hai lựa chọn sai: coi user là Premium (cấp nhầm quyền, thất thoát) hoặc free (giới hạn nhầm người đã trả). Nghiêng về free an toàn hơn về mặt doanh thu - tệ nhất là một Premium tạm bị giới hạn (sửa được khi DB hồi), không phải cấp Premium miễn phí hàng loạt. An toàn nghiêng về "ít quyền".

**Vì sao trigger là tín hiệu, không ép (DEC-BILL-24, §1 #8)?** Niềm tin là moat của SănDeal (hậu-Honey). Dark pattern ép upgrade (che nút đóng, hù dọa) thu được vài chuyển đổi ngắn hạn nhưng phá niềm tin dài hạn. Trigger đúng lúc (user vừa chạm giới hạn, đang thấy giá trị) là gợi ý lịch sự người dùng bỏ qua được - chuyển đổi từ giá trị thật, không từ ép buộc.

**Vì sao tách gating khỏi thanh toán (§1 #11)?** `Allow` chỉ trả lời "user hiện có quyền không" dựa subscription hiện tại. Nó không nên tự thu tiền hay kích hoạt Premium - đó là việc của checkout (FR-BILL-002) và IPN (FR-BILL-003). Tách bạch giữ gating đơn giản, dễ test, và tránh lẫn "kiểm quyền" với "đổi quyền".

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/bill/migrations/0005_plan_feature.sql
CREATE TABLE plan_feature (
  id          BIGSERIAL PRIMARY KEY,
  tier        TEXT     NOT NULL
                CHECK (tier IN ('free','premium_basic','premium_plus','premium_pro')),
  feature_key TEXT     NOT NULL,
  limit_value BIGINT   NOT NULL,   -- -1 = unlimited; 0 = không có quyền; >0 = giới hạn
  UNIQUE (tier, feature_key)
);

-- Tính năng lõi: free vẫn khả dụng (KHÔNG khóa cứng) (§1 #3)
INSERT INTO plan_feature (tier, feature_key, limit_value) VALUES
  ('free',          'price_tracking',   -1),   -- theo dõi giá: unlimited cho free
  ('free',          'fake_sale_detect', -1),   -- sale ảo: free
  ('free',          'price_chart',      -1),   -- biểu đồ: free
  ('free',          'wishlist_items',   20),   -- giới hạn số wishlist item (gating mềm)
  ('free',          'bottom_predict',    0),   -- dự đoán đáy nâng cao: chỉ Premium
  ('premium_basic', 'wishlist_items',  100),
  ('premium_basic', 'bottom_predict',   -1),
  ('premium_plus',  'wishlist_items',  500),
  ('premium_plus',  'bottom_predict',   -1),
  ('premium_pro',   'wishlist_items',   -1),
  ('premium_pro',   'bottom_predict',   -1)
ON CONFLICT (tier, feature_key) DO NOTHING;
```

### Gate (Go)

```go
// services/bill/internal/gating/gate.go

// Allow quyết định gating ở backend theo subscription tier (§1 #2).
// Fail-safe: lỗi đọc tier -> coi như free (§1 #4).
func (g *Gate) Allow(ctx context.Context, userID int64, featureKey string) (bool, error) {
    tier := "free"
    if sub, ok, err := g.subs.GetActive(ctx, userID); err == nil && ok {
        tier = g.plans.TierOf(sub.PlanID)
    } // lỗi/không active -> giữ "free" (fail-safe, §1 #4)

    limit, err := g.repo.LimitFor(ctx, tier, featureKey)
    if err != nil {
        return false, nil // fail-safe: từ chối an toàn thay vì cấp nhầm
    }
    switch {
    case limit == 0:
        return false, nil          // không có quyền (§1 #1)
    case limit < 0:
        return true, nil           // unlimited
    default:
        used, err := g.repo.CountUsage(ctx, userID, featureKey)
        if err != nil { return false, nil }
        return used < limit, nil   // còn dưới giới hạn (§1 #4)
    }
}
```

### Triggers (Go)

```go
// services/bill/internal/gating/triggers.go

type UpgradeSignal struct {
    Trigger       string `json:"trigger"`        // ví dụ "wishlist_limit_reached"
    SuggestedTier string `json:"suggested_tier"` // ví dụ "premium_basic"
}

// EvaluateTriggers trả tín hiệu CTA gợi ý (KHÔNG ép, KHÔNG thu tiền) (§1 #7,#8).
func (g *Gate) EvaluateTriggers(ctx context.Context, userID int64, ev UsageEvent) (*UpgradeSignal, error) {
    allowed, _ := g.Allow(ctx, userID, ev.FeatureKey)
    if !allowed && ev.FeatureKey == "wishlist_items" {
        return &UpgradeSignal{Trigger: "wishlist_limit_reached", SuggestedTier: "premium_basic"}, nil
    }
    if ev.FeatureKey == "bottom_predict" && !allowed {
        return &UpgradeSignal{Trigger: "premium_feature_touch", SuggestedTier: "premium_basic"}, nil
    }
    return nil, nil // không có tín hiệu -> không hiển thị gì
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `plan_feature` tồn tại với UNIQUE `(tier, feature_key)` + CHECK tier.
2. Seed -> tính năng lõi (`price_tracking`, `fake_sale_detect`, `price_chart`) có `limit_value = -1` cho `free` (không khóa cứng).
3. `Allow(free_user, "price_tracking")` -> `true` (tính năng lõi luôn khả dụng cho free).
4. `Allow(free_user, "wishlist_items")` khi đã có 20 item (đạt giới hạn free) -> `false` (`ErrLimitReached` ở tầng trên).
5. `Allow(premium_basic_user, "wishlist_items")` khi có 50 item -> `true` (giới hạn 100).
6. `Allow(free_user, "bottom_predict")` -> `false` (limit 0, chỉ Premium); `Allow(premium_basic_user, "bottom_predict")` -> `true`.
7. Client gửi cờ `tier=premium_pro` trong request -> bị bỏ qua; `Allow` vẫn dùng subscription thật từ backend (free user vẫn bị giới hạn).
8. Lỗi đọc tier (mock subs fail) -> `Allow` coi như free (fail-safe), không cấp nhầm Premium.
9. Đổi `plan_feature.limit_value` cho một tier -> hành vi gating đổi theo, KHÔNG cần sửa code (nguồn sự thật là bảng).
10. `EvaluateTriggers(free_user, wishlist khi đầy)` -> trả `UpgradeSignal{trigger:"wishlist_limit_reached"}`; khi chưa đầy -> `nil` (không CTA).
11. `EvaluateTriggers` chỉ trả tín hiệu; KHÔNG kích hoạt subscription hay thu tiền (kiểm không gọi checkout).
12. Metric `feature_gate_denied_total{feature_key}` tăng khi từ chối; `upgrade_trigger_shown_total` tăng khi phát tín hiệu.

---

## §5 - Kiểm thử (verification)

```go
// services/bill/internal/gating/gate_test.go
func TestAllow_CoreFeatureFreeUser(t *testing.T) {
    g, free := setupFreeUser(t)
    for _, f := range []string{"price_tracking", "fake_sale_detect", "price_chart"} {
        ok, _ := g.Allow(ctx, free, f)
        require.True(t, ok, "tính năng lõi %s phải khả dụng cho free", f) // §1 #3
    }
}

func TestAllow_WishlistLimit_Free(t *testing.T) {
    g, free := setupFreeUser(t)
    seedWishlistItems(t, free, 20) // đạt giới hạn free (20)
    ok, _ := g.Allow(ctx, free, "wishlist_items")
    require.False(t, ok) // >= limit -> chặn
}

func TestAllow_WishlistLimit_Premium(t *testing.T) {
    g, prem := setupPremiumUser(t, "premium_basic")
    seedWishlistItems(t, prem, 50)
    ok, _ := g.Allow(ctx, prem, "wishlist_items")
    require.True(t, ok) // giới hạn 100, mới 50
}

func TestAllow_PremiumOnlyFeature(t *testing.T) {
    g, free := setupFreeUser(t)
    ok, _ := g.Allow(ctx, free, "bottom_predict")
    require.False(t, ok) // limit 0 cho free
    g2, prem := setupPremiumUser(t, "premium_basic")
    ok2, _ := g2.Allow(ctx, prem, "bottom_predict")
    require.True(t, ok2)
}

func TestAllow_IgnoresClientTierFlag(t *testing.T) {
    g, free := setupFreeUser(t)
    // request "tự xưng" premium_pro không ảnh hưởng - Allow đọc subscription thật
    ok, _ := g.Allow(withClaimedTier(ctx, "premium_pro"), free, "bottom_predict")
    require.False(t, ok) // vẫn free (§1 #2)
}

func TestAllow_FailSafeFree(t *testing.T) {
    g, prem := setupPremiumUserButSubsErrors(t) // đọc tier lỗi
    ok, _ := g.Allow(ctx, prem, "bottom_predict")
    require.False(t, ok) // fail-safe coi như free (§1 #4)
}
```

```go
// services/bill/internal/gating/triggers_test.go
func TestTriggers_WishlistFull_ShowsCTA(t *testing.T) {
    g, free := setupFreeUser(t)
    seedWishlistItems(t, free, 20)
    sig, _ := g.EvaluateTriggers(ctx, free, UsageEvent{FeatureKey: "wishlist_items"})
    require.NotNil(t, sig)
    require.Equal(t, "wishlist_limit_reached", sig.Trigger)
}

func TestTriggers_NoSignalWhenWithinLimit(t *testing.T) {
    g, free := setupFreeUser(t)
    seedWishlistItems(t, free, 5)
    sig, _ := g.EvaluateTriggers(ctx, free, UsageEvent{FeatureKey: "wishlist_items"})
    require.Nil(t, sig) // chưa chạm giới hạn -> không CTA (§1 #8)
}

func TestTriggers_DoesNotChargeOrActivate(t *testing.T) {
    g, free := setupFreeUser(t)
    seedWishlistItems(t, free, 20)
    g.EvaluateTriggers(ctx, free, UsageEvent{FeatureKey: "wishlist_items"})
    require.Equal(t, 0, g.checkoutCalls()) // chỉ tín hiệu, không thu tiền (§1 #11)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration `0005_plan_feature.sql` (ánh xạ tier->feature->limit + seed; tính năng lõi free unlimited) -> `repo.go` (`LimitFor`, `CountUsage`) -> `gate.go` (`Allow` đọc tier FR-BILL-001 + tra plan_feature + fail-safe free) -> `triggers.go` (`EvaluateTriggers` trả tín hiệu CTA) -> tests. `Allow` dùng làm middleware/kiểm trong các handler có giới hạn (ví dụ FR-TRACK-002 wishlist gọi `Allow` trước khi thêm). Quyền lợi đổi bằng cập nhật `plan_feature`, không sửa code. Nâng cấp thật đi qua checkout (FR-BILL-002), không phải `Allow`.

---

## §7 - Phụ thuộc

- **FR-BILL-001** - `subscription` + `GetActive` + `plan_catalog` (tier); gating đọc tier hiện tại từ đây.
- **FR-BILL-003 (liên quan)** - subscription chuyển `active` sau payment paid; gating phản ánh tier mới ngay.
- **FR-INFRA-001 (gateway)** - gắn JWT + `user_id` cho các handler dùng `Allow`.
- **FR-TRACK-002** - wishlist là feature có giới hạn (`wishlist_items`); gọi `Allow` trước khi thêm item.
- **FR-DEAL-004** - dự đoán đáy nâng cao (`bottom_predict`) là tính năng Premium; gating kiểm tier.
- Lib: `pgx`.

---

## §8 - Payload ví dụ

### Kiểm gating khi thêm wishlist item (nội bộ, FR-TRACK-002 gọi)

```go
ok, _ := gate.Allow(ctx, userID, "wishlist_items")
if !ok {
    // tầng API ánh xạ sang 402/403 + gợi ý nâng cấp
    sig, _ := gate.EvaluateTriggers(ctx, userID, gating.UsageEvent{FeatureKey: "wishlist_items"})
    writeUpgradeHint(w, sig) // client hiển thị CTA, user bỏ qua được
    return
}
```

### Tín hiệu upgrade trả cho client (gợi ý, không ép)

```json
{
  "trigger": "wishlist_limit_reached",
  "suggested_tier": "premium_basic"
}
```

### Quyền lợi tier trong plan_feature

```sql
SELECT tier, feature_key, limit_value FROM plan_feature WHERE feature_key='wishlist_items' ORDER BY limit_value;
--  free          | wishlist_items | 20
--  premium_basic | wishlist_items | 100
--  premium_plus  | wishlist_items | 500
--  premium_pro   | wishlist_items | -1   (unlimited)
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- A/B test ngưỡng giới hạn free (20 wishlist hợp lý chưa) - điều chỉnh qua `plan_feature` khi có dữ liệu conversion; schema cho phép đổi không sửa code.
- Gamification sâu hơn (điểm, streak, mở khóa tạm tính năng Premium) - thêm lớp trên trigger; FR này là nền gating + tín hiệu.
- Giới hạn theo tần suất thời gian (ví dụ X alert/ngày cho free) - thêm feature_key dạng rate + cửa sổ; hiện limit là đếm tổng.
- Hiển thị CTA cụ thể (copy, vị trí) - thuộc FR-WEB/FR-MOBILE; backend chỉ trả tín hiệu trung lập.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Tin cờ tier client | ignores-client-flag test | Premium miễn phí qua sửa cờ | Gating ở backend, đọc subscription thật (§1 #2) |
| Khóa cứng tính năng lõi | core-feature test | phá free-tier, đuổi user | Lõi unlimited cho free (§1 #3) |
| Hardcode giới hạn rải rác | review + đổi-bảng test | đổi quyền lợi lệch | plan_feature nguồn sự thật (§1 #5) |
| Lỗi tier cấp nhầm Premium | fail-safe test | thất thoát | Fail-safe về free (§1 #4) |
| Dark pattern ép upgrade | review (§1 #8) | tổn hại niềm tin | CTA bỏ qua được, không ép (§1 #8) |
| Lỗi limit thô lọt API | mã lỗi xác định | UX xấu | ErrLimitReached -> 402/403 (§1 #6) |
| Gating tự thu tiền | no-charge test | lẫn kiểm quyền với đổi quyền | Tách Allow khỏi checkout (§1 #11) |
| Vượt giới hạn không chặn | wishlist-limit test | Premium value rò cho free | So used < limit (§1 #4) |
| Seed quyền lợi nhân đôi | ON CONFLICT | giữ một dòng/cặp | Idempotent seed (§1 #9) |

---

## §11 - Ghi chú

- Gating + upgrade trigger là cơ chế chuyển free->paid (conversion 2-5%, §4.1) - trụ thu tiền của mô hình free-tier mạnh.
- Gating ở backend (đọc subscription thật) đóng lỗ hổng "client tự xưng Premium"; client chỉ hiển thị theo backend cho phép.
- Tính năng lõi (theo dõi giá/sale ảo/biểu đồ) KHÔNG khóa cứng - free-tier mạnh là sức hút ở thị trường WTP thấp (§4.3).
- Quyền lợi ở `plan_feature` cho phép A/B test ngưỡng bằng cập nhật dữ liệu, không sửa code.
- Fail-safe về free khi lỗi: an toàn nghiêng về "ít quyền", không cấp nhầm Premium hàng loạt.
- Trigger là tín hiệu lịch sự bỏ qua được, không dark pattern - bảo vệ moat niềm tin.
- Tách gating khỏi thanh toán: `Allow` kiểm quyền, checkout (FR-BILL-002) đổi quyền.

---

*Hết FR-BILL-005. Status: ready_to_implement (mục tiêu audit 10/10).*
