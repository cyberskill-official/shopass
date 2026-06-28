---
fr_id: FR-SCRAPE-003
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. FR-SCRAPE-003 đặc tả Playwright headless farm + anti-fingerprint ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Phủ đúng yêu cầu §3.3 nguồn: spoof Canvas/WebGL/AudioContext readback ổn định theo seed, khớp JA3/JA4 TLS + HTTP/2 SETTINGS (dưới tầng JS, then chốt cho Akamai), fingerprint nhất quán nội bộ (timezone<->locale<->WebGL), gắn profile với proxy cùng nước, hành vi giống người trước khi trích. `RenderPrice` + `ChallengedError` nối sạch vào fallback adapter. Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Phát hiện 2 ký tự mũi tên U+2192 trong comment Vietnamese của khối test (`fingerprint.test.ts` dòng 202-203), vi phạm tiêu chí O. Đã sửa thành `->`. Quét lại: 0 mũi tên; không có em-dash/en-dash/nháy cong/ellipsis.

### ISS-002 - Đối chiếu fact anti-fingerprint với §3.3/§3.9 (xác nhận khớp)
Kiểm chéo: §3.3 yêu cầu spoof Canvas/WebGL/AudioContext + JA3/JA4 TLS + HTTP/2 settings; §3.9 nói Shopee hash thiết bị từ Canvas/WebGL, Lazada dùng Akamai (TLS/HTTP). FR §1 #1-#4 phủ đủ các điểm này. Không khiếm khuyết.

### ISS-003 - Cross-ref proxy + downstream (xác nhận)
§1 #5 gắn profile <-> proxy session cùng nước (FR-SCRAPE-004); §1 #8 `ChallengedError` chuyển FR-SCRAPE-005; snapshot ghi qua InsertSnapshot delta-only (FR-PRICE-002). Quy đổi giá DOM bằng số nguyên (render.ts `toSnapshot`), AC10. Các cross-ref khớp, không nhân bản logic.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 profile nhất quán nội bộ | AC1,AC2 | `fingerprint.ts::makeProfile/isCoherent` + test |
| #2 spoof readback ổn định seed | AC3 | `browser.ts::spoofScript` + `Canvas readback ổn định` |
| #3 ẩn automation webdriver | AC4 | `addInitScript` + `webdriver ẩn sau patch` |
| #4 khớp JA3/JA4 + HTTP/2 | AC5,AC6 | TLS client + WebGL test |
| #5 gắn proxy cùng nước | AC7 | `bindProfileProxy` guard |
| #6 hành vi giống người trước trích | AC8 | `behavior.ts::humanize` + thứ tự test |
| #7 RenderPrice fallback | AC10 | `render.ts::renderPrice` |
| #8 ErrChallenged | AC9 | `ChallengedError` + challenge test |
| #9 dọn tài nguyên page | AC11 | `page.close()` finally |
| #10 xoay profile pool (SHOULD) | - | config profile pool |
| #11 OTel metric (SHOULD) | AC12 | `farm_render_total` |
| #12 không lưu trang thô | AC10 (chỉ trích giá) | minimal extract |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Lớp chống ban (fingerprint + TLS + behavior) khóa bằng test, đúng §3.3/§3.9, gắn đúng vào fallback adapter và proxy. Một typography defect đã sửa. Score = 10/10. Verdict: PASS.

---

*Audit độc lập FR-SCRAPE-003 - hết.*
