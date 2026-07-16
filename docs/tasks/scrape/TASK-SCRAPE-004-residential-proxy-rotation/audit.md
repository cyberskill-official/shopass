---
fr_id: TASK-SCRAPE-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. TASK-SCRAPE-004 đặc tả residential proxy rotation + tiering + cost-guard ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Residential bắt buộc (datacenter vô dụng với Cloudflare/Akamai §3.3); phân tầng nhà cung cấp khớp số §3.3 nguồn (enterprise 8,5-12 / mid 3-6 / budget 1,75 USD/GB); `SelectTier` chọn theo độ khó target; cost-guard chặn trước khi tiêu, ưu tiên giữ hot hy sinh cold. Bảng `proxy_usage` và sổ chi phí khớp DATA-MODEL. Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Phát hiện 3 ký tự mũi tên U+2192 trong comment Vietnamese của khối code (`tier.go::SelectTier` dòng 147, `costguard.go::Evaluate` dòng 173 + 175), vi phạm tiêu chí O. Đã sửa thành `->`. Quét lại: 0 mũi tên; không có em-dash/en-dash/nháy cong/ellipsis.

### ISS-002 - Đối chiếu proxy_usage với DATA-MODEL (xác nhận khớp)
Kiểm cột-theo-cột: §3 `0002_proxy_usage.sql` có `day DATE`, `provider TEXT`, `tier TEXT`, `country TEXT`, `bytes_used BIGINT`, `cost_micro_usd BIGINT` (micro-USD số nguyên), `PRIMARY KEY (day, provider, tier, country)`. Trùng khít DATA-MODEL của `proxy_usage`. Tiền/chi phí là BIGINT, không float. AC1 + AC10 + `TestCost_IntegerMicroUSD` khóa `1_750_000 micro x 2GB = 3_500_000`. Không khiếm khuyết.

### ISS-003 - Tier giá khớp nguồn §3.3 (xác nhận)
So §3.3 dòng 90: Enterprise (Bright Data, Oxylabs) $8,5-12/GB; Mid (Decodo, SOAX, NetNut) $3-6/GB; Budget (IPRoyal) từ $1,75/GB. task §1 #2 + `tier.go` chép đúng dải giá và nhà cung cấp. `SelectTier(DiffAkamai/DiffByteDance)=enterprise`, `DiffShopeeJSON=budget`, `DiffUnknown=mid` (AC2,AC3). Khớp.

### ISS-004 - Geo nhất quán + cost-guard (xác nhận cross-ref)
§1 #6 cặp (proxy <-> profile cùng nước) với TASK-SCRAPE-003 (AC7 `GeoMatchesProfile`); §1 #7 `Evaluate` Proceed/Downgrade/BlockCold giữ hot (AC8,AC9). Cross-ref đúng.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 residential bắt buộc | AC1 | DEC-SCRAPE-15 + config |
| #2 tier nhà cung cấp (giá §3.3) | AC2,AC3 | `tier.go::Tier` |
| #3 SelectTier theo độ khó | AC2,AC3 | `SelectTier` + `TestSelectTier_ByDifficulty` |
| #4 Acquire (tier,country) + xoay IP | AC5 | `pool.go::Acquire/rotate` |
| #5 cooldown IP ban | AC6 | `MarkBanned` + `TestAcquire_BannedIPCooldown` |
| #6 geo khớp fingerprint | AC7 | `BindProfile` + `TestAcquire_GeoMatchesProfile` |
| #7 cost-guard hạ tier/block cold | AC8,AC9 | `costguard.go::Evaluate` + `TestCostGuard_DowngradeThenBlock` |
| #8 CanProceed | AC9 | `pool.go::Acquire` guard |
| #9 chi phí số nguyên micro-USD | AC10 | `costMicroUSD` + `TestCost_IntegerMicroUSD` |
| #10 OTel metric (SHOULD) | AC12 | counters/histogram |
| #11 creds Vault | AC4 | `vault.ProxyCreds` (TASK-INFRA-003) |
| #12 proxy_usage báo cáo §4.1 | AC11 | `0002_proxy_usage.sql` |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. `proxy_usage` + `cost_micro_usd` BIGINT khớp DATA-MODEL; tier giá khớp §3.3; cost-guard gắn ngưỡng §4.1; geo nhất quán TASK-SCRAPE-003. Một typography defect đã sửa. Score = 10/10. Verdict: PASS.

---

*Audit độc lập TASK-SCRAPE-004 - hết.*
