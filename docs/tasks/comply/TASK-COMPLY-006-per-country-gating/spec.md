---
id: TASK-COMPLY-006
title: "Khung per-country gating - luật voucher/affiliate/dữ liệu theo nước (VN/ID/TH/PH/MY/SG/TW), cấu hình khai báo + cổng kiểm tra runtime"
module: COMPLY
priority: MUST
status: done
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-INFRA-005, TASK-CART-004, TASK-COMPLY-007, TASK-AFFIL-004]
depends_on: [TASK-INFRA-005, TASK-CART-004]
blocks: [TASK-COMPLY-007]
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.7 (SEA sequencing, per-country gating bắt buộc: MY & PH bỏ stacking voucher 2025, quyền kênh affiliate khác, cookie window khác, bảo vệ dữ liệu khác)"
  - "docs/... §2 (per-country gating bắt buộc do luật khác theo nước), §5.5 (SEA: Indonesia PDP, Thailand PDPA)"
source_decisions:
  - "DEC-COMPLY-23: ma trận luật theo nước là cấu hình khai báo (country_rule) versioned; KHÔNG rải if-country rải rác trong mã"
  - "DEC-COMPLY-24: mặc định an toàn - nước chưa cấu hình -> tính năng có rủi ro pháp lý bị tắt (deny by default)"
  - "DEC-COMPLY-25: ba trục gating tối thiểu: voucher_stacking, affiliate_channel, data_protection_regime; mở rộng được"
  - "DEC-COMPLY-26: tái dùng region/feature-flag hook của TASK-INFRA-005; task này cung cấp lớp luật, INFRA-005 cung cấp lớp định tuyến vùng"

language: "PostgreSQL 16 + Go 1.22 (comply-svc)"
service: shopass/services/comply/
new_files:
  - services/comply/migrations/0007_country_rule.sql
  - services/comply/internal/gating/rule.go
  - services/comply/internal/gating/registry.go
  - services/comply/internal/gating/repo.go
  - services/comply/internal/gating/rule_test.go
  - services/comply/internal/gating/registry_test.go
modified_files:
  - services/comply/internal/gating/types.go      # struct CountryRule, GateKey
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - rải điều kiện if country == "MY" trong mã nghiệp vụ thay vì tra country_rule (vi phạm DEC-COMPLY-23)
  - mặc định cho phép tính năng ở nước chưa cấu hình (vi phạm DEC-COMPLY-24 deny-by-default)
  - hardcode cookie-window/stacking-rule trong CART thay vì lấy từ gating (vi phạm DEC-COMPLY-23)

effort_hours: 8
sub_tasks:
  - "0.5h: 0007_country_rule.sql - bảng country_rule (country, gate_key, value) versioned + seed 7 nước"
  - "1.0h: types.go + rule.go - enum GateKey, model CountryRule, parse value theo gate"
  - "1.5h: registry.go - Allow(country, gate) + Value(country, gate) deny-by-default"
  - "1.0h: repo.go - load ma trận, cache, reload khi đổi version"
  - "1.0h: seed 7 nước (VN/ID/TH/PH/MY/SG/TW) cho 3 trục gating"
  - "1.5h: rule_test.go - MY/PH no-stacking; VN stacking; nước lạ -> deny"
  - "1.5h: registry_test.go - data regime đúng nước; mở rộng gate mới không phá nước cũ"

risk_if_skipped: "Tài liệu nguồn nói rõ per-country gating là BẮT BUỘC (§2, §5.7): luật voucher, affiliate, dữ liệu khác nhau theo nước - MY & PH bỏ stacking voucher 2025, quyền kênh affiliate khác, cookie window khác, chế độ bảo vệ dữ liệu khác. Nếu rải if-country trong mã thì khi mở mỗi nước mới phải sửa khắp nơi, dễ sót và vi phạm luật địa phương. Cart optimizer (TASK-CART-004) áp sai luật stacking -> gợi ý sai, có thể vi phạm điều khoản sàn ở nước đó. Khung khai báo + deny-by-default biến luật theo nước thành cấu hình kiểm soát được, là điều kiện để mở rộng SEA an toàn (§8 roadmap P3)."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **MUST** cung cấp khung per-country gating khai báo: ma trận luật theo nước (voucher stacking, kênh affiliate, chế độ bảo vệ dữ liệu) lưu dạng cấu hình versioned, với cổng kiểm tra runtime deny-by-default. Mã nghiệp vụ **MUST** tra cứu khung này thay vì rải điều kiện theo nước. Hợp đồng:

1. **MUST** định nghĩa bảng `country_rule (id, country, gate_key, value, version, effective_from, created_at)` với `country` mã ISO 2 ký tự (VN/ID/TH/PH/MY/SG/TW), `gate_key` thuộc tập GateKey, `value` là JSON/text theo gate. Mỗi (country, gate_key, version) bất biến.
2. **MUST** định nghĩa tập GateKey tối thiểu (DEC-COMPLY-25): `voucher_stacking` (bool/policy), `affiliate_channel` (danh sách kênh được phép, ví dụ có cho extension hay không), `data_protection_regime` (mã chế độ: PDPL/PDP_ID/PDPA_TH...). Tập mở rộng được.
3. **MUST** seed ma trận cho 7 nước trong phạm vi SEA sequencing (§5.7), tối thiểu cho 3 gate trên. Ví dụ: VN `voucher_stacking = allow`; MY và PH `voucher_stacking = deny` (bỏ stacking 2025).
4. **MUST** cung cấp cổng `Allow(ctx, country string, gate GateKey) (bool, error)` và `Value(ctx, country string, gate GateKey) (RuleValue, error)` đọc từ ma trận.
5. **MUST** mặc định an toàn deny-by-default (DEC-COMPLY-24): nước chưa cấu hình cho một gate -> `Allow` trả false; tính năng rủi ro pháp lý bị tắt cho tới khi có luật được khai báo rõ.
6. **MUST** đảm bảo mã nghiệp vụ KHÔNG rải if-country (DEC-COMPLY-23): CART (TASK-CART-004) lấy luật stacking qua `Value(country, voucher_stacking)`; AFFIL (TASK-AFFIL-004) lấy kênh được phép qua `Value(country, affiliate_channel)`; quy ước này được audit (TASK-COMPLY-005 có thể bổ sung rule).
7. **MUST** tái dùng region/feature-flag hook của TASK-INFRA-005 (DEC-COMPLY-26): INFRA-005 quyết định "user thuộc nước nào"; task này quyết định "nước đó được làm gì". Hai lớp tách bạch.
8. **MUST** versioned + reload: đổi luật một nước tạo version mới; registry nạp lại ma trận hiệu lực mà không cần triển khai lại mã.
9. **MUST** validate `country` (ISO 2 ký tự, thuộc tập đã biết) và `gate_key` (thuộc GateKey); từ chối giá trị lạ bằng lỗi xác định.
10. **SHOULD** phát OTel metric `gating_denied_total{country,gate}` để thấy tính năng nào bị chặn ở nước nào; hỗ trợ rà soát mở rộng.
11. **MUST** đảm bảo `data_protection_regime` của một nước trỏ đúng chế độ để TASK-COMPLY-007 chọn adapter (PDPL cho VN, PDP cho ID, PDPA cho TH).
12. **MUST** đảm bảo thêm một GateKey mới hoặc một nước mới KHÔNG phá hành vi nước/gate cũ (mở rộng cộng dồn, không sửa ngược).

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao ma trận khai báo, không if-country rải rác (DEC-COMPLY-23)?** Luật khác nhau theo nước trên nhiều trục (stacking, affiliate, dữ liệu). Nếu mỗi nơi tự `if country == "MY"`, thì mở một nước mới phải sửa khắp CART, AFFIL, COMPLY - dễ sót chỗ và sai luật. Tập trung vào một bảng `country_rule` + cổng `Allow`/`Value` cho ra một nguồn sự thật: thêm nước = thêm dòng cấu hình, không đụng mã nghiệp vụ.

**Vì sao deny-by-default (DEC-COMPLY-24)?** Khi mở một nước mới mà chưa nghiên cứu xong luật, lựa chọn an toàn là TẮT tính năng có rủi ro pháp lý, không phải bật. Nếu mặc định cho phép, một nước chưa cấu hình sẽ chạy luật của nước khác - vi phạm. Deny-by-default biến "chưa biết luật" thành "tạm khóa", buộc phải khai báo rõ trước khi bật.

**Vì sao tách INFRA-005 (vùng) khỏi COMPLY-006 (luật) (DEC-COMPLY-26)?** "User ở nước nào" là bài toán định tuyến vùng (INFRA-005: region config, feature flags). "Nước đó được làm gì" là bài toán luật (task này). Trộn hai thứ làm cả hai khó bảo trì. Tách lớp: INFRA-005 trả `country`, COMPLY-006 trả `được phép gì` cho `country` đó.

**Vì sao ví dụ MY/PH no-stacking (§1 #3)?** Tài liệu nguồn ghi cụ thể MY và PH bỏ stacking voucher 2025 (§5.7). Đây là khác biệt luật thật, có hệ quả trực tiếp lên cart optimizer: ở VN có thể chồng voucher, ở MY/PH không. Seed đúng khác biệt này cho CART-004 áp luật đúng theo nước.

**Vì sao data_protection_regime trỏ adapter (§1 #11)?** SEA có nhiều chế độ bảo vệ dữ liệu (PDPL VN, PDP Indonesia, PDPA Thailand). TASK-COMPLY-007 cần biết nước này theo chế độ nào để chọn adapter tuân thủ tương ứng. `data_protection_regime` là điểm nối: gating khai báo chế độ, adapter thực thi chi tiết.

---

## §3 - Hợp đồng API / DDL

### Migration

```sql
-- services/comply/migrations/0007_country_rule.sql
CREATE TABLE country_rule (
  id             BIGSERIAL   PRIMARY KEY,
  country        TEXT        NOT NULL CHECK (char_length(country) = 2),
  gate_key       TEXT        NOT NULL,
  value          JSONB       NOT NULL,
  version        INTEGER     NOT NULL DEFAULT 1,
  effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (country, gate_key, version)
);

-- Seed 7 nuoc x 3 gate (vi du, version 1).
INSERT INTO country_rule (country, gate_key, value) VALUES
  ('VN', 'voucher_stacking', '{"mode":"allow"}'),
  ('ID', 'voucher_stacking', '{"mode":"allow"}'),
  ('TH', 'voucher_stacking', '{"mode":"allow"}'),
  ('PH', 'voucher_stacking', '{"mode":"deny"}'),   -- bo stacking 2025
  ('MY', 'voucher_stacking', '{"mode":"deny"}'),   -- bo stacking 2025
  ('SG', 'voucher_stacking', '{"mode":"allow"}'),
  ('TW', 'voucher_stacking', '{"mode":"allow"}'),
  ('VN', 'data_protection_regime', '{"regime":"PDPL"}'),
  ('ID', 'data_protection_regime', '{"regime":"PDP_ID"}'),
  ('TH', 'data_protection_regime', '{"regime":"PDPA_TH"}'),
  ('VN', 'affiliate_channel', '{"allow":["web","extension"]}'),
  ('MY', 'affiliate_channel', '{"allow":["web"]}');
```

### Registry (Go)

```go
// services/comply/internal/gating/registry.go
type GateKey string

const (
    GateVoucherStacking GateKey = "voucher_stacking"
    GateAffiliateChannel GateKey = "affiliate_channel"
    GateDataRegime       GateKey = "data_protection_regime"
)

// Value tra cau hinh luat; deny-by-default khi nuoc chua cau hinh.
func (r *Registry) Value(ctx context.Context, country string, g GateKey) (RuleValue, error) {
    if !validCountry(country) {
        return RuleValue{}, ErrUnknownCountry
    }
    v, ok := r.lookup(country, g)
    if !ok {
        return RuleValue{Denied: true}, nil // chua cau hinh -> tat an toan
    }
    return v, nil
}

// Allow tien loi cho gate dang bool/mode.
func (r *Registry) Allow(ctx context.Context, country string, g GateKey) (bool, error) {
    v, err := r.Value(ctx, country, g)
    if err != nil {
        return false, err
    }
    return v.Mode == "allow", nil // thieu cau hinh -> Mode rong -> false
}
```

---

## §4 - Acceptance criteria

1. Migration chạy sạch -> `country_rule` tồn tại; seed 7 nước cho `voucher_stacking`.
2. `Allow("VN", voucher_stacking)` -> true.
3. `Allow("MY", voucher_stacking)` -> false (bỏ stacking 2025).
4. `Allow("PH", voucher_stacking)` -> false.
5. `Allow("KR", voucher_stacking)` (nước chưa trong tập) -> lỗi `ErrUnknownCountry`.
6. `Value("VN", data_protection_regime)` -> regime `PDPL`.
7. `Value("ID", data_protection_regime)` -> regime `PDP_ID`.
8. `Value("TH", data_protection_regime)` -> regime `PDPA_TH`.
9. `Value("SG", affiliate_channel)` chưa seed -> `Denied = true` (deny-by-default), không lỗi.
10. INSERT `country` 3 ký tự -> lỗi CHECK constraint.
11. Thêm version 2 cho `(MY, voucher_stacking)` = allow -> registry reload trả allow; nước/gate khác không đổi.
12. Metric `gating_denied_total{country,gate}` tăng khi một gate bị deny.

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/gating/rule_test.go
func TestGate_MYNoStacking(t *testing.T) {
    r := setupSeeded(t)
    ok, _ := r.Allow(ctx, "MY", GateVoucherStacking)
    require.False(t, ok) // bo stacking 2025
}

func TestGate_VNStacking(t *testing.T) {
    r := setupSeeded(t)
    ok, _ := r.Allow(ctx, "VN", GateVoucherStacking)
    require.True(t, ok)
}

func TestGate_UnknownCountryRejected(t *testing.T) {
    r := setupSeeded(t)
    _, err := r.Allow(ctx, "KR", GateVoucherStacking)
    require.ErrorIs(t, err, ErrUnknownCountry)
}

func TestGate_DenyByDefault(t *testing.T) {
    r := setupSeeded(t)
    v, err := r.Value(ctx, "SG", GateAffiliateChannel) // chua seed
    require.NoError(t, err)
    require.True(t, v.Denied) // tat an toan
}

// services/comply/internal/gating/registry_test.go
func TestRegistry_DataRegimePerCountry(t *testing.T) {
    r := setupSeeded(t)
    vn, _ := r.Value(ctx, "VN", GateDataRegime)
    id, _ := r.Value(ctx, "ID", GateDataRegime)
    th, _ := r.Value(ctx, "TH", GateDataRegime)
    require.Equal(t, "PDPL", vn.Regime)
    require.Equal(t, "PDP_ID", id.Regime)
    require.Equal(t, "PDPA_TH", th.Regime)
}

func TestRegistry_NewVersionDoesNotBreakOthers(t *testing.T) {
    r := setupSeeded(t)
    seedVersion(t, r, "MY", GateVoucherStacking, 2, `{"mode":"allow"}`)
    r.Reload(ctx)
    my, _ := r.Allow(ctx, "MY", GateVoucherStacking)
    ph, _ := r.Allow(ctx, "PH", GateVoucherStacking)
    require.True(t, my)   // version moi
    require.False(t, ph)  // nuoc khac khong doi
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: migration 0007 (country_rule + seed) -> types.go (GateKey, RuleValue) -> rule.go (parse value theo gate) -> registry.go (Allow/Value deny-by-default + cache) -> repo.go (load + reload theo version) -> tests. Registry nạp ma trận hiệu lực vào bộ nhớ, reload khi version đổi (không cần triển khai lại mã nghiệp vụ). CART-004 và AFFIL-004 inject registry và gọi `Value`/`Allow`; không nơi nào hardcode luật theo nước. INFRA-005 cung cấp `country` của user; task này nhận `country` đó và trả luật.

---

## §7 - Phụ thuộc

- **TASK-INFRA-005** - region config xác định `country` của user (lớp định tuyến vùng).
- **TASK-CART-004** - cart optimizer lấy luật `voucher_stacking` qua `Value`; thay if-country bằng tra ma trận.
- **TASK-AFFIL-004 (liên quan)** - guardrails affiliate lấy kênh được phép qua `affiliate_channel`.
- **TASK-COMPLY-007 (downstream)** - adapter SEA chọn theo `data_protection_regime`.
- Lib: `encoding/json`, driver `pgx`.

---

## §8 - Payload ví dụ

### CART hỏi luật stacking cho nước của user

```go
v, err := gating.Value(ctx, user.Country, gating.GateVoucherStacking)
if err != nil { return err }
if v.Mode == "deny" {
    // MY/PH: khong goi y chong voucher
    return optimizeWithoutStacking(cart)
}
return optimizeWithStacking(cart) // VN/ID/SG/TW
```

### Khai báo luật cho một nước mới khi mở rộng (cấu hình, không sửa mã)

```json
POST /v1/comply/country-rule
[
  { "country": "TH", "gate_key": "affiliate_channel", "value": { "allow": ["web"] } },
  { "country": "TH", "gate_key": "voucher_stacking",  "value": { "mode": "allow" } }
]
```

---

## §9 - Câu hỏi mở

Đã chốt khung. Hoãn:
- Trục gating chi tiết hơn (cookie_window theo nước, ngưỡng giao dịch foreign platform) - thêm GateKey khi cần.
- Quy trình phê duyệt thay đổi luật (4-eyes) trước khi version mới hiệu lực - bổ sung kiểm soát thay đổi sau.
- Đồng bộ luật với nguồn pháp lý ngoài (theo dõi thay đổi quy định) - hiện cập nhật thủ công.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| if-country rải trong mã | review §1 #6 + audit | mở nước mới sót chỗ | Tập trung `country_rule`; gọi `Value` |
| Nước chưa cấu hình chạy luật nước khác | deny-by-default + AC #9 | vi phạm luật địa phương | `Value` trả Denied khi thiếu cấu hình |
| MY/PH áp stacking sai | AC #3,#4 + test | gợi ý vi phạm điều khoản sàn | Seed deny cho MY/PH; CART tra gating |
| country lạ | `ErrUnknownCountry` + AC #5 | hành vi không xác định | validate ISO 2 ký tự + tập đã biết |
| country sai định dạng | DB CHECK + AC #10 | dữ liệu bẩn | CHECK char_length = 2 |
| Đổi luật một nước phá nước khác | versioned + AC #11 | hồi quy chéo | Mở rộng cộng dồn; test không phá nước cũ |
| Adapter SEA chọn sai regime | `data_protection_regime` + AC #6,#7,#8 | tuân thủ sai chế độ | Regime khai báo rõ; TASK-COMPLY-007 đọc |
| Reload không cập nhật | repo reload §1 #8 | luật cũ còn hiệu lực | Reload theo version; test reload |

---

## §11 - Ghi chú

- Per-country gating là bắt buộc theo tài liệu nguồn (§2, §5.7): luật voucher/affiliate/dữ liệu khác nhau theo nước.
- Ma trận khai báo + cổng `Allow`/`Value` thay cho if-country rải rác: thêm nước = thêm cấu hình, không sửa mã nghiệp vụ.
- Deny-by-default: nước chưa cấu hình thì tắt tính năng rủi ro pháp lý, buộc khai báo luật rõ trước khi bật.
- Tách lớp với INFRA-005: INFRA-005 trả "user ở nước nào", COMPLY-006 trả "nước đó được làm gì".
- Khác biệt thật được seed: MY/PH bỏ stacking voucher 2025 (§5.7); CART-004 áp luật đúng theo nước.
- `data_protection_regime` là điểm nối sang TASK-COMPLY-007: gating khai báo chế độ (PDPL/PDP/PDPA), adapter thực thi chi tiết.

---

*Hết TASK-COMPLY-006. Status: ready_to_review (implementation ready for review).*
