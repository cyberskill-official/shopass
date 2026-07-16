---
fr_id: TASK-SCRAPE-008
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. TASK-SCRAPE-008 đặc tả adapter Lazada ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Khác biệt cốt lõi Akamai xử lý đúng §3.2/§3.9: TLS/HTTP2 fingerprint match của farm (TASK-SCRAPE-003 #4, kiểm trước khi JS chạy) + residential enterprise bắt buộc (`SelectTier(DiffAkamai)`, datacenter vô dụng với Akamai §3.3). JSON nhúng (window state `__moduleData__`) ưu tiên, fallback DOM, chờ render, né API ký. Giá VND số nguyên đồng bộ TASK-PRICE-002; là chân thứ ba của so sánh chéo 3 sàn (TASK-PRICE-004). Score = 10/10.

## §2 - Findings

### ISS-001 - Typography (xác nhận sạch, không cần sửa)
Quét file: 0 ký tự mũi tên U+2192, 0 em-dash, 0 en-dash, 0 nháy cong, 0 ellipsis. Toàn bộ dùng ASCII `->`. Không có khiếm khuyết typography ở task này (giống 007, khác 001-006 vốn có arrow glyph đã sửa).

### ISS-002 - Đối chiếu fact Lazada/Akamai với §3.2/§3.9 (xác nhận khớp)
Kiểm chéo nguồn: §3.2 dòng 17 - Lazada (Alibaba) dùng Akamai, ưu tiên đọc DOM đã render; §3.9 - Lazada Medium-High, Akamai (Alibaba), fingerprinting TLS/HTTP. task §1 #1-#3 phủ đúng: TLS match là then chốt (không tùy chọn), residential enterprise bắt buộc. Không khiếm khuyết.

### ISS-003 - Tiền tệ không float + cross-ref (xác nhận)
§1 #5 + #12 quy đổi bằng số nguyên (`parseVNDInt`), AC7 khóa `₫1.234.567 -> 1234567`; không float trên tiền. Cross-ref: proxy enterprise + datacenter bị từ chối qua TASK-SCRAPE-004 (AC2 `proxy datacenter bị từ chối`), TLS match qua TASK-SCRAPE-003 #4 (AC3), Akamai challenge -> ChallengedError chuyển TASK-SCRAPE-005 (AC8), báo monitor TASK-SCRAPE-006 (AC10), ghi qua InsertSnapshot delta-only (TASK-PRICE-002). Tất cả khớp.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 render qua farm | AC1 | `adapter.ts::fetch` (TASK-SCRAPE-003) |
| #2 TLS/HTTP2 match cho Akamai | AC3 | farm TLS (TASK-SCRAPE-003 #4) |
| #3 residential enterprise; datacenter từ chối | AC2 | `SelectTier(DiffAkamai)` + `proxy datacenter bị từ chối` |
| #4 JSON nhúng ưu tiên + fallback DOM | AC4,AC5 | `extract.ts::extractLazada` + tests |
| #5 map VND số nguyên | AC7 | `parseVNDInt` + `toSnapshot` |
| #6 hành vi giống người trước trích | AC (humanize) | `humanize` (TASK-SCRAPE-003) |
| #7 Akamai ChallengedError | AC8 | `Akamai challenge ném ChallengedError` |
| #8 báo monitor | AC10 | report outcome (TASK-SCRAPE-006) |
| #9 chờ render trước trích | AC9 | `waitForSelector(readyAnchor)` + test |
| #10 tối thiểu hóa | AC11 | chỉ trích trường giá |
| #11 OTel metric (SHOULD) | AC12 | counters/histogram |
| #12 quy đổi số nguyên | AC7 | `parse VND số nguyên` test |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Lớp Akamai (TLS match + residential enterprise) khóa bằng test, đúng §3.2/§3.9; tiền tệ số nguyên đồng bộ TASK-PRICE-002. Không có khiếm khuyết phải sửa ở task này. Score = 10/10. Verdict: PASS.

---

*Audit độc lập TASK-SCRAPE-008 - hết.*
