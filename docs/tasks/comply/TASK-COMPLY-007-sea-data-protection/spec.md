---
id: TASK-COMPLY-007
title: "Adapter bảo vệ dữ liệu SEA - Indonesia PDP Law + Thailand PDPA; lớp adapter theo regime trên khung consent/DSAR/breach VN"
module: COMPLY
priority: SHOULD
status: ready_to_review
verify: T
phase: P3
milestone: P3 - slice 1
slice: 1
owner: Stephen Cheng (Founder)
created: 2026-06-28
related_frs: [TASK-COMPLY-001, TASK-COMPLY-006, TASK-COMPLY-002, TASK-COMPLY-003, TASK-COMPLY-004]
depends_on: [TASK-COMPLY-001, TASK-COMPLY-006]
blocks: []
source_pages:
  - "docs/TÀI LIỆU NỀN TẢNG SẢN PHẨM SănDeal §5.5 (SEA khác biệt: Indonesia PDP Law, Thailand PDPA -> per-country compliance)"
  - "docs/... §5.7 (SEA sequencing VN -> ID/TH -> PH/MY -> SG/TW), §2 (per-country gating bắt buộc)"
source_decisions:
  - "DEC-COMPLY-27: bảo vệ dữ liệu SEA là lớp adapter theo regime trên khung lõi VN (TASK-COMPLY-001/003/004), KHÔNG viết lại từ đầu mỗi nước"
  - "DEC-COMPLY-28: adapter chọn theo data_protection_regime của TASK-COMPLY-006 (PDPL/PDP_ID/PDPA_TH); regime quyết định khác biệt theo nước"
  - "DEC-COMPLY-29: mỗi adapter khai báo khác biệt cụ thể (yêu cầu consent, hạn DSAR, hạn breach, localization) so với baseline VN; phần chung tái dùng"
  - "DEC-COMPLY-30: SHOULD ở P3 - chỉ kích hoạt adapter cho nước đã mở (gating bật); nước chưa mở giữ deny-by-default"

language: "Go 1.22 (comply-svc adapter layer)"
service: shopass/services/comply/
new_files:
  - services/comply/internal/regime/adapter.go
  - services/comply/internal/regime/vn_pdpl.go
  - services/comply/internal/regime/id_pdp.go
  - services/comply/internal/regime/th_pdpa.go
  - services/comply/internal/regime/registry.go
  - services/comply/internal/regime/adapter_test.go
  - services/comply/internal/regime/registry_test.go
modified_files:
  - services/comply/internal/regime/types.go      # interface RegimeAdapter, struct RegimeProfile
allowed_tools:
  - file_read: services/comply/**
  - file_write: services/comply/**
  - bash: cd services/comply && go test ./...
disallowed_tools:
  - viết lại toàn bộ consent/DSAR/breach cho mỗi nước thay vì adapter trên baseline (vi phạm DEC-COMPLY-27)
  - kích hoạt adapter cho nước chưa mở/chưa gating (vi phạm DEC-COMPLY-30)
  - hardcode regime trong mã thay vì lấy từ TASK-COMPLY-006 (vi phạm DEC-COMPLY-28)

effort_hours: 8
sub_tasks:
  - "0.5h: types.go - interface RegimeAdapter + RegimeProfile (khác biệt theo nước)"
  - "1.0h: adapter.go - baseline VN PDPL làm mặc định + khung override"
  - "1.0h: vn_pdpl.go - profile baseline (72h breach, DPIA 60d, consent PDPL)"
  - "1.0h: id_pdp.go - profile Indonesia PDP (khác biệt khai báo so baseline)"
  - "1.0h: th_pdpa.go - profile Thailand PDPA (khác biệt khai báo so baseline)"
  - "1.0h: registry.go - chọn adapter theo regime từ TASK-COMPLY-006; chỉ nước đã mở"
  - "1.5h: adapter_test.go - mỗi regime trả profile đúng; phần chung tái dùng baseline"
  - "1.0h: registry_test.go - nước chưa mở -> không adapter; regime lạ -> lỗi"

risk_if_skipped: "Khi mở SEA (P3, §5.7), Indonesia có PDP Law và Thailand có PDPA - mỗi nước có yêu cầu bảo vệ dữ liệu riêng (§5.5). Nếu áp nguyên khung PDPL VN cho ID/TH mà không điều chỉnh theo luật địa phương, có thể vi phạm quy định nước đó. Viết lại toàn bộ consent/DSAR/breach cho mỗi nước thì tốn kém và dễ lệch nhau. Lớp adapter trên baseline VN giữ phần chung (đa số nguyên tắc tương đồng) và chỉ khai báo khác biệt theo nước, là cách mở rộng SEA bền vững. SHOULD vì chỉ cần khi nước tương ứng được mở; nước chưa mở giữ deny-by-default (TASK-COMPLY-006)."
---

## §1 - Mô tả (BCP-14 normative)

Service COMPLY **SHOULD** cung cấp lớp adapter bảo vệ dữ liệu SEA: mỗi chế độ (regime) là một adapter khai báo khác biệt so với baseline PDPL VN, được chọn theo `data_protection_regime` của TASK-COMPLY-006, áp dụng cho nước đã mở. Hợp đồng:

1. **SHOULD** định nghĩa interface `RegimeAdapter` trả `RegimeProfile` - tập tham số tuân thủ theo nước: yêu cầu consent (granularity, ngôn ngữ), hạn DSAR phản hồi, hạn/ngưỡng thông báo breach, yêu cầu bản địa hóa thông báo.
2. **SHOULD** cung cấp adapter baseline `vn_pdpl` làm mặc định (DEC-COMPLY-27): phản ánh khung lõi đã có (breach 72h, DPIA 60 ngày + review 6 tháng, consent PDPL). Các adapter khác kế thừa baseline và override phần khác biệt.
3. **SHOULD** cung cấp adapter `id_pdp` (Indonesia PDP Law) khai báo khác biệt cụ thể so với baseline; phần không khác giữ nguyên baseline (DEC-COMPLY-29).
4. **SHOULD** cung cấp adapter `th_pdpa` (Thailand PDPA) khai báo khác biệt cụ thể so với baseline; phần không khác giữ nguyên baseline.
5. **SHOULD** chọn adapter theo regime từ TASK-COMPLY-006 (DEC-COMPLY-28): `registry.For(ctx, country)` đọc `data_protection_regime` của nước qua gating, trả adapter tương ứng. KHÔNG hardcode regime.
6. **SHOULD** chỉ kích hoạt adapter cho nước đã mở (DEC-COMPLY-30): nước chưa mở (gating deny-by-default) -> `For` trả lỗi/không adapter; tính năng SEA không chạy cho nước chưa sẵn sàng.
7. **SHOULD** đảm bảo phần chung tái dùng baseline: nếu một regime không override một tham số, giá trị baseline VN được dùng (tránh trùng lặp khai báo).
8. **SHOULD** expose `Profile(ctx, country) (RegimeProfile, error)` cho các task khác (consent UI, DSAR SLA, breach window) đọc tham số đúng theo nước user.
9. **SHOULD** validate regime trả từ gating thuộc tập đã hỗ trợ (PDPL/PDP_ID/PDPA_TH); regime lạ -> lỗi xác định, không im lặng dùng baseline sai.
10. **MAY** phát OTel metric `regime_profile_resolved_total{regime}` để thấy phân bố nước đang phục vụ.
11. **SHOULD** khai báo khác biệt dưới dạng dữ liệu (RegimeProfile field) chứ không nhánh điều kiện rải rác, để thêm regime mới là thêm một profile.
12. **SHOULD** ghi rõ trong từng adapter nguồn khác biệt (chú thích quy định) để rà soát pháp lý truy được về luật gốc.

---

## §2 - Vì sao thiết kế này (rationale cho người đọc)

**Vì sao adapter trên baseline, không viết lại mỗi nước (DEC-COMPLY-27)?** Các chế độ bảo vệ dữ liệu SEA chia sẻ phần lớn nguyên tắc (consent, quyền chủ thể, thông báo vi phạm). Viết lại consent/DSAR/breach cho từng nước tạo bốn bản gần giống nhau, lệch dần theo thời gian. Lấy PDPL VN làm baseline và để mỗi nước override phần khác biệt giữ một lõi chung + một lớp mỏng khác biệt - dễ bảo trì, khó lệch.

**Vì sao chọn adapter theo regime của TASK-COMPLY-006 (DEC-COMPLY-28)?** TASK-COMPLY-006 đã khai báo `data_protection_regime` cho mỗi nước (PDPL/PDP_ID/PDPA_TH). Tái dùng điểm nối đó: `registry.For(country)` hỏi gating "nước này chế độ gì" rồi trả adapter tương ứng. Không hardcode regime trong adapter layer tránh hai nguồn sự thật mâu thuẫn.

**Vì sao chỉ kích hoạt cho nước đã mở (DEC-COMPLY-30)?** SEA sequencing có thứ tự (VN -> ID/TH -> PH/MY -> SG/TW). Một nước chưa mở thì chưa nghiên cứu xong luật và gating để deny-by-default. Adapter không nên chạy cho nước chưa sẵn sàng. `For` trả lỗi khi gating chưa mở nước đó - nhất quán với deny-by-default của TASK-COMPLY-006.

**Vì sao khác biệt là dữ liệu, không nhánh điều kiện (§1 #11)?** Nếu khác biệt regime nằm trong `if regime == "PDPA_TH"` rải rác, thêm một nước SEA mới phải sửa khắp nơi. Đặt khác biệt vào `RegimeProfile` (các field tham số) làm việc thêm nước thành thêm một file profile khai báo - cùng triết lý khai báo của TASK-COMPLY-006.

**Vì sao SHOULD chứ không MUST (priority)?** Đây là tính năng P3, chỉ cần khi nước SEA tương ứng được mở. Trước khi mở ID/TH, baseline VN đủ phục vụ thị trường VN. Đánh dấu SHOULD phản ánh đúng: cần khi mở rộng, không chặn release P1/P2. Nhưng khi đã mở một nước SEA, adapter của nước đó trở thành bắt buộc cho nước đó.

---

## §3 - Hợp đồng API / DDL

### Interface + profile (Go)

```go
// services/comply/internal/regime/types.go
type RegimeProfile struct {
    Regime            string        // PDPL / PDP_ID / PDPA_TH
    BreachWindow      time.Duration // baseline 72h
    DPIAFilingWindow  time.Duration // baseline 60 ngay
    ConsentLanguages  []string      // ngon ngu thong bao consent bat buoc
    DSARResponseSLA   time.Duration // han phan hoi DSAR
    Notes             string        // nguon khac biet (quy dinh goc)
}

type RegimeAdapter interface {
    Profile() RegimeProfile
}
```

### Baseline + override (Go)

```go
// services/comply/internal/regime/vn_pdpl.go
func baseline() RegimeProfile {
    return RegimeProfile{
        Regime:           "PDPL",
        BreachWindow:     72 * time.Hour,
        DPIAFilingWindow: 60 * 24 * time.Hour,
        ConsentLanguages: []string{"vi"},
        DSARResponseSLA:  30 * 24 * time.Hour,
        Notes:            "Luat 91/2025/QH15 + NĐ 356/2025 (baseline VN)",
    }
}

// services/comply/internal/regime/th_pdpa.go
// Thailand PDPA: ke thua baseline, override phan khac biet (ngon ngu, ghi chu nguon).
func thPDPA() RegimeProfile {
    p := baseline()
    p.Regime = "PDPA_TH"
    p.ConsentLanguages = []string{"th", "en"} // ban dia hoa thong bao
    p.Notes = "Thailand PDPA - khac biet ngon ngu thong bao; tham so khac giu baseline"
    return p
}

// services/comply/internal/regime/registry.go
// For chon adapter theo regime tu TASK-COMPLY-006; chi nuoc da mo.
func (r *Registry) For(ctx context.Context, country string) (RegimeAdapter, error) {
    rv, err := r.gating.Value(ctx, country, gating.GateDataRegime)
    if err != nil {
        return nil, err
    }
    if rv.Denied {
        return nil, ErrCountryNotOpen // nuoc chua mo (deny-by-default)
    }
    a, ok := r.byRegime[rv.Regime]
    if !ok {
        return nil, ErrUnsupportedRegime // regime la -> khong im lang dung baseline sai
    }
    return a, nil
}
```

---

## §4 - Acceptance criteria

1. `Profile("VN")` -> regime PDPL, BreachWindow 72h, DPIAFilingWindow 60 ngày (baseline).
2. `Profile("TH")` -> regime PDPA_TH; ConsentLanguages chứa `th`; tham số không override = baseline.
3. `Profile("ID")` -> regime PDP_ID; khác biệt khai báo của adapter id_pdp được phản ánh.
4. `For("KR")` (nước chưa mở, gating deny) -> lỗi `ErrCountryNotOpen`.
5. Regime lạ (gating trả regime không hỗ trợ) -> `For` lỗi `ErrUnsupportedRegime`, KHÔNG im lặng dùng baseline.
6. Adapter TH không override `BreachWindow` -> giá trị 72h baseline được dùng (tái dùng phần chung).
7. Thêm một regime mới = thêm một file profile; nước/regime cũ không đổi hành vi.
8. `For("VN")` -> trả adapter vn_pdpl (baseline) đúng.
9. `Profile` đọc regime qua gating (TASK-COMPLY-006), KHÔNG hardcode trong adapter.
10. Mỗi adapter có `Notes` không rỗng trỏ nguồn khác biệt (truy về luật gốc).
11. Metric `regime_profile_resolved_total{regime}` tăng khi resolve một profile.
12. Phần chung không trùng lặp: override chỉ khai báo field khác baseline (review code).

---

## §5 - Kiểm thử (verification)

```go
// services/comply/internal/regime/adapter_test.go
func TestProfile_VNBaseline(t *testing.T) {
    p := baseline()
    require.Equal(t, "PDPL", p.Regime)
    require.Equal(t, 72*time.Hour, p.BreachWindow)
    require.Equal(t, 60*24*time.Hour, p.DPIAFilingWindow)
}

func TestProfile_THInheritsBaselineBreachWindow(t *testing.T) {
    p := thPDPA()
    require.Equal(t, "PDPA_TH", p.Regime)
    require.Contains(t, p.ConsentLanguages, "th")
    require.Equal(t, 72*time.Hour, p.BreachWindow) // khong override -> baseline
}

func TestProfile_AdaptersHaveSourceNotes(t *testing.T) {
    for _, p := range []RegimeProfile{baseline(), idPDP(), thPDPA()} {
        require.NotEmpty(t, p.Notes) // truy ve luat goc
    }
}

// services/comply/internal/regime/registry_test.go
func TestRegistry_CountryNotOpen(t *testing.T) {
    r := setupWithGating(t, denyAll())
    _, err := r.For(ctx, "KR")
    require.ErrorIs(t, err, ErrCountryNotOpen)
}

func TestRegistry_UnsupportedRegimeNotSilent(t *testing.T) {
    r := setupWithGating(t, regimeFor("XX", "PDPA_XX"))
    _, err := r.For(ctx, "XX")
    require.ErrorIs(t, err, ErrUnsupportedRegime) // khong dung baseline sai
}

func TestRegistry_ResolvesPerRegime(t *testing.T) {
    r := setupWithGating(t, openSEA()) // VN/ID/TH mo
    vn, _ := r.For(ctx, "VN")
    th, _ := r.For(ctx, "TH")
    require.Equal(t, "PDPL", vn.Profile().Regime)
    require.Equal(t, "PDPA_TH", th.Profile().Regime)
}
```

---

## §6 - Khung triển khai

Xem §3. Thứ tự: types.go (interface + RegimeProfile) -> vn_pdpl.go (baseline) -> id_pdp.go + th_pdpa.go (override) -> registry.go (chọn theo gating, chỉ nước đã mở) -> tests. Baseline là một hàm trả profile VN; mỗi adapter gọi `baseline()` rồi override field khác biệt - phần chung không lặp lại. `registry.For` đọc regime qua `gating.Value(country, GateDataRegime)` (TASK-COMPLY-006), giữ một nguồn sự thật. Các task consent UI / DSAR SLA / breach window đọc `Profile(country)` để áp tham số đúng nước.

---

## §7 - Phụ thuộc

- **TASK-COMPLY-006** - cung cấp `data_protection_regime` mỗi nước + deny-by-default cho nước chưa mở; `For` đọc từ đây.
- **TASK-COMPLY-001** - baseline consent PDPL; adapter điều chỉnh granularity/ngôn ngữ theo nước.
- **TASK-COMPLY-002 (liên quan)** - DPIA filing window có thể khác theo regime (profile field).
- **TASK-COMPLY-003 (liên quan)** - DSAR response SLA lấy từ profile theo nước.
- **TASK-COMPLY-004 (liên quan)** - breach window lấy từ profile (baseline 72h).
- Lib: chỉ stdlib `time`.

---

## §8 - Payload ví dụ

### Consent UI hỏi profile để bản địa hóa thông báo

```go
prof, err := regime.Profile(ctx, user.Country)
if err != nil { return err }
// TH: hien consent bang th/en; VN: vi
renderConsent(prof.ConsentLanguages, purpose)
```

### Breach service lấy window theo nước của sự cố

```go
prof, _ := regime.Profile(ctx, incident.Country)
authorityDue := incident.AcknowledgedAt.Add(prof.BreachWindow) // baseline 72h tru khi regime override
```

---

## §9 - Câu hỏi mở

Đã chốt khung adapter. Hoãn (cần xác minh pháp lý từng nước trước khi mở, §10 tài liệu nguồn):
- Khác biệt chi tiết hạn breach/DSAR của Indonesia PDP và Thailand PDPA so với 72h/30 ngày baseline - chốt với tư vấn pháp lý địa phương trước khi mở.
- Yêu cầu lưu trú dữ liệu (data residency) theo nước - thêm field profile khi có yêu cầu cụ thể.
- Adapter cho PH/MY/SG/TW - thêm khi sequencing tới các nước đó.

---

## §10 - Failure modes inventory

| Lỗi | Phát hiện | Hệ quả | Khắc phục |
|---|---|---|---|
| Viết lại consent/DSAR/breach mỗi nước | review §1 #2 | bốn bản lệch nhau | Adapter override trên baseline |
| Áp PDPL VN nguyên cho ID/TH | profile override + AC #2,#3 | vi phạm luật địa phương | RegimeProfile khai báo khác biệt |
| Kích hoạt adapter nước chưa mở | `ErrCountryNotOpen` + AC #4 | chạy luật chưa sẵn sàng | `For` tôn trọng deny-by-default |
| Regime lạ im lặng dùng baseline | `ErrUnsupportedRegime` + AC #5 | tuân thủ sai âm thầm | Lỗi xác định, không fallback baseline |
| Trùng lặp khai báo phần chung | review §1 #7 + AC #12 | bảo trì lệch | Override chỉ field khác baseline |
| Hardcode regime trong adapter | review §1 #5 | hai nguồn sự thật | Đọc regime qua gating TASK-COMPLY-006 |
| Mất truy vết nguồn khác biệt | `Notes` §1 #12 + AC #10 | rà soát pháp lý khó | Mỗi adapter ghi nguồn quy định |
| Thêm regime phá nước cũ | AC #7 | hồi quy chéo | Mở rộng cộng dồn; test nước cũ |

---

## §11 - Ghi chú

- Bảo vệ dữ liệu SEA là lớp adapter trên baseline PDPL VN, không viết lại mỗi nước: một lõi chung + lớp mỏng khác biệt.
- Adapter chọn theo `data_protection_regime` của TASK-COMPLY-006; giữ một nguồn sự thật về "nước nào chế độ gì".
- Chỉ kích hoạt adapter cho nước đã mở; nước chưa mở giữ deny-by-default - nhất quán với gating.
- Khác biệt regime là dữ liệu (RegimeProfile field), không nhánh điều kiện: thêm nước = thêm một profile.
- SHOULD vì là tính năng P3, cần khi mở nước SEA tương ứng; nhưng khi đã mở một nước, adapter của nước đó là bắt buộc cho nước đó.
- Khác biệt chi tiết hạn breach/DSAR của ID/TH cần xác minh với tư vấn pháp lý địa phương trước khi mở (§9, §10 tài liệu nguồn) - khung sẵn sàng nhận giá trị đã xác minh.

---

*Hết TASK-COMPLY-007. Status: ready_to_implement (mục tiêu audit 10/10).*
