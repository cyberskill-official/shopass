---
id: TASK-CART-003
title: "optimizeCart - tối ưu giỏ/voucher/freeship (knapsack-like): filterByMinSpend + chooseBestShopVoucherPerShop + validStack + applyCaps -> combo giảm lớn nhất; bám pseudo-code §3.5(3)"
module: CART
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-CART-001, TASK-CART-002, TASK-CART-004, TASK-INFRA-001]
depends_on: [TASK-CART-001, TASK-CART-002]
blocks: [TASK-CART-004, TASK-MOBILE-002]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5(3) (pseudo-code optimizeCart: filterByMinSpend, chooseBestShopVoucherPerShop, validStack, applyCaps)"
  - "docs/... §3.7 (POST /v1/cart/optimize {platform, items[], vouchers[]} -> best combo)"
source_decisions:
  - "DEC-CART-13: optimizeCart bám đúng cấu trúc pseudo-code §3.5(3) - duyệt (platformVoucher x freeship x shopCombo), lọc validStack + meetsMinSpend, applyCaps, giữ best.discount"
  - "DEC-CART-14: validStack ủy quyền cho luật stacking per-country (TASK-CART-004) qua interface StackRules; TASK-CART-003 KHÔNG hardcode luật nước - chỉ gọi rules.ValidStack"
  - "DEC-CART-15: mọi phép tính giảm giá dùng số nguyên VND (int64); percent tính giá*pct/100 bằng số nguyên; applyCaps kẹp theo cap từng voucher - không float"
  - "DEC-CART-16: chooseBestShopVoucherPerShop chọn voucher shop tốt nhất CHO TỪNG shop độc lập (một shop tối đa một voucher shop), trên các shop có trong giỏ"
  - "DEC-CART-17: optimizer là hàm thuần (input: cart_items, vouchers, rules) - không tự đọc DB; caller (handler /v1/cart/optimize) nạp giỏ (TASK-CART-002) + voucher còn hiệu lực (TASK-CART-001 ListActive) rồi gọi"
  - "DEC-CART-18: kết quả trả best {discount, combo[]} kèm phân rã từng phần (shop/platform/freeship) để client hiển thị minh bạch người dùng tiết kiệm bao nhiêu từ đâu"

language: "Go 1.22 (cart-svc); hàm thuần optimizer + handler POST /v1/cart/optimize"
service: shopass/services/cart/
new_files:
  - services/cart/internal/optimizer/optimize.go
  - services/cart/internal/optimizer/discount.go
  - services/cart/internal/optimizer/types.go
  - services/cart/internal/optimizer/rules.go
  - services/cart/internal/api/optimize.go
  - services/cart/internal/optimizer/optimize_test.go
  - services/cart/internal/optimizer/discount_test.go
modified_files:
  - services/cart/internal/api/router.go         # đăng ký POST /v1/cart/optimize
allowed_tools:
  - file_read: services/cart/**
  - file_write: services/cart/**
  - bash: cd services/cart && go test ./...
disallowed_tools:
  - hardcode luật stacking per-country trong optimizer (vi phạm DEC-CART-14 - thuộc TASK-CART-004 qua interface)
  - dùng float cho phép tính giảm giá/cap (vi phạm DEC-CART-15, sai số tổng giảm)
  - cho optimizer đọc DB trực tiếp (vi phạm DEC-CART-17 - là hàm thuần, caller nạp dữ liệu)
  - bỏ applyCaps (vi phạm pseudo-code §3.5(3), tổng giảm vượt trần voucher thực tế)

effort_hours: 12
sub_tasks:
  - "1.0h: types.go - CartItem, Voucher (từ TASK-CART-001/002), OptimizeResult{discount, combo, breakdown}"
  - "1.0h: rules.go - interface StackRules{ValidStack(pv, fs, shopVouchers) bool}; TASK-CART-004 hiện thực"
  - "2.0h: discount.go - discountShop, discountPlatform, freeshipValue, applyCaps (int64, percent nguyên)"
  - "1.5h: optimize.go - filterByMinSpend + chooseBestShopVoucherPerShop"
  - "2.0h: optimize.go - vòng duyệt (platformVoucher x freeship x shopCombo) bám pseudo-code §3.5(3), giữ best"
  - "1.0h: optimize.go - meetsMinSpend (tổng giỏ/đơn so min_spend), ghép breakdown minh bạch"
  - "1.0h: optimize.go (handler) - POST /v1/cart/optimize: nạp giỏ + voucher, gọi optimizer thuần, trả JSON"
  - "1.5h: optimize_test.go - ví dụ §3.5(3) (giỏ 2 shop, ra 110k VN-stack), best chọn đúng combo, cap kẹp đúng"
  - "1.0h: discount_test.go - percent int, applyCaps kẹp, freeshipValue, min_spend gate"

risk_if_skipped: "optimizeCart là trái tim của tính năng tối ưu giỏ hàng - khác biệt cạnh tranh mà tài liệu nêu BeeCost còn yếu (§5.6: chưa tối ưu giỏ đa sàn mạnh). Đây là thứ trả lời câu hỏi tiền-tươi của persona Linh: với giỏ này, tổ hợp voucher nào cho tôi giảm nhiều nhất. Nếu thuật toán sai (bỏ applyCaps, hoặc validStack sai) thì con số tiết kiệm hiển thị cho người dùng SAI - hoặc hứa giảm nhiều hơn thực tế (người dùng thất vọng lúc thanh toán), hoặc bỏ lỡ combo tốt (mất giá trị). Nếu dùng float thì sai số tích lũy khi cộng shop+platform+freeship rồi so cap - lệch vài đồng đủ để chọn sai combo. Nếu hardcode luật stacking VN vào optimizer thì khi mở MY/PH (đã bỏ stacking 2025) phải sửa lõi thuật toán thay vì chỉ đổi luật - vi phạm tách biệt và dễ sai per-country. Đây là task effort cao nhất của CART vì là lõi tính toán; phải bám đúng pseudo-code §3.5(3) và để luật nước ở TASK-CART-004."
---

## §1 - Mô tả (BCP-14 normative)

Service CART **MUST** hiện thực `optimizeCart` - thuật toán knapsack-like chọn tổ hợp voucher (shop + platform + freeship) cho tổng giảm lớn nhất trên một giỏ hàng, bám đúng cấu trúc pseudo-code §3.5(3), với luật stacking ủy quyền cho TASK-CART-004 và mọi phép tính bằng số nguyên VND. Hợp đồng:

1. **MUST** hiện thực `optimizeCart(cartItems, vouchers, rules)` là hàm thuần (DEC-CART-17): không đọc DB; nhận giỏ (từ TASK-CART-002) + voucher còn hiệu lực (từ TASK-CART-001 `ListActive`) + luật stacking, trả combo tối ưu.
2. **MUST** bám cấu trúc pseudo-code §3.5(3) (DEC-CART-13): lọc voucher shop theo min_spend (`filterByMinSpend`), rồi duyệt mọi tổ hợp `(platformVoucher hoặc none) x (freeship hoặc none) x chooseBestShopVoucherPerShop`, bỏ tổ hợp `!validStack` hoặc `!meetsMinSpend`, tính tổng giảm với `applyCaps`, giữ `best.discount`.
3. **MUST** dùng `chooseBestShopVoucherPerShop` chọn voucher shop tốt nhất cho TỪNG shop độc lập (một shop tối đa một voucher shop) trên các shop có trong giỏ (DEC-CART-16).
4. **MUST** kiểm `validStack(platformVoucher, freeship, shopVouchers)` qua interface `StackRules` do TASK-CART-004 hiện thực (DEC-CART-14); TASK-CART-003 KHÔNG hardcode luật stacking theo nước - chỉ gọi `rules.ValidStack`.
5. **MUST** kiểm `meetsMinSpend`: tổng giá trị đơn (hoặc đơn của shop tương ứng) phải đạt `min_spend` của voucher trước khi áp; voucher không đạt ngưỡng bị loại khỏi tổ hợp.
6. **MUST** tính giảm giá bằng số nguyên VND (DEC-CART-15): voucher `amount` giảm `discount_value`; voucher `percent` giảm `giá * discount_value / 100` (chia nguyên); freeship giảm `freeshipValue`. KHÔNG dùng float.
7. **MUST** áp `applyCaps`: mỗi voucher có `cap` (trần giảm tuyệt đối) thì phần giảm của nó bị kẹp `min(computed, cap)` (DEC-CART-15). Bỏ `applyCaps` là sai pseudo-code và làm tổng giảm vượt trần thực tế.
8. **MUST** trả `OptimizeResult{discount, combo[], breakdown}` (DEC-CART-18): `discount` là tổng giảm tối ưu (int64 VND); `combo` là danh sách voucher được chọn; `breakdown` phân rã giảm theo từng phần (shop/platform/freeship) để hiển thị minh bạch người dùng tiết kiệm bao nhiêu từ đâu.
9. **MUST** xử lý giỏ rỗng và trường hợp không voucher nào áp được: trả `discount = 0`, `combo` rỗng - không lỗi (không tối ưu được nghĩa là giảm 0, không phải lỗi).
10. **MUST** đảm bảo kết quả tất định: cùng input (giỏ + voucher + rules) luôn cho cùng combo (sắp xếp ổn định khi nhiều combo bằng điểm) - để test và để người dùng thấy kết quả nhất quán.
11. API `POST /v1/cart/optimize` **MUST** là caller: nạp giỏ (TASK-CART-002) + voucher còn hiệu lực (TASK-CART-001) cho `platform`, gọi `optimizeCart` thuần với `rules` của nước tương ứng (TASK-CART-004), trả JSON `OptimizeResult`. Handler nằm sau JWT gateway (TASK-INFRA-001).
12. **SHOULD** phát OTel `cart_optimize_duration_ms` (histogram) và `cart_optimize_discount_vnd` (histogram tổng giảm) để theo dõi hiệu năng + giá trị mang lại.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao bám đúng pseudo-code §3.5(3) (DEC-CART-13)?** Tài liệu nguồn đã đặc tả thuật toán: duyệt tổ hợp (platform x freeship x shop-combo), lọc stack hợp lệ + đạt min_spend, áp cap, giữ max. Đây là knapsack-like với không gian nhỏ (số voucher platform/freeship ít, shop voucher chọn tốt nhất per shop nên không bùng nổ). Bám đúng cấu trúc này giữ thuật toán đúng với thiết kế đã được cân nhắc và dễ đối chiếu.

**Vì sao luật stacking ủy quyền TASK-CART-004 (DEC-CART-14)?** Luật xếp chồng đổi theo nước (VN cho stack 1 shop + 1 platform + freeship; MY/PH 2025 bỏ stack - §3.9, §5.7). Nếu hardcode luật VN vào optimizer thì mở nước mới phải sửa lõi thuật toán - dễ sai và vi phạm tách biệt. `optimizeCart` chỉ gọi `rules.ValidStack`; luật cụ thể nằm ở TASK-CART-004 per-country. Lõi tính toán (chọn combo max) tách khỏi luật (combo nào hợp lệ).

**Vì sao số nguyên VND (DEC-CART-15)?** Optimizer cộng nhiều phần giảm (shop + platform + freeship) rồi so với cap và so giữa các combo. Float gây sai số tích lũy - cộng ba số float lệch vài đồng đủ để chọn sai combo hoặc vượt cap. Số nguyên VND chính xác tuyệt đối; percent tính `giá * pct / 100` bằng chia nguyên (sàn không có phần trăm thập phân).

**Vì sao chooseBestShopVoucherPerShop độc lập per shop (DEC-CART-16)?** Mỗi shop có voucher riêng; một đơn nhiều shop áp một voucher shop cho mỗi shop. Chọn voucher tốt nhất cho từng shop độc lập là tối ưu cục bộ đúng (voucher shop A không ảnh hưởng shop B). Điều này giữ không gian tìm kiếm nhỏ thay vì duyệt mọi tổ hợp voucher shop.

**Vì sao applyCaps bắt buộc (DEC-CART-15, §1 #7)?** Voucher thực tế có trần giảm (`cap`): "giảm 20% tối đa 50k". Bỏ cap thì optimizer tính giảm 20% của đơn lớn ra số vượt 50k - con số người dùng thấy SAI, lúc thanh toán sàn chỉ giảm 50k. `applyCaps` kẹp đúng trần để con số khớp thực tế.

**Vì sao trả breakdown minh bạch (DEC-CART-18)?** Người dùng cần thấy "bạn tiết kiệm 110k: 30k từ voucher shop A, 50k từ platform, 30k freeship" thay vì chỉ một con số. Minh bạch khoản giảm từ đâu vừa xây niềm tin (định vị hậu-Honey) vừa giúp người dùng hiểu vì sao combo này tốt.

**Vì sao hàm thuần, caller nạp DB (DEC-CART-17)?** Tách `optimizeCart` (logic thuần, dễ test với input cố định) khỏi I/O (handler nạp giỏ + voucher). Hàm thuần test được với mọi ca biên mà không cần DB; handler chỉ ghép dữ liệu rồi gọi.

---

## §3 - Hợp đồng API / DDL

### Interface luật stacking (rules.go) - TASK-CART-004 hiện thực

```go
// services/cart/internal/optimizer/rules.go
// StackRules trừu tượng luật xếp chồng per-country; TASK-CART-004 hiện thực cụ thể.
type StackRules interface {
    // ValidStack trả true nếu tổ hợp (platform voucher, freeship, các shop voucher) được phép
    // theo luật của nước; TASK-CART-003 KHÔNG biết luật cụ thể (DEC-CART-14).
    ValidStack(pv *Voucher, fs *Voucher, shopVouchers []Voucher) bool
}
```

### Tính giảm giá số nguyên (discount.go)

```go
// services/cart/internal/optimizer/discount.go

// discountAmount tính phần giảm của một voucher trên một giá trị đơn (int64 VND).
func discountAmount(v Voucher, base int64) int64 {
    var d int64
    switch v.DiscountType {
    case DiscountAmount:
        d = v.DiscountValue
    case DiscountPercent:
        d = base * v.DiscountValue / 100 // chia nguyên, KHÔNG float (DEC-CART-15)
    }
    return applyCap(d, v.Cap)
}

// applyCap kẹp phần giảm theo trần tuyệt đối của voucher (DEC-CART-15, §1 #7).
func applyCap(d int64, cap *int64) int64 {
    if cap != nil && d > *cap {
        return *cap
    }
    return d
}

func freeshipValue(fs *Voucher, base int64) int64 {
    if fs == nil {
        return 0
    }
    return discountAmount(*fs, base)
}
```

### optimizeCart (optimize.go) - bám pseudo-code §3.5(3)

```go
// services/cart/internal/optimizer/optimize.go

// optimizeCart bám đúng cấu trúc pseudo-code §3.5(3) (DEC-CART-13):
// duyệt (platformVoucher x freeship x shopCombo), lọc validStack + meetsMinSpend, applyCaps, giữ best.
func optimizeCart(items []CartItem, vouchers Vouchers, rules StackRules) OptimizeResult {
    cartTotal := totalOf(items)
    applicableShop := filterByMinSpend(vouchers.Shop, items)       // voucher shop đạt min_spend
    best := OptimizeResult{Discount: 0, Combo: nil}

    for _, pv := range append(vouchers.Platform, nil) {            // platformVoucher + [none]
        for _, fs := range append(vouchers.Freeship, nil) {       // freeship + [none]
            shopCombo := chooseBestShopVoucherPerShop(applicableShop, items) // tốt nhất per shop
            if !rules.ValidStack(pv, fs, shopCombo) {              // luật per-country (TASK-CART-004)
                continue
            }
            if !meetsMinSpend(shopCombo, pv, fs, items) {          // ngưỡng đơn
                continue
            }
            total := sumShopDiscounts(shopCombo, items) +
                discountOf(pv, cartTotal) + freeshipValue(fs, cartTotal) // applyCaps trong từng hàm
            if total > best.Discount {
                best = OptimizeResult{
                    Discount:  total,
                    Combo:     combine(shopCombo, pv, fs),
                    Breakdown: breakdownOf(shopCombo, pv, fs, items), // minh bạch (DEC-CART-18)
                }
            }
        }
    }
    return best // giỏ rỗng / không voucher -> Discount 0, Combo nil (§1 #9)
}

// chooseBestShopVoucherPerShop chọn voucher shop tốt nhất CHO TỪNG shop (DEC-CART-16).
func chooseBestShopVoucherPerShop(shopVouchers []Voucher, items []CartItem) []Voucher {
    bestPerShop := map[string]Voucher{}
    for _, v := range shopVouchers {
        base := shopSubtotal(items, *v.ShopID)
        d := discountAmount(v, base)
        cur, ok := bestPerShop[*v.ShopID]
        if !ok || d > discountAmount(cur, shopSubtotal(items, *cur.ShopID)) {
            bestPerShop[*v.ShopID] = v
        }
    }
    return sortedValues(bestPerShop) // tất định (§1 #10)
}
```

---

## §4 - Acceptance criteria

1. `optimizeCart` là hàm thuần: grep `services/cart/internal/optimizer/**`: KHÔNG có truy vấn DB/SQL (caller nạp dữ liệu).
2. Cấu trúc bám pseudo-code §3.5(3): có `filterByMinSpend`, vòng duyệt `(platform x freeship x shopCombo)`, `validStack`, `meetsMinSpend`, `applyCaps`, giữ `best.discount`.
3. `chooseBestShopVoucherPerShop` chọn đúng một voucher tốt nhất cho mỗi shop trong giỏ (test giỏ 2 shop).
4. `validStack` gọi `rules.ValidStack` (interface TASK-CART-004); grep: optimizer KHÔNG hardcode luật nước (không có hằng VN/MY/PH).
5. Ví dụ §3.5(3): giỏ shop A (300k) + B (250k), voucher shop A "-30k đơn>=250k", platform "-50k đơn>=500k", freeship "<=30k", luật VN-stack -> tổng giảm = 110k (30k + 50k + 30k).
6. Phép tính percent dùng chia nguyên (`base * pct / 100`); grep: không float trong tính giảm/cap.
7. `applyCap` kẹp phần giảm theo `cap`: voucher "giảm 20% tối đa 50k" trên đơn 400k -> giảm 50k (không 80k).
8. `meetsMinSpend` loại voucher khi đơn không đạt ngưỡng (voucher platform "-50k đơn>=500k" không áp cho giỏ 400k).
9. Giỏ rỗng hoặc không voucher áp được -> `Discount = 0`, `Combo` rỗng, không lỗi.
10. Kết quả tất định: chạy hai lần cùng input cho cùng `combo` (sắp xếp ổn định).
11. `OptimizeResult.Breakdown` phân rã giảm theo shop/platform/freeship khớp tổng `Discount`.
12. `go test ./...` xanh.

---

## §5 - Kiểm thử (verification)

```go
// services/cart/internal/optimizer/optimize_test.go

// Ví dụ minh họa §3.5(3): VN-stack -> 110k.
func TestOptimize_VNStackExample_110k(t *testing.T) {
    items := []CartItem{
        {ShopID: ptr("A"), Qty: 1, UnitPrice: 300_000},
        {ShopID: ptr("B"), Qty: 1, UnitPrice: 250_000},
    }
    vouchers := Vouchers{
        Shop:     []Voucher{shopV("A", DiscountAmount, 30_000, ptr(int64(250_000)))}, // -30k đơn>=250k
        Platform: []Voucher{platV(DiscountAmount, 50_000, ptr(int64(500_000)))},      // -50k đơn>=500k
        Freeship: []Voucher{freeV(30_000, nil)},                                      // freeship <=30k
    }
    res := optimizeCart(items, vouchers, VNStackRules{}) // VN cho stack
    require.Equal(t, int64(110_000), res.Discount)       // 30k + 50k + 30k
}

func TestOptimize_ApplyCap_KepsDiscount(t *testing.T) {
    items := []CartItem{{ShopID: ptr("A"), Qty: 1, UnitPrice: 400_000}}
    vouchers := Vouchers{
        Shop: []Voucher{shopV("A", DiscountPercent, 20, ptr(int64(50_000)))}, // 20% tối đa 50k
    }
    res := optimizeCart(items, vouchers, VNStackRules{})
    require.Equal(t, int64(50_000), res.Discount) // 20% của 400k = 80k, kẹp về cap 50k
}

func TestOptimize_MinSpendGate_ExcludesVoucher(t *testing.T) {
    items := []CartItem{{ShopID: ptr("A"), Qty: 1, UnitPrice: 400_000}} // <500k
    vouchers := Vouchers{
        Platform: []Voucher{platV(DiscountAmount, 50_000, ptr(int64(500_000)))}, // cần >=500k
    }
    res := optimizeCart(items, vouchers, VNStackRules{})
    require.Equal(t, int64(0), res.Discount) // không đạt min_spend -> không áp
}

func TestOptimize_BestShopVoucherPerShop(t *testing.T) {
    items := []CartItem{{ShopID: ptr("A"), Qty: 1, UnitPrice: 300_000}}
    vouchers := Vouchers{Shop: []Voucher{
        shopV("A", DiscountAmount, 20_000, nil),
        shopV("A", DiscountAmount, 35_000, nil), // tốt hơn cho shop A
    }}
    combo := chooseBestShopVoucherPerShop(vouchers.Shop, items)
    require.Len(t, combo, 1)
    require.Equal(t, int64(35_000), combo[0].DiscountValue) // chọn voucher tốt hơn
}

func TestOptimize_EmptyCart_ZeroDiscount(t *testing.T) {
    res := optimizeCart(nil, Vouchers{}, VNStackRules{})
    require.Equal(t, int64(0), res.Discount)
    require.Empty(t, res.Combo)
}

func TestOptimize_Deterministic(t *testing.T) {
    items := []CartItem{{ShopID: ptr("A"), Qty: 1, UnitPrice: 300_000}}
    v := Vouchers{Shop: []Voucher{shopV("A", DiscountAmount, 30_000, nil)}}
    a := optimizeCart(items, v, VNStackRules{})
    b := optimizeCart(items, v, VNStackRules{})
    require.Equal(t, a.Combo, b.Combo) // cùng input -> cùng combo
}
```

```go
// services/cart/internal/optimizer/discount_test.go
func TestDiscountPercent_IntegerDivision(t *testing.T) {
    v := Voucher{DiscountType: DiscountPercent, DiscountValue: 15}
    require.Equal(t, int64(45_000), discountAmount(v, 300_000)) // 15% của 300k = 45k (chia nguyên)
}

func TestApplyCap(t *testing.T) {
    require.Equal(t, int64(50_000), applyCap(80_000, ptr(int64(50_000)))) // kẹp
    require.Equal(t, int64(30_000), applyCap(30_000, ptr(int64(50_000)))) // dưới cap giữ nguyên
    require.Equal(t, int64(80_000), applyCap(80_000, nil))                // không cap
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `types.go` (CartItem, Voucher, Vouchers, OptimizeResult, Breakdown) -> `rules.go` (interface StackRules, TASK-CART-004 hiện thực) -> `discount.go` (`discountAmount`, `applyCap`, `freeshipValue` - int64, percent chia nguyên, test thuần) -> `optimize.go` (`filterByMinSpend`, `chooseBestShopVoucherPerShop`, vòng duyệt bám §3.5(3), `meetsMinSpend`, `breakdownOf`) -> `optimize.go` handler (`POST /v1/cart/optimize`: nạp giỏ TASK-CART-002 + voucher TASK-CART-001, gọi optimizer thuần với rules nước) -> tests. Optimizer là package thuần, test được với input cố định gồm cả ví dụ §3.5(3). Handler ghép I/O (DB) quanh hàm thuần. Luật `VNStackRules` dùng trong test là stub đơn giản; hiện thực đầy đủ per-country ở TASK-CART-004.

---

## §7 - Phụ thuộc

- **TASK-CART-001** - `voucher_catalog` + `ListActive` cấp tập voucher còn hiệu lực (shop/platform/freeship, cap, min_spend, stack_group) cho optimizer (depends_on cứng).
- **TASK-CART-002** - `cart_snapshot` + `cart_item` cấp giỏ (qty, unit_price, shop_id) để tối ưu (depends_on cứng).
- **TASK-CART-004 (downstream)** - hiện thực `StackRules.ValidStack` per-country (VN stack vs MY/PH no-stack); optimizer gọi interface này, không hardcode luật.
- **TASK-INFRA-001 (gateway)** - JWT auth trước handler `/v1/cart/optimize`.
- Extension/lib: thuần Go (không lib ngoài cho lõi); handler dùng `net/http`, `encoding/json`, `pgx` (qua repo).

---

## §8 - Payload ví dụ

### Request POST /v1/cart/optimize

```json
{
  "platform_id": 1,
  "snapshot_id": 5012,
  "items": [
    { "shop_id": "A", "qty": 1, "unit_price": 300000 },
    { "shop_id": "B", "qty": 1, "unit_price": 250000 }
  ]
}
```

### Response (best combo, VN-stack) - breakdown minh bạch

```json
{
  "discount": 110000,
  "combo": [
    { "code": "SHOPA30K", "type": "shop", "shop_id": "A" },
    { "code": "GIAM50K", "type": "platform" },
    { "code": "FREESHIP", "type": "freeship" }
  ],
  "breakdown": { "shop": 30000, "platform": 50000, "freeship": 30000 }
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Tối ưu gom đơn để đạt min_spend (gợi ý thêm SKU) - tính năng group-buy (§6 #4), slice sau.
- Voucher điều kiện bậc thang trong optimizer - chờ TASK-CART-001 mô hình hóa bậc thang.
- Tối ưu khi số voucher platform/freeship lớn (cận biên hiệu năng) - đo trước, thêm cắt tỉa nếu cần.
- Đa tiền tệ - giữ int64 minor unit; luật stacking đã tách per-country (TASK-CART-004).

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Hardcode luật nước trong optimizer | grep hằng VN/MY | mở nước phải sửa lõi | Ủy quyền StackRules (DEC-CART-14) |
| Float trong tính giảm/cap | grep + discount test | sai số tổng giảm, chọn sai combo | Số nguyên VND (DEC-CART-15) |
| Bỏ applyCaps | optimize test cap | tổng giảm vượt trần thực tế | applyCap kẹp (§1 #7) |
| Optimizer đọc DB | grep SQL | không test được thuần | Hàm thuần, caller nạp (DEC-CART-17) |
| min_spend không kiểm | min-spend test | áp voucher không đủ ngưỡng | meetsMinSpend gate (§1 #5) |
| Chọn nhiều voucher một shop | best-per-shop test | vi phạm luật một-voucher-shop | chooseBestShopVoucherPerShop (DEC-CART-16) |
| Kết quả không tất định | deterministic test | UX/test bất ổn | Sắp xếp ổn định (§1 #10) |
| Giỏ rỗng panic | empty-cart test | crash | Trả Discount 0 (§1 #9) |
| Con số thiếu minh bạch | review breakdown | mất niềm tin | Breakdown shop/platform/freeship (DEC-CART-18) |

---

## §11 - Ghi chú

- `optimizeCart` là trái tim tối ưu giỏ - khác biệt cạnh tranh mà BeeCost còn yếu (§5.6); trả lời câu hỏi tiền-tươi của persona Linh.
- Bám đúng pseudo-code §3.5(3): duyệt (platform x freeship x shop-combo), lọc validStack + min_spend, applyCaps, giữ max.
- Luật stacking ủy quyền TASK-CART-004 qua interface - lõi tính toán (chọn combo max) tách khỏi luật (combo nào hợp lệ theo nước).
- Số nguyên VND tránh sai số khi cộng shop+platform+freeship rồi so cap/so combo; percent chia nguyên.
- applyCaps khớp con số với thực tế sàn (giảm có trần); breakdown minh bạch xây niềm tin.
- Hàm thuần test được với input cố định gồm ví dụ §3.5(3); handler ghép I/O quanh nó.

---

*Hết TASK-CART-003. Status: ready_to_implement (mục tiêu audit 10/10).*
