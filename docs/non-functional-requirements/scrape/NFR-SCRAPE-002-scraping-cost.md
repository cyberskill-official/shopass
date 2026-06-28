---
id: NFR-SCRAPE-002
title: "SCRAPE chi phí scraping/proxy - biến phí ~0,1-0,2 USD/user/tháng; chia sẻ SKU giảm chi phí cận biên ở quy mô"
module: SCRAPE
category: cost
priority: MUST
verification: T
phase: P1
slo: "Tổng biến phí scraping (proxy + headless farm) <= ~0,1-0,2 USD/user/tháng theo §4.1; chi phí cận biên/SKU giảm khi nhiều user chia sẻ cùng SKU"
owner: Stephen Cheng (Founder)
created: 2026-06-27
related_frs: [FR-SCRAPE-001, FR-SCRAPE-003, FR-SCRAPE-004, FR-SCRAPE-005, FR-PRICE-002]
source: "docs/... §4.1 (unit economics: proxy/scraping ~$30-60/1.000 users; headless farm ~$20-40; tổng ~0,1-0,2 USD/user; chia sẻ SKU giảm cận biên), §3.3 (tiering proxy + delta-only)"
---

## §1 - Statement (BCP-14 normative)

1. Tổng biến phí scraping (proxy residential + headless farm) **MUST** giữ <= ~0,1-0,2 USD/user/tháng theo unit economics §4.1, ở cả mốc 1.000 users và 100.000 users.
2. Chi phí cận biên mỗi SKU theo dõi **MUST** giảm khi nhiều user cùng theo dõi một SKU: một SKU được nhiều người quan tâm chỉ cần quét một lần, chia sẻ cho mọi user theo dõi nó (không quét lại per-user).
3. Hệ **MUST** dùng các đòn bẩy chi phí đã thiết kế để đạt ngưỡng: scan-frequency tiering (FR-SCRAPE-001, không quét mọi SKU mỗi phút), delta-only writes (FR-PRICE-002, không ghi khi giá không đổi), proxy tiering + cost-guard (FR-SCRAPE-004, dùng budget/mid khi đủ, enterprise chỉ khi cần), pacing tránh request thừa (FR-SCRAPE-005).
4. Chi phí proxy theo GB **MUST** được đo và quy về số nguyên (micro-USD) per provider/ngày (`proxy_usage` của FR-SCRAPE-004) để báo cáo unit economics đáng tin.
5. Cost-guard **MUST** đặt trần ngân sách/ngày và hạ tier / dừng tier cold khi gần trần (FR-SCRAPE-004 #7) - không để một sự cố (retry-loop, đợt sàn siết) đẩy chi phí vượt §4.1.
6. Hệ **SHOULD** ưu tiên đường rẻ trước đường đắt: gọi internal JSON (FR-SCRAPE-002, rẻ) trước khi dùng headless farm (FR-SCRAPE-003, đắt hơn nhiều lần); farm chỉ khi buộc.
7. Chi phí/user **SHOULD** giảm theo quy mô (lợi thế chia sẻ dữ liệu): ở 100.000 users, dù tổng proxy tăng, chi phí cận biên/user giảm nhờ độ phủ SKU dùng chung lớn hơn.

## §2 - Vì sao ràng buộc này

Mô hình kinh doanh của SănDeal là free-tier mạnh tài trợ bằng affiliate (§4.1, §4.3) - willingness-to-pay ở VN thấp, nên phần lớn user không trả tiền. Điều đó chỉ bền nếu biến phí phục vụ một user đủ thấp để affiliate (và một phần nhỏ Premium) bù được. Scraping/proxy là biến phí lớn nhất phía vận hành; nếu nó vượt ~0,1-0,2 USD/user, mô hình free-tier sụp. Lợi thế cốt lõi là chia sẻ dữ liệu: giá một SKU là tài sản dùng chung - quét một lần phục vụ mọi người theo dõi nó. Càng nhiều user, độ trùng SKU càng cao, chi phí cận biên/user càng giảm. Đây là lý do quy mô là bạn của SănDeal về mặt chi phí. Các đòn bẩy (tiering, delta-only, proxy tiering, cost-guard) không phải tối ưu lặt vặt mà là điều kiện để đơn vị kinh tế đứng vững.

## §3 - Đo lường (measurement)

- Counter `proxy_cost_usd_total{provider}` + `proxy_gb_used_total{provider, tier}` (FR-SCRAPE-004) -> chi phí proxy thực.
- Gauge `farm_render_total{platform}` + chi phí farm phân bổ (CPU/giờ) -> chi phí headless.
- Báo cáo dẫn xuất: chi phí scraping/user/tháng = (proxy_cost + farm_cost) / số user hoạt động - đối chiếu ngưỡng §4.1.
- Gauge `sku_shared_ratio` = số (user theo dõi) / số (SKU duy nhất quét) - đo hiệu quả chia sẻ; tỷ lệ cao = chi phí cận biên thấp.
- Counter `snapshot_skipped_total` vs `snapshot_written_total` (FR-PRICE-002/FR-SCRAPE-005) - hiệu quả delta-only phía storage/ghi.
- Phân bổ chi phí theo tier proxy (enterprise/mid/budget) - xác nhận không lạm dụng enterprise.

## §4 - Verification

- Cost-model test (T): với tải mô phỏng ở 1.000 và 100.000 users (độ trùng SKU thực tế), tính chi phí scraping/user -> <= ~0,2 USD/user/tháng.
- Sharing test (T): N user theo dõi cùng một SKU -> SKU đó được quét một lần (không N lần); `sku_shared_ratio` tăng theo quy mô giả lập.
- Tiering test (T): xác nhận phần lớn GB đi qua budget/mid; enterprise chỉ cho TikTok/Lazada (FR-SCRAPE-004) - không enterprise tràn lan.
- Cost-guard test (T): vượt 80% ngân sách -> hạ tier; chạm trần -> dừng cold (FR-SCRAPE-004 #7) - chi phí không vượt trần dù có sự cố.
- Delta-only contribution: với 80%+ lần quét no-change, xác nhận ghi thực thấp (giảm chi phí storage time-series phối hợp NFR-PRICE-001).

## §5 - Xử lý khi vi phạm

- Chi phí scraping/user vượt §4.1 -> sev-2; kiểm: tiering có chạy? delta-only có hiệu quả (skip-rate >= 80%)? enterprise có bị lạm dụng? cost-guard có đặt trần?
- `sku_shared_ratio` thấp bất thường (quét trùng per-user) -> sev-3; kiểm orchestrator có dedup theo SKU (một SKU một job, FR-SCRAPE-001 PK `product_id`).
- Chi phí enterprise vượt tỷ lệ kỳ vọng -> sev-3; xem `SelectTier` có hạ tier đúng cho target dễ (Shopee JSON -> budget).
- Cost-guard không đặt trần (chi phí ngày vượt ngân sách) -> sev-2; đây là van an toàn cuối, phải hoạt động.
- Farm (đắt) gọi quá thường (fallback rate cao) -> sev-3; kiểm vì sao JSON path hụt nhiều (sàn siết? adapter drift FR-SCRAPE-006?), vì farm đắt hơn JSON nhiều lần.

---

*Hết NFR-SCRAPE-002.*
