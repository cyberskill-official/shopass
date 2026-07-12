---
id: FR-INFRA-005
title: "Per-country region config - gating flags theo nước (VN/ID/TH/PH/MY/SG/TW): voucher stacking, affiliate channel allowed, data residency; config loader + feature flags"
module: INFRA
priority: MUST
status: done
verify: T
phase: P0
milestone: P0 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-INFRA-002, FR-CART-004, FR-AFFIL-004, FR-COMPLY-006, FR-COMPLY-007]
depends_on: [FR-INFRA-002]
blocks: [FR-CART-004, FR-COMPLY-006]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §2 (per-country gating bắt buộc - luật voucher/affiliate/dữ liệu khác theo nước)"
  - "docs/... §5.7 (SEA sequencing VN->ID/TH->PH/MY->SG/TW), §3.9 (MY/PH bỏ stacking 2025)"
source_decisions:
  - "DEC-INFRA-21: cấu hình per-country là dữ liệu (loader + flag), KHÔNG rải if-country khắp code; mỗi nước có một CountryPolicy"
  - "DEC-INFRA-22: CountryPolicy mang voucher_stacking_allowed, affiliate_channels_allowed, data_residency_region (tối thiểu)"
  - "DEC-INFRA-23: mặc định an toàn - nước chưa cấu hình -> policy hạn chế nhất (no-stack, không bật affiliate, residency mặc định) thay vì mở"
  - "DEC-INFRA-24: feature flag gắn scope theo country; truy vấn flag luôn kèm country context"
  - "DEC-INFRA-25: country của một thao tác suy ra từ platform.country (FR-INFRA-002) hoặc locale người dùng; không đoán từ IP một mình"

language: "Go 1.22 (shared region package); cấu hình YAML/JSON versioned trong repo + override runtime"
service: shopass/region/
new_files:
  - region/policy.go
  - region/loader.go
  - region/flags.go
  - region/config/countries.yaml
  - region/policy_test.go
  - region/loader_test.go
  - region/flags_test.go
modified_files: []
allowed_tools:
  - file_read: region/**
  - file_write: region/**
  - bash: cd region && go test ./...
disallowed_tools:
  - rải if country=='MY' khắp code nghiệp vụ (vi phạm DEC-INFRA-21; phải qua CountryPolicy)
  - mặc định mở (bật stacking/affiliate) cho nước chưa cấu hình (vi phạm DEC-INFRA-23)
  - đoán country chỉ từ IP (vi phạm DEC-INFRA-25)

effort_hours: 6
sub_tasks:
  - "1.0h: policy.go - kiểu CountryPolicy + enum trục (stacking, affiliate channel, residency)"
  - "1.0h: loader.go - đọc countries.yaml, validate, trả map country->policy; mặc định an toàn cho nước thiếu"
  - "1.0h: flags.go - feature flag có scope country; Lookup(flag, country) bool"
  - "0.5h: countries.yaml - 7 nước VN/ID/TH/PH/MY/SG/TW với policy đúng (VN stack, MY/PH no-stack)"
  - "1.0h: policy_test.go - VN cho stack, MY/PH không; nước lạ -> policy hạn chế nhất"
  - "0.5h: loader_test.go - YAML hỏng -> lỗi; thiếu nước -> default an toàn"
  - "1.0h: flags_test.go - flag bật ở nước A tắt ở nước B; truy vấn không kèm country -> lỗi/false"

risk_if_skipped: "Luật voucher, affiliate và dữ liệu khác nhau theo nước là ràng buộc cứng (§2). MY/PH đã bỏ stacking voucher 2025 - nếu optimizer (FR-CART-003) áp luật VN cho MY/PH thì gợi ý sai luật, làm hỏng giỏ của người dùng và có thể vi phạm điều khoản sàn. Affiliate channel 'browser extension' bị cấm ở một số nước (ví dụ Shopee Indonesia) - bật nhầm là vi phạm ToS, bị reject affiliate. Data residency khác theo nước (Indonesia PDP, Thailand PDPA) - sai là vi phạm luật. Không tập trung gating thì các luật này rải khắp code, dễ sai và khó kiểm. Đây là nền cho mở SEA an toàn."
---

## §1 - Mô tả (BCP-14 normative)

Service INFRA **MUST** cung cấp một lớp cấu hình per-country: mỗi nước có một CountryPolicy dữ liệu hóa, một loader, và feature flag có scope theo nước, để luật voucher/affiliate/dữ liệu được tập trung thay vì rải khắp code. Hợp đồng:

1. Cấu hình per-country **MUST** là dữ liệu, không phải `if country == ...` rải rác (DEC-INFRA-21). Mỗi nước ánh xạ tới một `CountryPolicy`; code nghiệp vụ đọc policy, không hard-code điều kiện nước.
2. `CountryPolicy` **MUST** mang tối thiểu (DEC-INFRA-22): `voucher_stacking_allowed bool`, `affiliate_channels_allowed []Channel`, `data_residency_region string`.
3. Loader **MUST** đọc cấu hình 7 nước VN/ID/TH/PH/MY/SG/TW từ file versioned (`countries.yaml`), validate cấu trúc, và trả map `country -> CountryPolicy`.
4. Cấu hình **MUST** phản ánh luật thực: `VN.voucher_stacking_allowed = true`; `MY.voucher_stacking_allowed = false` và `PH.voucher_stacking_allowed = false` (đã bỏ stacking 2025, §3.9).
5. Với nước CHƯA cấu hình, loader **MUST** trả policy hạn chế nhất (DEC-INFRA-23): `voucher_stacking_allowed = false`, `affiliate_channels_allowed = []` (rỗng), `data_residency_region` mặc định an toàn. KHÔNG mặc định mở.
6. Feature flag **MUST** có scope theo nước (DEC-INFRA-24): `Lookup(flag, country) bool`. Một flag có thể bật ở nước này, tắt ở nước khác.
7. Mọi truy vấn flag/policy **MUST** kèm country context (DEC-INFRA-25). API KHÔNG cho truy vấn flag mà thiếu country (tránh lấy nhầm mặc định toàn cục).
8. Country của một thao tác **MUST** suy ra từ `platform.country` (FR-INFRA-002) hoặc locale người dùng, KHÔNG đoán chỉ từ IP (DEC-INFRA-25). IP chỉ là tín hiệu phụ.
9. `affiliate_channels_allowed` **MUST** dùng enum kênh rõ ràng (ví dụ `web`, `extension`, `coupon`, `app`); FR-AFFIL-004 đọc đây để quyết định kênh nào hợp lệ ở nước hiện tại.
10. Loader **MUST** validate: country là ISO-3166 alpha-2; trùng country trong file là lỗi; channel ngoài enum là lỗi. File hỏng -> lỗi rõ ràng lúc khởi tạo (fail-fast), KHÔNG chạy với cấu hình mơ hồ.
11. Lớp region **SHOULD** hỗ trợ override runtime (ví dụ bật/tắt một flag cho một nước mà không deploy lại) qua một nguồn override (KV/Redis), với file là mặc định nền.
12. Lớp region **MUST** là điểm tham chiếu duy nhất cho FR-CART-004 (stacking), FR-AFFIL-004 (channel), FR-COMPLY-006/007 (residency) - một nguồn sự thật per-country.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

Vì sao gating là dữ liệu, không phải if-country (DEC-INFRA-21)? SănDeal mở 7 thị trường với luật khác nhau về voucher, affiliate và dữ liệu. Nếu mỗi luật là một `if country == 'MY'` rải trong optimizer, affiliate, compliance, thì thêm một nước hay đổi một luật là sửa hàng chục chỗ, dễ sót. Dữ liệu hóa thành CountryPolicy gom luật vào một nơi: thêm nước là thêm một mục cấu hình.

Vì sao mặc định hạn chế nhất (DEC-INFRA-23)? Khi mở một nước mới mà chưa kịp nghiên cứu luật, mặc định an toàn là tắt: không stack voucher, không bật affiliate, residency thận trọng. Mặc định mở là mời rủi ro pháp lý (gợi ý stacking trái luật, bật kênh affiliate bị cấm). Tắt mặc định rồi bật có chủ đích khi đã xác nhận luật là tư thế an toàn.

Vì sao stacking khác theo nước (§1 #4)? Đây là khác biệt luật cụ thể và đắt nếu sai: VN cho phép stack 1 shop + 1 platform + freeship; MY/PH đã bỏ stacking 2025. Optimizer áp luật VN cho MY/PH sẽ gợi ý tổ hợp voucher không áp được, làm hỏng giỏ và mất niềm tin. Policy per-country cho optimizer biết luật đúng của nước đang tính.

Vì sao truy vấn flag bắt buộc kèm country (§1 #7)? Một lỗi tinh vi là quên truyền country và vô tình lấy mặc định toàn cục. Bắt buộc country trong chữ ký `Lookup(flag, country)` biến lỗi này thành lỗi biên dịch/runtime rõ ràng thay vì hành vi sai âm thầm.

Vì sao country từ platform/locale, không từ IP (DEC-INFRA-25)? IP dễ sai: VPN, du lịch, proxy. Một người dùng VN dùng VPN không nên đột nhiên chịu luật nước khác. `platform.country` (sàn nào, nước nào) và locale người dùng là tín hiệu chủ đích, ổn định hơn. IP chỉ bổ trợ, không quyết một mình.

Vì sao một nguồn sự thật per-country (§1 #12)? CART, AFFIL, COMPLY đều cần biết luật nước. Nếu mỗi module giữ bản sao luật riêng, chúng sẽ lệch nhau theo thời gian. Tập trung vào lớp region cho một nơi cập nhật, mọi module đọc cùng sự thật.

---

## §3 - Hợp đồng API / DDL

### CountryPolicy + enum (Go)

```go
// region/policy.go
type Channel string
const (
    ChannelWeb       Channel = "web"
    ChannelExtension Channel = "extension"
    ChannelCoupon    Channel = "coupon"
    ChannelApp       Channel = "app"
)

type CountryPolicy struct {
    Country               string    // ISO-3166 alpha-2
    VoucherStackingAllowed bool
    AffiliateChannelsAllowed []Channel
    DataResidencyRegion   string    // ví dụ "ap-southeast-1"
}

// restrictivePolicy là mặc định an toàn cho nước chưa cấu hình (DEC-INFRA-23).
func restrictivePolicy(country string) CountryPolicy {
    return CountryPolicy{
        Country:               country,
        VoucherStackingAllowed: false,
        AffiliateChannelsAllowed: nil, // rỗng = không bật kênh nào
        DataResidencyRegion:   defaultResidency,
    }
}
```

### Loader (§1 #3, #5, #10)

```go
// region/loader.go
type Registry struct{ byCountry map[string]CountryPolicy }

func Load(path string) (*Registry, error) {
    raw, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var doc struct{ Countries []CountryPolicy `yaml:"countries"` }
    if err := yaml.Unmarshal(raw, &doc); err != nil { return nil, err }
    reg := &Registry{byCountry: map[string]CountryPolicy{}}
    for _, p := range doc.Countries {
        if !isAlpha2(p.Country) { return nil, fmt.Errorf("country không hợp lệ: %q", p.Country) }
        if _, dup := reg.byCountry[p.Country]; dup {
            return nil, fmt.Errorf("trùng country: %s", p.Country) // §1 #10
        }
        if err := validateChannels(p.AffiliateChannelsAllowed); err != nil { return nil, err }
        reg.byCountry[p.Country] = p
    }
    return reg, nil
}

// Policy trả policy của nước; nước chưa cấu hình → mặc định hạn chế nhất.
func (r *Registry) Policy(country string) CountryPolicy {
    if p, ok := r.byCountry[country]; ok { return p }
    return restrictivePolicy(country) // §1 #5
}
```

### Feature flag scope-country (§1 #6, #7)

```go
// region/flags.go
// Lookup luôn cần country; không có overload thiếu country (DEC-INFRA-24, §1 #7).
func (r *Registry) Lookup(flag string, country string) bool {
    if country == "" { return false } // thiếu country → an toàn = false
    return r.flagFor(flag, country) // override runtime > mặc định file
}
```

### countries.yaml (trích)

```yaml
# region/config/countries.yaml
countries:
  - country: VN
    voucherStackingAllowed: true
    affiliateChannelsAllowed: [web, extension, app]
    dataResidencyRegion: ap-southeast-1
  - country: MY
    voucherStackingAllowed: false      # bỏ stacking 2025
    affiliateChannelsAllowed: [web, app]
    dataResidencyRegion: ap-southeast-1
  - country: PH
    voucherStackingAllowed: false      # bỏ stacking 2025
    affiliateChannelsAllowed: [web, app]
    dataResidencyRegion: ap-southeast-1
  - country: ID
    voucherStackingAllowed: true
    affiliateChannelsAllowed: [web, app]  # Shopee ID hạn chế kênh extension/coupon
    dataResidencyRegion: ap-southeast-3   # Indonesia PDP — residency riêng
```

---

## §4 - Acceptance criteria

1. `Load("countries.yaml")` hợp lệ -> trả Registry với >=4 nước cấu hình (VN/MY/PH/ID tối thiểu).
2. `Policy("VN").VoucherStackingAllowed` = `true`.
3. `Policy("MY").VoucherStackingAllowed` = `false`; `Policy("PH").VoucherStackingAllowed` = `false`.
4. `Policy("XX")` (nước chưa cấu hình) -> policy hạn chế nhất: stacking false, channels rỗng.
5. `Policy("VN").AffiliateChannelsAllowed` chứa `extension`; `Policy("ID")` KHÔNG chứa `extension` (Shopee ID hạn chế).
6. `Lookup("real_sale_v2", "VN")` và `Lookup("real_sale_v2", "MY")` có thể khác nhau (flag scope country).
7. `Lookup(flag, "")` (thiếu country) -> `false` (an toàn).
8. YAML hỏng (cú pháp sai) -> `Load` trả lỗi (fail-fast).
9. File có trùng country (hai mục `VN`) -> `Load` trả lỗi.
10. File có channel ngoài enum (`telegram`) -> `Load` trả lỗi.
11. File có country không phải alpha-2 (`Vietnam`) -> `Load` trả lỗi.
12. `Policy("ID").DataResidencyRegion` khác `Policy("VN")` (residency per-country phản ánh Indonesia PDP).

---

## §5 - Kiểm thử (verification)

```go
// region/policy_test.go
func TestPolicy_VNStacks_MYPHDoNot(t *testing.T) {
    reg := mustLoad(t, "config/countries.yaml")
    require.True(t,  reg.Policy("VN").VoucherStackingAllowed)
    require.False(t, reg.Policy("MY").VoucherStackingAllowed)
    require.False(t, reg.Policy("PH").VoucherStackingAllowed)
}

func TestPolicy_UnknownCountry_Restrictive(t *testing.T) {
    reg := mustLoad(t, "config/countries.yaml")
    p := reg.Policy("XX")
    require.False(t, p.VoucherStackingAllowed)
    require.Empty(t, p.AffiliateChannelsAllowed) // không bật kênh nào
}

func TestPolicy_AffiliateChannel_PerCountry(t *testing.T) {
    reg := mustLoad(t, "config/countries.yaml")
    require.Contains(t,    reg.Policy("VN").AffiliateChannelsAllowed, ChannelExtension)
    require.NotContains(t, reg.Policy("ID").AffiliateChannelsAllowed, ChannelExtension)
}

// region/loader_test.go
func TestLoad_DuplicateCountry_Errors(t *testing.T) {
    _, err := Load(writeTemp(t, dupCountryYAML))
    require.Error(t, err)
}

func TestLoad_BadChannel_Errors(t *testing.T) {
    _, err := Load(writeTemp(t, badChannelYAML)) // channel "telegram"
    require.Error(t, err)
}

func TestLoad_NonAlpha2_Errors(t *testing.T) {
    _, err := Load(writeTemp(t, `countries: [{country: Vietnam}]`))
    require.Error(t, err)
}

// region/flags_test.go
func TestLookup_PerCountry(t *testing.T) {
    reg := regWithFlag(t, "real_sale_v2", map[string]bool{"VN": true, "MY": false})
    require.True(t,  reg.Lookup("real_sale_v2", "VN"))
    require.False(t, reg.Lookup("real_sale_v2", "MY"))
}

func TestLookup_MissingCountry_False(t *testing.T) {
    reg := mustLoad(t, "config/countries.yaml")
    require.False(t, reg.Lookup("any_flag", "")) // thiếu country = false
}
```

---

## §6 - Khung triển khai

Xem §3. Loader đọc `countries.yaml` lúc khởi tạo, fail-fast nếu hỏng. `Policy(country)` trả mặc định hạn chế nhất cho nước thiếu. Flag scope-country với override runtime (KV/Redis) chồng lên mặc định file: thay đổi nóng một flag cho một nước không cần deploy. Country của thao tác lấy từ `platform.country` hoặc locale ở tầng gọi; lớp region nhận country làm tham số, không tự đoán. Mọi module per-country (CART/AFFIL/COMPLY) đọc qua lớp này, không giữ bản sao luật riêng.

---

## §7 - Phụ thuộc

- FR-INFRA-002 - `platform.country` là nguồn suy ra country của thao tác.
- FR-CART-004 (downstream) - đọc `VoucherStackingAllowed` cho engine stacking per-country.
- FR-AFFIL-004 (downstream) - đọc `AffiliateChannelsAllowed` để chặn kênh bị cấm theo nước.
- FR-COMPLY-006 / FR-COMPLY-007 (downstream) - đọc `DataResidencyRegion` cho gating dữ liệu SEA.
- Hạ tầng: file YAML versioned + nguồn override runtime (KV/Redis) tuỳ chọn.

---

## §8 - Payload ví dụ

### Optimizer hỏi luật stacking của nước hiện tại (nội bộ)

```go
pol := region.Policy(platform.Country) // ví dụ "MY"
if !pol.VoucherStackingAllowed {
    // MY/PH: chỉ chọn max(platform, freeship) thay vì stack (FR-CART-004)
}
```

### Affiliate kiểm kênh hợp lệ theo nước

```go
pol := region.Policy(country)
if !slices.Contains(pol.AffiliateChannelsAllowed, region.ChannelExtension) {
    return ErrChannelNotAllowed // ví dụ Shopee Indonesia cấm kênh extension
}
```

---

## §9 - Câu hỏi mở

Đã chốt. Hoãn:
- Bản đồ luật chi tiết hơn (cookie window per-network, ngưỡng giao dịch MOIT) - mở rộng CountryPolicy khi FR-COMPLY-008 cần.
- Quản trị flag qua UI (thay vì sửa YAML/KV) - công cụ vận hành giai đoạn sau.
- Phân vùng dữ liệu vật lý theo residency (DB per-region) - quyết định hạ tầng khi mở ID/TH thật.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| if-country rải khắp code | review | thêm nước phải sửa nhiều nơi | Dữ liệu hóa CountryPolicy (§1 #1) |
| Mặc định mở cho nước mới | policy test | gợi ý/affiliate trái luật | Mặc định hạn chế nhất (§1 #5) |
| Áp luật VN cho MY/PH | cart test | giỏ hỏng, vi phạm điều khoản | Stacking per-country (§1 #4) |
| Quên truyền country | chữ ký Lookup | lấy nhầm mặc định toàn cục | Bắt buộc country tham số (§1 #7) |
| Đoán country từ IP | review | sai với VPN/du lịch | Suy từ platform/locale (§1 #8) |
| YAML hỏng chạy lén | loader fail-fast | hành vi mơ hồ | Lỗi lúc khởi tạo (§1 #10) |
| Trùng country trong file | loader validate | policy bất định | Lỗi trùng (§1 #10) |
| Channel ngoài enum | loader validate | kênh không xác định | Lỗi channel lạ (§1 #10) |
| Module giữ bản sao luật riêng | review | luật lệch theo thời gian | Một nguồn sự thật region (§1 #12) |

---

## §11 - Ghi chú

- Dữ liệu hóa gating thành CountryPolicy gom luật voucher/affiliate/residency vào một nơi: thêm nước là thêm một mục cấu hình, không phải sửa hàng chục if.
- Mặc định hạn chế nhất là tư thế an toàn khi mở nước mới chưa kịp nghiên cứu luật - tắt rồi bật có chủ đích.
- Khác biệt stacking VN vs MY/PH là luật cụ thể và đắt nếu sai; policy per-country cho optimizer biết luật đúng của nước đang tính.
- Bắt buộc country trong `Lookup` biến lỗi "quên truyền country" thành lỗi rõ ràng thay vì hành vi sai âm thầm.
- Country từ platform/locale ổn định hơn IP (VPN/du lịch); IP chỉ bổ trợ.
- Một nguồn sự thật per-country cho CART/AFFIL/COMPLY tránh các bản sao luật lệch nhau theo thời gian.

---

*Hết FR-INFRA-005. Status: ready_to_implement (mục tiêu audit 10/10).*
