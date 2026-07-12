---
id: FR-CART-004
title: "Engine luật stacking per-country - VN cho stack 1 shop + 1 platform + freeship (vd 110k); MY/PH 2025 bỏ stacking, freeship gộp nhóm platform (vd 80k); hiện thực StackRules đọc CountryPolicy (FR-INFRA-005)"
module: CART
priority: MUST
status: done
verify: T
phase: P2
milestone: P2 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [FR-CART-003, FR-INFRA-005, FR-CART-001, FR-COMPLY-006]
depends_on: [FR-CART-003, FR-INFRA-005]
blocks: [FR-COMPLY-006]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §3.5(3) (VN: 1 shop + 1 platform + freeship = 110k; MY/PH 2025 bỏ stacking, freeship gộp nhóm platform = 80k)"
  - "docs/... §5.7 (SEA sequencing + per-country gating: MY & PH bỏ stacking voucher 2025), §3.9 (Shopee MY/PH 2025: tối đa 1 shop + 1 platform, freeship & cashback gộp nhóm platform)"
source_decisions:
  - "DEC-CART-19: luật stacking đọc từ CountryPolicy của FR-INFRA-005 (voucher_stacking_allowed + cách gộp nhóm) - KHÔNG hardcode if-country trong optimizer (DEC-CART-14); engine là hiện thực StackRules"
  - "DEC-CART-20: VN policy - cho stack: 1 voucher shop (mỗi shop) + 1 voucher platform + 1 freeship đồng thời (vd 30k + 50k + 30k = 110k §3.5(3))"
  - "DEC-CART-21: MY/PH policy (2025) - bỏ stacking đa nhóm: tối đa 1 shop voucher + 1 platform voucher; freeship & cashback GỘP vào nhóm platform (cùng stack_group) nên loại trừ với platform voucher - chọn max(platform, freeship) (vd max(50k,30k)+30k shop = 80k §3.5(3))"
  - "DEC-CART-22: gộp nhóm thể hiện qua stack_group (FR-CART-001 DEC-CART-04) - hai voucher cùng stack_group loại trừ lẫn nhau; MY/PH gán freeship cùng stack_group với platform; VN để khác nhóm (stack được)"
  - "DEC-CART-23: mặc định an toàn (DEC-INFRA-23) - nước chưa cấu hình -> policy hạn chế nhất (no-stack) thay vì cho stack, tránh hiển thị giảm cao hơn luật thực tế nước đó"

language: "Go 1.22 (cart-svc); hiện thực interface StackRules (FR-CART-003) đọc CountryPolicy (FR-INFRA-005)"
service: shopass/services/cart/
new_files:
  - services/cart/internal/optimizer/stacking/vn_rules.go
  - services/cart/internal/optimizer/stacking/mypph_rules.go
  - services/cart/internal/optimizer/stacking/policy_rules.go
  - services/cart/internal/optimizer/stacking/factory.go
  - services/cart/internal/optimizer/stacking/vn_rules_test.go
  - services/cart/internal/optimizer/stacking/mypph_rules_test.go
  - services/cart/internal/optimizer/stacking/factory_test.go
modified_files:
  - services/cart/internal/api/optimize.go        # chọn rules theo country của platform
allowed_tools:
  - file_read: services/cart/**
  - file_read: region/**
  - file_write: services/cart/**
  - bash: cd services/cart && go test ./...
disallowed_tools:
  - hardcode if-country trong optimizer lõi (vi phạm DEC-CART-14/19 - luật ở engine này, đọc CountryPolicy)
  - mặc định cho stack khi nước chưa cấu hình (vi phạm DEC-CART-23, hiển thị giảm cao hơn luật thực tế)
  - cho MY/PH stack freeship cùng platform (vi phạm DEC-CART-21, sai luật 2025)
  - tính giảm bằng float (vi phạm đồng nhất DEC-CART-15)

effort_hours: 6
sub_tasks:
  - "0.75h: policy_rules.go - PolicyStackRules đọc CountryPolicy (voucher_stacking_allowed, freeship_grouped_with_platform)"
  - "1.0h: vn_rules.go - VN cho stack 1 shop + 1 platform + freeship (khác stack_group)"
  - "1.5h: mypph_rules.go - MY/PH bỏ stack đa nhóm; freeship gộp stack_group platform -> loại trừ; max(platform, freeship)"
  - "0.75h: factory.go - chọn rules theo country (VN/MY/PH/...) từ CountryPolicy; mặc định no-stack (DEC-CART-23)"
  - "1.0h: vn_rules_test.go - ví dụ §3.5(3) VN -> ValidStack cho combo 110k true"
  - "1.5h: mypph_rules_test.go - ví dụ §3.5(3) MY/PH -> combo platform+freeship cùng lúc bị loại; ra 80k; default no-stack"

risk_if_skipped: "Luật stacking per-country là điều kiện bắt buộc để mở SEA mà không sai con số tiền cho người dùng: tài liệu nêu rõ MY & PH đã bỏ stacking voucher 2025 (§5.7, §3.9), trong khi VN vẫn cho stack. Nếu áp luật VN cho MY/PH thì optimizer tính ra tổng giảm 110k trong khi luật MY/PH thực tế chỉ cho 80k - hiển thị con số cao hơn thực tế, người dùng thất vọng lúc thanh toán và mất niềm tin (đụng định vị minh bạch). Ngược lại áp no-stack cho VN thì bỏ lỡ giá trị, người dùng VN thấy SănDeal kém hơn đối thủ. Đây là FR gating per-country mà §2 tài liệu nhấn mạnh là bắt buộc. Nếu hardcode if-country vào optimizer lõi thì mỗi nước mới phải sửa lõi thuật toán (FR-CART-003) thay vì thêm một policy - vi phạm tách biệt và dễ sai. Mặc định cho stack khi nước chưa cấu hình là nguy hiểm nhất: hiển thị giảm cao hơn luật thực tế ở nước chưa kiểm - phải mặc định no-stack (an toàn)."
---

## §1 - Mô tả (BCP-14 normative)

Service CART **MUST** hiện thực engine luật stacking per-country (interface `StackRules` của FR-CART-003) đọc `CountryPolicy` của FR-INFRA-005: VN cho stack 1 shop + 1 platform + freeship; MY/PH (2025) bỏ stacking đa nhóm và gộp freeship vào nhóm platform; nước chưa cấu hình mặc định no-stack. Hợp đồng:

1. **MUST** hiện thực `StackRules.ValidStack(pv, fs, shopVouchers)` (interface FR-CART-003) cho từng chính sách nước, đọc cấu hình từ `CountryPolicy` (FR-INFRA-005) - KHÔNG hardcode `if country == ...` trong optimizer lõi (DEC-CART-19).
2. Chính sách VN **MUST** cho stack đồng thời (DEC-CART-20): tối đa 1 voucher shop cho mỗi shop + 1 voucher platform + 1 voucher freeship. `ValidStack` trả `true` cho tổ hợp này khi các voucher khác `stack_group`.
3. Chính sách MY/PH (2025) **MUST** bỏ stacking đa nhóm (DEC-CART-21): tối đa 1 voucher shop + 1 voucher platform; freeship (và cashback) GỘP vào nhóm platform nên loại trừ với voucher platform - không được dùng platform voucher VÀ freeship cùng lúc. Optimizer chỉ chọn `max(platform, freeship)` cộng voucher shop.
4. Việc gộp nhóm **MUST** thể hiện qua `stack_group` (DEC-CART-22, FR-CART-001 DEC-CART-04): voucher cùng `stack_group` loại trừ lẫn nhau. Với MY/PH, freeship mang cùng `stack_group` với platform; với VN, freeship khác nhóm (stack được). `ValidStack` từ chối tổ hợp có hai voucher cùng `stack_group`.
5. Engine **MUST** đọc cờ từ `CountryPolicy`: tối thiểu `voucher_stacking_allowed` (FR-INFRA-005 DEC-INFRA-22) và quy tắc gộp freeship-platform; KHÔNG nhúng luật cứng tách rời cấu hình nước.
6. `factory` **MUST** chọn `StackRules` đúng theo country của thao tác (suy từ `platform.country`, FR-INFRA-005 DEC-INFRA-25): VN -> VN rules; MY/PH -> MY/PH rules; nước khác -> theo `CountryPolicy` của nước đó.
7. Nước chưa cấu hình `CountryPolicy` **MUST** mặc định policy hạn chế nhất (no-stack) (DEC-CART-23, FR-INFRA-005 DEC-INFRA-23) - tối đa một voucher mỗi loại, không stack đa nhóm; KHÔNG mặc định cho stack.
8. `ValidStack` **MUST** từ chối tổ hợp vượt số voucher cho phép theo policy: VN cho (shop-per-shop + platform + freeship); MY/PH cho (shop-per-shop + một-trong-{platform, freeship}).
9. Engine **MUST NOT** tự tính giá trị giảm (đó là việc của FR-CART-003 `discount.go`); engine CHỈ quyết tổ hợp nào hợp lệ (`ValidStack` trả bool). Tách quyết-định-hợp-lệ khỏi tính-giá-trị.
10. Engine **MUST** đảm bảo kết quả per-country khớp ví dụ §3.5(3): cùng giỏ + voucher, VN rules cho phép combo ra 110k; MY/PH rules loại tổ hợp platform+freeship đồng thời nên combo tối ưu ra 80k.
11. Mọi quyết định **MUST** tất định theo (policy, voucher set) - cùng input cho cùng kết quả `ValidStack`.
12. **SHOULD** phát OTel `stack_rules_country{country}` và `stack_rejected_total{country, reason}` để theo dõi luật áp theo nước và tổ hợp bị loại.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao luật đọc CountryPolicy, không hardcode (DEC-CART-19)?** Tài liệu nhấn mạnh per-country gating là bắt buộc (§2): luật voucher khác theo nước và đổi theo thời gian (MY/PH bỏ stack 2025). Nếu hardcode `if country == "VN"` rải khắp thì mở nước mới hoặc đổi luật phải sửa code khắp nơi - dễ sai. Đọc `CountryPolicy` (dữ liệu, FR-INFRA-005) tách luật khỏi code: thêm nước là thêm cấu hình, không sửa lõi.

**Vì sao VN cho stack, MY/PH không (DEC-CART-20/21)?** Đây là sự thật thị trường: §3.9 ghi Shopee MY/PH 2025 cho tối đa 1 shop + 1 platform, freeship & cashback gộp nhóm platform; VN vẫn cho stack 1 shop + 1 platform + freeship. Áp sai luật là hiển thị con số sai - cao hơn thực tế ở MY/PH (người dùng thất vọng lúc thanh toán) hoặc thấp hơn ở VN (kém đối thủ). Mỗi nước một policy đúng với luật của nó.

**Vì sao gộp nhóm qua stack_group (DEC-CART-22)?** stack_group (FR-CART-001) là nhãn loại trừ: voucher cùng nhóm không dùng cùng nhau. MY/PH gộp freeship vào nhóm platform = gán cùng stack_group = loại trừ. VN để freeship khác nhóm = stack được. Cùng một cơ chế (stack_group loại trừ) thể hiện cả hai luật chỉ bằng cách gán nhãn khác nhau - không cần logic riêng cho từng nước.

**Vì sao mặc định no-stack cho nước chưa cấu hình (DEC-CART-23)?** Mặc định an toàn (FR-INFRA-005 DEC-INFRA-23): nếu một nước chưa được kiểm luật, cho stack là rủi ro - hiển thị giảm cao hơn luật thực tế nước đó, sai con số tiền và lệch compliance. No-stack (chỉ một voucher mỗi loại) là phía an toàn: thà hiển thị giảm thấp hơn rồi mở dần khi xác minh luật, còn hơn hứa quá rồi sai.

**Vì sao engine chỉ quyết hợp lệ, không tính giá trị (§1 #9)?** Tách bạch: FR-CART-003 `discount.go` tính giá trị giảm (int64, applyCaps); engine này chỉ trả `ValidStack` bool (tổ hợp này có được phép không). Hai trách nhiệm riêng dễ test riêng và không lẫn. Optimizer ghép: với mỗi tổ hợp hợp lệ (`ValidStack`), tính giá trị, giữ max.

**Vì sao khớp ví dụ §3.5(3) (§1 #10)?** Tài liệu cho hai con số kiểm chứng: VN 110k, MY/PH 80k trên cùng giỏ. Đây là test vàng cho engine: VN rules phải cho combo ra 110k; MY/PH rules phải loại tổ hợp platform+freeship đồng thời nên combo tối ưu ra 80k. Khớp hai con số này là bằng chứng luật per-country đúng.

---

## §3 - Hợp đồng API / DDL

### Policy rules đọc CountryPolicy (policy_rules.go)

```go
// services/cart/internal/optimizer/stacking/policy_rules.go
// PolicyStackRules đọc CountryPolicy (FR-INFRA-005) thay vì hardcode luật (DEC-CART-19).
type PolicyStackRules struct {
    StackingAllowed       bool // CountryPolicy.voucher_stacking_allowed
    FreeshipGroupedWithPF bool // freeship gộp nhóm platform (MY/PH 2025)
}

func (r PolicyStackRules) ValidStack(pv *Voucher, fs *Voucher, shopVouchers []Voucher) bool {
    // một shop tối đa một voucher shop (đã đảm bảo bởi chooseBestShopVoucherPerShop)
    if !r.StackingAllowed {
        // no-stack: tối đa một voucher ngoài shop (platform XOR freeship)
        if pv != nil && fs != nil {
            return false
        }
    }
    if r.FreeshipGroupedWithPF && pv != nil && fs != nil {
        return false // MY/PH: freeship gộp nhóm platform -> loại trừ (DEC-CART-21)
    }
    return !hasSameStackGroupConflict(pv, fs, shopVouchers) // stack_group loại trừ (DEC-CART-22)
}
```

### VN và MY/PH rules (qua factory)

```go
// services/cart/internal/optimizer/stacking/vn_rules.go
// VN: cho stack 1 shop + 1 platform + freeship (DEC-CART-20).
func newVNRules() StackRules {
    return PolicyStackRules{StackingAllowed: true, FreeshipGroupedWithPF: false}
}

// services/cart/internal/optimizer/stacking/mypph_rules.go
// MY/PH 2025: bỏ stack đa nhóm; freeship gộp nhóm platform (DEC-CART-21).
func newMYPHRules() StackRules {
    return PolicyStackRules{StackingAllowed: true, FreeshipGroupedWithPF: true}
    // StackingAllowed=true cho phép shop + một-trong-{platform,freeship};
    // FreeshipGroupedWithPF=true loại trừ platform & freeship dùng cùng lúc.
}
```

### Factory chọn rules theo country (factory.go)

```go
// services/cart/internal/optimizer/stacking/factory.go
import "shopass/region" // CountryPolicy (FR-INFRA-005)

// RulesForCountry chọn StackRules theo policy của nước; mặc định no-stack (DEC-CART-23).
func RulesForCountry(country string, policy region.CountryPolicy) StackRules {
    switch country {
    case "VN":
        return newVNRules()
    case "MY", "PH":
        return newMYPHRules()
    default:
        // nước khác: đọc CountryPolicy; chưa cấu hình -> hạn chế nhất (no-stack)
        return PolicyStackRules{
            StackingAllowed:       policy.VoucherStackingAllowed, // mặc định false nếu chưa set
            FreeshipGroupedWithPF: !policy.VoucherStackingAllowed,
        }
    }
}
```

---

## §4 - Acceptance criteria

1. `ValidStack` đọc cờ từ `CountryPolicy`/policy struct; grep `optimizer/stacking/**`: KHÔNG có chuỗi country hardcode trong logic `ValidStack` (factory map country -> rules là chấp nhận; logic luật không if-country).
2. VN rules: `ValidStack` trả `true` cho tổ hợp (shop-per-shop + platform + freeship) khi khác stack_group.
3. MY/PH rules: `ValidStack` trả `false` khi cả `pv` và `fs` khác nil (platform + freeship đồng thời) - freeship gộp nhóm platform.
4. MY/PH rules: `ValidStack` trả `true` cho (shop + platform) hoặc (shop + freeship), không cả hai.
5. stack_group loại trừ: hai voucher cùng `stack_group` -> `ValidStack` trả `false` (cả VN lẫn MY/PH).
6. `factory.RulesForCountry("VN", ...)` trả VN rules; `("MY"|"PH", ...)` trả MY/PH rules.
7. Nước chưa cấu hình (policy mặc định) -> no-stack (StackingAllowed false) - `ValidStack` từ chối platform + freeship đồng thời.
8. Ví dụ §3.5(3) đầu-cuối: với optimizer (FR-CART-003) + VN rules, combo tối ưu ra 110k; với MY/PH rules, combo tối ưu ra 80k (platform+freeship bị loại, chọn max).
9. Engine không tính giá trị giảm (chỉ trả bool); grep: không có phép nhân/cộng tiền trong `ValidStack`.
10. Kết quả tất định: cùng (policy, voucher set) cho cùng `ValidStack`.
11. `go test ./...` xanh.

---

## §5 - Kiểm thử (verification)

```go
// services/cart/internal/optimizer/stacking/vn_rules_test.go
func TestVN_AllowsStack_ShopPlatformFreeship(t *testing.T) {
    r := newVNRules()
    pv := platV(DiscountAmount, 50_000, ptr(int64(500_000)), "platform-grp")
    fs := freeV(30_000, "freeship-grp")  // khác stack_group platform
    shop := []Voucher{shopV("A", DiscountAmount, 30_000, "shop-A-grp")}
    require.True(t, r.ValidStack(&pv, &fs, shop)) // VN cho stack cả ba (DEC-CART-20)
}

// Ví dụ §3.5(3) VN -> 110k (qua optimizer FR-CART-003).
func TestVN_OptimizeExample_110k(t *testing.T) {
    items := exampleCart() // shop A 300k, shop B 250k
    v := exampleVouchers() // shop A -30k, platform -50k>=500k, freeship 30k
    res := optimizeCart(items, v, newVNRules())
    require.Equal(t, int64(110_000), res.Discount)
}
```

```go
// services/cart/internal/optimizer/stacking/mypph_rules_test.go
func TestMYPH_RejectsPlatformPlusFreeship(t *testing.T) {
    r := newMYPHRules()
    pv := platV(DiscountAmount, 50_000, ptr(int64(500_000)), "platform-grp")
    fs := freeV(30_000, "platform-grp") // MY/PH: freeship cùng nhóm platform
    shop := []Voucher{shopV("A", DiscountAmount, 30_000, "shop-A-grp")}
    require.False(t, r.ValidStack(&pv, &fs, shop)) // không stack platform + freeship (DEC-CART-21)
}

func TestMYPH_AllowsShopPlusOne(t *testing.T) {
    r := newMYPHRules()
    pv := platV(DiscountAmount, 50_000, ptr(int64(500_000)), "platform-grp")
    shop := []Voucher{shopV("A", DiscountAmount, 30_000, "shop-A-grp")}
    require.True(t, r.ValidStack(&pv, nil, shop)) // shop + platform OK
}

// Ví dụ §3.5(3) MY/PH -> 80k (platform+freeship bị loại, chọn max(50k,30k)+30k shop).
func TestMYPH_OptimizeExample_80k(t *testing.T) {
    items := exampleCart()
    v := exampleVouchersMYPH() // freeship gán cùng stack_group platform
    res := optimizeCart(items, v, newMYPHRules())
    require.Equal(t, int64(80_000), res.Discount) // max(50k, 30k) + 30k shop = 80k
}
```

```go
// services/cart/internal/optimizer/stacking/factory_test.go
func TestFactory_DefaultsToNoStack(t *testing.T) {
    r := RulesForCountry("XX", region.CountryPolicy{}) // nước chưa cấu hình
    pv := platV(DiscountAmount, 50_000, nil, "p")
    fs := freeV(30_000, "f")
    require.False(t, r.ValidStack(&pv, &fs, nil)) // mặc định no-stack (DEC-CART-23)
}

func TestFactory_SelectsByCountry(t *testing.T) {
    require.IsType(t, newVNRules(), RulesForCountry("VN", region.CountryPolicy{}))
    require.IsType(t, newMYPHRules(), RulesForCountry("MY", region.CountryPolicy{}))
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: `policy_rules.go` (`PolicyStackRules` đọc cờ stacking + gộp nhóm) -> `vn_rules.go` + `mypph_rules.go` (cấu hình policy cho VN và MY/PH) -> `factory.go` (`RulesForCountry` chọn theo country, mặc định no-stack) -> cập nhật handler `optimize.go` (suy country từ `platform.country` rồi chọn rules) -> tests. Engine hiện thực interface `StackRules` của FR-CART-003 - optimizer lõi gọi `ValidStack`, không biết luật cụ thể. `CountryPolicy` đến từ FR-INFRA-005 (region package); country suy từ `platform.country` (DEC-INFRA-25). Test vàng là hai ví dụ §3.5(3): VN ra 110k, MY/PH ra 80k - chạy qua optimizer thật (FR-CART-003) với rules tương ứng.

---

## §7 - Phụ thuộc

- **FR-CART-003** - định nghĩa interface `StackRules` mà engine này hiện thực; optimizer lõi gọi `ValidStack` (depends_on cứng).
- **FR-INFRA-005** - cung cấp `CountryPolicy` (`voucher_stacking_allowed`, quy tắc gộp) + cách suy country từ `platform.country`; engine đọc policy thay vì hardcode (depends_on cứng).
- **FR-CART-001** - `stack_group` trên voucher là cơ chế thể hiện gộp nhóm/loại trừ mà engine đọc.
- **FR-COMPLY-006 (đồng hướng, P3)** - khung per-country gating tổng quát; luật stacking là một mặt của gating này.
- Extension/lib: thuần Go; đọc `region.CountryPolicy`.

---

## §8 - Payload ví dụ

### Cùng giỏ, khác nước -> khác combo (qua /v1/cart/optimize)

VN (platform.country = VN):

```json
{ "discount": 110000, "breakdown": { "shop": 30000, "platform": 50000, "freeship": 30000 } }
```

MY (platform.country = MY, freeship gộp nhóm platform):

```json
{ "discount": 80000, "breakdown": { "shop": 30000, "platform": 50000, "freeship": 0 } }
```

### CountryPolicy đọc từ FR-INFRA-005 (minh họa)

```yaml
# region/config/countries.yaml (FR-INFRA-005)
VN: { voucher_stacking_allowed: true,  freeship_grouped_with_platform: false }
MY: { voucher_stacking_allowed: true,  freeship_grouped_with_platform: true }
PH: { voucher_stacking_allowed: true,  freeship_grouped_with_platform: true }
# nước chưa liệt kê -> mặc định no-stack (DEC-CART-23)
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Luật stacking các nước SEA còn lại (ID/TH/SG/TW) - thêm policy khi mở từng nước (FR-COMPLY-007), mặc định no-stack tới khi xác minh.
- Cashback gộp nhóm (MY/PH gộp cả cashback nhóm platform) - mô hình khi có FR cashback (FR-AFFIL-005).
- Luật giới hạn số voucher shop trên một đơn nếu sàn thay đổi - thêm cờ vào CountryPolicy.
- Theo dõi thay đổi luật theo thời gian (versioned policy) - bám config versioned của FR-INFRA-005.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Hardcode if-country trong optimizer | grep + factory | mở nước phải sửa lõi | Đọc CountryPolicy (DEC-CART-19) |
| Áp luật VN cho MY/PH | mypph test 80k | hiển thị 110k sai thực tế | MY/PH rules riêng (DEC-CART-21) |
| MY/PH stack platform + freeship | RejectsPlatformPlusFreeship | con số cao hơn luật 2025 | Loại trừ qua FreeshipGroupedWithPF |
| Mặc định cho stack nước mới | factory default test | giảm cao hơn luật thực tế | Mặc định no-stack (DEC-CART-23) |
| stack_group không loại trừ | same-group test | combo sai | hasSameStackGroupConflict (DEC-CART-22) |
| Engine tự tính tiền | grep nhân/cộng | lẫn trách nhiệm | Chỉ trả bool (§1 #9) |
| Kết quả không tất định | determinism | test/UX bất ổn | Tất định theo policy (§1 #11) |
| Sai country suy ra | review factory | áp nhầm luật | Suy từ platform.country (DEC-INFRA-25) |
| Áp no-stack cho VN | vn test 110k | kém đối thủ | VN rules cho stack (DEC-CART-20) |

---

## §11 - Ghi chú

- Luật stacking per-country là gating bắt buộc để mở SEA đúng con số: VN cho stack (110k), MY/PH 2025 bỏ stack (80k) - §3.5(3), §5.7, §3.9.
- Engine đọc CountryPolicy (FR-INFRA-005) thay vì hardcode - thêm nước là thêm cấu hình, không sửa lõi optimizer.
- Gộp nhóm thể hiện qua stack_group (cùng cơ chế loại trừ): MY/PH gán freeship cùng nhóm platform; VN để khác nhóm.
- Mặc định no-stack cho nước chưa cấu hình là phía an toàn - thà giảm thấp hơn rồi mở dần khi xác minh luật.
- Engine chỉ quyết tổ hợp hợp lệ (bool); tính giá trị giảm là việc của FR-CART-003 - hai trách nhiệm tách riêng.
- Test vàng là hai con số §3.5(3): VN 110k, MY/PH 80k qua optimizer thật - bằng chứng luật per-country đúng.

---

*Hết FR-CART-004. Status: ready_to_implement (mục tiêu audit 10/10).*
