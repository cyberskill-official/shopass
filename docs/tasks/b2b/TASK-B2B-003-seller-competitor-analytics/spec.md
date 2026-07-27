---
id: TASK-B2B-003
title: "Seller-facing competitor price analytics - seller theo dõi vị thế giá của chính mình so với phân vị thị trường (median/p25/p75) theo category từ market_trend_daily, KHÔNG lộ giá đối thủ đơn lẻ"
module: B2B
priority: COULD
status: done
verify: T
phase: P3
milestone: P3 - slice 2
slice: 2
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-B2B-001, TASK-B2B-002, TASK-PRICE-001]
depends_on: [TASK-B2B-001]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §6 mục 8 (Seller-facing analytics - theo dõi giá đối thủ, SaaS B2B)"
  - "docs/... §6 mục 7 (B2B data ẩn danh), §5.5 (PDPL Luật 91/2025)"
source_decisions:
  - "DEC-B2B-20: seller chỉ thấy vị thế giá CỦA CHÍNH SELLER so với phân vị thị trường ẩn danh (median/p25/p75) của category - KHÔNG bao giờ thấy giá của một đối thủ cụ thể, một shop cụ thể"
  - "DEC-B2B-21: giá của chính seller lấy từ tracked_product/price_daily của các SKU mà seller xác nhận sở hữu (seller_owned_sku liên kết bằng shop_id đã xác minh) - KHÔNG cho seller nhập tay tùy ý SKU của người khác"
  - "DEC-B2B-22: vị thế biểu diễn bằng percentile_rank của giá seller trong dải phân vị category (vd 'giá bạn ở khoảng p60 thị trường') - tổng hợp, không quy về cá thể"
  - "DEC-B2B-23: nếu category của SKU seller có ô market_trend_daily bị suppress (dưới k-anonymity) thì KHÔNG render vị thế cho ô đó - trả 'không đủ dữ liệu thị trường'"
  - "DEC-B2B-24: liên kết quyền sở hữu shop phải qua xác minh (verified_seller) trước khi seller xem analytics SKU của shop đó - chống seller dò giá shop người khác"
  - "DEC-B2B-25: KHÔNG ghi log truy vấn analytics của seller theo cách lộ ra seller đang theo dõi category nào của ai - log chỉ giữ ở mức tổng hợp/ẩn danh"

language: "PostgreSQL 16 + Go 1.22 (b2b-svc)"
service: shopass/services/b2b/
new_files:
  - services/b2b/migrations/0003_seller_owned_sku.sql
  - services/b2b/internal/seller/position.go
  - services/b2b/internal/seller/ownership.go
  - services/b2b/internal/api/seller_handler.go
  - services/b2b/internal/seller/position_test.go
  - services/b2b/internal/seller/ownership_test.go
modified_files:
  - services/b2b/internal/api/router.go            # đăng ký route GET /v1/b2b/seller/position
allowed_tools:
  - file_read: services/b2b/**
  - file_write: services/b2b/**
  - bash: cd services/b2b && go test ./...
disallowed_tools:
  - trả giá của một SKU/shop đối thủ cụ thể cho seller (vi phạm DEC-B2B-20, lộ giá đối thủ + rủi ro PDPL)
  - cho seller xem analytics của shop chưa xác minh quyền sở hữu (vi phạm DEC-B2B-24)
  - render vị thế khi ô market_trend_daily nguồn bị suppress (vi phạm DEC-B2B-23)

effort_hours: 8
sub_tasks:
  - "1.0h: 0003_seller_owned_sku.sql - bảng liên kết seller-shop đã xác minh + SKU sở hữu"
  - "1.5h: ownership.go - xác minh quyền sở hữu shop trước khi phục vụ analytics; chặn shop chưa verified"
  - "1.5h: position.go - tính percentile_rank của giá seller trong dải phân vị category (từ market_trend_daily)"
  - "1.5h: seller_handler.go - HTTP: xác thực seller, kiểm ownership, gọi position, 200/403/422"
  - "1.5h: ownership_test.go - shop verified xem được; shop chưa verified -> 403; SKU người khác -> 403"
  - "1.0h: position_test.go - vị thế đúng dải; ô suppress -> 'không đủ dữ liệu'; không trả giá cá thể"
  - "0.5h: OTel metric seller_position_served_total + seller_position_denied_total{reason}"

risk_if_skipped: "TASK-B2B-003 là sản phẩm SaaS B2B hướng seller (§6 mục 8) - cho người bán biết giá của họ đang ở đâu so với mặt bằng thị trường để định giá cạnh tranh. Rủi ro chí mạng là vượt ranh giới ẩn danh: nếu vô tình để seller suy ra giá của một đối thủ cụ thể thì SănDeal vừa biến thành công cụ do thám giá vừa vi phạm PDPL (§5.5) và phá moat niềm tin. Thiết kế phải tuyệt đối: seller chỉ thấy VỊ THẾ của chính mình trong dải phân vị tổng hợp, không bao giờ thấy con số của người khác. Thiếu xác minh quyền sở hữu shop thì một seller có thể dò analytics SKU của shop đối thủ - một lỗ hổng nghiêm trọng. Vì priority COULD nên task này ở slice sau, nhưng ranh giới bảo mật của nó là bắt buộc ngay từ thiết kế."
---

## §1 - Mô tả (BCP-14 normative)

Service B2B **MUST** cho seller xem vị thế giá CỦA CHÍNH SELLER so với dải phân vị thị trường ẩn danh (`median_p/p25_p/p75_p`) theo category, đọc từ `market_trend_daily`, sau khi xác minh quyền sở hữu shop, và **MUST NOT** bao giờ lộ giá của một đối thủ/shop/SKU cụ thể. Hợp đồng:

1. **MUST** chỉ trả cho seller: (a) giá SKU của chính seller, (b) dải phân vị thị trường tổng hợp của category (`median_p/p25_p/p75_p` từ `market_trend_daily`), (c) percentile_rank của giá seller trong dải đó. **MUST NOT** trả giá của bất kỳ SKU/shop nào khác (DEC-B2B-20).
2. **MUST** định nghĩa bảng `seller_owned_sku (id, seller_org_id, shop_id, product_id, verified, linked_at)`; chỉ SKU có `verified=true` mới được dùng (DEC-B2B-21).
3. **MUST** xác minh quyền sở hữu shop trước khi phục vụ analytics (DEC-B2B-24): seller chỉ xem được SKU thuộc `shop_id` mà tổ chức của họ đã được xác minh sở hữu. Yêu cầu xem SKU/shop chưa xác minh -> `403`.
4. **MUST** biểu diễn vị thế bằng `percentile_rank` của giá seller trong dải phân vị category (DEC-B2B-22), ví dụ "giá của bạn ở khoảng p60 thị trường". **MUST NOT** quy vị thế về một cá thể đối thủ.
5. **MUST** chỉ render vị thế khi ô `market_trend_daily` nguồn của category đó `suppressed=false` (DEC-B2B-23). Nếu ô bị suppress -> trả `422` với lý do `"insufficient_market_data"`; **MUST NOT** dựng vị thế trên ô suppress.
6. **MUST** đọc dải phân vị thị trường DUY NHẤT từ `market_trend_daily` qua repo của TASK-B2B-001 (ô đã phát hành); giá của chính seller đọc từ `price_daily` của các `product_id` đã xác minh sở hữu.
7. **MUST** expose endpoint `GET /v1/b2b/seller/position?shop_id=...&category_id=...&platform_id=...&day=...` trả vị thế cho SKU seller trong category.
8. **MUST** phân biệt mã trạng thái: `200` (có vị thế), `403` (chưa xác minh sở hữu shop/SKU), `422` (không đủ dữ liệu thị trường - ô suppress), `400` (tham số sai).
9. **MUST** không log truy vấn theo cách lộ seller đang theo dõi ai (DEC-B2B-25): log/metric chỉ giữ ở mức tổng hợp (đếm theo tier/lý do), không ghi cặp (seller -> đối thủ).
10. **SHOULD** phát OTel metric: `seller_position_served_total` (counter), `seller_position_denied_total{reason}` (counter), `seller_position_build_ms` (histogram).
11. **MUST** resolve `seller_org_id` từ phiên/claim của caller; **MUST NOT** tin `seller_org_id`/`shop_id` do client tự khai mà không đối chiếu `seller_owned_sku`.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao chỉ trả vị thế của chính seller, không giá đối thủ (DEC-B2B-20)? Đây là ranh giới sống còn. "Theo dõi giá đối thủ" (§6 mục 8) phải hiểu là "biết giá của mình ở đâu so với mặt bằng", KHÔNG phải "xem giá cụ thể của shop X". Cái sau biến SănDeal thành công cụ do thám giá, vi phạm PDPL và phá niềm tin. Vị thế phân vị tổng hợp cho seller đủ tín hiệu để định giá mà không lộ bất kỳ cá thể nào.

Vì sao xác minh quyền sở hữu shop (DEC-B2B-24)? Nếu bất kỳ seller nào cũng truy vấn analytics cho bất kỳ shop_id nào thì họ dò được vị thế shop đối thủ - lách ngay nguyên tắc trên bằng cách "đóng vai" shop khác. Bắt buộc `verified=true` cho cặp (seller_org, shop) đảm bảo seller chỉ soi chính mình.

Vì sao chỉ dùng dải phân vị từ market_trend_daily đã phát hành (DEC-B2B-23, §1 #6)? Dải median/p25/p75 đã qua cổng k-anonymity (TASK-B2B-001) nên tự nó an toàn - không quy về cá thể. Nếu category quá hẹp khiến ô bị suppress thì cũng nghĩa là dải phân vị không đủ "đông" để an toàn; render vị thế lúc đó có thể vô tình lộ. Nên khi ô suppress, trả 422 thay vì cố dựng.

Vì sao không log cặp seller -> đối thủ (DEC-B2B-25)? Bản thân hành vi "seller A đang xem category mỹ phẩm tầm giá X" là thông tin nhạy cảm về chiến lược của A. Giữ log ở mức tổng hợp tránh tạo ra một kho dữ liệu cạnh tranh thứ cấp mà chính SănDeal lại phải bảo vệ.

Vì sao percentile_rank thay vì khoảng cách tới một SKU (DEC-B2B-22)? percentile_rank ("bạn ở p60") là phát biểu về vị trí trong phân phối, không tham chiếu sản phẩm nào. Nó cho seller hiểu "đắt hơn 60% thị trường" mà không cần và không thể chỉ ra "đắt hơn shop nào".

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/b2b/migrations/0003_seller_owned_sku.sql
CREATE TABLE seller_owned_sku (
  id           BIGSERIAL   PRIMARY KEY,
  seller_org_id BIGINT     NOT NULL,        -- tham chiếu tổ chức B2B (b2b_subscription.id hoặc org riêng)
  shop_id      TEXT        NOT NULL,
  product_id   BIGINT      NOT NULL,        -- tracked_product.id
  verified     BOOLEAN     NOT NULL DEFAULT false,
  linked_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (seller_org_id, product_id)
);

CREATE INDEX idx_sos_verified ON seller_owned_sku (seller_org_id, shop_id)
  WHERE verified = true;
```

### Vị thế (Go)

```go
// services/b2b/internal/seller/position.go
type Position struct {
    CategoryID     int64    `json:"category_id"`
    PlatformID     int16    `json:"platform_id"`
    Day            string   `json:"day"`
    SellerPrice    int64    `json:"seller_price"`      // giá của CHÍNH seller
    MarketMedian   int64    `json:"market_median_p"`   // dải tổng hợp ẩn danh
    MarketP25      int64    `json:"market_p25_p"`
    MarketP75      int64    `json:"market_p75_p"`
    PercentileRank float64  `json:"percentile_rank"`   // 0..100, vị thế của seller
    // KHÔNG có trường nào chứa giá/khóa của đối thủ
}

// rank tính vị thế của giá seller trong dải phân vị category (xấp xỉ tuyến tính
// p25->p75); chỉ dùng số tổng hợp, không tham chiếu SKU đối thủ.
func rank(sellerPrice, p25, median, p75 int64) float64 {
    switch {
    case sellerPrice <= p25:
        return 25 * float64(sellerPrice) / float64(maxI(p25, 1))
    case sellerPrice <= median:
        return 25 + 25*float64(sellerPrice-p25)/float64(maxI(median-p25, 1))
    case sellerPrice <= p75:
        return 50 + 25*float64(sellerPrice-median)/float64(maxI(p75-median, 1))
    default:
        return 75 + min(25, 25*float64(sellerPrice-p75)/float64(maxI(p75, 1)))
    }
}
```

### Ownership gate (§1 #3)

```go
// services/b2b/internal/seller/ownership.go
// assertOwned chặn truy vấn nếu seller chưa được xác minh sở hữu shop_id.
func (o *Ownership) assertOwned(ctx context.Context, sellerOrgID int64, shopID string) error {
    var n int
    err := o.pool.QueryRow(ctx,
        `SELECT count(*) FROM seller_owned_sku
         WHERE seller_org_id=$1 AND shop_id=$2 AND verified=true`,
        sellerOrgID, shopID).Scan(&n)
    if err != nil {
        return err
    }
    if n == 0 {
        return ErrNotVerifiedOwner{ShopID: shopID} // -> 403
    }
    return nil
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `seller_owned_sku` tồn tại với UNIQUE `(seller_org_id, product_id)`.
2. Seller đã xác minh sở hữu shop -> `200` + vị thế (giá seller + dải phân vị + percentile_rank).
3. Seller truy vấn `shop_id` chưa xác minh sở hữu -> `403`.
4. Seller truy vấn `product_id` thuộc shop khác -> `403`.
5. Category có ô `market_trend_daily` bị suppress -> `422` `"insufficient_market_data"`; KHÔNG render vị thế.
6. Response KHÔNG chứa giá/khóa của bất kỳ SKU/shop nào khác (chỉ giá của chính seller + dải tổng hợp).
7. `percentile_rank` nằm trong 0..100 và tăng đơn điệu theo giá seller (giá cao hơn -> rank cao hơn) trong cùng dải.
8. `seller_org_id` lấy từ phiên caller; tham số tự khai không khớp `seller_owned_sku` bị từ chối.
9. Log/metric KHÔNG ghi cặp (seller -> category/đối thủ cụ thể); chỉ đếm tổng hợp theo lý do.
10. Dải phân vị đọc từ `market_trend_daily` (ô phát hành); giá seller đọc từ `price_daily` của SKU đã xác minh.

---

## §5 - Kiểm thử (verification)

```go
// services/b2b/internal/seller/ownership_test.go
func TestOwnership_Verified_OK(t *testing.T) {
    o := setupOwnership(t, seedOwned(1, "shopA", 100, true))
    require.NoError(t, o.assertOwned(ctx, 1, "shopA"))
}

func TestOwnership_NotVerified_403(t *testing.T) {
    o := setupOwnership(t, seedOwned(1, "shopA", 100, false)) // chưa verified
    err := o.assertOwned(ctx, 1, "shopA")
    require.ErrorAs(t, err, &ErrNotVerifiedOwner{})
}

func TestOwnership_OtherShop_403(t *testing.T) {
    o := setupOwnership(t, seedOwned(1, "shopA", 100, true))
    err := o.assertOwned(ctx, 1, "shopB") // seller 1 không sở hữu shopB
    require.ErrorAs(t, err, &ErrNotVerifiedOwner{})
}

// services/b2b/internal/seller/position_test.go
func TestPosition_SuppressedMarket_422(t *testing.T) {
    h := setupHandler(t, withTrendCell(/* category 991 suppressed */), withOwned(1, "shopA", 100, true))
    res := h.get("/v1/b2b/seller/position?shop_id=shopA&category_id=991&platform_id=1&day=2026-06-20")
    require.Equal(t, 422, res.Code)
}

func TestPosition_NoCompetitorPrice(t *testing.T) {
    body := servePosition(t /* dải median 300k, seller 320k */)
    // chỉ có giá seller + dải tổng hợp
    require.Contains(t, body, "seller_price")
    require.Contains(t, body, "market_median_p")
    require.NotContains(t, body, "competitor")
    require.NotContains(t, body, "shop_id\":\"shopB") // không lộ shop khác
}

func TestRank_Monotonic(t *testing.T) {
    r1 := rank(240_000, 250_000, 300_000, 410_000)
    r2 := rank(320_000, 250_000, 300_000, 410_000)
    r3 := rank(500_000, 250_000, 300_000, 410_000)
    require.Less(t, r1, r2)
    require.Less(t, r2, r3)
    require.GreaterOrEqual(t, r1, 0.0)
    require.LessOrEqual(t, r3, 100.0)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0003 (seller_owned_sku) -> ownership.go (cổng xác minh) -> position.go (percentile_rank từ dải phân vị) -> seller_handler.go -> tests. Handler chạy sau gateway (TASK-INFRA-001). Xác minh quyền sở hữu shop (gắn `verified=true`) là quy trình ngoài phạm vi task này (vd seller chứng minh qua kênh chính thức của sàn) - task chỉ tiêu thụ cờ `verified`. Dải phân vị lấy qua `QueryCells` của TASK-B2B-001.

---

## §7 - Phụ thuộc

- TASK-B2B-001 - `market_trend_daily` + `QueryCells` (ô đã phát hành) là nguồn dải phân vị thị trường.
- TASK-PRICE-001 - `tracked_product` (product_id, shop_id) để liên kết SKU seller.
- TASK-B2B-002 (cùng module) - chia sẻ khái niệm tổ chức/subscription B2B.
- TASK-INFRA-001 (gateway) - xác thực + rate-limit.
- Extension/lib: driver `pgx`.

---

## §8 - Payload ví dụ

### Vị thế (200)

```json
{
  "category_id": 7,
  "platform_id": 1,
  "day": "2026-06-20",
  "seller_price": 320000,
  "market_median_p": 320000,
  "market_p25_p": 250000,
  "market_p75_p": 410000,
  "percentile_rank": 50.0
}
```

### Chưa xác minh sở hữu (403)

```json
{ "error": "not_verified_owner", "shop_id": "shopB", "message": "Cần xác minh quyền sở hữu shop trước khi xem analytics." }
```

### Không đủ dữ liệu thị trường (422)

```json
{ "error": "insufficient_market_data", "category_id": 991, "message": "Category này chưa đủ dữ liệu thị trường ẩn danh để hiển thị vị thế." }
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Vị thế theo chuỗi thời gian (xu hướng percentile_rank của seller qua nhiều ngày) - mở rộng từ một ngày sang khoảng ở slice sau.
- Gợi ý định giá (vd "giảm 5% để vào p40") - thêm khi có nhu cầu, vẫn chỉ dựa dải tổng hợp.
- Quy trình tự động xác minh quyền sở hữu shop qua OAuth của sàn - tích hợp khi sàn mở API seller.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Lộ giá đối thủ cụ thể | TestPosition_NoCompetitorPrice | do thám giá, vi phạm PDPL | Response chỉ giá seller + dải tổng hợp (DEC-B2B-20) |
| Seller dò shop chưa sở hữu | TestOwnership_NotVerified_403 + OtherShop | xem vị thế đối thủ | Cổng verified=true (DEC-B2B-24) |
| Render trên ô suppress | TestPosition_SuppressedMarket_422 | rủi ro lộ qua dải thưa | 422 khi ô suppress (DEC-B2B-23) |
| Tin shop_id client tự khai | review handler | giả mạo quyền sở hữu | Đối chiếu seller_owned_sku (§1 #11) |
| Log lộ cặp seller -> đối thủ | review log/metric | kho dữ liệu cạnh tranh thứ cấp | Log mức tổng hợp (DEC-B2B-25) |
| percentile_rank phi lý (>100/<0) | TestRank_Monotonic | hiển thị sai | Kẹp 0..100 + đơn điệu |
| Đọc giá đối thủ từ price_daily trực tiếp | review position.go | lách k-anonymity | Dải lấy từ market_trend_daily, không từ raw đối thủ |

---

## §11 - Ghi chú

- Ranh giới sống còn: "theo dõi giá đối thủ" = biết vị thế của mình trong dải tổng hợp, KHÔNG = xem giá shop cụ thể.
- Hai cổng độc lập: xác minh quyền sở hữu shop (seller chỉ soi chính mình) + dải phân vị đã qua k-anonymity (tự an toàn).
- Khi ô thị trường bị suppress, từ chối render (422) thay vì cố dựng - dải thưa là dải có rủi ro lộ.
- percentile_rank là phát biểu về phân phối, không tham chiếu sản phẩm nào - cho tín hiệu định giá mà không lộ cá thể.
- Không biến chính SănDeal thành kho dữ liệu cạnh tranh: không log cặp seller -> đối thủ.

---

*Hết TASK-B2B-003. Status: ready_to_review (awaiting HITL) (mục tiêu audit 10/10).*

*HITL accept (operator merge-then-continue): feature PR #106 merge `dfcc99751fb3481680beab21cb997fcfb733a87e` → done.*
