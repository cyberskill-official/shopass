---
fr_id: FR-SCRAPE-006
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. FR-SCRAPE-006 đặc tả giám sát DOM/selector drift + adapter health ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Rolling window đếm outcome per (platform, adapter_version); baseline động (không hằng số tuyệt đối); state machine healthy->degraded->broken có hysteresis + số mẫu tối thiểu (3 lớp chống báo động giả); phân biệt parse_fail (lỗi của ta, sửa adapter) với challenge/network (lỗi môi trường) để alert đúng địa chỉ; hạ tải target broken bảo vệ proxy. Đúng mitigation §5.2 (rủi ro phụ thuộc nền tảng) + §3.2 (Shopee A/B test DOM). Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Phát hiện 4 ký tự mũi tên U+2192 trong comment Vietnamese của khối code (`state.go::Next` dòng 142 + 152) và test (dòng 226, 230), vi phạm tiêu chí O. Đã sửa thành `->`. Quét lại: 0 mũi tên; không có em-dash/en-dash/nháy cong/ellipsis.

### ISS-002 - Đối chiếu §3.2/§5.2 (xác nhận khớp)
Kiểm chéo: §3.2 nói DOM giỏ hàng Shopee thay đổi theo A/B test; §5.2 là rủi ro phụ thuộc nền tảng existential (sàn đổi DOM/API). FR §1 + §11 neo đúng vào hai mục này; DEC-SCRAPE-24..27 hợp lý. Không khiếm khuyết.

### ISS-003 - Cross-ref orchestrator + observability (xác nhận)
§1 #4 alert qua FR-INFRA-004; §1 #6 `ShouldThrottle` cho orchestrator FR-SCRAPE-001 nhân `next_run_at` khi broken (AC9); §1 #9 phân biệt parse_fail vs challenge (challenge đến từ FR-SCRAPE-005), AC10 `TestMonitor_ParseFailNotChallenge`. Bảng §10 non-trivial (10 dòng, đủ phát hiện + khắc phục). FR này không sở hữu bảng DB nên không có ràng buộc schema/tiền tệ - phù hợp.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 đếm outcome theo window | AC1 | `window.go::Record/ParseFailRate` + `TestWindow_RateAndCount` |
| #2 baseline động | AC2 | `monitor.go` baseline |
| #3 state machine + hysteresis | AC4,AC5 | `state.go::Next` + `TestNext_*` |
| #4 alert đột biến | AC7 | `alert.go` |
| #5 dedup + cooldown | AC8 | `TestMonitor_AlertDedup` |
| #6 throttle khi broken | AC9 | `monitor.go::ShouldThrottle` |
| #7 số mẫu tối thiểu | AC3 | `minSamples` + `TestNext_MinSamplesGuard` |
| #8 tự hồi phục | AC6 | `Next` về Healthy + `TestNext_Recovers` |
| #9 phân loại lỗi | AC10 | `TestMonitor_ParseFailNotChallenge` |
| #10 OTel metric (SHOULD) | AC12 | gauges/counter |
| #11 race-safe | AC11 | `TestWindow_RaceSafe` (-race) |
| #12 state transition log (SHOULD) | - (§9/§11) | history log |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Ba lớp chống báo động giả (baseline động + hysteresis + min mẫu) và phân loại lỗi đúng địa chỉ khóa bằng test, đúng §3.2/§5.2. Một typography defect đã sửa. Score = 10/10. Verdict: PASS.

---

*Audit độc lập FR-SCRAPE-006 - hết.*
