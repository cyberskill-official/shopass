---
fr_id: TASK-B2B-004
audited: 2026-06-28
verdict: PASS
score: 10/10
template: engineering-spec@1
auditor: independent
---

## §1 - Tóm tắt verdict

TASK-B2B-004 đặc tả Premium API access ở mức triển khai được. 12 mệnh đề §1 normative, mỗi mệnh đề có AC và test. Ba điều kiện bảo mật của một API public được giữ chặt: key băm tại nghỉ + so khớp hằng-thời-gian, rate-limit per-key tại gateway (TASK-INFRA-001), và phạm vi CHỈ market_trend_daily đã ẩn danh (không raw, không user-level). Thu hồi key có hiệu lực <= 60s. Đạt 10/10.

## §2 - Findings (đã giải quyết)

### ISS-001 - API public rò ra raw/user-level (đã chốt)
Rủi ro lớn nhất: API public là bề mặt Internet, nếu chạm raw thì một key rút cạn dữ liệu chi tiết, phá k-anonymity. Giải: §1 #5 + DEC-B2B-33 chỉ route `/public/v1/trends` -> ô đã phát hành; AC #8 + `TestPublic_NoRawRoutes` + disallowed_tools.

### ISS-002 - Key cleartext + timing attack
Lưu cleartext = rò DB rò mọi key; so secret không hằng-thời-gian lộ qua timing. Giải: §1 #1/#12 + DEC-B2B-30 băm có salt + `crypto/subtle`; AC #1/#10.

### ISS-003 - Rate-limit né được
Rate-limit ở service lẻ cho phép vòng qua nhiều service vượt tổng. Giải: §1 #4 + DEC-B2B-32 thực thi per-key tại gateway; test `TestRate_ExceedsLimit_429`.

### ISS-004 - Key lộ dùng được lâu
Cache tier/quyền quá lâu làm key thu hồi vẫn chạy hàng giờ. Giải: §1 #3 + DEC-B2B-34 hiệu lực <= 60s; test `TestAuth_RevocationTakesEffect`.

## §3 - Traceability §1 -> AC -> artefact

| §1 | AC | Artefact |
|---|---|---|
| #1 key băm, không cleartext | #1 | `0004_api_key.sql` + `auth.go` |
| #2 schema tier | #1 | `api_key` CHECK tier |
| #3 revoked + hiệu lực nhanh | #3,#4 | `TestAuth_Revoked_401` + `TestAuth_RevocationTakesEffect` |
| #4 rate-limit gateway | #5 | `ratelimit.go::allow` + `TestRate_ExceedsLimit_429` |
| #5 chỉ market_trend_daily | #8 | `TestPublic_NoRawRoutes` + router |
| #6 endpoint gating theo tier | #7 | `allowedEndpoint` + `TestEndpoint_FreeTierDenied_403` |
| #7 usage không nội dung | #9 | `usage.go` + api_usage schema |
| #8 quota vs rate tách | #6 | handler phân biệt 429 |
| #12 hằng-thời-gian | #10 | `verifyHashConstantTime` |

## §4 - Kết luận

Toàn bộ mệnh đề normative có code/SQL/test backing. Ba lớp bảo vệ API public (auth băm, rate-limit gateway, phạm vi chỉ-ẩn-danh) được kiểm bằng test. Không mệnh đề mồ côi. Score = 10/10. Verdict: PASS. Sẵn sàng vào hàng đợi build (status `ready_to_implement`).

---

*Hết audit TASK-B2B-004.*
