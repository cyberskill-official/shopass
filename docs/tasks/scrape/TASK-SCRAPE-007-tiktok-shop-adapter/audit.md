---
fr_id: TASK-SCRAPE-007
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. TASK-SCRAPE-007 đặc tả adapter TikTok Shop ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST + một MUST NOT ở #2), mỗi cái testable có AC §4 và test §5. Đường an toàn khóa đúng §3.2/§3.9: đọc DOM-render qua farm (TASK-SCRAPE-003), né hoàn toàn API ký `msToken`/`_signature`/`X-Bogus` (không cuộc đua vũ trang với ByteDance), proxy enterprise cho rủi ro ban High (`SelectTier(DiffByteDance)`), JSON nhúng ưu tiên với fallback DOM, chờ SPA hydrate trước khi trích. Giá VND số nguyên (parseVNDInt), đồng bộ TASK-PRICE-002. Score = 10/10.

## §2 - Findings

### ISS-001 - Typography (xác nhận sạch, không cần sửa)
Quét file: 0 ký tự mũi tên U+2192, 0 em-dash, 0 en-dash, 0 nháy cong, 0 ellipsis. Toàn bộ dùng ASCII `->`. Không có khiếm khuyết typography ở task này (khác với 001-006 vốn có arrow glyph đã sửa).

### ISS-002 - Đối chiếu fact TikTok với §3.2/§3.9 (xác nhận khớp)
Kiểm chéo nguồn: §3.2 dòng 70 - TikTok ký request `msToken`/`_signature`/`X-Bogus` + app attestation mạnh -> ưu tiên đọc DOM render; §3.9 - TikTok Shop risk High. task §1 #1-#3 + #2 (MUST NOT ký) phủ đúng, chọn prefer-DOM. Con số 41,31% GMV trong risk_if_skipped khớp nguồn (dòng 9 + 30). Không khiếm khuyết.

### ISS-003 - Tiền tệ không float + cross-ref (xác nhận)
§1 #5 + #12 quy đổi bằng số nguyên (`parseVNDInt` bỏ '₫'/'.'), AC7 `parse VND số nguyên` khóa `₫1.234.567 -> 1234567`; không float trên tiền. Cross-ref: proxy enterprise qua TASK-SCRAPE-004 (AC2), challenge -> ChallengedError chuyển TASK-SCRAPE-005 (AC8), báo outcome cho monitor TASK-SCRAPE-006 (AC10), ghi qua InsertSnapshot delta-only (TASK-PRICE-002). Tất cả khớp. AC3 `không gọi endpoint ký` khóa MUST NOT của #2.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 render qua farm | AC1 | `adapter.ts::fetch` (TASK-SCRAPE-003) |
| #2 MUST NOT ký API | AC3 | `không gọi endpoint ký` test |
| #3 proxy enterprise | AC2 | `SelectTier(DiffByteDance)` |
| #4 JSON nhúng ưu tiên + fallback DOM | AC4,AC5 | `extract.ts::extractTikTok` + tests |
| #5 map VND số nguyên | AC7 | `parseVNDInt` + `toSnapshot` |
| #6 hành vi giống người trước trích | AC (humanize) | `humanize` (TASK-SCRAPE-003) |
| #7 ChallengedError | AC8 | `challenge ném ChallengedError` |
| #8 báo monitor | AC10 | report outcome (TASK-SCRAPE-006) |
| #9 chờ SPA hydrate | AC9 | `waitForSelector(readyAnchor)` + test |
| #10 tối thiểu hóa | AC11 | chỉ trích trường giá |
| #11 OTel metric (SHOULD) | AC12 | counters/histogram |
| #12 quy đổi số nguyên | AC7 | `parse VND số nguyên` test |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Né API ký + proxy enterprise + prefer-DOM khóa bằng test, đúng §3.2/§3.9; tiền tệ số nguyên đồng bộ TASK-PRICE-002. Không có khiếm khuyết phải sửa ở task này. Score = 10/10. Verdict: PASS.

---

*Audit độc lập TASK-SCRAPE-007 - hết.*
