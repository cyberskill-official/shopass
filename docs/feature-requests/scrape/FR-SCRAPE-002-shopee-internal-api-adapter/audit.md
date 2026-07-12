---
fr_id: FR-SCRAPE-002
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

Re-derive từ file hiện tại. FR-SCRAPE-002 đặc tả adapter Shopee internal-API ở mức triển khai được: 12 mệnh đề §1 normative (đủ MUST), mỗi cái testable có AC §4 và test §5. Đường giá chính là `/api/v4/pdp/get_pc` truy cập `is_login:false` (khớp §3.2 nguồn), `recommend` là nguồn phụ; fallback Playwright khi gặp HTML challenge (anti-bot Shopee Medium-High §3.9). Quy đổi giá micro-VND bằng phép chia số nguyên, đồng bộ BIGINT VND của FR-PRICE-002. Ranh giới token-not-on-server (không cookie người dùng ở backend) được khóa bằng test. Score = 10/10.

## §2 - Findings

### ISS-001 - Arrow glyph trong comment code (đã sửa)
Phát hiện 2 ký tự mũi tên U+2192 trong comment Vietnamese của khối code (`parse.go::toSnapshot` dòng 134, `adapter.go::Fetch` dòng 172), vi phạm tiêu chí O. Đã sửa thành `->`. Quét lại: 0 mũi tên còn lại; không có em-dash/en-dash/nháy cong/ellipsis.

### ISS-002 - Đối chiếu fact Shopee với §3.2/§3.9 (xác nhận khớp)
Kiểm chéo nguồn: endpoint `/api/v4/pdp/get_pc` và `/api/v4/recommend/recommend`, một số truy cập được khi `is_login:false` (§3.2 dòng 69); anti-bot Shopee Medium-High với slider/puzzle CAPTCHA (§3.9). FR mô tả đúng. Không khiếm khuyết.

### ISS-003 - Tiền tệ không float (xác nhận BIGINT, chia số nguyên)
§1 #5 + `toSnapshot` dùng `it.Price / shopeePriceUnit` (chia int64), kiểu `price.PriceSnapshot.Price` là int64; AC4 + `TestParse_IntegerDivision_NoFloatError` khóa `333_333_00000 / 100000 = 333_333`. Không có float trung gian trên tiền. Đúng DEC-PRICE-05.

### ISS-004 - Cross-ref ghi DB (xác nhận)
Adapter chỉ trả `PriceSnapshot`; ghi delta-only do orchestrator gọi `InsertSnapshot` (FR-SCRAPE-001 #8), AC12 kiểm tích hợp. Adapter không tự ghi SQL. Đúng.

## §3 - Traceability §1 -> AC -> artefact

| §1 mệnh đề | §4 AC | §5 test / §3 artefact |
|---|---|---|
| #1 thỏa PlatformAdapter | AC1 | compile-assert `var _ PlatformAdapter` |
| #2 endpoint pdp/get_pc | AC2 | `endpoints.go::pdpURL` |
| #3 is_login:false, no cookie | AC8 | `TestFetch_NoUserCookieSent` |
| #4 parse JSON -> snapshot | AC3 | `parse.go::toSnapshot` + `TestParse_ValidPDP` |
| #5 quy đổi số nguyên | AC4 | chia `/100000` + `TestParse_IntegerDivision_NoFloatError` |
| #6 fallback challenge | AC6 | `adapter.go::Fetch` + `TestFetch_HTMLChallenge_FallsBackToFarm` |
| #7 trả error không panic | AC7 | error path Fetch |
| #8 recommend nguồn phụ (SHOULD) | - | `recommendPath` |
| #9 ErrItemGone | AC5 | `TestParse_ItemGone` |
| #10 OTel metric (SHOULD) | AC11 | counters |
| #11 tối thiểu hóa phản hồi | AC10 | chỉ trích trường giá |
| #12 ts + ProductID | AC9 | `toSnapshot` set ProductID/TS |

## §4 - Kết luận

Mọi mệnh đề normative có AC + test/artefact; không mệnh đề mồ côi. Fact Shopee khớp §3.2/§3.9, tiền tệ BIGINT chia số nguyên, ranh giới backend vs extension khóa bằng test. Một typography defect đã sửa. Score = 10/10. Verdict: PASS.

---

*Audit độc lập FR-SCRAPE-002 - hết.*
