---
id: NFR-SCRAPE-001
title: "SCRAPE anti-bot resilience per-sàn + bảng xếp hạng rủi ro ban (Low/Medium/High) - Shopee Med-High, TikTok High, Lazada Med-High (scraping read-only); đọc giỏ qua extension Low; tự động xu/voucher High"
module: SCRAPE
category: reliability
priority: MUST
verification: T
phase: P1
slo: "Tỷ lệ ban/chặn IP per-sàn dưới ngưỡng vận hành; success-rate scraping read-only >= mục tiêu per-sàn; phục hồi sau sàn đổi DOM trong thời gian giới hạn"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-001, FR-SCRAPE-002, FR-SCRAPE-003, FR-SCRAPE-004, FR-SCRAPE-005, FR-SCRAPE-006, FR-SCRAPE-007, FR-SCRAPE-008]
source: "docs/... §3.9 (phân tích anti-bot per-sàn + xếp hạng rủi ro ban Low/Medium/High), §3.2 (cơ chế per-sàn), §5.2 (rủi ro phụ thuộc nền tảng)"
---

## §1 - Statement (BCP-14 normative)

1. Hệ scraping **MUST** chịu được anti-bot per-sàn ở mức cho phép vận hành liên tục, theo bảng xếp hạng rủi ro ban (§3.9):

   | Hoạt động | Shopee | TikTok Shop | Lazada |
   |---|---|---|---|
   | Scraping giá read-only quy mô lớn | Medium-High | High | Medium-High |
   | Đọc giỏ hàng client-side qua extension (cookie chính user) | Low | Low | Low |
   | Tự động hóa xu/voucher | High | High | High |

2. Với scraping read-only, hệ **MUST** dùng đủ lớp chống ban tương ứng mức rủi ro: residential proxy bắt buộc (FR-SCRAPE-004), anti-fingerprint + TLS match (FR-SCRAPE-003), pacing/jitter (FR-SCRAPE-005), concurrency cap per-platform (FR-SCRAPE-001). TikTok (High) và Lazada/Akamai (Medium-High) **MUST** dùng proxy tier enterprise.
3. Tỷ lệ request bị ban/challenge per-sàn **MUST** giữ dưới ngưỡng vận hành cấu hình; vượt ngưỡng kéo dài **MUST** kích hoạt hạ tải (FR-SCRAPE-006) thay vì tiếp tục đốt proxy vào request bị chặn.
4. Hệ **MUST** coi đọc giỏ hàng qua extension là Low-risk vì first-party (cookie của chính người dùng, hành vi giống người dùng) - KHÔNG áp hạ tầng chống-ban scraping cho đường này; nhưng vẫn tôn trọng rủi ro ToS (đường này tách hẳn backend scraping).
5. Hệ **MUST NOT** tự động hóa xu/voucher (mức High, dễ bị coi là bot/abuse) - chỉ làm checklist nhắc nhở + auto-test mã do người dùng khởi tạo (giới hạn ở CART/AFFIL phase sau), KHÔNG tự click xu.
6. Khi một sàn đổi DOM/API (rủi ro existential §5.2), hệ **MUST** phát hiện qua giám sát drift (FR-SCRAPE-006) và phục hồi (người sửa adapter + tự bỏ throttle) trong thời gian giới hạn, không để dữ liệu sàn đó chết âm thầm.
7. Hệ **SHOULD** phân tán rủi ro qua đa sàn: ban/chặn một sàn KHÔNG được làm sập toàn bộ giá trị sản phẩm (so sánh chéo vẫn chạy với các sàn còn lại; web app độc lập extension).

## §2 - Vì sao ràng buộc này

Anti-bot per-sàn là tuyến phòng thủ của đối thủ trực tiếp chống lại mô hình SănDeal. §3.9 phân loại rõ: scraping read-only của Shopee/Lazada là Medium-High, TikTok là High - đây là phần khó nhất và là nơi đốt nhiều công sức kỹ thuật nhất. Nếu không chịu được, dữ liệu giá - nền tảng của mọi tính năng lõi - không tồn tại. Ngược lại, đọc giỏ hàng qua extension là Low vì nó là hành vi first-party của chính người dùng, không phải scraping; gộp nhầm hai đường này sẽ hoặc làm extension nặng nề thừa, hoặc làm backend scraping lộ. Tự động xu/voucher là High và SănDeal chủ động không làm - đây là ranh giới tự đặt để không bị coi là công cụ abuse. Rủi ro phụ thuộc nền tảng là existential (§5.2): sàn có thể đổi DOM hoặc gửi C&D bất cứ lúc nào, nên khả năng phát hiện sớm và phân tán đa sàn quyết định sự sống còn.

## §3 - Đo lường (measurement)

- Gauge `scrape_ban_rate{platform}` = (challenge + block) / tổng request, theo cửa sổ trượt - đối chiếu ngưỡng vận hành per-sàn.
- Gauge `scrape_success_rate{platform}` cho scraping read-only.
- Counter `proxy_ip_banned_total{provider, platform}` (FR-SCRAPE-004) - tốc độ đốt IP.
- Gauge `adapter_health_state{platform, version}` (FR-SCRAPE-006) - healthy/degraded/broken.
- Counter `adapter_alert_total{platform, transition}` + thời gian từ alert tới phục hồi (MTTR adapter).
- Phân biệt rõ `parse_fail` (sàn đổi DOM) với `challenge` (anti-bot) trong metric để chẩn đoán đúng nguyên nhân.

## §4 - Verification

- Resilience test (T): chạy scraping read-only mỗi sàn ở tải mục tiêu qua residential enterprise (TikTok/Lazada) + anti-fingerprint, đo `scrape_ban_rate` dưới ngưỡng trong cửa sổ dài.
- Tiering test (T): xác nhận TikTok/Lazada được cấp proxy enterprise (`SelectTier` đúng), Shopee-JSON dùng budget/mid (FR-SCRAPE-004).
- DOM-drift recovery test (T): mô phỏng sàn đổi selector -> monitor (FR-SCRAPE-006) chuyển broken + alert + throttle; sau khi "sửa adapter", tự về healthy và bỏ throttle; đo MTTR dưới ngưỡng.
- Risk-boundary test: xác nhận đường đọc giỏ extension KHÔNG đi qua hạ tầng scraping; xác nhận không có code tự-click xu/voucher.
- Multi-platform degradation test: ban một sàn -> so sánh chéo + web vẫn chạy với sàn còn lại (không sập toàn cục).

## §5 - Xử lý khi vi phạm

- `scrape_ban_rate{platform}` vượt ngưỡng kéo dài -> sev-2; kiểm proxy tier (đúng enterprise cho TikTok/Lazada?), pacing/jitter, concurrency cap; hạ tải tạm qua FR-SCRAPE-006.
- Một sàn `broken` (DOM đổi) quá MTTR mục tiêu -> sev-2; ưu tiên sửa adapter; throttle giữ proxy không bị đốt trong lúc chờ.
- Tốc độ đốt IP bất thường (`proxy_ip_banned_total` tăng vọt) -> sev-3; xem lại fingerprint nhất quán + xoay vòng IP + cooldown.
- Phát hiện code tự-click xu/voucher (vi phạm §1 #5) -> chặn merge; đây là ranh giới compliance cứng (rủi ro bị coi là abuse + ToS).
- Một sàn gửi C&D / chặn kỹ thuật toàn diện (§5.2) -> sev-1 chiến lược; kích hoạt phân tán đa sàn, dựa web app độc lập, đánh giá lại đường thu thập.

---

*Hết NFR-SCRAPE-001.*
